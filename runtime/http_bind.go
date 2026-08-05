package runtime

import (
	"fmt"
	"io"
	"strings"
	"time"

	httplib "flag-lang/libraries/http"
)

var (
	GoBind_http_Request = NewFunction(adaptHTTPRequest)
	GoBind_http_Header  = NewFunction(adaptHTTPHeader)
	GoBind_http_Body    = NewFunction(adaptHTTPBody)
	GoBind_http_Send    = NewFunction(adaptHTTPSend)
	GoBind_http_Get     = NewFunction(adaptHTTPGet)
	GoBind_http_Post    = NewFunction(adaptHTTPPost)
)

func adaptHTTPRequest(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic(fmt.Sprintf("http/request expects 2 or 3 arguments, got %d", len(args)))
	}
	req := httplib.NewRequest(goArgString("http/request", 0, args[0]), goArgString("http/request", 1, args[1]))
	if len(args) == 3 {
		req = applyHTTPRequestOptions("http/request", req, args[2])
	}
	return httpRequestValue(req)
}

func adaptHTTPHeader(args ...Value) Value {
	goArgArityExact("http/header", args, 3)
	req := httpRequestFromValue("http/header", args[0])
	key := httpHeaderKey(args[1])
	value := httpHeaderValue("http/header", args[2])
	req.Headers = req.Headers.Clone()
	req.Headers.Add(key, value)
	return httpRequestValue(req)
}

func adaptHTTPBody(args ...Value) Value {
	goArgArityExact("http/body", args, 2)
	req := httpRequestFromValue("http/body", args[0])
	req.Body = httpBodyBytes("http/body", args[1])
	return httpRequestValue(req)
}

func adaptHTTPSend(args ...Value) Value {
	goArgArityExact("http/send", args, 1)
	resp := httpSendFromValue("http/send", args[0])
	return httpResponseValue(resp)
}

func adaptHTTPGet(args ...Value) Value {
	if len(args) != 1 && len(args) != 2 {
		panic(fmt.Sprintf("http/get expects 1 or 2 arguments, got %d", len(args)))
	}
	req := httplib.NewRequest("GET", goArgString("http/get", 0, args[0]))
	if len(args) == 2 {
		req = applyHTTPRequestOptions("http/get", req, args[1])
	}
	resp, err := httplib.Send(req)
	if err != nil {
		panic(err.Error())
	}
	return httpResponseValue(resp)
}

func adaptHTTPPost(args ...Value) Value {
	if len(args) != 2 && len(args) != 3 {
		panic(fmt.Sprintf("http/post expects 2 or 3 arguments, got %d", len(args)))
	}
	req := httplib.NewRequest("POST", goArgString("http/post", 0, args[0]))
	req.Body = httpBodyBytes("http/post", args[1])
	if len(args) == 3 {
		req = applyHTTPRequestOptions("http/post", req, args[2])
	}
	resp, err := httplib.Send(req)
	if err != nil {
		panic(err.Error())
	}
	return httpResponseValue(resp)
}

func httpRequestFromValue(name string, v Value) httplib.Request {
	if v.tag != TagMap {
		panic(fmt.Sprintf("%s expects a request map, got %s", name, ValueToString(v)))
	}
	method := requestFieldString(name, v, "method", "GET")
	url := requestFieldString(name, v, "url", "")
	req := httplib.NewRequest(method, url)

	if headers := Get(v, NewKeyword("headers")); headers.tag != TagNil {
		req.Headers = requestHeadersFromValue(name, headers)
	}
	if body := Get(v, NewKeyword("body")); body.tag != TagNil {
		req.Body = httpBodyBytes(name, body)
	}
	if timeout := Get(v, NewKeyword("timeout-ms")); timeout.tag != TagNil {
		req.Timeout = time.Duration(goArgInt64(name, 0, timeout)) * time.Millisecond
	}
	if follow := Get(v, NewKeyword("follow-redirects")); follow.tag != TagNil {
		req.FollowRedirects = goArgBool(name, 0, follow)
	}
	return req
}

func applyHTTPRequestOptions(name string, req httplib.Request, opts Value) httplib.Request {
	if opts.tag == TagNil {
		return req
	}
	if opts.tag != TagMap {
		panic(fmt.Sprintf("%s options: expected map, got %s", name, ValueToString(opts)))
	}
	if headers := Get(opts, NewKeyword("headers")); headers.tag != TagNil {
		req.Headers = requestHeadersFromValue(name, headers)
	}
	if body := Get(opts, NewKeyword("body")); body.tag != TagNil {
		req.Body = httpBodyBytes(name, body)
	}
	if timeout := Get(opts, NewKeyword("timeout-ms")); timeout.tag != TagNil {
		req.Timeout = time.Duration(goArgInt64(name, 0, timeout)) * time.Millisecond
	}
	if follow := Get(opts, NewKeyword("follow-redirects")); follow.tag != TagNil {
		req.FollowRedirects = goArgBool(name, 0, follow)
	}
	return req
}

