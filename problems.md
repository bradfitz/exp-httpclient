# Problems with the net/http Client API

This page collects problems with the existing `net/http` client interface.

Consider the following typical code you see Go programmers write:

```go
func GetFoo() (*T, error) {
    res, err := http.Get("http://example.com")
    if err != nil {
        return nil, err
    }
    t := new(T)
    if err := json.NewDecoder(res.Body).Decode(t); err != nil {
        return nil, err
    }
    return t, nil
```

There are several problems with that code, listed below.

## Too easy to not call Response.Body.Close.

The code above forgets to call `Response.Body.Close`. But if we read
to EOF, the close isn't strictly required, but if the JSON above is
invalid then it leaks the Transport's internal TCP connection, keeping
it open forever and tying up some memory and a file descriptor.

Unlike the HTTP server's
[`Handler`](https://golang.org/pkg/net/http/#Handler), on the client
side we have no scope after which we can do cleanup for the caller.

Fortunately the `Response.Body` is defined to always be non-nil so to
fix that, we defer a `Close` to cover both exit paths:

```go
func GetFoo() (*T, error) {
    res, err := http.Get("http://example.com")
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    t := new(T)
    if err := json.NewDecoder(res.Body).Decode(t); err != nil {
        return nil, err
    }
    return t, nil
```


## Too easy to not check return status codes

The code above forgets to check the HTTP status in
`Response.StatusCode`. We probably only cared about 2xx responses, so:

```go
func GetFoo() (*T, error) {
    res, err := http.Get("http://example.com")
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    if res.StatusCode < 200 || res.StatusCode > 299 {
        return nil, fmt.Errorf("bogus status: got %v", res.Status)
    }
    t := new(T)
    if err := json.NewDecoder(res.Body).Decode(t); err != nil {
        return nil, err
    }
    return t, nil
```

## Untyped HTTP Statuses

The status code is just an untyped integer. It'd be better as type so it could have methods to ask
which class it's in, and get a `String` representation without having two redundant fields:

* https://golang.org/pkg/net/http/#Response.StatusCode (the int)
* https://golang.org/pkg/net/http/#Response.Status (the string, which we have to synthesize anyway for HTTP/2 where it doesn't appear on the wire)

## Proper usage is too many lines of boilerplate

* NewRequest returning an error
* Contexts
* Error checks
* Status checks
* Closing body




## Overloaded package types

The [`net/http`](https://golang.org/pkg/net/http/) package reuses several types
(notably [`Request`](https://golang.org/pkg/net/http/#Request)) for both Server and Client,
with differing semantics on the struct fields.

Examples:

* https://golang.org/pkg/net/http/#Request.URL -- "For server requests the URL is parsed from the URI
  supplied on the Request-Line as stored in RequestURI.  For
  most requests, fields other than Path and RawQuery will be
  empty. For client requests, the URL's Host specifies the server to
  connect to, while the Request's Host field optionally
  specifies the Host header value to send in the HTTP
  request"
* https://golang.org/pkg/net/http/#Request.Header -- different fields are special for client vs server and omitted or prioritized from other fields (or automatic). This regularly confuses people.
* https://golang.org/pkg/net/http/#Request.Body -- ReadCloser, but package closes for client, but user closes for server
* https://golang.org/pkg/net/http/#Request.GetBody -- unused by server
* https://golang.org/pkg/net/http/#Request.ContentLength -- Negative one means unknown, but for client requests 0 also means unknown, or might mean actually zero. So we had to introduce the [`http.NoBody` variable](https://golang.org/pkg/net/http/#NoBody) to disambiguate.
* https://golang.org/pkg/net/http/#Request.TransferEncoding -- effectively unused, as chunking (the only common Transfer-Encoding) is automatic
* https://golang.org/pkg/net/http/#Request.Close -- used by clients, but not server (and its use by clients is a bit weird with HTTP/2). This would be better handled with some connection pool abstraction
* https://golang.org/pkg/net/http/#Request.Host -- for servers, what the client sent in HTTP/1 Host header or in HTTP/2 `:authority` (unspecified what happens if both are present). For clients, it's optional and overrides the Host header sent in HTTP/1 requests, and maybe the HTTP/2 `:authority`. TODO: look it up.
* https://golang.org/pkg/net/http/#Request.Form -- just a place for ParseForm (as called by server handlers) to stash stuff. Ignored by the client.
* https://golang.org/pkg/net/http/#Request.PostForm -- same as Form
* https://golang.org/pkg/net/http/#Request.MultipartForm -- same as Form
* https://golang.org/pkg/net/http/#Request.Trailer -- reasonably consistent, but complicated: for clients, a map that must be half populated (keys) at the beginning, and then fully populated (the values) before the body returns EOF. For servers, the same: the map gets populated at body EOF.
* https://golang.org/pkg/net/http/#Request.RemoteAddr -- only for servers, and loosely defined ("and has no defined format") but usually "ip:port"-ish.
* https://golang.org/pkg/net/http/#Request.RequestURI -- only for servers
* https://golang.org/pkg/net/http/#Request.TLS -- only for servers
* https://golang.org/pkg/net/http/#Request.Cancel -- only for clients, and deprecated (see below)
* https://golang.org/pkg/net/http/#Request.Response -- only for clients

## Types too transparent

The HTTP Request, Response, and Header types are too transparent and
generate too much garbage even when callers aren't interested in any
of their fields. We can't lazily parse or construct things with the
current API. The [fasthttp](https://github.com/valyala/fasthttp) package is a response to this, which claims:

> Fast HTTP package for Go. Tuned for high performance. Zero memory allocations in hot paths. Up to 10x faster than net/http.

## HTTP Header case

HTTP headers are defined as case insensitive, but Go [defines them](https://golang.org/pkg/net/http/#Header) as:

```go
// A Header represents the key-value pairs in an HTTP header.
type Header map[string][]string
```

That generally works, as long as users know about
[`CanonicalHeaderKey`](https://golang.org/pkg/net/http/#CanonicalHeaderKey),
but it regularly surprises people.

Also, the `[]string` value type could probably be its own named type
to permit methods to search for case insensitive HTTP
comma-and-whitespace-delimited tokens that are common in many protocols.

## Client vs. Transport distinction confuses people

Our HTTP client has two main types, which people are regularly confused by:

* `Client` is light, stateless, and mostly only handles redirect policy and timeouts.
* `Transport` is heavy and caches connections (it's more like a "connection pool", if we had that type) and does all the real work, but doesn't follow redirects.

To add to the confusion, we also have a `RoundTripper` interface,
which `Transport` implements, and `Client` almost implements, but has
a different method name (`Do` instead of `RoundTrip`).

The actual type that people pass around in their varies between the
three.

## Four ways to cancel requests

Four generations of HTTP cancelation:

* Go 1.1: https://golang.org/pkg/net/http/#Transport.CancelRequest
* Go 1.3: https://golang.org/pkg/net/http/#Client.Timeout
* Go 1.5: https://golang.org/pkg/net/http/#Request.Cancel
* Go 1.7: https://golang.org/pkg/net/http/#Request.WithContext

That's a lot of API bloat for users to read, and a pain for us to maintain.

## Context support is oddly bolted on

Context support was added late (in Go 1.7) with, and the only way to make a request with a context
is to make an expensive not-fully-deep but not-super-shallow clone of a Request with
[`Request.WithContext`](https://golang.org/pkg/net/http/#Request.WithContext).

It should be much easier. Timeouts should also be much easier
per-request for people who don't want to make a new
`context.WithTimeout` and remember to cancel it.

## HTTP/2 support is oddly bolted on

* No HTTP/2-specific API
* Magic and confusing auto-upgrading to HTTP/2

## Errors are not consistent or well defined

* TODO: bug reference
* TODO: reference issue of returning non-zero for both (Response, error) on body write error with header response (e.g. Unauthorized on a large POST)


