package transfer

import (
	"io"
	"mime/quotedprintable"
)

// NewQuotedPrintableEncoder will transform all bytes written to the returned
// io.WriteCloser into quoted-printable form and write them to the given
// io.Writer.
func NewQuotedPrintableEncoder(w io.Writer) io.WriteCloser {
	qpw := quotedprintable.NewWriter(w)
	return &writer{qpw, qpw}
}

// NewQuotedPrintableDecoder will read bytes from the given io.Reader and return
// them in the returned io.Reader after decoding them from quoted-printable
// format.
func NewQuotedPrintableDecoder(r io.Reader) io.Reader {
	return quotedprintable.NewReader(r)
}
