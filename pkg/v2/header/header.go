package header

import (
	"errors"
	"mime"
	"net/mail"
	"strings"
	"time"

	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/pkg/v2/param"
)

// Errors returned by various header methods and functions.
var (
	// ErrNoSuchField is returned by Header methods when the operation
	// being performed failed because the header named does not exist.
	ErrNoSuchField = errors.New("not such header field")

	// ErrNoSuchFieldParameter is returned by Header methods when the
	// operation being performed failed because the header exists, but a
	// sub-field of the header does not exist.
	ErrNoSuchFieldParameter = errors.New("no such header field parameter")

	// ErrManyFields is returned by Header methods when the operation
	// being performed failed because the there are multiple fields with the
	// given name.
	ErrManyFields = errors.New("many header fields found")

	// ErrWrongAddressType is returned by address setting methods that accept
	// either a string or an addr.AddressList when something other than those
	// types is provided.
	ErrWrongAddressType = errors.New("incorrect address type during write")
)

// Helpful constants to common headers.
const (
	ContentDisposition = "Content-disposition"
	ContentType        = "Content-type"
	Date               = "Date"
	Subject            = "Subject"
	To                 = "To"
	Cc                 = "Cc"
	Bcc                = "Bcc"
	From               = "From"
	ReplyTo            = "Reply-To"
)

type Header struct {
	Base

	// valueCache holds the semantic value for a header. As of this time, we
	// assume that all headers that have a semantic value are singular, which is
	// safe for content-type, content-disposition, from, to, date, cc, bcc, etc.
	// These fields are not typically repeated in an email header. If this were
	// to add received as a semantic header, then we'd have a problem with this,
	// though that might be resolved by just storing a slice of parsed received
	// headers or something.
	//
	// REMEMBER: This must only be used to hold "immutable" types. If a type can
	// be modified outside, we can have inconsistencies between what is stored
	// in valueCache and what is set in simple.Header
	valueCache map[string]any
}

// func NewHeader(in mail.Header) *Header {
// 	if in == nil {
// 		return &Header{}
// 	}
//
// 	var h Header
// 	for name, bodies := range in {
// 		for _, body := range bodies {
// 			h.InsertBeforeField(h.Size(), name, body)
// 		}
// 	}
//
// 	return &h
// }

func (h *Header) Get(name string) (string, error) {
	ixs := h.GetIndexesNamed(name)
	if len(ixs) == 0 {
		return "", ErrNoSuchField
	}

	if len(ixs) > 1 {
		return "", ErrManyFields
	}

	return h.GetField(ixs[0]).Body(), nil
}

func (h *Header) getValue(name string) (any, bool) {
	n := strings.ToLower(name)
	v, found := h.valueCache[n]
	return v, found
}

func (h *Header) setValue(name string, value any) {
	if h.valueCache == nil {
		h.valueCache = make(map[string]any, h.Size())
	}
	n := strings.ToLower(name)
	h.valueCache[n] = value
}

func (h *Header) getTime(name string) (time.Time, error) {
	f := h.GetFieldNamed(name, 0)
	if f == nil {
		return time.Time{}, ErrNoSuchField
	}

	t, err := mail.ParseDate(f.Body())
	if err != nil {
		return time.Time{}, err
	}

	h.setValue(name, t)

	return t, nil
}

func (h *Header) GetTime(name string) (time.Time, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getTime(name)
	}

	t, isTime := v.(time.Time)
	if !isTime {
		return h.getTime(name)
	}

	return t, nil
}

func (h *Header) getAddressList(name string) (addr.AddressList, error) {
	f := h.GetFieldNamed(name, 0)
	if f == nil {
		return nil, ErrNoSuchField
	}

	al, err := addr.ParseEmailAddressList(f.Body())
	if err != nil {
		al = parseEmailAddressList(f.Body())
	}

	h.setValue(name, al)

	return al, nil
}

func (h *Header) GetAddressList(name string) (addr.AddressList, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getAddressList(name)
	}

	al, isAddrList := v.(addr.AddressList)
	if !isAddrList {
		return h.getAddressList(name)
	}

	return al, nil
}

