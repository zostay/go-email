package email

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/zostay/go-addr/pkg/addr"
)

var (
	// ErrContinuationStart error is returned by ParseHeader when the first line
	// in the message is prefixed with space. The email header will still be
	// parsed, but the contents of the parse may be suspicious.
	ErrContinuationStart = errors.New("header starts with a continuation")
)

// HeaderParseError is returned when an error occurs during parse. This may
// include multiple errors. These errors will be accumulated in this error.
type HeaderParseError struct {
	Errs []error // The accumulated parse errors
}

// Error returns all the errors that occurred during header parsing.
func (err *HeaderParseError) Error() string {
	errs := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errs[i] = e.Error()
	}
	return "error parsing email header: " + strings.Join(errs, ", ")
}

// Header represents an email message header. The header object stores enough
// detail that the original header can be recreated byte for byte for
// roundtripping.
type Header struct {
	fields []*HeaderField // The list of fields
	lb     []byte         // The line break string to use
}

// HeaderField represents an individual field in the message header. When taken
// from a parsed header, it will preserve the original field, byte-for-byte.
type HeaderField struct {
	match    string
	name     string
	body     string
	original []byte
	cache    map[string]interface{}
}

// ParseHeader will parse the given string into an email header. It assumes that
// the entire string given represents the header. It will assume "\n" as the
// line break character to use during parsing.
func ParseHeader(m []byte) (*Header, error) {
	return ParseHeaderLB(m, []byte("\x0d"))
}

// ParseHeaderLB will parse the given string into an email header using the
// given line break string. It will assume the entire string given represents
// the header to be parsed.
func ParseHeaderLB(m, lb []byte) (*Header, error) {
	lines, err := ParseHeaderLines(m, lb)
	errs := make([]error, 0, 1)
	if err != nil {
		errs = append(errs, err)
	}

	h := Header{
		fields: make([]*HeaderField, len(lines)),
		lb:     lb,
	}

	for i, line := range lines {
		h.fields[i], err = ParseHeaderField(line, lb)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return &h, &HeaderParseError{errs}
	}
	return &h, nil
}

// ParseHeaderLines is used by ParseHeader and ParseHeaderLB to create a list of
// lines where each line represents a single field, including and folded
// continuation lines.
//
// Often, mail tools (like some versions of Microsoft Outlook or Exchange
// *eyeroll*) will incorrectly fold a line. As such, a field line is consider
// the start of a field when it does not start with a space AND it contains a
// colon. Otherwise, we will treat it as a fold even though this is not correct
// according to RFC 5322.
func ParseHeaderLines(m, lb []byte) ([][]byte, error) {
	h := make([][]byte, 0, len(m)/80)
	var err error
	for _, line := range bytes.SplitAfter(m, lb) {
		if line[0] == '\t' || line[0] == ' ' || !bytes.Contains(line, []byte(":")) {
			// Start with a continuation? Weird, uh...
			if len(h) == 0 {
				err = ErrContinuationStart
				h = append(h, line)
				continue
			}

			h[len(h)-1] = append(h[len(h)-1], line...)
		} else {
			h = append(h, line)
		}
	}

	return h, err
}

// ParseHeaderField will take a single header field, including any folded
// continuation lines. This will then construct a header field object.
func ParseHeaderField(f, lb []byte) (*HeaderField, error) {
	parts := bytes.SplitN(f, []byte(":"), 2)
	if len(parts) < 2 {
		name := UnfoldValue(f, lb)
		return &HeaderField{"", string(name), "", f, nil}, fmt.Errorf("header field %q missing body", name)
	}

	name := strings.TrimSpace(string(UnfoldValue(parts[0], lb)))
	body := strings.TrimSpace(string(UnfoldValue(parts[1], lb)))
	return &HeaderField{"", name, body, f, nil}, nil
}

// Break returns the line break string associated with this header.
func (h *Header) Break() []byte {
	if h.lb == nil {
		return []byte{'\n', '\r'}
	} else {
		return h.lb
	}
}

// String will return the string representation of the header. If the header was
// parsed from an email header and not modified, this will output the original
// header, preserved byte-for-byte.
func (h *Header) String() string {
	var out strings.Builder
	for _, f := range h.fields {
		out.WriteString(f.String())
	}
	return out.String()
}

// HeaderNames will return the unique header names found in the mail header.
func (h *Header) HeaderNames() []string {
	seen := map[string]struct{}{}
	names := make([]string, 0, len(h.fields))
	for _, f := range h.fields {
		if _, ok := seen[f.Match()]; ok {
			continue
		}
		names = append(names, f.Name())
	}
	return names
}

// HeaderGet will find the first header field with a matching name and return body
// value. It will return an empty string if no such header is present.
func (h *Header) HeaderGet(n string) string {
	f := h.HeaderGetField(n)
	if f != nil {
		return f.Body()
	}
	return ""
}

