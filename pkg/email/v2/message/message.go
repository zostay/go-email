package message

import (
	"io"

	"github.com/zostay/go-email/pkg/email/v2/header"
)

// Message is the base-level email message interface. It is simply a header
// and a message body, very similar to the net/mail message implementation.
type Message struct {
	header.Header
	io.Reader
}

// WriteTo writes the Message header and body to the destination
// io.Writer.
func (b *Message) WriteTo(w io.Writer) (int64, error) {
	hb := b.Header.Bytes()
	hn, err := w.Write(hb)
	if err != nil {
		return int64(hn), err
	}

	bn, err := io.Copy(w, b.Reader)
	return int64(hn) + bn, err
}

// IsMultipart always returns false.
func (m *Message) IsMultipart() bool {
	return false
}

// GetHeader returns the header for the message.
func (m *Message) GetHeader() *header.Header {
	return &m.Header
}

// GetReader returns the reader containing the body of the message.
func (m *Message) GetReader() (io.Reader, error) {
	return m.Reader, nil
}

// GetParts always returns nil and ErrNotMultipart.
func (m *Message) GetParts() ([]MimePart, error) {
	return nil, ErrNotMultipart
}
