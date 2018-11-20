# Problems with the net/http Client API

An incomplete list:

## Overloaded package types

The [`net/http`](https://golang.org/pkg/net/http/) package reuses several types (notably Request) for both Server and Client, with differing semantics on the struct fields.

Examples:

* https://golang.org/pkg/net/http/#Request.URL
* https://golang.org/pkg/net/http/#Request.Body
* https://golang.org/pkg/net/http/#Request.Header
* https://golang.org/pkg/net/http/#Request.Close
* https://golang.org/pkg/net/http/#Request.Host
* https://golang.org/pkg/net/http/#Request.Form

## Too easy to not call Response.Body.Close.

It's too easy to not close a Response.Body and leak or not reuse connections.

... no lifetime/scope after which the package can clean up for the user.

## Too easy to not check return status codes

...

## Proper usage is too many lines of boilerplate

* NewRequest returning an error
* Contexts
* Error checks
* Status checks
* Closing body

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

## Untyped HTTP Statuses

