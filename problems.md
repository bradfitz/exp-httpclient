**Alternate form:** [2018-12-18 slides](https://docs.google.com/presentation/d/e/2PACX-1vTTFQjMSxai7TvhBJgkJf4K3RrT3tJrP7KUQ3rZB8e4UL7grCnxQh7o4yYYvyYugnkcfVwvTrwA23B0/pub?start=false&loop=false&delayms=3000&slide=id.gc6f73a04f_0_0)

# Problems with the net/http Client API

This page collects problems with the existing `net/http` client interface.

Consider the following typical code you see Go programmers write:

```go
func GetFoo() (*T, error) {
    res, err := http.Get("http://foo/t.json")
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

The code above forgets to call `Response.Body.Close`, which means we
leak a TCP connection, some goroutines, and a file descriptor.

We say that closing a `Response.Body` is the responsibility of the
user, but for better or worse we don't require it when the body was
read to EOF. In that case we immediately recycle the connection to the
internal connection pool. That might train people to forget to close
it. In any case, it's not an interesting detail that users should need
to worry about.

The code above may leak the Body for two reasons:

1. invalid JSON causes the Decoder to return early
2. the JSON is valid, but after the decoded value (say, a JSON object), there is an unbuffered `"\n"`, which the JSON decoder doesn't need to read to return, but the keeps the connection alive. ([#20528](https://github.com/golang/go/issues/20528))

Unlike the HTTP server's
[`Handler`](https://golang.org/pkg/net/http/#Handler), on the client
side we have no scope after which we can do cleanup for the caller.

Fortunately the `Response.Body` is defined to always be non-nil so to
fix that, we defer a `Close` to cover both exit paths:

```go
func GetFoo() (*T, error) {
    res, err := http.Get("http://foo/t.json")
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
    res, err := http.Get("http://foo/t.json")
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

## Context support is oddly bolted on

The code above doesn't use contexts.

Context support was added late (in Go 1.7) with, and the only way to make a request with a context
is to make an expensive not-fully-deep but not-super-shallow clone of a Request with
[`Request.WithContext`](https://golang.org/pkg/net/http/#Request.WithContext).

If you knew your requests shouldn't take longer than 5 seconds but you
always wanted to accept a context for cancelation, you'd write
something like this today:Today you'd write somethin to write:

```go
func GetFoo(ctx context.Context) (*T, error) {
    req, err := http.NewRequest("GET", "http://foo/t.json", nil)
    if err != nil {
        return nil, err 
    }
    req = req.WithContext(ctx)
    res, err := http.DefaultClient.Do(req)
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

## Proper usage is too many lines of boilerplate

The latest code in the prior section is finally complete, but it's
super verbose.

It's no surprise that people commonly skip some of it until their
omission causes problems.

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

## HTTP/2 support is oddly bolted on

* No HTTP/2-specific API
* Magic and confusing auto-upgrading to HTTP/2
  * https://golang.org/issue/21336 - `bogus greeting when providing TLSClientConfig`
* The connection pool management (especially for new connections of
  unknown types) between HTTP/1 and HTTP/2 is ... special. And people
  want more control, but we lack the types to give them control, given
  our weird split over two packages.

Amusingly, the HTTP/2 support works because it latches onto an
otherwise-unused mechanism
([`Transport.RegisterProtocol`](https://golang.org/pkg/net/http/#Transport.RegisterProtocol))
we added to support non-HTTP client support, such as "file" or "ftp".
Turns out nobody used that. We should probably remove it when making HTTP/2 more integrated.

## Errors are not consistent or well defined

* many exported error variables are no longer used
* TODO: bug reference
* TODO: reference issue of returning non-zero for both (Response, error) on body write error with header response (e.g. Unauthorized on a large POST)

## httputil.ClientConn

Prior to Go 1, the net/http package had a ClientConn type (a precursor
to the Transport) that we moved to [`net/http/httputil.ClientConn`](https://golang.org/pkg/net/http/httputil/#ClientConn)
just before releasing Go 1. We should've deleted it instead. Today it's documented like:

> ClientConn is an artifact of Go's early HTTP implementation. It is
> low-level, old, and unused by Go's current HTTP stack. We should
> have deleted it before Go 1.
>
> Deprecated: Use Client or Transport in package net/http instead.

At least that documentation seems to have stopped the bug reports and
feature requests.

Unfortunately it lives in the same package as `ReverseProxy`, which is
high quality, maintained, and widely used. To have such an old buggy
relic next to a nice type surely gives some users second thoughts
about the nice part.
