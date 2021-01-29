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
	body []byte
}

// NewMessage builds a new basic email message from the given header and body.
func NewMessage(h *Header, body []byte) *Message {
	return &Message{*h, body}
}

// BodyString returns the message body as a string.
func (m *Message) BodyString() string { return string(m.body) }

// Body returns the message body.
func (m *Message) Body() []byte { return []byte(m.BodyString()) }

// SetBody sets the message body.
func (m *Message) SetBody(b []byte) { m.body = b }

// SetBodyString sets the message body from a string.
func (m *Message) SetBodyString(s string) { m.body = []byte(s) }

// String returns the email message as a string.
func (m *Message) String() string {
	var out strings.Builder
	out.WriteString(m.Header.String())
	out.Write(m.Header.lb)
	out.Write(m.body)
	return out.String()
}
