package transfer

import (
	"encoding/base64"
	"io"
)

// NewBase64Encoder will translate all bytes written to the returned
// io.WriteCloser into base64 encoding and write those to the give io.Writer.
func NewBase64Encoder(w io.Writer) io.WriteCloser {
	return &writer{
		base64.NewEncoder(base64.StdEncoding, w),
		true,
	}
}

// NewBase64Decoder will translate all bytes read from the given io.Reader as
// base64 and return the binary data to the returned io.Reader.
func NewBase64Decoder(r io.Reader) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, r)
}
