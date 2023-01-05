package header

import (
	"bytes"
	"errors"
	"strings"

	"github.com/zostay/go-email/pkg/v2/header/field"
)

var (
	// ErrIndexOutOfRange when an attempt is made to access a header field index
	// that is too large or to small.
	ErrIndexOutOfRange = errors.New("header field index is out of range")
)

// Base represents a basic email message header. It is a low-level interface
// to headers, but with the ability to apply field folding during output.
type Base struct {
	lbr    Break
	vf     *field.FoldEncoding
	fields []*field.Field
}

// initBase initializes the Break and fields values lazily.
func (h *Base) initBase() {
	if h.lbr == "" {
		h.lbr = LF
	}
	if h.fields == nil {
		h.fields = make([]*field.Field, 0, 10)
	}
}

// FoldEncoding returns the value folder used by this header during rendering.
func (h *Base) FoldEncoding() *field.FoldEncoding {
	if h.vf == nil {
		h.vf = field.DefaultFoldEncoding
	}
	return h.vf
}

// SetFoldEncoding changes the value folder used by this header during rendering.
func (h *Base) SetFoldEncoding(vf *field.FoldEncoding) {
	h.vf = vf
}

// Break returns the line break used to separate header fields and terminate the
// header.
func (h *Base) Break() Break {
	if h.lbr == "" {
		h.lbr = LF
	}
	return h.lbr
}

// SetBreak changes the line break to use with this header.
func (h *Base) SetBreak(lbr Break) {
	h.lbr = lbr
}

// GetField returns the nth field.
func (h *Base) GetField(n int) *field.Field {
	if n >= len(h.fields) {
		return nil
	}
	return h.fields[n]
}

// Size returns the number of header fields in the header.
func (h *Base) Size() int {
	return len(h.fields)
}

// GetFieldNamed returns the nth (0-indexed) with the given name or nil if no such
// header field is set.
func (h *Base) GetFieldNamed(name string, n int) *field.Field {
	for _, f := range h.fields {
		if strings.EqualFold(f.Name(), name) {
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
func (h *Base) GetAllFieldsNamed(name string) []*field.Field {
	fs := make([]*field.Field, 0, 10)
	for _, f := range h.fields {
		if strings.EqualFold(f.Name(), name) {
			fs = append(fs, f)
		}
	}
	return fs
}

// GetIndexesNamed returns the indexes of fields with the given name.
func (h *Base) GetIndexesNamed(name string) []int {
	is := make([]int, 0, 10)
	for i, f := range h.fields {
		if strings.EqualFold(f.Name(), name) {
			is = append(is, i)
		}
	}
	return is
}

// ListFields returns all the fields in the header.
func (h *Base) ListFields() []*field.Field {
	fs := make([]*field.Field, len(h.fields))
	for i := range h.fields {
		fs[i] = h.fields[i]
	}
	return fs
}

// Bytes returns the header as a slice of bytes.
func (h *Base) Bytes() []byte {
	var buf bytes.Buffer
	for _, f := range h.fields {
		foldedField := h.FoldEncoding().Fold(f.Bytes(), h.lbr.Bytes())
		buf.Write(foldedField)
		buf.Write(h.lbr.Bytes())
	}
	buf.Write(h.lbr.Bytes())
	return buf.Bytes()
}

// String returns the header as a string.
func (h *Base) String() string {
	return string(h.Bytes())
}

// InsertBeforeField will insert the given name and body values into the header
// at the given index.
func (h *Base) InsertBeforeField(
	n int,
	name,
	body string,
) {
	h.initBase()

	// cap the range of n to 0..len(h.fields)
	if n < 0 {
		n = 0
	}
	if n > len(h.fields) {
		n = len(h.fields)
	}

	// create the new field
	f := field.New(name, body)

	// make room for the new field
	h.fields = append(h.fields, nil)

	// move existing fields out of the way
	copy(h.fields[n+1:], h.fields[n:])

	// insert
	h.fields[n] = f
}

// ClearFields removes all fields from the header.
func (h *Base) ClearFields() {
	h.initBase()
	h.fields = h.fields[:0]
}

// DeleteField removes the nth field from the header. Fails with an error if the
// given index is out of range.
func (h *Base) DeleteField(n int) error {
	h.initBase()

	// bounds check
	if n < 0 || n >= len(h.fields) {
		return ErrIndexOutOfRange
	}

	// copy over the removed field
	copy(h.fields[n:], h.fields[n+1:])

	// shorten the slice by one
	h.fields = h.fields[:len(h.fields)-1]

	return nil
}
