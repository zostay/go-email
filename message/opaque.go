package message

import (
	"io"

	"github.com/zostay/go-email/v2/header"
	"github.com/zostay/go-email/v2/transfer"
)

// Opaque is the base-level email message interface. It is simply a header
// and a message body, very similar to the net/mail message implementation.
type Opaque struct {
	header.Header
	io.Reader

	// encoded tracks whether the body has had the content-transfer-encoding
	// still encoded or not...
	//
	// - parsing leaves encoding in place by default (unless
	// DecodeTransferEncoding() option is specified)
	//
	// - creating an opaque with a buffer will leave this false unless the
	// object is constructed using OpaqueAlreadyEncoded
	encoded bool
}

// WriteTo writes the Opaque header and body to the destination
// io.Writer.
//
// If the bytes head in io.Reader have had the Content-transfer-encoding decoded
// (e.g., the message was parsed with the DecodeTransferEncoding() option or was
// created via a Buffer), then this will encode the data as it is being written.
//
// This can only be safely called once as it will consume the io.Reader.
func (m *Opaque) WriteTo(w io.Writer) (int64, error) {
	var tw io.WriteCloser
	if !m.encoded {
		tw = transfer.ApplyTransferEncoding(&m.Header, w)
		defer func() { _ = tw.Close() }()
	}

	hn, err := m.Header.WriteTo(w)
	if err != nil {
		return hn, err
	}

	if tw != nil {
		w = tw
	}

	bn, err := io.Copy(w, m.Reader)
	return hn + bn, err
}

// IsMultipart always returns false.
func (m *Opaque) IsMultipart() bool {
	return false
}

// IsEncoded returns true if the Content-transfer-encoding has not been decoded
// for the bytes returned by the associated io.Reader. It will return false if
// that decoding has been performed.
//
// Be aware that a false value here does not mean any actually changes to the
// bytes have been made. If the Content-type of this message is a "multipart/*"
// type, then any Content-transfer-encoding is ignored. If the
// Content-transfer-encoding is set to something like "8bit", the transfer
// encoding returns the bytes as-is and no transformation of the data is
// performed anyway.
//
// However, if this returns true, then reading the data from io.Reader will
// return exactly the same bytes as would be written via WriteTo().
func (m *Opaque) IsEncoded() bool {
	return m.encoded
}

// GetHeader returns the header for the message.
func (m *Opaque) GetHeader() *header.Header {
	return &m.Header
}

// GetReader returns the reader containing the body of the message.
//
// If IsEncoded() returns false, the data returned by reading this io.Reader
// may differ from the data that would be written via WriteTo(). This is
// because the data here will have been decoded, but WriteTo() will encode the
// data anew as it writes.
func (m *Opaque) GetReader() io.Reader {
	return m.Reader
}

// GetParts always returns nil and ErrNotMultipart.
func (m *Opaque) GetParts() []Part {
	return nil
}
