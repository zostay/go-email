package email

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrContinuationStart = errors.New("header starts with a continuation")
)

type HeaderParseError struct {
	Errs []error
}

func (err *HeaderParseError) Error() string {
	errs := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errs[i] = e.Error()
	}
	return "error parsing email header: " + strings.Join(errs, ", ")
}

type Header struct {
	Fields []*HeaderField
	Break  string
}

type HeaderField struct {
	match    string
	name     string
	body     string
	original string
}

func ParseHeader(m string) (*Header, error) {
	return ParseHeaderLB(m, "\x0d")
}

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

func (h *Header) String() string {
	var out strings.Builder
	for _, f := range h.Fields {
		out.WriteString(f.String())
	}
	return out.String()
}

func (f *HeaderField) Match() string {
	if f.match != "" {
		return f.match
	}

	f.match = strings.ToLower(strings.TrimSpace(f.name))
	return f.match
}

func (f *HeaderField) Name() string     { return f.name }
func (f *HeaderField) Body() string     { return f.body }
func (f *HeaderField) Original() string { return f.original }
func (f *HeaderField) String() string   { return f.original }

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

func (f *HeaderField) SetNameUnsafe(n string) {
	f.original = n + f.original[len(f.name):]
	f.name = n
}

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

func (f *HeaderField) SetBodyUnsafe(b, lb string) {
	f.original = FoldValue(f.original[:len(f.name)+1]+" "+b, lb)
	f.body = b
}
