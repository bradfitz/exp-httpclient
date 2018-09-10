# Problems with the net/http Client API

An incomplete list:

## Overloaded package types

The [`net/http`](https://golang.org/pkg/net/http/) package reuses several types (notably Request) for both Server and Client, with differing semantics on the struct fields.

## Too easy to not call Response.Body.Close.

It's too easy to not close a Response.Body and leak or not reuse connections.

... no lifetime/scope after which the package can clean up for the user.

## Too easy to not check return status codes

...

## Proper usage is too many lines of boilerplate

.. NewRequest returning an error

## Types too transparent

* hard to optimize, generate too much garbage

## Client vs. Transport distinction confuses people

## Three ways to cancel requests

## Context support is oddly bolted on

## HTTP/2 support is oddly bolted on

* No HTTP/2-specific API
* Magic and confusing auto-upgrading to HTTP/2

## Errors are not consistent or well defined

* TODO: bug reference
* TODO: reference issue of returning non-zero for both (Response, error) on body write error with header response (e.g. Unauthorized on a large POST)

## Untyped HTTP Statuses

