package http

// Status is an HTTP status code.
//
// See RFC xxxx
//
// 1xx:
// 2xx:
// 3xx:
// ...
// TODO
type Status struct {
	code uint16
	text string
}

func NewStatus(code int, text string) Status {
	return Status{code: uint16(code), text: text}
}

// Code returns the numeric HTTP status code.
func (s Status) Code() int { return int(s.code) }

// Text returns the textual description of the status code.
//
// For HTTP/1.x responses, this is exactly what the server sent in the
// "reason-phrase" of the "status-line", per RFC 7230 Section 3.1.2.
//
// For HTTP/2 responses, which dropped reason phrases, this is from a
// table of common status codes.
//
// The text does not contain the number.
func (s Status) Text() string {
	if s.text != "" {
		return s.text
	}
	panic("TODO")
}

// String returns a combination of the Code and Text in the form of
// "200 OK".
func (s Status) String() string {
	panic("TODO")
}

// PermitsResponseBody reports whether the status code s permits
// having a response body.
func (s Status) PermitsResponseBody() bool { panic("TODO") }

// IsSuccess reports whether s is in the Successful (2xx) status code class,
// as defined by RFC 7231 section 6.3.
func (s Status) IsSuccess() bool {
	return s.code >= 200 && s.code <= 299
}

// IsClientError reports whether s is in the Client Error (4xx) status code class,
// as defined by RFC 7231 section 6.5.
func (s Status) IsClientError() bool {
	return s.code >= 400 && s.code <= 499
}

// IsServerError reports whether s is in the Server Error (5xx) status code class,
// as defined by RFC 7231 section 6.6.
func (s Status) IsServerError() bool {
	return s.code >= 500 && s.code <= 599
}

// IsNotModified reports whether s is the 304 Not Modified status.
func (s Status) IsNotModified() bool { return s.code == 304 }
