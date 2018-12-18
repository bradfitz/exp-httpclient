// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "crypto/tls"

// Protocol represents the HTTP version and TLS connection state.
type Protocol struct {
	major, minor byte
	tls          *tls.ConnectionState // non-nil if TLS
}

func (p Protocol) Major() int {
	return int(p.major)
}

func (p Protocol) Minor() int {
	return int(p.minor)
}

func (p Protocol) IsTLS() bool { return p.tls != nil }

// TODO
