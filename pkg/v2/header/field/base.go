package field

import (
	"fmt"
)

// Base implements an email.Field with a baseline
// implementation that does not implement folding. The body is only capable of
// holding opaque values stored as strings.
type Base struct {
	name string
	body string
}

// Name returns the name of the header field.
func (f *Base) Name() string {
	return f.name
}

// SetName updates the name of the header field.
func (f *Base) SetName(name string) {
	f.name = name
}

// Body returns teh value of the header field as a string.
func (f *Base) Body() string {
	return f.body
}

// SetBody updates the body fo the header field.
func (f *Base) SetBody(body string) {
	f.body = body
}

// String returns the complete header field as a string.
func (f *Base) String() string {
	return fmt.Sprintf("%s: %s", f.name, Encode(f.body))
}

// Bytes returns the complete header field as a slice of bytes.
func (f *Base) Bytes() []byte {
	return []byte(f.String())
}
