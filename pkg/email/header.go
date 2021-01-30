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

// Constants for use when selecting a line break to use with a new header. If
// you don't know what to pick, choose the StandardLineBreak.
const (
	StandardLineBreak   = "\x0d\x0a"
	LinuxLineBreak      = "\x0a"
	ClassicMacLineBreak = "\x0d"
	WeirdoLineBreak     = "\x0a\x0d"
)

// BadStartError is returned when the header begins with junk text that does not
// appear to be a header. This text is preserved in the error object.
type BadStartError struct {
	BadStart []byte // the text skipped at the start of header
}

// Error returns the error message.
func (err *BadStartError) Error() string {
	return "header starts with text that does not appear to be a header"
}

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
//
// If the header starts with text that does not appear to be a header. The start
// text will be skipped for header parsing. However, it will be accumulated into
// a BadStartError and returned with the parsed header.
func ParseHeaderLines(m, lb []byte) ([][]byte, error) {
	h := make([][]byte, 0, len(m)/80)
	var err *BadStartError
	for _, line := range bytes.SplitAfter(m, lb) {
		if len(line) == 0 {
			break
		}
		if line[0] == '\t' || line[0] == ' ' || !bytes.Contains(line, []byte(":")) {
			// Start with a continuation? Weird, uh...
			if len(h) == 0 {
				if err != nil {
					err.BadStart = append(err.BadStart, line...)
				} else {
					err = &BadStartError{line}
				}
				continue
			}

			h[len(h)-1] = append(h[len(h)-1], line...)
		} else {
			h = append(h, line)
		}
	}

	if err != nil {
		return h, err
	} else {
		return h, nil
	}
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

// NewHeader creates a new header.
func NewHeader(lb string) *Header {
	return &Header{lb: []byte(lb)}
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
		m := f.Match()
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		names = append(names, f.Name())
	}
	return names
}

// HeaderGet will find the first header field with a matching name and return
// the body value. It will return an empty string if no such header is present.
func (h *Header) HeaderGet(n string) string {
	hf, _ := h.HeaderGetFieldN(n, 0)
	if hf != nil {
		return hf.Body()
	}
	return ""
}

// HeaderGetN will find the (ix+1)th header field with a matching name and
// return the body value. It will return an empty string with an error if no
// such header is present.
//
// If ix is negative, it will return the (-ix)th body value from the end.
func (h *Header) HeaderGetN(n string, ix int) (string, error) {
	hf, err := h.HeaderGetFieldN(n, ix)
	if hf != nil {
		return hf.Body(), err
	}

	return "", err
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
	if i := h.headerFieldIndex(n, 0, false); i > -1 {
		return h.fields[i]
	}
	return nil
}

func (h *Header) headerFieldIndex(n string, ix int, fb bool) int {
	m := makeMatch(n)
	lasti := -1
	if ix < 0 {
		count := -1
		for i := len(h.fields) - 1; i >= 0; i-- {
			f := h.fields[i]
			if f.Match() == m {
				lasti = i
				if count == ix {
					return i
				}
				count--
			}
		}
	} else {
		count := 0
		for i, f := range h.fields {
			if f.Match() == m {
				lasti = i
				if count == ix {
					return i
				}
				count++
			}
		}
	}

	if fb {
		return lasti
	} else {
		return -1
	}
}

