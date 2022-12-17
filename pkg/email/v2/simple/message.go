package simple

import (
	"bytes"

	"github.com/zostay/go-email/pkg/email/v2"
)

// Message implements email.Message. The message itself is an opaque set of
// bytes and header fields are also treated as strings.
type Message struct {
	*Body
	*Header
}

// NewMessage constructs and returns an empty message.
func NewMessage() *Message {
	return &Message{
		Body: NewBody([]byte{}),
		Header: NewHeader(email.LF),
	}
}

// Bytes renders the complete message.
func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(m.Header.Bytes())
	buf.Write(m.Body.Bytes())
	return buf.Bytes()
}

// String renders the complete message.
func (m *Message) String() string {
	return string(m.Bytes())
}

var _ email.Message = &Message{}
