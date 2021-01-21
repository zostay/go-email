package email

import (
	"errors"
	"fmt"
	"strings"
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
	Fields []*HeaderField // The list of fields
	Break  string         // The line break string to use
}

// HeaderField represents an individual field in the message header. When taken
// from a parsed header, it will preserve the original field, byte-for-byte.
type HeaderField struct {
	match    string
	name     string
	body     string
	original string
}

// ParseHeader will parse the given string into an email header. It assumes that
// the entire string given represents the header. It will assume "\n" as the
// line break character to use during parsing.
func ParseHeader(m string) (*Header, error) {
	return ParseHeaderLB(m, "\x0d")
}

// ParseHeaderLB will parse the given string into an email header using the
// given line break string. It will assume the entire string given represents
// the header to be parsed.
func ParseHeaderLB(m, lb string) (*Header, error) {
	lines, err := ParseHeaderLines(m, lb)
	errs := make([]error, 0, 1)
	if err != nil {
		errs = append(errs, err)
	}

	h := Header{
		Fields: make([]*HeaderField, len(lines)),
		Break:  lb,
	}

	for i, line := range lines {
		h.Fields[i], err = ParseHeaderField(line, lb)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return &h, &HeaderParseError{errs}
}

// ParseHeaderLines is used by ParseHeader and ParseHeaderLB to create a list of
// lines where each line represents a single field, including and folded
// continuation lines.
func ParseHeaderLines(m, lb string) ([]string, error) {
	h := make([]string, 0, len(m)/80)
	var err error
	for _, line := range strings.SplitAfter(m, lb) {
		if strings.HasPrefix(line, " ") {
			// Start with a continuation? Weird, uh...
			if len(h) == 0 {
				err = ErrContinuationStart
				h = append(h, line)
				continue
			}

			h[len(h)-1] += line
		} else {
			h = append(h, line)
		}
	}

	return h, err
}

// ParseHeaderField will take a single header field, including any folded
// continuation lines. This will then construct a header field object.
func ParseHeaderField(f, lb string) (*HeaderField, error) {
	parts := strings.SplitN(f, ":", 2)
	if len(parts) < 2 {
		name := UnfoldValue(f, lb)
		return &HeaderField{"", name, "", f}, fmt.Errorf("header field %q missing body", name)
	}

	name := UnfoldValue(parts[0], lb)
	body := UnfoldValue(parts[1], lb)
	return &HeaderField{"", name, body, f}, nil
}

// String will return the string representation of the header. If the header was
// parsed from an email header and not modified, this will output the original
// header, preserved byte-for-byte.
func (h *Header) String() string {
	var out strings.Builder
	for _, f := range h.Fields {
		out.WriteString(f.String())
	}
	return out.String()
}

// Names will return the unique header names found in the mail header.
func (h *Header) Names() []string {
	seen := map[string]struct{}{}
	names := make([]string, 0, len(h.Fields))
	for _, f := range h.Fields {
		if _, ok := seen[f.Match()]; ok {
			continue
		}
		names = append(names, f.Name())
	}
	return names
}

// Get will find the first header with a matching name and return body value. It
// will return an empty string if no such header is present.
func (h *Header) Get(n string) string {
	m := makeMatch(n)
	for _, f := range h.Fields {
		if f.Match() == m {
			return f.Body()
		}
	}

	return ""
}

// GetAll will find all headers with a matching name and return a list of body
// values. Returns nil if no matching headers are present.
func (h *Header) GetAll(n string) []string {
	bs := make([]string, 0)
	m := makeMatch(n)
	for _, f := range h.Fields {
		if f.Match() == m {
			bs = append(bs, f.Body())
		}
	}

	return bs
}

// Set will find the first header with a matching name and replace it with the
// given body. If no header by that name is set, it will add a new header with
// that name and body.
func (h *Header) Set(n, b string) error {
	m := makeMatch(n)
	for _, f := range h.Fields {
		if f.Match() == m {
			return f.SetBody(b, h.Break)
		}
	}

	f, err := NewHeaderField(n, b, h.Break)
	if err != nil {
		return err
	}

	h.Fields = append(h.Fields, f)

	return nil
}

// Add will add a new header with the given name and body value. If an existing
// header with the same value is already present, this will add the new field
// before the first field with the same name.
func (h *Header) Add(n, b string) error {
	f, err := NewHeaderField(n, b, h.Break)
	if err != nil {
		return err
	}

	m := makeMatch(n)
	for i, f := range h.Fields {
		if f.Match() == m {
			b := h.Fields[:i]
			a := h.Fields[i:]
			h.Fields = append(b, f)
			h.Fields = append(h.Fields, a...)
			return nil
		}
	}

	h.Fields = append(h.Fields, f)
	return nil
}

// NewHeaderFields constructs a new header field using the given name, body, and
// line break string.
func NewHeaderField(n, b, lb string) (*HeaderField, error) {
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
func (f *HeaderField) Original() string { return f.original }

// String is an alias for Original.
func (f *HeaderField) String() string { return f.original }

// SetName will rename a field. This first checks to make sure no illegal
// characters are present in the field name. It will return an error if the
// given string contains a colon or any character outside the printable ASCII
// range.
func (f *HeaderField) SetName(n string) error {
	allowedNameChars := func(c rune) bool {
		if c == ':' {
			return false
		}
		if c >= 33 && c <= 126 {
			return true
		}
		return false
	}

	if strings.IndexFunc(n, allowedNameChars) > -1 {
		return errors.New("header name contains illegal character")
	}

	f.SetNameUnsafe(n)
	return nil
}

// SetNameUnsafe will rename a field without checks.
func (f *HeaderField) SetNameUnsafe(n string) {
	f.original = n + f.original[len(f.name):]
	f.name = n
}

// SetBody will update the body of the field. You must supply the line break to
// be used for folding. Before setting the field, the value will be checked to
// make sure it is legal. It is only permitted to contain printable ASCII,
// space, and tab characters.
func (f *HeaderField) SetBody(b, lb string) error {
	allowedBodyChars := func(c rune) bool {
		if c == ' ' || c == '\t' {
			return true
		}
		if c >= 33 && c <= 126 {
			return true
		}
		return false
	}

	if strings.IndexFunc(b, allowedBodyChars) > -1 {
		return errors.New("body name contains illegal character")
	}

	f.SetBodyUnsafe(b, lb)
	return nil
}

// SetBodyUnsafe will update the body of the field without checking to make sure
// it is valid. This can be used to provide a prefolded body (though, it will
// still be folded further if any long lines are found).
func (f *HeaderField) SetBodyUnsafe(b, lb string) {
	f.original = FoldValue(f.original[:len(f.name)+1]+" "+b, lb)
	f.body = b
}

// SetBodyNoFold will update the body of the field without checking to make sure
// it is valid and without performing any folding.
func (f *HeaderField) SetBodyNoFold(b string) {
	f.original = f.original[:len(f.name)+1] + " " + b
	f.body = b
}
