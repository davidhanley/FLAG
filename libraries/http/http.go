package httplib

import (
	"bytes"
	"fmt"
	"io"
	stdhttp "net/http"
	"strings"
	"time"
)

type Request struct {
	Method          string
	URL             string
	Headers         stdhttp.Header
	Body            []byte
	Timeout         time.Duration
	FollowRedirects bool
}

type Response struct {
	StatusCode int
	Status     string
	URL        string
	Method     string
	Headers    stdhttp.Header
	Body       []byte
}

func NewRequest(method, url string) Request {
	return Request{
		Method:          strings.ToUpper(strings.TrimSpace(method)),
		URL:             strings.TrimSpace(url),
		Headers:         stdhttp.Header{},
		FollowRedirects: true,
	}
}

func (r Request) WithHeader(key string, values ...string) Request {
	clone := r.clone()
	clone.Headers.Add(stdhttp.CanonicalHeaderKey(strings.TrimSpace(key)), strings.Join(values, ", "))
	return clone
}

func (r Request) WithBody(body []byte) Request {
	clone := r.clone()
	clone.Body = append([]byte(nil), body...)
	return clone
}

func (r Request) WithHeaders(headers stdhttp.Header) Request {
	clone := r.clone()
	for key, values := range headers {
		for _, value := range values {
			clone.Headers.Add(stdhttp.CanonicalHeaderKey(key), value)
		}
	}
	return clone
}

func (r Request) clone() Request {
	clone := r
	clone.Headers = cloneHeader(r.Headers)
	clone.Body = append([]byte(nil), r.Body...)
	return clone
}

func Send(req Request) (Response, error) {
	if req.Method == "" {
		req.Method = stdhttp.MethodGet
	}
	if req.URL == "" {
		return Response{}, fmt.Errorf("request url cannot be empty")
	}
	body := bytes.NewReader(req.Body)
	httpReq, err := stdhttp.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return Response{}, err
	}
	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	client := &stdhttp.Client{}
	if req.Timeout > 0 {
		client.Timeout = req.Timeout
	}
	if !req.FollowRedirects {
		client.CheckRedirect = func(_ *stdhttp.Request, _ []*stdhttp.Request) error {
			return stdhttp.ErrUseLastResponse
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		URL:        resp.Request.URL.String(),
		Method:     resp.Request.Method,
		Headers:    cloneHeader(resp.Header),
		Body:       bodyBytes,
	}, nil
}

func Get(url string) (Response, error) {
	return Send(NewRequest(stdhttp.MethodGet, url))
}

func Post(url string, body []byte) (Response, error) {
	req := NewRequest(stdhttp.MethodPost, url)
	req.Body = append([]byte(nil), body...)
	return Send(req)
}

func cloneHeader(header stdhttp.Header) stdhttp.Header {
	clone := make(stdhttp.Header, len(header))
	for key, values := range header {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}
