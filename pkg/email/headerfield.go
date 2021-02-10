package email

import (
	"strings"
)

// HeaderField represents an individual field in the message header. When taken
// from a parsed header, it will preserve the original field, byte-for-byte.
type HeaderField struct {
	match    string
	name     string
	body     string
	original []byte
	cache    map[string]interface{}
}

// NewHeaderField constructs a new header field using the given name, body, and
// line break string.
func NewHeaderField(n, b string, lb []byte) *HeaderField {
	f := HeaderField{
		original: []byte(": "),
	}

	f.SetName(n, lb)
	f.SetBody(b, lb)

	return &f
}

// NewHeaderFieldParsed constructs a new header field using the given name,
// body, line break, and original. No checks are performed on the name or body.
func NewHeaderFieldParsed(n, b string, original []byte) *HeaderField {
	return &HeaderField{"", n, b, original, nil}
}

// Match returns a string useful for matching this header. It will be the
// name string converted to lowercase.
func (f *HeaderField) Match() string {
	if f.match != "" {
		return f.match
	}

	f.match = MakeHeaderFieldMatch(f.name)
	return f.match
}

func MakeHeaderFieldMatch(n string) string {
	return strings.ToLower(strings.TrimSpace(n))
}

// Name returns the field name.
func (f *HeaderField) Name() string { return f.name }

// Body returns the field body.
func (f *HeaderField) Body() string { return f.body }

// Original returns the original text of the field or the newly set rendered
// text for the field.
func (f *HeaderField) Original() []byte { return f.original }

// CacheGet retrieves a value from a structured data cache associated with the
// header. Fields in the structured data cache are cleared when any setter
// method is called on the header field
func (f *HeaderField) CacheGet(k string) interface{} {
	if f.cache == nil {
		return nil
	}

	return f.cache[k]
}

// CacheSet sets a value in the structured data cache associated with the header
// field. The intention is for this to be set to structured data associated with
// the header value. If the name or body of the header is changed, this cache
// will be cleared.
func (f *HeaderField) CacheSet(k string, v interface{}) {
	if f.cache == nil {
		f.cache = make(map[string]interface{})
	}

	f.cache[k] = v
}

// String is an alias for Original.
func (f *HeaderField) String() string { return string(f.original) }

// Bytes returns the original as bytes.
func (f *HeaderField) Bytes() []byte { return f.original }

// SetName will rename a field. This first checks to make sure no illegal
// characters are present in the field name. The line break parameter must be
// passed so it can refold the line as needed.
func (f *HeaderField) SetName(n string, lb []byte) {
	f.cache = nil
	f.match = ""
	f.original = FoldValue(append([]byte(n), f.original[len(f.name):]...), lb)
	f.name = n
}

// SetNameNoFold will rename a field without checks and without folding. The
// name will be set as is.
func (f *HeaderField) SetNameNoFold(n string) {
	f.cache = nil
	f.match = ""
	f.original = append([]byte(n), f.original[len(f.name):]...)
	f.name = n
}

// SetBody will update the body of the field. You must supply the line break to
// be used for folding.
func (f *HeaderField) SetBody(b string, lb []byte) {
	newOrig := append(f.original[:len(f.name)+1], ' ')
	newOrig = append(newOrig, []byte(b)...)
	newOrig = append(newOrig, lb...)
	f.cache = nil
	f.original = FoldValue(newOrig, lb)
	f.body = b
}

// SetBodyNoFold will update the body of the field without checking to make sure
// it is valid and without performing any folding.
func (f *HeaderField) SetBodyNoFold(b string) {
	newOrig := append(f.original[:len(f.name)+1], ' ')
	newOrig = append(newOrig, []byte(b)...)
	f.cache = nil
	f.original = newOrig
	f.body = b
}