func (h *Header) getParamValue(name string) (*param.Value, error) {
	f := h.GetFieldNamed(name, 0)
	if f == nil {
		return nil, ErrNoSuchField
	}

	pv, err := param.Parse(f.Body())
	if err != nil {
		return nil, err
	}

	h.setValue(name, pv)

	return pv, nil
}

func (h *Header) GetParamValue(name string) (*param.Value, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getParamValue(name)
	}

	pv, isPV := v.(*param.Value)
	if !isPV {
		return h.getParamValue(name)
	}

	if pv == nil {
		return pv, nil
	}

	// return a copy to prevent the cached value from being modified
	return pv.Clone(), nil
}

func forbiddenBodyChar(c rune) bool {
	if c == ' ' || c == '\t' {
		return false
	}
	if c >= 33 && c <= 126 {
		return false
	}
	return true
}

func (h *Header) Set(name, body string) {
	// if the given header body contains forbidden characters, encode the header
	if strings.IndexFunc(body, forbiddenBodyChar) > -1 {
		body = mime.BEncoding.Encode("utf-8", body)
	}

	// Check for existing fields
	ixs := h.GetIndexesNamed(name)

	// if none, insert the new field and we're done
	if len(ixs) == 0 {
		h.InsertBeforeField(h.Size(), name, body)
		return
	}

	// if more than one, we're setting so delete any but the first
	if len(ixs) > 1 {
		for i := len(ixs) - 1; i > 0; i-- {
			// ignore out of range errors, we don't make that mistake here
			_ = h.DeleteField(ixs[i])
		}
	}

	// get the field we want to modify or replace
	f := h.GetField(ixs[0])
	f.SetName(name)
	f.SetBody(body)
}

func (h *Header) SetTime(name string, body time.Time) {
	h.setValue(name, body)
	bodyStr := body.Format(time.RFC1123Z)
	h.Set(name, bodyStr)
}

func (h *Header) SetAddressList(name string, body addr.AddressList) {
	h.setValue(name, body)
	body[0].Address()
	bodyStr := body.String()
	h.Set(name, bodyStr)
}

func (h *Header) SetParamValue(name string, body *param.Value) {
	h.setValue(name, body)
	bodyStr := body.String()
	h.Set(name, bodyStr)
}

func (h *Header) getParamValueValue(name string) (string, error) {
	pv, err := h.GetParamValue(name)
	if err != nil {
		return "", err
	}

	return pv.Value(), nil
}

func (h *Header) setParamValueValue(name, v string) {
	pv, err := h.GetParamValue(name)
	if err != nil {
		if errors.Is(err, ErrNoSuchField) {
			pv := param.New(v)
			h.SetParamValue(name, pv)
		}
		h.Set(name, v)
	}

	newPv := param.Modify(pv, param.Change(v))

	h.SetParamValue(name, newPv)
}

func (h *Header) getParamValueParam(name, p string) (string, error) {
	pv, err := h.GetParamValue(name)
	if err != nil {
		return "", err
	}

	if v := pv.Parameter(p); v != "" {
		return v, nil
	}

	return "", ErrNoSuchFieldParameter
}

func (h *Header) setParamValueParam(name, p, v string) error {
	pv, err := h.GetParamValue(name)
	if err != nil {
		return err
	}

	newPv := param.Modify(pv, param.Set(p, v))
	h.SetParamValue(name, newPv)

	return nil
}

// GetContentType returns the MIME type set in the Content-type header (other
// parameters will not be returned).
func (h *Header) GetContentType() (string, error) {
	return h.getParamValueValue(ContentType)
}

// SetContentType sets the MIME type on the Content-type header, creating it if
// it has not been set yet.
func (h *Header) SetContentType(mt string) {
	h.setParamValueValue(ContentType, mt)
}

// GetCharset gets the charset from the Content-type header or returns an error
// if either the Content-type is missing or charset is not set as a parameter on
// it.
func (h *Header) GetCharset() (string, error) {
	return h.getParamValueParam(ContentType, param.Charset)
}