// HeaderGetAddressList returns addresses for a header. If the header is not set or
// empty, it will return nil and no error. If the header has a value, but cannot
// be parsed as an address list, it will return nil and an error. If the header
// can be parsed as an email list, the email addresses will be returned.
//
// This only returns the addresses for the first occurence of a header, as the
// email address headers are only permitted a single time in email.
func (h *Header) HeaderGetAddressList(n string) (addr.AddressList, error) {
	b := h.HeaderGet(n)
	if b == "" {
		return nil, nil
	}

	return addr.ParseEmailAddressList(b)
}

// HeaderDate parses and returns the date in the email. This will read the header
// named "Date". As this header is always required, it will return the time.Time
// zero value and an error if this method is called and no value is present. If
// the date header is present, it will returned the parsed value or an error if
// the date cannot be parsed.
func (h *Header) HeaderDate() (time.Time, error) {
	b := h.HeaderGet("Date")
	if b == "" {
		return time.Time{}, nil
	}

	return mail.ParseDate(b)
}

// HeaderGetField will find the first header field and return the header field object
// itself. It will return nil if no such header is present.
func (h *Header) HeaderGetField(n string) *HeaderField {
	m := makeMatch(n)
	for _, f := range h.fields {
		if f.Match() == m {
			return f
		}
	}
	return nil
}

// HeaderGetAll will find all header fields with a matching name and return a list of body
// values. Returns nil if no matching headers are present.
func (h *Header) HeaderGetAll(n string) []string {
	hfs := h.HeaderGetAllFields(n)
	bs := make([]string, len(hfs))
	for i, f := range hfs {
		bs[i] = f.Body()
	}
	return bs
}

// HeaderGetAllFields will find all the header fields with a matching name and return
// the list of field objects. It will return any empty slice if no headers with
// this name are present.
func (h *Header) HeaderGetAllFields(n string) []*HeaderField {
	hfs := make([]*HeaderField, 0)
	m := makeMatch(n)
	for _, f := range h.fields {
		if f.Match() == m {
			hfs = append(hfs, f)
		}
	}
	return hfs
}

// HeaderSet will find the first header with a matching name and replace it with the
// given body. If no header by that name is set, it will add a new header with
// that name and body.
func (h *Header) HeaderSet(n, b string) error {
	m := makeMatch(n)
	for _, f := range h.fields {
		if f.Match() == m {
			return f.SetBody(b, h.Break())
		}
	}

	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	h.fields = append(h.fields, f)

	return nil
}

// HeaderAdd will add a new header with the given name and body value. If an existing
// header with the same value is already present, this will add the new field
// before the first field with the same name.
func (h *Header) HeaderAdd(n, b string) error {
	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	m := makeMatch(n)
	for i, f := range h.fields {
		if f.Match() == m {
			b := h.fields[:i]
			a := h.fields[i:]
			h.fields = append(b, f)
			h.fields = append(h.fields, a...)
			return nil
		}
	}

	h.fields = append(h.fields, f)
	return nil
}

// NewHeaderFields constructs a new header field using the given name, body, and
// line break string.
func NewHeaderField(n, b string, lb []byte) (*HeaderField, error) {
	f := HeaderField{}

	var err error
	err = f.SetName(n)
	if err != nil {
		return nil, err
	}

	err = f.SetBody(b, lb)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// Match returns a string useful for matching this header. It will be the
// name string converted to lowercase.
func (f *HeaderField) Match() string {
	if f.match != "" {
		return f.match
	}

	f.match = makeMatch(f.name)
	return f.match
}

func makeMatch(n string) string {
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

// SetName will rename a field. This first checks to make sure no illegal
// characters are present in the field name. It will return an error if the
// given string contains a colon or any character outside the printable ASCII
// range.
func (f *HeaderField) SetName(n string) error {
	forbiddenNameChars := func(c rune) bool {
		if c == ':' {
			return true
		}
		if c >= 33 && c <= 126 {
			return false
		}
		return true
	}

	if strings.IndexFunc(n, forbiddenNameChars) > -1 {
		return errors.New("header name contains illegal character")
	}

	f.SetNameUnsafe(n)
	return nil
}

// SetNameUnsafe will rename a field without checks.
func (f *HeaderField) SetNameUnsafe(n string) {
	f.cache = nil
	f.original = append([]byte(n), f.original[len(f.name):]...)
	f.name = n
}

// SetBody will update the body of the field. You must supply the line break to
// be used for folding. Before setting the field, the value will be checked to
// make sure it is legal. It is only permitted to contain printable ASCII,
// space, and tab characters.
func (f *HeaderField) SetBody(b string, lb []byte) error {
	forbiddenBodyChars := func(c rune) bool {
		if c == ' ' || c == '\t' {
			return false
		}
		if c >= 33 && c <= 126 {
			return false
		}
		return true
	}

	if strings.IndexFunc(b, forbiddenBodyChars) > -1 {
		return errors.New("body name contains illegal character")
	}

	f.SetBodyUnsafe(b, lb)
	return nil
}

// SetBodyUnsafe will update the body of the field without checking to make sure
// it is valid. This can be used to provide a prefolded body (though, it will
// still be folded further if any long lines are found).
func (f *HeaderField) SetBodyUnsafe(b string, lb []byte) {
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
