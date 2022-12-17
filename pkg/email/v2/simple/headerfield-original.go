package simple

import (
	"bytes"
	"errors"

	"github.com/zostay/go-email/pkg/email/v2"
)

var (
	// ErrOriginalTooShort is the error returned by the header field parser to
	// indicate that the given header field is either empty or two short to
	// be a parseable header. The shortest possible header would be three bytes
	// long, e.g., []byte("a:\n").
	ErrOriginalTooShort = errors.New("header field is empty or too short")

	// ErrOriginalMissingColon is the error that indicates that the header field
	// parser is unable to find a colon in the input.
	ErrOriginalMissingColon = errors.New("header field is missing colon separating name from body in original")

	// ErrOriginalMissingBreak is the error that indicates that the header field
	// parser is unable to find a line break in the input.
	ErrOriginalMissingBreak = errors.New("header field is missing line break at the end")
)

// HeaderFieldOriginal is a email.HeaderField implementation that presents the
// parsed original value. Objects of this type are immutable.
type HeaderFieldOriginal struct {
	field []byte // complete original field
	colon int    // the index of the colon
}

// ParseOriginal will parse the given field and return a HeaderFieldOriginal. If
// something goes wrong in parsing the field, it will return an error instead.
func ParseOriginal(field []byte) (*HeaderFieldOriginal, error) {
	if len(field) < 3 {
		return nil, ErrOriginalTooShort
	}

	// locate the colon separator
	colon := bytes.Index(field, []byte{':'})
	if colon < 0 {
		return nil, ErrOriginalMissingColon
	}

	return &HeaderFieldOriginal{
		field: field,
		colon: colon,
	}, nil
}

// String returns the original as a string.
func (f *HeaderFieldOriginal) String() string {
	return string(f.field)
}

// Bytes returns the original.
func (f *HeaderFieldOriginal) Bytes() []byte {
	return f.field
}

// Name returns the name part of the original.
func (f *HeaderFieldOriginal) Name() string {
	return string(f.field[:f.colon])
}

// Body returns the body part of the original as bytes.
func (f *HeaderFieldOriginal) Body() string {
	return string(f.field[f.colon+1:])
}

var _ email.HeaderField = &HeaderFieldOriginal{}
var _ email.Outputter = &HeaderFieldOriginal{}