func httpSendFromValue(name string, v Value) httplib.Response {
	req := httpRequestFromValue(name, v)
	resp, err := httplib.Send(req)
	if err != nil {
		panic(err.Error())
	}
	return resp
}

func requestFieldString(name string, req Value, field string, fallback string) string {
	value := Get(req, NewKeyword(field))
	if value.tag == TagNil {
		return fallback
	}
	return goArgString(name, 0, value)
}

func httpHeaderKey(v Value) string {
	switch v.tag {
	case TagSymbol:
		symbol := v.SymbolObject()
		return strings.ToLower(symbol.Name)
	case TagString:
		return strings.ToLower(v.StringValue())
	default:
		return strings.ToLower(anyToString(ValueToAny(v)))
	}
}

func httpHeaderValue(name string, v Value) string {
	return anyToString(v)
}

func requestHeadersFromValue(name string, v Value) map[string][]string {
	if v.tag == TagNil {
		return nil
	}
	if v.tag != TagMap {
		panic(fmt.Sprintf("%s headers: expected map, got %s", name, ValueToString(v)))
	}
	headers := map[string][]string{}
	for _, entry := range v.MapEntries() {
		key := httpHeaderKey(entry.Key)
		for _, value := range headerValuesFromValue(name, entry.Value) {
			headers[key] = append(headers[key], value)
		}
	}
	return headers
}

func headerValuesFromValue(name string, v Value) []string {
	switch v.tag {
	case TagNil:
		return nil
	case TagArray:
		values := v.ArrayValues()
		out := make([]string, 0, len(values))
		for _, item := range values {
			out = append(out, httpHeaderValue(name, item))
		}
		return out
	case TagList:
		values := v.ListValues()
		out := make([]string, 0, len(values))
		for _, item := range values {
			out = append(out, httpHeaderValue(name, item))
		}
		return out
	default:
		return []string{httpHeaderValue(name, v)}
	}
}

func httpBodyBytes(name string, v Value) []byte {
	switch v.tag {
	case TagNil:
		return nil
	case TagFile:
		reader := goArgReader(name, 0, v)
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(fmt.Sprintf("%s body: %v", name, err))
		}
		return body
	default:
		if reader, ok := ValueToAny(v).(io.Reader); ok {
			body, err := io.ReadAll(reader)
			if err != nil {
				panic(fmt.Sprintf("%s body: %v", name, err))
			}
			return body
		}
		return []byte(anyToString(v))
	}
}

func httpRequestValue(req httplib.Request) Value {
	headers := NewMap()
	for key, values := range req.Headers {
		switch len(values) {
		case 0:
			continue
		case 1:
			headers = Assoc(headers, NewKeyword(strings.ToLower(key)), NewString(values[0]))
		default:
			items := make([]Value, 0, len(values))
			for _, value := range values {
				items = append(items, NewString(value))
			}
			headers = Assoc(headers, NewKeyword(strings.ToLower(key)), NewArray(items...))
		}
	}
	out := NewMap(
		NewKeyword("method"), NewString(req.Method),
		NewKeyword("url"), NewString(req.URL),
		NewKeyword("headers"), headers,
		NewKeyword("body"), NewString(string(req.Body)),
		NewKeyword("timeout-ms"), NewLong(int64(req.Timeout/time.Millisecond)),
		NewKeyword("follow-redirects"), NewBool(req.FollowRedirects),
	)
	return out
}

func httpResponseValue(resp httplib.Response) Value {
	headers := NewMap()
	for key, values := range resp.Headers {
		switch len(values) {
		case 0:
			continue
		case 1:
			headers = Assoc(headers, NewKeyword(strings.ToLower(key)), NewString(values[0]))
		default:
			items := make([]Value, 0, len(values))
			for _, value := range values {
				items = append(items, NewString(value))
			}
			headers = Assoc(headers, NewKeyword(strings.ToLower(key)), NewArray(items...))
		}
	}
	return NewMap(
		NewKeyword("status"), NewLong(int64(resp.StatusCode)),
		NewKeyword("status-text"), NewString(resp.Status),
		NewKeyword("url"), NewString(resp.URL),
		NewKeyword("method"), NewString(resp.Method),
		NewKeyword("ok"), NewBool(resp.StatusCode >= 200 && resp.StatusCode < 300),
		NewKeyword("headers"), headers,
		NewKeyword("body"), NewString(string(resp.Body)),
	)
}
