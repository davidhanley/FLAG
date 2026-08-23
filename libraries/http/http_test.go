package httplib

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "yes" {
			t.Fatalf("expected request header, got %q", got)
		}
		w.Header().Set("X-Reply", "pong")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	req := NewRequest(http.MethodPost, server.URL).WithHeader("X-Test", "yes").WithBody([]byte("body"))
	resp, err := Send(req)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "hello" {
		t.Fatalf("expected response body hello, got %q", resp.Body)
	}
	if got := resp.Headers.Get("X-Reply"); got != "pong" {
		t.Fatalf("expected response header pong, got %q", got)
	}
}

func TestRouterMatchesParamsAndMethods(t *testing.T) {
	router, err := NewRouter([]Route{
		{
			Method: "GET",
			Path:   "/users/:id",
			Handler: func(req ServerRequest) ServerResponse {
				return ServerResponse{
					StatusCode: http.StatusOK,
					Body:       []byte(req.Params["id"]),
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("NewRouter returned error: %v", err)
	}

	ok := router.Handle(ServerRequest{Method: "GET", Path: "/users/42"})
	if ok.StatusCode != http.StatusOK || string(ok.Body) != "42" {
		t.Fatalf("expected matched route body 42, got %d %q", ok.StatusCode, ok.Body)
	}

	method := router.Handle(ServerRequest{Method: "POST", Path: "/users/42"})
	if method.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", method.StatusCode)
	}
	if got := method.Headers.Get("Allow"); got != "GET" {
		t.Fatalf("expected Allow=GET, got %q", got)
	}

	missing := router.Handle(ServerRequest{Method: "GET", Path: "/missing"})
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", missing.StatusCode)
	}
}

func TestListenServesRouter(t *testing.T) {
	router, err := NewRouter([]Route{
		{
			Method: "GET",
			Path:   "/hello/:name",
			Handler: func(req ServerRequest) ServerResponse {
				return ServerResponse{
					Headers: http.Header{"X-Name": []string{req.Params["name"]}},
					Body:    []byte("hi " + req.Params["name"]),
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("NewRouter returned error: %v", err)
	}

	server, err := Listen("127.0.0.1:0", router.Handler())
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	defer func() { _ = server.Close() }()

	resp, err := http.Get("http://" + server.Addr() + "/hello/flag")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Name"); got != "flag" {
		t.Fatalf("expected X-Name=flag, got %q", got)
	}
	if string(body) != "hi flag" {
		t.Fatalf("expected body hi flag, got %q", body)
	}
}
