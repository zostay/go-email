package message

import (
	"bytes"

	"github.com/zostay/go-email/pkg/email/v2/header"
)

// Buffer provides an interface for building email messages. It provides the
// message header for manipulation and can be written to set the body of the
// message.
type Buffer struct {
	header.Header
	buf bytes.Buffer
}

// Write implements io.Writer so you can write the message to this buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	return b.buf.Write(p)
}

// Message returns the message that has been created by this buffer.
func (b *Buffer) Message() *Message {
	return &Message{b.Header, &b.buf}
}
