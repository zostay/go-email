package simple

import (
	"fmt"

	"github.com/zostay/go-email/pkg/email/v2"
)

// HeaderFieldBase implements an email.HeaderField with a baseline
// implementation that does not implement folding. The body is only capable of
// holding opaque values stored as strings.
type HeaderFieldBase struct {
	name string
	body string
}

// NewHeaderFieldBase constructs a header field with the given name and string
// values and returns the newly created object.
func NewHeaderFieldBase(
	name,
	body string,
) *HeaderFieldBase {
	return &HeaderFieldBase{name, body}
}

// Name returns the name of the header field.
func (f *HeaderFieldBase) Name() string {
	return f.name
}

// SetName updates the name of the header field.
func (f *HeaderFieldBase) SetName(name string) {
	f.name = name
}

// Body returns teh value of the header field as a string.
func (f *HeaderFieldBase) Body() string {
	return f.body
}

// SetBody updates the body fo the header field.
func (f *HeaderFieldBase) SetBody(body string) {
	f.body = body
}

// String returns the complete header field as a string.
func (f *HeaderFieldBase) String() string {
	return fmt.Sprintf("%s: %s", f.name, f.body)
}

// Bytes returns the complete header field as a slice of bytes.
func (f *HeaderFieldBase) Bytes() []byte {
	return []byte(f.String())
}

var _ email.MutableHeaderField = &HeaderFieldBase{}
var _ email.Outputter = &HeaderFieldBase{}
