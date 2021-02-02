// Package email is provides generic tools for representing an email and for
// parsing email headers. This is used in conjuncion with the simple and mime
// packages to provide simple email parsing and MIME message handling.
//
// None of the operations of this library are intended to be thread safe. No
// guarantees are made that the object will be kept in a consistent state while
// within a method, so you shouldn't try to manipulate or access the data of
// these objects concurrently.
package email

import (
	"bytes"
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

// Bytes returns the email message as a slice of bytes.
func (m *Message) Bytes() []byte {
	var out bytes.Buffer
	out.WriteString(m.Header.String())
	out.Write(m.Header.lb)
	out.Write(m.body)
	return out.Bytes()
}
