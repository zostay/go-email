package field

import (
	"bytes"
	"errors"

	"github.com/zostay/go-email/pkg/email/v2"
)

var (
	// ErrRawFieldTooShort is the error returned by the header field parser to
	// indicate that the given header field is either empty or two short to
	// be a parseable header. The shortest possible header would be three bytes
	// long, e.g., []byte("a:\n").
	ErrRawFieldTooShort = errors.New("header field is empty or too short")

	// ErrRawFieldMissingColon is the error that indicates that the header field
	// parser is unable to find a colon in the input.
	ErrRawFieldMissingColon = errors.New("header field is missing colon separating name from body in Raw")

	// ErrRawFieldMissingBreak is the error that indicates that the header field
	// parser is unable to find a line break in the input.
	ErrRawFieldMissingBreak = errors.New("header field is missing line break at the end")
)

// Raw is a email.Field implementation that presents the
// parsed Raw value. Objects of this type are immutable.
type Raw struct {
	field []byte // complete Raw field
	colon int    // the index of the colon
}

// Parse will parse the given field and return a Raw. If
// something goes wrong in parsing the field, it will return an error instead.
func Parse(field []byte) (*Raw, error) {
	if len(field) < 3 {
		return nil, ErrRawFieldTooShort
	}

	// locate the colon separator
	colon := bytes.Index(field, []byte{':'})
	if colon < 0 {
		return nil, ErrRawFieldMissingColon
	}

	return &Raw{
		field: field,
		colon: colon,
	}, nil
}

// String returns the Raw as a string.
func (f *Raw) String() string {
	return string(f.field)
}

// Bytes returns the Raw.
func (f *Raw) Bytes() []byte {
	return f.field
}

// Name returns the name part of the Raw.
func (f *Raw) Name() string {
	return string(f.field[:f.colon])
}

// Body returns the body part of the Raw as bytes.
func (f *Raw) Body() string {
	return string(f.field[f.colon+1:])
}

var _ email.HeaderField = &Raw{}
var _ email.Outputter = &Raw{}