// HeaderGetFieldN locates the (ix+1)th named header and returns the header
// field object. If no such header exists, the field is returned as nil and an
// error is returned.
//
// If ix is negative, it will return the (-ix)th header from the end.
func (h *Header) HeaderGetFieldN(n string, ix int) (*HeaderField, error) {
	if i := h.headerFieldIndex(n, ix, false); i > -1 {
		return h.fields[i], nil
	}
	return nil, fmt.Errorf("unable to find index %d of header named %q", ix, n)
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

// HeaderSet will find the first header with a matching name and replace it
// with the given body. If no header by that name is set, it will add a new
// header with that name and body.
func (h *Header) HeaderSet(n, b string) error {
	if i := h.headerFieldIndex(n, 0, false); i > -1 {
		f := h.fields[i]
		return f.SetBody(b, h.Break())
	}

	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	h.fields = append(h.fields, f)

	return nil
}

// HeaderSetN will find the (ix+1)th header matching the given name and replace
// the body value with the new value given.
//
// If ix is negative, it will replace the body value of the (-ix)th value from
// the end.
//
// If no header with that number is available, no change will be made and an
// error will be returned.
func (h *Header) HeaderSetN(n, b string, ix int) error {
	hf, err := h.HeaderGetFieldN(n, ix)
	if err != nil {
		return err
	}

	return hf.SetBody(b, h.lb)
}

// HeaderSetAll does a full header replacement. This performs a number of
// combined operations.
//
// 1. If no body values are given, this is equivalent to HeaderDeleteAll.
//
// 2. If some body values are given and there are some headers already present.
//    This is equivalent to calling HeaderSet for each of those values.
//
// 3. If there are fewer body values than headers already present, the remaining
//    headers will be deleted.
//
// 4. If there are more body values than headers already present, this is
//    equivalent to calling HeaderAdd for each of those headers.
//
// Basically, it's going to make sure all the given headers are set and will
// start by changing the ones already in place, removing any additional ones
// that aren't updated, and adding new ones if necessary.
//
// If the operation is successful, it returns nil. If there is an error, then
// the object will be unchanged and an error returned.
func (h *Header) HeaderSetAll(n string, bs ...string) error {
	// no values, so delete all
	if len(bs) == 0 {
		h.HeaderDeleteAll(n)
		return nil
	}

	hfs := make([]*HeaderField, 0, len(h.fields))
	m := makeMatch(n)
	bi := 0
	for _, hf := range h.fields {
		if hf.Match() == m {
			// Set existing field
			if bi < len(bs) {
				err := hf.SetBody(bs[bi], h.lb)
				if err != nil {
					return err
				}

				hfs = append(hfs, hf)
				bi++
			}
			// else, skip and delete the field
		} else {
			// unrelated field, copy it through
			hfs = append(hfs, hf)
		}
	}

	orig := h.fields
	h.fields = hfs

	// Add a new field
	for i := bi; i < len(bs); i++ {
		err := h.HeaderAdd(n, bs[i])
		if err != nil {
			h.fields = orig
			return err
		}
	}

	return nil
}

// HeaderAdd will add a new header field with the given name and body value to
// the bottom of the header. Existing headers will be left alone as-is.
//
// Returns an error if the given header name or body value is not legal.
func (h *Header) HeaderAdd(n, b string) error {
	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	h.fields = append(h.fields, f)
	return nil
}

// HeaderAddBefore will add a new header field with the given name and body value to
// teh top of the header. Existing headers will be left alone as-is.
//
// Returns an error if the given header name or body value is not legal.
func (h *Header) HeaderAddBefore(n, b string) error {
	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	h.fields = append([]*HeaderField{f}, h.fields...)
	return nil
}

// HeaderAddN will add a new header field after the (ix+1)th instance of the
// given header. If there is any header field with the given name, but not
// a (ix+1)th header, it will be added after the lsat one.
//
// If ix is negative, then it will be added after the (-ix)th header field from
// the end. If there is at least one header field with the given name, but no
// (-ix)th header, it will be added after the first one.
//
// If no header field with the given name is present, the header will be added
// to the bottom of the header.
//
// If the given field name or body value is not legal, an error will be returned
// and the header will not be modified.
func (h *Header) HeaderAddN(n, b string, ix int) error {
	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	if i := h.headerFieldIndex(n, ix, true); i > -1 {
		var a, b []*HeaderField
		b = h.fields[:i]
		if len(h.fields) > i+1 {
			a = h.fields[i+1:]
		} else {
			a = []*HeaderField{}
		}
		h.fields = append(b, f)
		h.fields = append(h.fields, a...)
		return nil
	}

	h.fields = append(h.fields, f)
	return nil
}

// HeaderAddBeforeN will add a new header field before the (ix+1)th instance of the
// given header. If there is any header field with the given name, but not
// a (ix+1)th header, it will be added before the lsat one.
//
// If ix is negative, then it will be added before the (-ix)th header field from
// the end. If there is at least one header field with the given name, but no
// (-ix)th header, it will be added before the first one.
//
// If no header field with the given name is present, the header will be added
// to the top of the header.
//
// If the given field name or body value is not legal, an error will be returned
// and the header will not be modified.
func (h *Header) HeaderAddBeforeN(n, b string, ix int) error {
	f, err := NewHeaderField(n, b, h.Break())
	if err != nil {
		return err
	}

	if i := h.headerFieldIndex(n, ix, true); i > -1 {
		var a, b []*HeaderField
		if i > 0 {
			b = h.fields[:i-1]
		} else {
			b = []*HeaderField{}
		}
		a = h.fields[i:]
		h.fields = append(b, f)
		h.fields = append(h.fields, a...)
		return nil
	}

	h.fields = append(h.fields, f)
	return nil
}

// HeaderDelete will remove the (ix+1)th header matching the given name. Returns
// an error if no such header exists.
func (h *Header) HeaderDelete(n string, ix int) error {
	count := 0
	m := makeMatch(n)
	for i, f := range h.fields {
		if f.Match() == m {
			if count == ix {
				h.fields = append(h.fields[:i-1], h.fields[i+1:]...)
				return nil
			}
			count++
		}
	}
	return fmt.Errorf("cannot delete index %d of the %q header", ix, n)
}

// HeaderDeleteAll will remove all headers matching the given name.
func (h *Header) HeaderDeleteAll(n string) {
	m := makeMatch(n)
	nf := make([]*HeaderField, 0, len(h.fields))
	for _, f := range h.fields {
		if f.Match() != m {
			nf = append(nf, f)
		}
	}
	h.fields = nf
}

// HeaderFields will return a slice containing all header fields.
func (h *Header) HeaderFields() []*HeaderField {
	return h.fields
}

// NewHeaderField constructs a new header field using the given name, body, and
// line break string.
func NewHeaderField(n, b string, lb []byte) (*HeaderField, error) {
	f := HeaderField{
		original: []byte(": "),
	}

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
