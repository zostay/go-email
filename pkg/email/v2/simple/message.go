package simple

import "github.com/zostay/go-email/pkg/email/v2"

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

var _ email.Message = &Message{}
