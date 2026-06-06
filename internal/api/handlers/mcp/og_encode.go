package mcp

import (
	"image"
	"image/png"
	"io"
)

// pngEncode is a tiny indirection so tests can swap it for an in-memory
// encoder. Production uses the standard library PNG encoder.
var pngEncode = func(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}
