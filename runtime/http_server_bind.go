package runtime

import (
	"fmt"
	stdhttp "net/http"
	"strings"
	"sync"

	httplib "flag-lang/libraries/http"
)

var (
	GoBind_http_Listen  = NewFunction(adaptHTTPListen)
	GoBind_http_Stop    = NewFunction(adaptHTTPStop)
	GoBind_http_Address = NewFunction(adaptHTTPAddress)
)

var httpServerRegistry = struct {
	mu      sync.Mutex
	nextID  int64
	servers map[int64]*httplib.Server
}{
	servers: map[int64]*httplib.Server{},
}

func adaptHTTPListen(args ...Value) Value {
	goArgArityExact("http/listen", args, 1)
	opts := args[0]
	if opts.tag != TagMap {
		panic(fmt.Sprintf("http/listen expects an options map, got %s", ValueToString(opts)))
	}
	addr := httpListenAddress(opts)
	handler := httpListenHandler(opts)
	server, err := httplib.Listen(addr, handler)
	if err != nil {
		panic(err.Error())
	}
	return httpServerHandleValue(storeHTTPServer(server), server.Addr())
}

func adaptHTTPStop(args ...Value) Value {
	goArgArityExact("http/stop", args, 1)
	id := httpServerID("http/stop", args[0])
	server := loadHTTPServer("http/stop", id)
	if err := server.Close(); err != nil {
		panic(err.Error())
	}
	deleteHTTPServer(id)
	return NilValue()
}

func adaptHTTPAddress(args ...Value) Value {
	goArgArityExact("http/address", args, 1)
	id := httpServerID("http/address", args[0])
	return NewString(loadHTTPServer("http/address", id).Addr())
}

