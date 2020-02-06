package tctypes

import (
	"bytes"
	"net/http"
	"net/http/httputil"
)

// DefaultTransport is the default transport used by the HTTP logger.
var DefaultTransport = http.DefaultTransport

// HTTPLogger provides a logging http.RoundTripper transport.
//
// Handles logging of HTTP requests and responses to standard logging funcs.
type HTTPLogger struct {
	transport http.RoundTripper
	reqf      func([]byte)
	resf      func([]byte)
}

// NewHTTPLogger creates a new HTTP transport.
func NewHTTPLogger(transport http.RoundTripper, reqf, resf func([]byte)) *HTTPLogger {
	return &HTTPLogger{
		transport: transport,
		reqf:      reqf,
		resf:      resf,
	}
}

// NewHTTPLogf creates a new HTTP transport that logs to the provided logging
// function for the provided transport.
//
// Prefixes "-> " and "<- " to each line of the HTTP request, response, and an
// additional blank line ("\n\n") to the output.
func NewHTTPLogf(transport http.RoundTripper, logf func(string, ...interface{})) *HTTPLogger {
	nl := []byte("\n")
	f := func(prefix []byte, buf []byte) {
		buf = append(prefix, bytes.ReplaceAll(buf, nl, append(nl, prefix...))...)
		logf("%s\n\n", string(buf))
	}
	return NewHTTPLogger(
		transport,
		func(buf []byte) { f([]byte("-> "), buf) },
		func(buf []byte) { f([]byte("<- "), buf) },
	)
}

// RoundTrip satisfies the http.RoundTripper interface.
func (hl *HTTPLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	trans := hl.transport
	if trans == nil {
		trans = DefaultTransport
	}

	reqBody, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	res, err := trans.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resBody, err := httputil.DumpResponse(res, true)
	if err != nil {
		return nil, err
	}

	hl.reqf(reqBody)
	hl.resf(resBody)

	return res, err
}
