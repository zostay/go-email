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

func NewMessage(h *Header, body []byte) *Message {
	return &Message{*h, body}
}

func (m *Message) BodyString() string { return string(m.body) }
func (m *Message) Body() []byte       { return []byte(m.BodyString()) }

func (m *Message) SetBody(b []byte)       { m.body = b }
func (m *Message) SetBodyString(s string) { m.body = []byte(s) }

func (m *Message) String() string {
	var out strings.Builder
	out.WriteString(m.Header.String())
	out.Write(m.Header.lb)
	out.Write(m.Header.lb)
	out.Write(m.body)
	return out.String()
}
