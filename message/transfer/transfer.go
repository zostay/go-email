package transfer

import (
	"io"

	"github.com/zostay/go-email/v2/message/header"
)

const (
	None            = ""                 // bytes will be left as-is
	Bit7            = "7bit"             // bytes will be left as-is
	Bit8            = "8bit"             // bytes will be left as-is
	Binary          = "binary"           // bytes will be left as-is
	QuotedPrintable = "quoted-printable" // bytes will be transformed between quoted-printable and binary data
	Base64          = "base64"           // bytes will be transformed between base64 and binary data
)

// writer is an internal type to make as-is writers work properly.
type writer struct {
	io.Writer
	performClose bool
}

// Close will close the nested writer if performClose is true.
func (w *writer) Close() error {
	if c, isCloser := w.Writer.(io.Closer); w.performClose && isCloser {
		return c.Close()
	}
	return nil
}

// Transcoding is a pair of functions that can be used to transform to and from
// a transfer encoding.
type Transcoding struct {
	// Encoder returns an io.WriteCloser, which will encode binary data and
	// write the encoded form to the given io.Writer. You must call Close() on
	// the returned io.WriteCloser when you are finished.
	Encoder func(io.Writer) io.WriteCloser

	// Decoder returns an io.Reader, which will read from the given io.Reader
	// when read and decode the encoded data back into binary form the encoded
	// form.
	Decoder func(io.Reader) io.Reader
}

// AsIsTranscoder is just a shortcut to a no-op encoder/decoder.
var AsIsTranscoder = Transcoding{NewAsIsEncoder, NewAsIsDecoder}

// Transcodings defines the supported Content-transfer-encodings and how to
// handle them. It can be modified to change the global handling of transfer
// encodings.
var Transcodings = map[string]Transcoding{
	None:            AsIsTranscoder,
	Bit7:            AsIsTranscoder,
	Bit8:            AsIsTranscoder,
	Binary:          AsIsTranscoder,
	QuotedPrintable: {NewQuotedPrintableEncoder, NewQuotedPrintableDecoder},
	Base64:          {NewBase64Encoder, NewBase64Decoder},
}

// ApplyTransferEncoding is a helper that will check the given header to see if
// transfer encoding ought to be performed. It will return an io.WritCloser that
// will write the encoding (or just pass data through if no encoding is
// necessary).
//
// You must call Close() on the returned io.WriteCloser when you are finished
// writing.
func ApplyTransferEncoding(h *header.Header, w io.Writer) io.WriteCloser {
	cte, err := h.GetTransferEncoding()
	if err != nil {
		return &writer{w, false}
	}

	tc, hasCode := Transcodings[cte]
	if hasCode {
		return tc.Encoder(w)
	}

	return &writer{w, false}
}

// ApplyTransferDecoding returns an io.Reader that will modify incoming bytes
// according to the transfer encoding detected from the given header. (Or the
// io.Reader will leave the bytes as is if there's no transfer encoding or the
// transfer encoding is one that is interpreted as-is).
func ApplyTransferDecoding(h *header.Header, r io.Reader) io.Reader {
	// check to see if the content-type is permitted to have
	// content-transfer-encoding, it's allowed if:
	// |-> Content-type is missing or unreadable
	// |-> Content-type is not a "multipart/*" type
	ct, err := h.GetContentType()
	if err == nil && ct != nil && ct.Type() == "multipart" {
		return r
	}

	// check to see if content-transfer-encoding is set and readable, if not
	// don't continue
	cte, err := h.GetTransferEncoding()
	if err != nil {
		return r
	}

	// check to see if we have a decoder for it and build and return it if we do
	tc, hasCode := Transcodings[cte]
	if hasCode {
		return tc.Decoder(r)
	}

	return r
}
