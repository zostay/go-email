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

// WriteTo writes the Message message header and body to the destination io.Writer.
func (b *Message) WriteTo(w io.Writer) (int64, error) {
	hb := b.Header.Bytes()
	hn, err := w.Write(hb)
	if err != nil {
		return int64(hn), err
	}

	bn, err := io.Copy(w, b.Reader)
	return int64(hn) + bn, err
}
