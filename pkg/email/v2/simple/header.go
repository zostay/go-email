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

// Break returns the line break used to separate header fields and terminate the
// header.
func (h *Header) Break() email.Break {
	return h.lbr
}

// SetBreak changes the line break to use with this header.
func (h *Header) SetBreak(lbr email.Break) {
	h.lbr = lbr
}

// GetField returns the first header field for the given name or nil if no such
// field is set.
func (h *Header) GetField(name string) email.HeaderField {
	for _, f := range h.fields {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// GetFieldN returns the nth (0-indexed) with the given name or nil if no such
// header field is set.
func (h *Header) GetFieldN(name string, n int) email.HeaderField {
	for _, f := range h.fields {
		if f.Name() == name {
			if n == 0 {
				return f
			}
			n--
		}
	}
	return nil
}

// GetAllFields returns all the fields with the given name or nil if no fields
// are set with that name.
func (h *Header) GetAllFields(name string) []email.HeaderField {
	fs := make([]email.HeaderField, 0, 10)
	for _, f := range h.fields {
		if f.Name() == name {
			fs = append(fs, f)
		}
	}
	return fs
}
