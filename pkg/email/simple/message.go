package simple

import (
	"bytes"
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

type Message struct {
	Header
	email.Body
}

// NewMessage builds a new simple email message from the given header and body.
func NewMessage(h *Header, body []byte) *Message {
	return &Message{*h, *email.NewBody(body)}
}

// String returns the email message as a string.
func (m *Message) String() string {
	var out strings.Builder
	out.WriteString(m.Header.String())
	out.Write(m.Header.Break())
	out.WriteString(m.Body.String())
	return out.String()
}

// Bytes returns the email message as a slice of bytes.
func (m *Message) Bytes() []byte {
	var out bytes.Buffer
	out.WriteString(m.Header.String())
	out.Write(m.Header.Break())
	out.Write(m.Body.Bytes())
	return out.Bytes()
}
