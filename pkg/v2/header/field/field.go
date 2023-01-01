package field

import (
	"bytes"

	"github.com/zostay/go-email/pkg/email/v2"
)

// Field provides a low-level interface to manage a single email header field.
// Every Field contains name and body values from the embedded Base object. In
// addition to this base object, it may also contain an even lower-level
// representation in Raw. This allows the object to maintain a decoded logical
// string value as well as a fully encoded raw value.
//
// The Name() and Body() methods will always surface the Base field.
//
// The Strings() and Bytes() methods will always work on the Raw field if
// present, but fallback to using the Base field if not present.
//
// The SetName() and SetBody() methods will always update Base and result in Raw
// being cleared, if present. The SetRaw() method is provided to create a new
// Raw value after such an edit.
//
// Both Raw and Base are exposed if something more low-level or nuanced is
// needed.
type Field struct {
	Base
	*Raw
}

// New constructs a new field with no original value.
func New(name, body string) *Field {
	return &Field{Base{name, body}, nil}
}

// String returns the Raw.String() if Raw is not nil. It returns the
// Base.String() otherwise.
func (f *Field) String() string {
	if f.Raw != nil {
		return f.Raw.String()
	}
	return f.Base.String()
}

// Bytes returns the Raw.Bytes() if Raw is not nil. It returns the Base.Bytes()
// otherwise.
func (f *Field) Bytes() []byte {
	if f.Raw != nil {
		return f.Raw.Bytes()
	}
	return f.Base.Bytes()
}

// Name returns the Base.Name().
func (f *Field) Name() string {
	return f.Base.Name()
}

// Body returns the Base.Body().
func (f *Field) Body() string {
	return f.Base.Body()
}

// SetName sets the name of the field by calling Base.SetName(). Calling this
// will also result in Raw being set to nil.
func (f *Field) SetName(n string) {
	f.Raw = nil
	f.Base.SetName(n)
}

// SetBody sets the body of the field by calling Base.SetBody(). Calling this
// will also result in Raw being set to nil.
func (f *Field) SetBody(b string) {
	f.Raw = nil
	f.Base.SetBody(b)
}

// SetRaw replaces Raw with a new value. The value will be run through
// field.Parse(). If there's a problem parsing the field, this method will fail
// with an error.
func (f *Field) SetRaw(o []byte) {
	ix := bytes.IndexRune(o, ':')
	if ix < 0 {
		ix = len(o)
	}
	f.Raw = &Raw{o, ix}
}

var _ email.MutableHeaderField = &Field{}
var _ email.Outputter = &Field{}
