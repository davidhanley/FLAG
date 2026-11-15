package runtime

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPBridgeSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "yes" {
			t.Fatalf("expected request header, got %q", got)
		}
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("expected method POST, got %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if string(body) != "hello" {
			t.Fatalf("expected request body hello, got %q", string(body))
		}
		w.Header().Set("X-Reply", "pong")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("world"))
	}))
	defer server.Close()

	req := Call(GoBind_http_Request, NewString(http.MethodPost), NewString(server.URL))
	req = Call(GoBind_http_Header, req, NewKeyword("x-test"), NewString("yes"))
	req = Call(GoBind_http_Body, req, NewString("hello"))

	resp := Call(GoBind_http_Send, req)
	if !isType[*mapImpl](resp) {
		t.Fatalf("expected response map, got %s", resp.String())
	}
	if got := Get(resp, NewKeyword("status")); !isType[longImpl](got) || got.Long() != http.StatusCreated {
		t.Fatalf("expected 201 status, got %s", got.String())
	}
	if got := Get(resp, NewKeyword("ok")); !isType[boolImpl](got) || !got.Bool() {
		t.Fatalf("expected ok response, got %s", got.String())
	}
	if got := Get(resp, NewKeyword("body")); !isType[*stringImpl](got) || got.StringValue() != "world" {
		t.Fatalf("expected response body world, got %s", got.String())
	}
	headers := Get(resp, NewKeyword("headers"))
	if got := Get(headers, NewKeyword("x-reply")); !isType[*stringImpl](got) || got.StringValue() != "pong" {
		t.Fatalf("expected response header pong, got %s", got.String())
	}
}

func TestHTTPBridgeListenRoutes(t *testing.T) {
	handler := NewFunction(func(args ...Value) Value {
		if len(args) != 1 {
			t.Fatalf("expected one request arg, got %d", len(args))
		}
		req := args[0]
		if got := Get(req, NewKeyword("method")); !isType[*stringImpl](got) || got.StringValue() != http.MethodGet {
			t.Fatalf("expected GET method, got %s", got.String())
		}
		if got := Get(Get(req, NewKeyword("query")), NewKeyword("lang")); !isType[*stringImpl](got) || got.StringValue() != "flag" {
			t.Fatalf("expected query lang=flag, got %s", got.String())
		}
		params := Get(req, NewKeyword("params"))
		name := Get(params, NewKeyword("name"))
		if !isType[*stringImpl](name) {
			t.Fatalf("expected route param name, got %s", name.String())
		}
		return NewMap(
			NewKeyword("status"), NewLong(http.StatusAccepted),
			NewKeyword("headers"), NewMap(NewKeyword("x-name"), name),
			NewKeyword("body"), NewString("hello "+name.StringValue()),
		)
	})

	route := NewMap(
		NewKeyword("method"), NewString("GET"),
		NewKeyword("path"), NewString("/hello/:name"),
		NewKeyword("handler"), handler,
	)
	server := Call(GoBind_http_Listen, NewMap(
		NewKeyword("addr"), NewString("127.0.0.1:0"),
		NewKeyword("routes"), NewArray(route),
	))
	addr := Call(GoBind_http_Address, server)
	if !isType[*stringImpl](addr) || addr.StringValue() == "" {
		t.Fatalf("expected server address, got %s", addr.String())
	}

	resp, err := http.Get("http://" + addr.StringValue() + "/hello/world?lang=flag")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Name"); got != "world" {
		t.Fatalf("expected X-Name=world, got %q", got)
	}
	if string(body) != "hello world" {
		t.Fatalf("expected body hello world, got %q", body)
	}

	methodResp, err := http.Post("http://"+addr.StringValue()+"/hello/world", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("Post returned error: %v", err)
	}
	defer methodResp.Body.Close()
	if methodResp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", methodResp.StatusCode)
	}

	_ = Call(GoBind_http_Stop, server)
}
