package runtime

import (
	"io"
	"net/http"
	"net/http/httptest"
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
	if resp.tag != TagMap {
		t.Fatalf("expected response map, got %s", ValueToString(resp))
	}
	if got := Get(resp, NewKeyword("status")); got.tag != TagLong || got.Long() != http.StatusCreated {
		t.Fatalf("expected 201 status, got %s", ValueToString(got))
	}
	if got := Get(resp, NewKeyword("ok")); got.tag != TagBool || !got.Bool() {
		t.Fatalf("expected ok response, got %s", ValueToString(got))
	}
	if got := Get(resp, NewKeyword("body")); got.tag != TagString || got.StringValue() != "world" {
		t.Fatalf("expected response body world, got %s", ValueToString(got))
	}
	headers := Get(resp, NewKeyword("headers"))
	if got := Get(headers, NewKeyword("x-reply")); got.tag != TagString || got.StringValue() != "pong" {
		t.Fatalf("expected response header pong, got %s", ValueToString(got))
	}
}
