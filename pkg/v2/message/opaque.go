package message

import (
	"io"

	"github.com/zostay/go-email/pkg/v2/header"
)

// Opaque is the base-level email message interface. It is simply a header
// and a message body, very similar to the net/mail message implementation.
type Opaque struct {
	header.Header
	io.Reader
}

// WriteTo writes the Opaque header and body to the destination
// io.Writer.
func (b *Opaque) WriteTo(w io.Writer) (int64, error) {
	hb := b.Header.Bytes()
	hn, err := w.Write(hb)
	if err != nil {
		return int64(hn), err
	}

	bn, err := io.Copy(w, b.Reader)
	return int64(hn) + bn, err
}

// IsMultipart always returns false.
func (m *Opaque) IsMultipart() bool {
	return false
}

// GetHeader returns the header for the message.
func (m *Opaque) GetHeader() *header.Header {
	return &m.Header
}

// GetReader returns the reader containing the body of the message.
func (m *Opaque) GetReader() (io.Reader, error) {
	return m.Reader, nil
}

// GetParts always returns nil and ErrNotMultipart.
func (m *Opaque) GetParts() ([]Part, error) {
	return nil, ErrNotMultipart
}