// SetCharset sets the charset on the Content-type header or returns an error if
// Content-type header is not set at all. You must set a MIME type in the
// Content-type header before setting the charset.
func (h *Header) SetCharset(c string) error {
	return h.setParamValueParam(ContentType, param.Charset, c)
}

// GetBoundary gets the boundary fro the Content-type header or returns an error
// if either the Content-type is missing or boundary is not set as a parameter
// on it.
func (h *Header) GetBoundary() (string, error) {
	return h.getParamValueParam(ContentType, param.Boundary)
}

// SetBoundary sets the boundary on the Content-type header or returns ane error
// if the Content-type header is not set at all. You must set a MIME type in the
// Content-type header before setting the boundary.
func (h *Header) SetBoundary(b string) error {
	return h.setParamValueParam(ContentType, param.Boundary, b)
}

// GetContentDisposition returns the primary value of the Content-disposition
// header, describing what the function of this part of the message is. This
// method returns an error if there is no Content-disposition header.
func (h *Header) GetContentDisposition() (string, error) {
	return h.getParamValueValue(ContentDisposition)
}

// SetContentDisposition sets the disposition value of the Content-disposition
// header field.
func (h *Header) SetContentDisposition(d string) {
	h.setParamValueValue(ContentDisposition, d)
}

// GetFilename gets the filename parameter of the Content-disposition header.
// It returns an error instead if either the Content-disposition header is not
// set or the filename parameter is not set on the header.
func (h *Header) GetFilename() (string, error) {
	return h.getParamValueParam(ContentDisposition, param.Filename)
}

// SetFilename sets the filename parameter of the Content-disposition header.
// Returns an error if the Content-disposition header does not exist yet. A
// disposition must be set before setting a filename.
func (h *Header) SetFilename(f string) error {
	return h.setParamValueParam(ContentDisposition, param.Filename, f)
}

// GetDate retrieves the Date header as a time.Time value.
func (h *Header) GetDate() (time.Time, error) {
	return h.GetTime(Date)
}

// SetDate updates the Date header from the given time.Time value.
func (h *Header) SetDate(d time.Time) {
	h.SetTime(Date, d)
}

// GetSubject returns the header subject or returns an error if the Subject
// header is not set.
func (h *Header) GetSubject() (string, error) {
	return h.Get(Subject)
}

// SetSubject sets the subject header.
func (h *Header) SetSubject(s string) {
	h.Set(Subject, s)
}

func (h *Header) setAddress(n string, a any) error {
	switch v := a.(type) {
	case string:
		h.Set(n, v)
	case addr.AddressList:
		h.SetAddressList(n, v)
	default:
		return ErrWrongAddressType
	}
	return nil
}

// GetTo returns the To address field as an addr.AddressList or returns an error
// if the To header does not exist or there's a problem decoding the addresses.
func (h *Header) GetTo() (addr.AddressList, error) {
	return h.GetAddressList(To)
}

// SetTo sets the To address field with either an addr.AddressList or a string.
// It will fail with an error returned if something other than those types is
// provided.
func (h *Header) SetTo(a any) error {
	return h.setAddress(To, a)
}

// GetCc returns the Cc address field as an addr.AddressList or returns an error
// if the Cc header does not exist or there's a problem decoding the addresses.
func (h *Header) GetCc() (addr.AddressList, error) {
	return h.GetAddressList(Cc)
}

// SetCc sets the Cc address field with either an addr.AddressList or a string.
// It will fail with an error returned if something other than those types is
// provided.
func (h *Header) SetCc(a any) error {
	return h.setAddress(Cc, a)
}

// GetBcc returns the Bcc address field as an addr.AddressList or returns an error
// if the Bcc header does not exist or there's a problem decoding the addresses.
func (h *Header) GetBcc() (addr.AddressList, error) {
	return h.GetAddressList(Bcc)
}

// SetBcc sets the Bcc address field with either an addr.AddressList or a string.
// It will fail with an error returned if something other than those types is
// provided.
func (h *Header) SetBcc(a any) error {
	return h.setAddress(Bcc, a)
}

