package simple

import (
	"bytes"

	"github.com/zostay/go-email/pkg/email/v2"
)

// Header represents an email message header.
type Header struct {
	lbr    email.Break
	vf *ValueFolder
	fields []*HeaderField
}

// NewHeader constructs a new header from the given break and header fields.
func NewHeader(lbr email.Break, f ...*HeaderField) *Header {
	if f == nil {
		f = make([]*HeaderField, 0, 10)
	}
	vf := NewDefaultValueFolder()
	return &Header{lbr, vf, f}
}

// ValueFolder returns the value folder used by this header during rendering.
func (h *Header) ValueFolder() *ValueFolder {
	return h.vf
}

// SetValueFolder changes the value folder used by this header during rendering.
func (h *Header) SetValueFolder(vf *ValueFolder) {
	h.vf = vf
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

// GetField returns the nth field.
func (h *Header) GetField(n int) email.HeaderField {
	if n >= len(h.fields) {
		return nil
	}
	return h.fields[n]
}

// Size returns the number of header fields in the header.
func (h *Header) Size() int {
	return len(h.fields)
}

// GetFieldNamed returns the nth (0-indexed) with the given name or nil if no such
// header field is set.
func (h *Header) GetFieldNamed(name string, n int) email.HeaderField {
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

// GetAllFieldsNamed returns all the fields with the given name or nil if no fields
// are set with that name.
func (h *Header) GetAllFieldsNamed(name string) []email.HeaderField {
	fs := make([]email.HeaderField, 0, 10)
	for _, f := range h.fields {
		if f.Name() == name {
			fs = append(fs, f)
		}
	}
	return fs
}

// ListFields returns all the fields in the header.
func (h *Header) ListFields() []email.HeaderField {
	fs := make([]email.HeaderField, len(h.fields))
	for i := range h.fields {
		fs[i] = h.fields[i]
	}
	return fs
}

// Bytes returns the header as a slice of bytes.
func (h *Header) Bytes() []byte {
	var buf bytes.Buffer
	for _, f := range h.fields {
		foldedField := h.vf.Fold(f.Bytes(), h.lbr)
		buf.Write(foldedField)
		buf.Write(h.lbr.Bytes())
	}
	buf.Write(h.lbr.Bytes())
	return buf.Bytes()
}

// String returns the header as a string.
func (h *Header) String() string {
	return string(h.Bytes())
}

// InsertBeforeField will insert the given name and body values into the header
// at the given index.
func (h *Header) InsertBeforeField(
	n int,
	name,
	body string,
) {
	// cap the range of n to 0..len(h.fields)
	if n < 0 {
		n = 0
	}
	if n > len(h.fields) {
		n = len(h.fields)
	}

	// create the new field
	f := &HeaderField{
		HeaderFieldBase: HeaderFieldBase{
			name: name,
			body: body,
		},
	}

	// make room for the new field
	h.fields = append(h.fields, nil)

	// move existing fields out of the way
	copy(h.fields[n+1:], h.fields[n:])

	// insert
	h.fields[n] = f
}

// ClearFields removes all fields from the header.
func (h *Header) ClearFields() {
	h.fields = h.fields[:0]
}

// DeleteField removes the nth field from the header.
func (h *Header) DeleteField(n int) error {
	// bounds check
	if n < 0 || n >= len(h.fields) {
		return email.ErrIndexOutOfRange
	}

	// copy over the removed field
	copy(h.fields[n:], h.fields[n+1:])

	// shorten the slice by one
	h.fields = h.fields[:len(h.fields)-1]

	return nil
}

var _ email.WithMutableBreak = &Header{}
var _ email.MutableHeader = &Header{}
