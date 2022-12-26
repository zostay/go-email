package email

import (
	"io"

	"github.com/zostay/go-email/pkg/email/v2/header"
)

// InterfaceMessage represents an email message. It is composed of a Header and a Body.
type InterfaceMessage interface {
	Body
	BasicHeader
	Outputter
}

// Message is the base-level email message interface. It is simply a header
// and a message body, very similar to the net/mail implementation.
type Message struct {
	header.Header
	io.Reader
}
