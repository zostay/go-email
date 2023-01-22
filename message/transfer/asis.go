package transfer

import "io"

// NewAsIsEncoder returns an io.WriteCloser that writes bytes as-is.
func NewAsIsEncoder(w io.Writer) io.WriteCloser {
	return &writer{w, nil}
}

// NewAsIsDecoder returns an io.Reader that reads bytes as-is.
func NewAsIsDecoder(r io.Reader) io.Reader {
	return r
}
