package httplib

import (
	"context"
	"fmt"
	"io"
	"net"
	stdhttp "net/http"
	"sort"
	"strings"
	"time"
)

type ServerRequest struct {
	Method     string
	URL        string
	Path       string
	RawPath    string
	RawQuery   string
	Query      map[string][]string
	Headers    stdhttp.Header
	Body       []byte
	Host       string
	RemoteAddr string
	Scheme     string
	Params     map[string]string
}

type ServerResponse struct {
	StatusCode int
	Headers    stdhttp.Header
	Body       []byte
}

type Handler func(ServerRequest) ServerResponse

type Route struct {
	Method  string
	Path    string
	Handler Handler
}

type Server struct {
	server   *stdhttp.Server
	listener net.Listener
}

type Router struct {
	routes []compiledRoute
}

type compiledRoute struct {
	method  string
	path    string
	parts   []routePart
	handler Handler
}

type routePartKind int

const (
	routeLiteral routePartKind = iota
	routeParam
	routeWildcard
)

type routePart struct {
	kind  routePartKind
	value string
}

func NewRouter(routes []Route) (*Router, error) {
	compiled := make([]compiledRoute, 0, len(routes))
	for i, route := range routes {
		if route.Handler == nil {
			return nil, fmt.Errorf("route %d: handler cannot be nil", i+1)
		}
		method := strings.ToUpper(strings.TrimSpace(route.Method))
		if method == "" {
			method = "*"
		}
		path, parts, err := compileRoutePath(route.Path)
		if err != nil {
			return nil, fmt.Errorf("route %d: %w", i+1, err)
		}
		compiled = append(compiled, compiledRoute{
			method:  method,
			path:    path,
			parts:   parts,
			handler: route.Handler,
		})
	}
	return &Router{routes: compiled}, nil
}

func (r *Router) Handler() Handler {
	if r == nil {
		return nil
	}
	return r.Handle
}

func (r *Router) Handle(req ServerRequest) ServerResponse {
	if r == nil {
		return notFoundResponse()
	}
	allowed := make([]string, 0, 4)
	seen := map[string]bool{}
	for _, route := range r.routes {
		params, ok := matchRouteParts(route.parts, req.Path)
		if !ok {
			continue
		}
		if route.method == "*" || route.method == strings.ToUpper(req.Method) {
			req.Params = params
			return route.handler(req)
		}
		if route.method != "" && route.method != "*" && !seen[route.method] {
			seen[route.method] = true
			allowed = append(allowed, route.method)
		}
	}
	if len(allowed) > 0 {
		sort.Strings(allowed)
		return ServerResponse{
			StatusCode: stdhttp.StatusMethodNotAllowed,
			Headers:    stdhttp.Header{"Allow": []string{strings.Join(allowed, ", ")}},
			Body:       []byte("method not allowed"),
		}
	}
	return notFoundResponse()
}

func Listen(addr string, handler Handler) (*Server, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}
	if strings.TrimSpace(addr) == "" {
		addr = "127.0.0.1:0"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		server: &stdhttp.Server{
			Handler: stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				req, err := newServerRequest(r)
				if err != nil {
					stdhttp.Error(w, err.Error(), stdhttp.StatusInternalServerError)
					return
				}
				writeServerResponse(w, handler(req))
			}),
		},
		listener: listener,
	}
	go func() {
		_ = server.server.Serve(listener)
	}()
	return server, nil
}

func (s *Server) Addr() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) Close() error {
	if s == nil || s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err == stdhttp.ErrServerClosed {
		return nil
	}
	return err
}

func newServerRequest(r *stdhttp.Request) (ServerRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ServerRequest{}, err
	}
	_ = r.Body.Close()
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = path
	}
	query := make(map[string][]string, len(r.URL.Query()))
	for key, values := range r.URL.Query() {
		query[key] = append([]string(nil), values...)
	}
	scheme := requestScheme(r)
	url := r.URL.RequestURI()
	if r.Host != "" {
		url = scheme + "://" + r.Host + r.URL.RequestURI()
	}
	return ServerRequest{
		Method:     r.Method,
		URL:        url,
		Path:       path,
		RawPath:    rawPath,
		RawQuery:   r.URL.RawQuery,
		Query:      query,
		Headers:    cloneHeader(r.Header),
		Body:       body,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		Scheme:     scheme,
		Params:     map[string]string{},
	}, nil
}

func requestScheme(r *stdhttp.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if value := r.Header.Get("X-Forwarded-Proto"); value != "" {
		return value
	}
	return "http"
}

func writeServerResponse(w stdhttp.ResponseWriter, resp ServerResponse) {
	status := resp.StatusCode
	if status == 0 {
		status = stdhttp.StatusOK
	}
	headers := cloneHeader(resp.Headers)
	if len(resp.Body) > 0 && headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", "text/plain; charset=utf-8")
	}
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(status)
	if len(resp.Body) > 0 {
		_, _ = w.Write(resp.Body)
	}
}

func notFoundResponse() ServerResponse {
	return ServerResponse{
		StatusCode: stdhttp.StatusNotFound,
		Body:       []byte("not found"),
	}
}

func compileRoutePath(path string) (string, []routePart, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "", nil, fmt.Errorf("path must start with /")
	}
	if path == "/" {
		return path, nil, nil
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	parts := make([]routePart, 0, len(segments))
	for i, segment := range segments {
		if segment == "" {
			return "", nil, fmt.Errorf("path %q contains an empty segment", path)
		}
		switch {
		case strings.HasPrefix(segment, "*"):
			name := strings.TrimPrefix(segment, "*")
			if name == "" {
				return "", nil, fmt.Errorf("path %q has wildcard without a name", path)
			}
			if i != len(segments)-1 {
				return "", nil, fmt.Errorf("path %q has a wildcard before the end", path)
			}
			parts = append(parts, routePart{kind: routeWildcard, value: name})
		case strings.HasPrefix(segment, ":"):
			name := strings.TrimPrefix(segment, ":")
			if name == "" {
				return "", nil, fmt.Errorf("path %q has a parameter without a name", path)
			}
			parts = append(parts, routePart{kind: routeParam, value: name})
		default:
			parts = append(parts, routePart{kind: routeLiteral, value: segment})
		}
	}
	return path, parts, nil
}

func matchRouteParts(parts []routePart, path string) (map[string]string, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if path == "/" {
		return map[string]string{}, len(parts) == 0
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	params := map[string]string{}
	i := 0
	for ; i < len(parts); i++ {
		if i >= len(segments) {
			return nil, false
		}
		part := parts[i]
		switch part.kind {
		case routeLiteral:
			if segments[i] != part.value {
				return nil, false
			}
		case routeParam:
			params[part.value] = segments[i]
		case routeWildcard:
			params[part.value] = strings.Join(segments[i:], "/")
			return params, true
		}
	}
	if i != len(segments) {
		return nil, false
	}
	return params, true
}
