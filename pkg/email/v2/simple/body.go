package simple

import "github.com/zostay/go-email/pkg/email/v2"

// Body is a simple email message body. A simple message body is opaque and has
// few additional semantics other than being a slice of bytes.
type Body struct {
	content []byte
}

// NewBody will construct a message body and return it.
func NewBody(c []byte) *Body {
	return &Body{c}
}

// Content will return the content as a slice of bytes.
func (b *Body) Content() []byte {
	return b.content
}

// SetContent will change the content to the given slice of bytes.
func (b *Body) SetContent(c []byte) {
	b.content = c
}

// ContentString will return the contet as a string.
func (b *Body) ContentString() string {
	return string(b.content)
}

// SetContentString will set the content from a string.
func (b *Body) SetContentString(c string) {
	b.content = []byte(c)
}

// String returns the content as a string.
func (b *Body) String() string {
	return b.ContentString()
}

// Bytes returns the content as a slice of bytes.
func (b *Body) Bytes() []byte {
	return b.Content()
}

var _ email.Body = &Body{}
var _ email.Outputter = &Body{}
