package mcp

import (
	"bytes"
	"io"
	"net/url"
)

// bytesReader returns a *bytes.Reader for the given bytes. Defined as a tiny
// helper so executor.go doesn't have to import bytes/io directly.
func bytesReader(b []byte) io.Reader {
	return bytes.NewReader(b)
}

// newURL parses a path into a *url.URL. Used internally for synthetic
// request URLs in tests.
func newURL(path string) *url.URL {
	return &url.URL{Path: path}
}
