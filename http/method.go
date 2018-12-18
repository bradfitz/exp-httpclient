// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// Method is an HTTP method. Although HTTP methods are case insensitive,
// values of this type must contain only capital letters.
type Method string

// Common HTTP methods.
//
// Unless otherwise noted, these are defined in RFC 7231 section 4.3.
const (
	Get     Method = "GET"
	Head    Method = "HEAD"
	Post    Method = "POST"
	Put     Method = "PUT"
	Patch   Method = "PATCH" // RFC 5789
	Delete  Method = "DELETE"
	Connect Method = "CONNECT"
	Options Method = "OPTIONS"
	Trace   Method = "TRACE"
)

func (m Method) RequestBodyAllowed() bool {
	panic("TODO")
}

func (m Method) RequestBodyCommon() bool {
	panic("TODO")
}

func (m Method) ResponseBodyAllowed() bool {
	panic("TODO")
}
