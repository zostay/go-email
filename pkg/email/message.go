package email

import (
	"strings"
)

// Message represents an email message and body. The message object stores
// enough detail that the original message can be roundtripped and preserved
// byte-for-byte while still providing useful tools for reading the header
// fields and other information.
type Message struct {
	Header
	body string
}

func NewMessage(h *Header, body string) *Message {
	return &Message{*h, body}
}

func (m *Message) Body() string { return m.body }

func (m *Message) String() string {
	var out strings.Builder
	out.WriteString(m.Header.String())
	out.WriteString(m.Header.lb)
	out.WriteString(m.Header.lb)
	out.WriteString(m.body)
	return out.String()
}
