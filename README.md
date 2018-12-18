# Rethinking Go's HTTP client

This repository explores redesigning the API for
the [Go language](https://golang.org/)'s
[`net/http`](https://golang.org/pkg/net/http/)
[`Client`](https://golang.org/pkg/net/http/#Client) and 
[`Transport`](https://golang.org/pkg/net/http/#Transport).

# FAQ

## What's wrong with Go's HTTP client?

See the [list of problems](problems.md).

## What about the Server?

This repo does not aim to address the server side of the `net/http`
package. The server half is in better shape than the client, and it's
also easier to fix the client half without fragmenting the
ecosystem. Changing the Server interface needs to be done much more
carefully.

But even long term, it's almost certainly best for the client and server to
live in separate packages. They might share some types & code from
shared HTTP package(s).

## Who's leading this effort?

Brad Fitzpatrick, [@bradfitz](https://github.com/bradfitz). I've owned
the net/http package for over 8 years and have plenty of gripes about
it. I welcome all input. If we're going to finally change it, we
should get it right, so there's no need to rush this process.

## What's the plan?

Roughly:

* Iterate on the API & godoc repeatedly until it looks right (with a fake, `panic("TODO")`-only implementation)
* Discuss, revise.
* Add a temporary implementation (likely inefficient), wrapping the existing net/http Client.
* Port code to use it. See if we're still happy.
* Discuss, revise.
* Copy `net/http` and `golang.org/x/net/http2` code into `httpclient` (likely several packages).
* Benchmark, tune, revise API as needed.
* Redo the "legacy" `net/http` and `golang.org/x/net/http2` client APIs as wrappers around `httpclient`

Of course, this is all up for debate.