// GetFrom returns the From address field as an addr.AddressList or returns an error
// if the From header does not exist or there's a problem decoding the addresses.
func (h *Header) GetFrom() (addr.AddressList, error) {
	return h.GetAddressList(From)
}

// SetFrom sets the From address field with either an addr.AddressList or a string.
// It will fail with an error returned if something other than those types is
// provided.
func (h *Header) SetFrom(a any) error {
	return h.setAddress(From, a)
}

// GetReplyTo returns the ReplyTo address field as an addr.AddressList or returns an error
// if the ReplyTo header does not exist or there's a problem decoding the addresses.
func (h *Header) GetReplyTo() (addr.AddressList, error) {
	return h.GetAddressList(ReplyTo)
}

// SetReplyTo sets the ReplyTo address field with either an addr.AddressList or a string.
// It will fail with an error returned if something other than those types is
// provided.
func (h *Header) SetReplyTo(a any) error {
	return h.setAddress(ReplyTo, a)
}

// parseEmailAddressList is a fallback method for email address parsing. The
// parser in github.com/zostay/go-addr is a strict parser, which is useful for
// getting good accurate parsing of email addresses, especially for validating
// data entry. However, when working with the mess that is the Internet, you
// want to get something useful (strict out/liberal in), even if its technically
// wrong, well, this method can be used to clean up the mess.
//
// It works as follows:
//
// 1. Split the string up by commas.
// 2. Each string resulting from the split is trimmed of whitespace.
// 3. The comments are stripped from each string and held.
// 4. All the words at the start are treated as the display name.
// 5. The last word at the end is treated as the email address.
//
// As some address fields have something other than an address in it because
// people on the Internet are weird, the result will be wrong sometimes.
//
// We stuff whatever we get into an addr.Mailbox and call it good. As they are
// so rare, we will assume we are never dealing with groups. This may lead to
// oddness if a group is encountered.
func parseEmailAddressList(v string) addr.AddressList {
	extractComments := func(s string) (string, string) {
		var clean, comment strings.Builder
		nestLevel := 0
		for _, c := range s {
			if c == '(' {
				nestLevel++
				if nestLevel == 1 {
					continue
				} else {
					comment.WriteRune(c)
				}
			} else if c == ')' {
				nestLevel--
				if nestLevel == 0 {
					continue
				} else if nestLevel < 0 {
					nestLevel = 0
					clean.WriteRune(c)
				} else {
					comment.WriteRune(c)
				}
			} else if nestLevel > 0 {
				comment.WriteRune(c)
			} else {
				clean.WriteRune(c)
			}
		}

		return clean.String(), comment.String()
	}

	mbs := strings.Split(v, ",")
	as := make(addr.AddressList, 0, len(mbs))
	for _, orig := range mbs {
		mb, com := extractComments(orig)

		mb = strings.TrimSpace(mb)
		com = strings.TrimSpace(com)

		parts := strings.Fields(mb)

		var dn, email string
		if len(parts) == 0 {
			email = ""
		} else if len(parts) > 1 {
			dn = strings.Join(parts[:len(parts)-1], " ")
			email = parts[len(parts)-1]
		} else {
			email = parts[0]
		}

		if email != "" {
			var addrSpec *addr.AddrSpec
			if i := strings.Index(email, "@"); i > -1 {
				addrSpec = addr.NewAddrSpecParsed(
					email[:i],
					email[i+1:],
					email,
				)
			} else {
				addrSpec = addr.NewAddrSpecParsed(
					email,
					"",
					email,
				)
			}

			mailbox, err := addr.NewMailboxParsed(dn, addrSpec, com, orig)
			if err != nil {
				mailbox, _ = addr.NewMailboxParsed(dn, addrSpec, "", orig)
			}

			as = append(as, mailbox)
		}
	}

	return as
}

// var _ email.BasicHeader = &Header{}
// var _ email.Outputter = &Header{}
// var _ email.WithMutableBreak = &Header{}
