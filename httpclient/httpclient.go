// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httpclient is an experimental design for a new Go HTTP client.
//
// It does not work. Do not use it. It exists for brainstorming,
// prototyping, and reviewing godoc only.
package httpclient // import "inet.af/httpclient"

import (
	"context"
	"io"
	nethttp "net/http"
	"net/url"
	"strings"
	"time"

	"inet.af/http"
)

// Request is an HTTP client request.
//
// It can only be used once.
type Request struct {
	method string
	url    string

	body io.ReadCloser
}

// NewGet returns a new GET request to the provided URL.
func NewGet(url string) *Request { return NewRequest("GET", url) }

// NewHead returns a new HEAD request to the provided URL.
//
// A Head response never contains a body, so any attempt to read its
// body is an error.
func NewHead(url string) *Request { return NewRequest("HEAD", url) }

// NewPost returns a new POST request to the provided URL.
func NewPost(url string) *Request { return NewRequest("POST", url) }

// NewPut returns a new PUT request to the provided URL.
func NewPut(url string) *Request { return NewRequest("PUT", url) }

// NewRequest returns a new request to the provided URL using the
// provided HTTP method.
func NewRequest(method, url string) *Request {
	return &Request{method: method, url: url}
}

// Body sets the body for the request. If the Body also implements
// io.Closer, it is closed at the end of a request.
//
// If the body implements io.Seeker (such as *bytes.Reader and
// *strings.Reader), its seek position is remembered and restored on
// any necessary automatic retries. As a special case, *bytes.Buffer
// is recognized and promoted to a *bytes.Reader so it is restartable.
// For all other types, RestartableBody should be used instead, so
// requests can be retried.
func (r *Request) Body(body io.Reader) *Request {
	panic("TODO")
	//if rc, ok := r.(io.ReadCloser); ok {
	//r.body = r
	return r
}

// FormValues sets the Request's Content-Type to
// "application/x-www-form-urlencoded" and encodes the provided data
// values as its body.
func (r *Request) FormValues(data url.Values) *Request {
	r.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	r.Body(strings.NewReader(data.Encode()))
	return r

}

// RestartableBody is like Body, but sets a func which returns the
// Body as needed. The function may be called 0, 1, or multiple
// times.
func (r *Request) RestartableBody(getBody func() io.Reader) *Request {
	panic("TODO")
	return r
}

// SetHeader sets the header k to the value v, overwriting any previous values.
func (r *Request) SetHeader(k, v string) *Request {
	panic("TODO")
	return r
}

// SetTrailer sets the trailer k to the value v, overwriting any previous values.
func (r *Request) SetTrailer(k, v string) *Request {
	panic("TODO")
	return r
}

// AddHeader appends the value v to the header k.
func (r *Request) AddHeader(k, v string) *Request {
	panic("TODO")
	return r
}

// AddTrailer appends the value v to the trailer k.
func (r *Request) AddTrailer(k, v string) *Request {
	panic("TODO")
	return r
}

// LimitBytes limits the response bytes.
//
// By default, LimitBytes is bounded to 16 MB, except when writing to disk.
//
// To disable the limit, use a negative number.
func (r *Request) LimitBytes(n int64) *Request {
	panic("TODO")
	return r
}

type RedirectPolicy func(RedirectState) error

type RedirectState struct {
	// TODO: history
}

// RedirectPolicy ...
//
// By default, N redirects are followed.
//
// TODO: specifies cookies.
//
// As a special case, the nil redirect policy disables all redirects.
func (r *Request) RedirectPolicy(policy RedirectPolicy) *Request {
	panic("TODO")
	return r
}

func (r *Request) Jar(jar nethttp.CookieJar) *Request {
	panic("TODO")
	return r
}

type Pool interface {

	// unexported method so we can extend this interface over time
	// without breaking people. Implementers must embed a concrete
	// type from elsewhere.
	unexported()
}

var defaultPool Pool // = TODO

func DefaultPool() Pool {
	return defaultPool
}

// Pool sets the connection pool to use with this connection.
//
//
//
// As a special case, the nil pool disables connection reuse.
func (r *Request) Pool(pool Pool) *Request {
	panic("TODO")
	return r
}

// Header are the response headers.
type Header struct {
	// * opaque value type (try for small struct)
	// * response headers only, so only headers
	// * lazily parse by default

	_ [0]byte
}

func (h Header) Get(key string) string           { panic("TODO") }
func (h Header) GetMultiple(key string) []string { panic("TODO") }
func (h Header) ContainsToken(key, token string) { panic("TODO") }

type Connection struct {
	// * opaque value type

	// TLS info
}

func (c Connection) Protocol() http.Protocol {
	panic("TODO")
}

// Close immediately closes the underlying connection, even if it's
// still in use. To shut it down as soon as gracefully possible, ... TODO.
func (c Connection) Close() error {
	panic("TODO")
}

// Timeout sets the timeout for the entire request, including any
// redirects and reading the response body.
//
// If the context expires first, the request still fails; the Timeout
// cannot extend the context's lifetime.
func (r *Request) Timeout(d time.Duration) *Request {
	panic("TODO")
	return r
}

// Do executes the HTTP request and returns the response.
//
// On error, the the error will be one of:
//
//   -
func (r *Request) Do(ctx context.Context, h Handler) (ResponseData, error) {
	panic("TODO")
}

func JSONUnmarshal(dst interface{}) Handler {
	return HandlerFunc(func(s HandlerState) (ResponseData, error) {
		panic("TODO")
	})
}

// ResponseData represents the response body in its possibly
// unmarhsaled form. The concrete type depends on the the Handler
// used.
//
// If Go gets generics, this type would go away. For now it serves
// as documentation only.
type ResponseData interface{}

// HandlerState is the interface available for Handlers while processing
// an HTTP response from a server.
type HandlerState interface {
	Connection() Connection
	Status() http.Status
	Header() Header
	Body() io.Reader
	Trailer() (Header, error)

	unexported()
}

// Handler is the interface for something that can process an HTTP
// response from a server.
type Handler interface {
	ReadHTTP(HandlerState) (ResponseData, error)
}

// HandlerFunc implements Handler using the underlying func.
type HandlerFunc func(HandlerState) (ResponseData, error)

func (hf HandlerFunc) ReadHTTP(s HandlerState) (ResponseData, error) {
	return hf(s)
}

// StatusError is the error type returned when the status was not 2xx.
type StatusError struct {
	Status http.Status
}
