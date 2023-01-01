package simple

import "github.com/zostay/go-email/pkg/email/v2"

// Header represents an email message header.
type Header struct {
	lbr    email.Break
	fields []*HeaderField
}

// NewHeader constructs a new header from the given break and header fields.
func NewHeader(lbr email.Break, f ...*HeaderField) *Header {
	return &Header{lbr, f}
}

// Break returns the line break used with these header fields.
