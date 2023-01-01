package param

import (
	"fmt"
	"mime"
	"sort"
	"strings"
)

const (
	// Charset is the name of the charset parameter that may be present in the
	// Content-type header.
	Charset = "charset"

	// Boundary is the name of the boundary paramter that may be present in the
	// Content-type header.
	Boundary = "boundary"

	// Filename is the name of the filename parameter that may be present in the
	// Content-disposition header.
	Filename = "filename"
)

// Value represents a parsed parameterized header field, such as is used in the
// Content-type and Content-disposition headers. A Value object is immutable:
// You cannot change it in place. However, a Modify() function is provided to
// perform transformation of a Value into a new Value.
type Value struct {
	v  string
	ps map[string]string
}

// Parse takes a header field body, parses it as a Value and returns it. If an
// error occurs in the process, it returns an error.
func Parse(v string) (*Value, error) {
	mt, ps, err := mime.ParseMediaType(v)
	if err != nil {
		return nil, err
	}

	return &Value{mt, ps}, nil
}

// New creates a new parameterized header field with no parameters.
func New(v string) *Value {
	return &Value{v, map[string]string{}}
}

// NewWithParams creates a new parametersized header field with the given
// parameters.
func NewWithParams(v string, ps map[string]string) *Value {
	return &Value{v, ps}
}

// Modifier is a modification to apply to a Value when calling the Modify()
// function.
type Modifier func(*Value)

// Change is a Modifier that replaces the primary value of the Value.
func Change(value string) Modifier {
	return func(pv *Value) {
		pv.v = value
	}
}

// Set is a Modifier that sets a parameter with the given name on the Value.
func Set(name, value string) Modifier {
	return func(pv *Value) {
		pv.ps[name] = value
	}
}

// Delete is a Modifier that removes the parameter with the given name from the
// Value.
func Delete(name string) Modifier {
	return func(pv *Value) {
		delete(pv.ps, name)
	}
}

// Modify clones a Value, applies the given modifications (if any) and returns
// the new Value. You can pass multiple changes to this function:
//
//	v, _ := value.Parse("multipart/mixed; boundary=abc123; charset=latin1")
//	nv := value.Modify(v, Change("multipart/alternate"), Set("charset", "utf-8"))
func Modify(pv *Value, changes ...Modifier) *Value {
	copy := pv.Clone()
	for _, change := range changes {
		change(copy)
	}
	return copy
}

// Value returns the primary value of the Value. This is the value before the
// first semi-colon.
func (pv *Value) Value() string {
	return pv.v
}

// Disposition is a synonym for Value() and returns the Content-disposition,
// either "inline" or "attachment".
func (pv *Value) Disposition() string {
	return pv.v
}

// MediaType is a synonym for Value() and returns the Content-type value, e.g.,
// "text/html", "image/jpeg", "multipart/mixed", etc.
func (pv *Value) MediaType() string {
	return pv.v
}

// Type is only intended for use with the Content-type header. It searches the
// MediaType() for a slash. If found, it will return the string before that
// slash. If no slash is found, it returns an empty string.
//
// For example, if MediaType() returns "image/jpeg", this method will return
// "image".
func (pv *Value) Type() string {
	if ix := strings.IndexRune(pv.v, '/'); ix >= 0 {
		return pv.v[:ix]
	}
	return ""
}

// Subtype is only intended for use with teh Content-type header. It searches
// the MediaType() for a slash. If found, it will return the string after that
// slash. If no slash is found, it returns an empty string.
//
// For example, if MediaType() returns "text/html", this method will return
// "html".
func (pv *Value) Subtype() string {
	if ix := strings.IndexRune(pv.v, '/'); ix >= 0 {
		return pv.v[ix+1:]
	}
	return ""
}

// Parameters returns the parameters encoded on this Value as a map. Do not
// modify this map. The behavior if you do is not defined and may change in the
// future. If you need to modify it, make a copy first.
func (pv *Value) Parameters() map[string]string {
	return pv.ps
}

// Parameter returns the value of the parameter with the given name.
func (pv *Value) Parameter(k string) string {
	return pv.ps[k]
}

// Filename returns the value of the "filename" parameter. It is intended for
// use with the Content-disposition header.
func (pv *Value) Filename() string {
	return pv.ps[Filename]
}

// Charset returns the value of the "charset" parameter. It is intended for use
// with the Content-type header.
func (pv *Value) Charset() string {
	return pv.ps[Charset]
}

// Boundary returns the value of the "boundary" parameter. It is intended for
// use with the Content-type header.
func (pv *Value) Boundary() string {
	return pv.ps[Boundary]
}

// String returns the serialized value of the Value including the primary value
// and all parameters.
func (pv *Value) String() string {
	pks := make([]string, 0, len(pv.ps))
	for k := range pv.ps {
		pks = append(pks, k)
	}
	sort.Strings(pks)

	parts := make([]string, len(pv.ps)+1)
	parts[0] = pv.v

	for n, k := range pks {
		parts[n+1] = fmt.Sprintf("%s=%s", k, pv.ps[k])
	}

	return strings.Join(parts, "; ")
}

// Bytes returns the serialized value of the Value including the primary value
// and all parameters.
func (pv *Value) Bytes() []byte {
	return []byte(pv.String())
}

// Clone returns a deep copy of the Value.
func (pv *Value) Clone() *Value {
	var copy Value
	copy.v = pv.v
	copy.ps = make(map[string]string, len(pv.ps))
	for k, v := range pv.ps {
		copy.ps[k] = v
	}
	return &copy
}
