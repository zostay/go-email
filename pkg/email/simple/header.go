package simple

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

var (
	ErrIllegalFieldName = errors.New("field name contains illegal character")
	ErrIllegalBodyValue = errors.New("body value contains illegal character")
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

// BodySetter is a helper that is used to convert complex types used to set
// headers into strings and then set the field body value.
type BodySetter func(*email.HeaderField, interface{}, []byte) error

func defaultBodySetter(hf *email.HeaderField, b interface{}, lb []byte) error {
	var sb string
	switch bv := b.(type) {
	case string:
		sb = bv
	case []byte:
		sb = string(bv)
	case fmt.Stringer:
		sb = bv.String()
	default:
		return errors.New("unsupported value type set on header body")
	}

	if strings.IndexFunc(sb, forbiddenBodyChars) > -1 {
		return ErrIllegalBodyValue
	}

	hf.SetBody(sb, lb)

	return nil
}

// Header provides a simple header interface for dealing with email headers
// without worrying about MIME details.
type Header struct {
	// SetBody is set to a BodySetter that handles string, slices and arrays of
	// bytes, and fmt.Stringer objects.
	SetBody BodySetter
	email.Header
}

// NewHeader will build a new simple header. The strings provied are alternating
// name/value pairs.
//
// This will return an error if an odd number of strings is passed. It may also
// return an error if the given header names or body values are not limited to
// legal characters.
func NewHeader(lb string, hs ...string) (*Header, error) {
	if len(hs)%2 != 0 {
		return nil, errors.New("header field name provided with no body value")
	}

	h := Header{defaultBodySetter, *email.NewHeader(lb)}

	var n string
	for i, v := range hs {
		if i%2 == 0 {
			n = v
		} else {
			err := h.HeaderAdd(n, v)
			if err != nil {
				return nil, err
			}
		}
	}

	return &h, nil
}

// ParseHeader will parse the given string into an email header. It assumes that
// the entire string given represents the header. It will assume "\n" as the
// line break character to use during parsing.
func ParseHeader(m []byte) (*Header, error) {
	return ParseHeaderLB(m, []byte(email.LF))
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

	h := Header{defaultBodySetter, *email.NewHeader(string(lb))}

	for _, line := range lines {
		hf, err := ParseHeaderField(line, lb)
		if err != nil {
			errs = append(errs, err)
		}

		if hf != nil {
			h.Fields = append(h.Fields, hf)
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
func ParseHeaderField(f, lb []byte) (*email.HeaderField, error) {
	parts := bytes.SplitN(f, []byte(":"), 2)
	if len(parts) < 2 {
		name := email.UnfoldValue(f)
		return email.NewHeaderFieldParsed(string(name), "", f), fmt.Errorf("header field %q missing body", name)
	}

	name := strings.TrimSpace(string(email.UnfoldValue(parts[0])))
	body := strings.TrimSpace(string(email.UnfoldValue(parts[1])))

	return email.NewHeaderFieldParsed(name, body, f), nil
}

// HeaderNames will return the unique header names found in the mail header.
func (h *Header) HeaderNames() []string {
	seen := map[string]struct{}{}
	names := make([]string, 0, len(h.Fields))
	for _, f := range h.Fields {
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

// HeaderRename will find the first header with a matching name and give it a
// new name.
//
// Returns an error if that name is not legal or if no header with the given
// name is found.
func (h *Header) HeaderRename(oldName, newName string) error {
	hf, err := h.HeaderGetFieldN(oldName, 0)
	if err != nil {
		return err
	}

	hf.SetName(newName, h.Break())
	return nil
}

// HeaderRenameN will find the (ix+1)th from the top or the (-ix)th from the
// bottom and give it a new name.
//
// Returns an error if that name is not legal or if no header with the given
// name is found.
func (h *Header) HeaderRenameN(oldName, newName string, ix int) error {
	hf, err := h.HeaderGetFieldN(oldName, ix)
	if err != nil {
		return err
	}

	hf.SetName(newName, h.Break())
	return nil
}

// HeaderRenameAll will rename all headers with the first name to have the name
// given second. If no headers with the given name are found, this method does
// nothing.
//
// Returns an error if the name is not legal of if no header wwith the given
// name is found.
func (h *Header) HeaderRenameAll(oldName, newName string) error {
	found := false
	m := email.MakeHeaderFieldMatch(oldName)
	for _, hf := range h.Fields {
		if hf.Match() == m {
			found = true
			hf.SetName(newName, h.Break())
		}
	}

	if !found {
		return fmt.Errorf("No header named %q found for renaming.", oldName)
	}

	return nil
}

func forbiddenNameChars(c rune) bool {
	if c == ':' {
		return true
	}
	if c >= 33 && c <= 126 {
		return false
	}
	return true
}

func checkName(n string) error {
	if strings.IndexFunc(n, forbiddenNameChars) > -1 {
		return ErrIllegalFieldName
	}
	return nil
}

func forbiddenBodyChars(c rune) bool {
	if c == ' ' || c == '\t' {
		return false
	}
	if c >= 33 && c <= 126 {
		return false
	}
	return true
}

// HeaderSet will find the first header with a matching name and replace it
// with the given body. If no header by that name is set, it will add a new
// header with that name and body.
func (h *Header) HeaderSet(n string, b interface{}) error {
	if err := checkName(n); err != nil {
		return err
	}

	if i := h.HeaderFieldIndex(n, 0, false); i > -1 {
		f := h.Fields[i]
		if err := h.SetBody(f, b, h.Break()); err != nil {
			return err
		}
		return nil
	}

	f := email.NewHeaderField(n, "", h.Break())
	if err := h.SetBody(f, b, h.Break()); err != nil {
		return err
	}

	h.Fields = append(h.Fields, f)

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
func (h *Header) HeaderSetN(n string, b interface{}, ix int) error {
	hf, err := h.HeaderGetFieldN(n, ix)
	if err != nil {
		return err
	}

	if err := h.SetBody(hf, b, h.Break()); err != nil {
		return err
	}

	return nil
}

// HeaderSetAll does a full header replacement. This performs a number of
// combined operations.
//
// 1. If no body values are given, this is equivalent to HeaderDeleteAll.
//
//  2. If some body values are given and there are some headers already present.
//     This is equivalent to calling HeaderSet for each of those values.
//
//  3. If there are fewer body values than headers already present, the remaining
//     headers will be deleted.
//
//  4. If there are more body values than headers already present, this is
//     equivalent to calling HeaderAdd for each of those headers.
//
// Basically, it's going to make sure all the given headers are set and will
// start by changing the ones already in place, removing any additional ones
// that aren't updated, and adding new ones if necessary.
//
// If the operation is successful, it returns nil. If there is an error, then
// the object will be unchanged and an error returned.
func (h *Header) HeaderSetAll(n string, bs ...interface{}) error {
	if err := checkName(n); err != nil {
		return err
	}

	// no values, so delete all
	if len(bs) == 0 {
		h.HeaderDeleteAll(n)
		return nil
	}

	hfs := make([]*email.HeaderField, 0, len(h.Fields))
	m := email.MakeHeaderFieldMatch(n)
	bi := 0
	for _, hf := range h.Fields {
		if hf.Match() == m {
			// Set existing field
			if bi < len(bs) {
				err := h.SetBody(hf, bs[bi], h.Break())
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

	orig := h.Fields
	h.Fields = hfs

	// Add a new field
	for i := bi; i < len(bs); i++ {
		f := email.NewHeaderField(n, "", h.Break())
		err := h.SetBody(f, bs[bi], h.Break())
		if err != nil {
			h.Fields = orig
			return err
		}
		h.Fields = append(h.Fields, f)
	}

	return nil
}

// HeaderAdd will add a new header field with the given name and body value to
// the bottom of the header. Existing headers will be left alone as-is.
//
// Returns an error if the given header name or body value is not legal.
func (h *Header) HeaderAdd(n string, b interface{}) error {
	if err := checkName(n); err != nil {
		return err
	}

	f := email.NewHeaderField(n, "", h.Break())
	if err := h.SetBody(f, b, h.Break()); err != nil {
		return err
	}

	h.Fields = append(h.Fields, f)
	return nil
}

// HeaderAddBefore will add a new header field with the given name and body value to
// teh top of the header. Existing headers will be left alone as-is.
//
// Returns an error if the given header name or body value is not legal.
func (h *Header) HeaderAddBefore(n string, b interface{}) error {
	if err := checkName(n); err != nil {
		return err
	}

	f := email.NewHeaderField(n, "", h.Break())
	if err := h.SetBody(f, b, h.Break()); err != nil {
		return err
	}

	h.Fields = append([]*email.HeaderField{f}, h.Fields...)
	return nil
}

// HeaderAddN will add a new header field after the (ix+1)th instance of the
// given header. If there is any header field with the given name, but not
// a (ix+1)th header, it will be added after the last one.
//
// If ix is negative, then it will be added after the (-ix)th header field from
// the end. If there is at least one header field with the given name, but no
// (-ix)th header, it will be added after the first one.
//
// If no header field with the given name is present, the new header field will
// be added to the bottom of the header.
//
// If the given field name or body value is not legal, an error will be returned
// and the header will not be modified.
func (h *Header) HeaderAddN(n string, b interface{}, ix int) error {
	if err := checkName(n); err != nil {
		return err
	}

	f := email.NewHeaderField(n, "", h.Break())
	if err := h.SetBody(f, b, h.Break()); err != nil {
		return err
	}

	if i := h.HeaderFieldIndex(n, ix, true); i > -1 {
		// make space for the new entry
		h.Fields = append(h.Fields, nil)

		// move everything after i back one
		copy(h.Fields[i+1:], h.Fields[i:])

		// Insert the new entry at i+1 (after i)
		h.Fields[i+1] = f

		return nil
	}

	h.Fields = append(h.Fields, f)
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
// If no header field with the given name is present, the new header field will
// be added to the top of the header.
//
// If the given field name or body value is not legal, an error will be returned
// and the header will not be modified.
func (h *Header) HeaderAddBeforeN(n string, b interface{}, ix int) error {
	if err := checkName(n); err != nil {
		return err
	}

	f := email.NewHeaderField(n, "", h.Break())
	if err := h.SetBody(f, b, h.Break()); err != nil {
		return err
	}

	if i := h.HeaderFieldIndex(n, ix, true); i > -1 {
		// make space for the new entry
		h.Fields = append(h.Fields, nil)

		// move everything after i back one
		copy(h.Fields[i+1:], h.Fields[i:])

		// Insert the new entry at i (before i)
		h.Fields[i] = f

		return nil
	}

	h.Fields = append(h.Fields, nil)
	copy(h.Fields[1:], h.Fields[0:])
	h.Fields[0] = f
	return nil
}

// HeaderDelete will remove the (ix+1)th header matching the given name. Returns
// an error if no such header exists.
func (h *Header) HeaderDelete(n string, ix int) error {
	count := 0
	m := email.MakeHeaderFieldMatch(n)
	for i, f := range h.Fields {
		if f.Match() == m {
			if count == ix {
				if i == 0 {
					h.Fields = h.Fields[1:]
				} else {
					h.Fields = append(h.Fields[:i], h.Fields[i+1:]...)
				}
				return nil
			}
			count++
		}
	}
	return fmt.Errorf("cannot delete index %d of the %q header", ix, n)
}

// HeaderDeleteAll will remove all headers matching the given name.
func (h *Header) HeaderDeleteAll(n string) {
	m := email.MakeHeaderFieldMatch(n)
	nf := make([]*email.HeaderField, 0, len(h.Fields))
	for _, f := range h.Fields {
		if f.Match() != m {
			nf = append(nf, f)
		}
	}
	h.Fields = nf
}
