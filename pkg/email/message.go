package email

import (
	"strings"
)

// Message represents an email message and body. The message object stores
// enough detail that the original message can be roundtripped and preserved
// byte-for-byte while still providing useful tools for reading the header
// fields and other information.
type Message struct {
	Header *Header
	Body   string
}

func (m *Message) String() string {
	var out strings.Builder
	out.WriteString(m.Header.String())
	out.WriteString(m.Header.Break)
	out.WriteString(m.Header.Break)
	out.WriteString(m.Body)
	return out.String()
}