func httpListenAddress(opts Value) string {
	if addr := Get(opts, NewKeyword("addr")); addr.tag != TagNil {
		return goArgString("http/listen", 0, addr)
	}
	host := "127.0.0.1"
	if v := Get(opts, NewKeyword("host")); v.tag != TagNil {
		host = goArgString("http/listen", 0, v)
	}
	port := int64(0)
	if v := Get(opts, NewKeyword("port")); v.tag != TagNil {
		port = goArgInt64("http/listen", 0, v)
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func httpListenHandler(opts Value) httplib.Handler {
	if handler := Get(opts, NewKeyword("handler")); handler.tag != TagNil {
		return httpHandlerFromValue("http/listen", handler)
	}
	if router := Get(opts, NewKeyword("router")); router.tag != TagNil {
		return httpRouterFromValue("http/listen", router)
	}
	if routes := Get(opts, NewKeyword("routes")); routes.tag != TagNil {
		return httpRouterFromRoutesValue("http/listen", routes)
	}
	panic("http/listen expects :handler, :router, or :routes")
}

func httpHandlerFromValue(name string, handler Value) httplib.Handler {
	if handler.tag != TagFunction {
		panic(fmt.Sprintf("%s handler: expected function, got %s", name, ValueToString(handler)))
	}
	return func(req httplib.ServerRequest) httplib.ServerResponse {
		resp := Call(handler, httpServerRequestValue(req))
		return httpServerResponseFromValue(name, resp)
	}
}

func httpRouterFromValue(name string, router Value) httplib.Handler {
	if router.tag != TagMap {
		panic(fmt.Sprintf("%s router: expected map, got %s", name, ValueToString(router)))
	}
	routes := Get(router, NewKeyword("routes"))
	if routes.tag == TagNil {
		panic(fmt.Sprintf("%s router: missing :routes", name))
	}
	return httpRouterFromRoutesValue(name, routes)
}

func httpRouterFromRoutesValue(name string, routes Value) httplib.Handler {
	items := httpRouteValues(name, routes)
	compiled := make([]httplib.Route, 0, len(items))
	for i, route := range items {
		if route.tag != TagMap {
			panic(fmt.Sprintf("%s route %d: expected map, got %s", name, i+1, ValueToString(route)))
		}
		handler := Get(route, NewKeyword("handler"))
		if handler.tag != TagFunction {
			panic(fmt.Sprintf("%s route %d: missing function :handler", name, i+1))
		}
		compiled = append(compiled, httplib.Route{
			Method: strings.ToUpper(strings.TrimSpace(requestFieldString(name, route, "method", "*"))),
			Path:   requestFieldString(name, route, "path", ""),
			Handler: func(fn Value) httplib.Handler {
				return func(req httplib.ServerRequest) httplib.ServerResponse {
					resp := Call(fn, httpServerRequestValue(req))
					return httpServerResponseFromValue(name, resp)
				}
			}(handler),
		})
	}
	router, err := httplib.NewRouter(compiled)
	if err != nil {
		panic(err.Error())
	}
	return router.Handler()
}

func httpRouteValues(name string, v Value) []Value {
	switch v.tag {
	case TagArray:
		return v.ArrayValues()
	case TagList:
		return v.ListValues()
	case TagNil:
		return nil
	default:
		panic(fmt.Sprintf("%s routes: expected array or list, got %s", name, ValueToString(v)))
	}
}

func httpServerRequestValue(req httplib.ServerRequest) Value {
	out := NewMap(
		NewKeyword("method"), NewString(req.Method),
		NewKeyword("url"), NewString(req.URL),
		NewKeyword("path"), NewString(req.Path),
		NewKeyword("raw-path"), NewString(req.RawPath),
		NewKeyword("query-string"), NewString(req.RawQuery),
		NewKeyword("query"), httpStringListsValue(req.Query, false),
		NewKeyword("headers"), httpStringListsValue(map[string][]string(req.Headers), true),
		NewKeyword("body"), NewString(string(req.Body)),
		NewKeyword("host"), NewString(req.Host),
		NewKeyword("remote-addr"), NewString(req.RemoteAddr),
		NewKeyword("scheme"), NewString(req.Scheme),
		NewKeyword("params"), httpStringMapValue(req.Params, false),
	)
	return out
}

func httpServerResponseFromValue(name string, v Value) httplib.ServerResponse {
	if v.tag == TagNil {
		return httplib.ServerResponse{StatusCode: stdhttp.StatusOK}
	}
	if v.tag != TagMap {
		return httplib.ServerResponse{
			StatusCode: stdhttp.StatusOK,
			Body:       httpBodyBytes(name, v),
		}
	}
	status := int64(stdhttp.StatusOK)
	if raw := Get(v, NewKeyword("status")); raw.tag != TagNil {
		status = goArgInt64(name, 0, raw)
	}
	headers := stdhttp.Header{}
	if raw := Get(v, NewKeyword("headers")); raw.tag != TagNil {
		headers = stdhttp.Header(requestHeadersFromValue(name, raw))
	}
	body := []byte(nil)
	if raw := Get(v, NewKeyword("body")); raw.tag != TagNil {
		body = httpBodyBytes(name, raw)
	}
	return httplib.ServerResponse{
		StatusCode: int(status),
		Headers:    headers,
		Body:       body,
	}
}

func httpStringListsValue(values map[string][]string, lowerKeys bool) Value {
	out := NewMap()
	for key, items := range values {
		if lowerKeys {
			key = strings.ToLower(key)
		}
		switch len(items) {
		case 0:
			continue
		case 1:
			out = Assoc(out, NewKeyword(key), NewString(items[0]))
		default:
			vals := make([]Value, 0, len(items))
			for _, item := range items {
				vals = append(vals, NewString(item))
			}
			out = Assoc(out, NewKeyword(key), NewArray(vals...))
		}
	}
	return out
}

func httpStringMapValue(values map[string]string, lowerKeys bool) Value {
	out := NewMap()
	for key, value := range values {
		if lowerKeys {
			key = strings.ToLower(key)
		}
		out = Assoc(out, NewKeyword(key), NewString(value))
	}
	return out
}

func httpServerHandleValue(id int64, addr string) Value {
	return NewMap(
		NewKeyword("server-id"), NewLong(id),
		NewKeyword("addr"), NewString(addr),
	)
}

func httpServerID(name string, handle Value) int64 {
	if handle.tag != TagMap {
		panic(fmt.Sprintf("%s expects a server handle map, got %s", name, ValueToString(handle)))
	}
	id := Get(handle, NewKeyword("server-id"))
	if id.tag != TagLong {
		panic(fmt.Sprintf("%s expects a server handle with :server-id, got %s", name, ValueToString(handle)))
	}
	return id.Long()
}

func storeHTTPServer(server *httplib.Server) int64 {
	httpServerRegistry.mu.Lock()
	defer httpServerRegistry.mu.Unlock()
	httpServerRegistry.nextID++
	id := httpServerRegistry.nextID
	httpServerRegistry.servers[id] = server
	return id
}

func loadHTTPServer(name string, id int64) *httplib.Server {
	httpServerRegistry.mu.Lock()
	defer httpServerRegistry.mu.Unlock()
	server := httpServerRegistry.servers[id]
	if server == nil {
		panic(fmt.Sprintf("%s: unknown server id %d", name, id))
	}
	return server
}

func deleteHTTPServer(id int64) {
	httpServerRegistry.mu.Lock()
	defer httpServerRegistry.mu.Unlock()
	delete(httpServerRegistry.servers, id)
}
