package header

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/v2/message/header/param"
)

// Errors returned by various header methods and functions.
var (
	// ErrNoSuchField is returned by Header methods when the operation
	// being performed failed because the header named does not exist.
	ErrNoSuchField = errors.New("no such header field")

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

// These are standard headers defined in RFC 5322.
const (
	Bcc                     = "Bcc"
	Cc                      = "Cc"
	Comments                = "Comments"
	ContentDisposition      = "Content-disposition"
	ContentTransferEncoding = "Content-transfer-encoding"
	ContentType             = "Content-type"
	Date                    = "Date"
	From                    = "From"
	InReplyTo               = "In-reply-to"
	Keywords                = "Keywords"
	MessageID               = "Message-id"
	References              = "References"
	ReplyTo                 = "Reply-to"
	Sender                  = "Sender"
	Subject                 = "Subject"
	To                      = "To"
)

// Even more custom date formats, built from those seen in the wild that the
// usual parsers have trouble with.
const (
	// UnixDateWithEarlyYear is a weird one, eh?
	UnixDateWithEarlyYear = "Mon Jan 02 15:04:05 2006 MST"
)

// Header wraps a Base, which does the actual storage and low-level field
// manipulation. This provides several methods to make reading and manipulating
// the header more convenient and some caching for complex values parsed from
// header fields.
//
// The getter methods of this object will return an error if the field being
// fetched has not been set on the header. The error returned will be
// ErrNoSuchField.
type Header struct {
	// Base provides the low-level storage of header fields.
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

// Clone returns a deep copy of the header object.
func (h *Header) Clone() *Header {
	// the value cache objects are immutable, so they may be copied as-is
	vc := make(map[string]any, len(h.valueCache))
	for k, v := range h.valueCache {
		vc[k] = v
	}

	return &Header{
		Base:       *h.Base.Clone(),
		valueCache: vc,
	}
}

// getValue retrieves the cached value. The first value is the cached value
// (which may be nil). The second value is a boolean that returns true if the
// cache value was set.
func (h *Header) getValue(name string) (any, bool) {
	n := strings.ToLower(name)
	v, found := h.valueCache[n]
	return v, found
}

// setValue replaces the cached value for the given name.
func (h *Header) setValue(name string, value any) {
	if h.valueCache == nil {
		h.valueCache = make(map[string]any, h.Len())
	}
	n := strings.ToLower(name)
	h.valueCache[n] = value
}

// Get retrieves the string value of the named field.
//
// If the named field is not set in the header, it will return an empty string
// with ErrNoSuchField. If there are multiple headers for the given named field,
// it will return the first value found and return ErrManyFields.
func (h *Header) Get(name string) (string, error) {
	ixs := h.GetIndexesNamed(name)
	if len(ixs) == 0 {
		return "", ErrNoSuchField
	}

	b := h.GetField(ixs[0]).Body()
	if len(ixs) > 1 {
		return b, ErrManyFields
	}

	return b, nil
}

// ParseTime is a function that provides the time parsing used by GetTime() and
// GetDate() to parse dates to be used on any field body. This will attempt to
// parse the date using the format specified by RFC 5322 first and fallback to
// parsing it in many other formats.
//
// It either returns a parsed time or the parse error.
func ParseTime(body string) (time.Time, error) {
	t, err := mail.ParseDate(body)
	if err == nil {
		return t, nil
	}

	t, err = dateparse.ParseAny(body)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse(UnixDateWithEarlyYear, body)
	if err == nil {
		return t, nil
	}

	return t, fmt.Errorf("time string %q cannot be parsed", body)
}

// getTime parses the header body as a date and caches the result.
func (h *Header) getTime(name string) (time.Time, error) {
	body, err := h.Get(name)
	if err != nil {
		return time.Time{}, err
	}

	t, err := ParseTime(body)
	if err != nil {
		return t, err
	}

	h.setValue(name, t)

	return t, nil
}

// GetTime gets the given date header field as a time.Time. It will attempt to
// parse the date in many formats, not just the format specified by RFC 5322
// (though, it will try that first).
//
// It will return an error if it is unable to parse the time value from the date
// header. It will return the zero value and ErrNoSuchField if the header does
// not exist. It will return the zero value and ErrManyFields if more than one
// field with the name is set on the header.
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

// ParseAddressList provides the same address parsing functionality build into
// the GetAddressList() and GetAllAddressLists() and can be used to parse any
// field body. It will attempt a strict parse of the email address list.
// However, if that fails, an extremely lenient parsing will be attempted, which
// might result in results that can only be described as "weird" in the effort
// to provide some kind of result. It is so forgiving, it will return some kind
// of value for any input.
//
// It will either return an addr.AddressList or an error describing the parse error.
func ParseAddressList(body string) addr.AddressList {
	al, err := addr.ParseEmailAddressList(body)
	if err != nil {
		al = parseEmailAddressList(body)
	}

	return al
}

// getAddressList will parse an addr.AddressList out of the field or return an
// error. This falls back onto parseEmailAddressList() if
// addr.ParseEmailAddrList() lets us down.
func (h *Header) getAddressList(name string) (addr.AddressList, error) {
	body, err := h.Get(name)
	if err != nil {
		return nil, err
	}

	al := ParseAddressList(body)
	h.setValue(name, al)

	return al, nil
}

// GetAddressList will return an addr.AddressList for the named field. This
// method works hard to avoid parse errors and tries to accept anything. As such
// a badly formatted address field might return a weird address value.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
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

// getAllAddressLists will return a slice of addr.AddressList for all headers
// with the given name or return an error.
func (h *Header) getAllAddressLists(name string) ([]addr.AddressList, error) {
	bs, err := h.GetAll(name)
	if err != nil {
		return nil, err
	}

	allAl := make([]addr.AddressList, 0, 10)
	for _, b := range bs {
		al := ParseAddressList(b)
		allAl = append(allAl, al)
	}

	h.setValue(name, allAl)

	return allAl, nil
}

// GetAllAddressLists will return a slice of addr.AddressList for all headers
// with the given name.
//
// This uses a very forgiving parser for email addresses, so it won't error on
// weird and wonky addresses, but do its best to return them, so you may get
// weird results from this.
//
// If the named field does not exist in the header, this will return nil with
// ErrNoSuchField.
func (h *Header) GetAllAddressLists(name string) ([]addr.AddressList, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getAllAddressLists(name)
	}

	als, isAddrLists := v.([]addr.AddressList)
	if !isAddrLists {
		return h.getAllAddressLists(name)
	}

	return als, nil
}

// getParamValue will parse a param.Value out of the given field or return an
// error.
func (h *Header) getParamValue(name string) (*param.Value, error) {
	body, err := h.Get(name)
	if err != nil {
		return nil, err
	}

	pv, err := param.Parse(body)
	if err != nil {
		return nil, err
	}

	h.setValue(name, pv)

	return pv, nil
}

// GetParamValue will return a param.Value for the header field matching the
// given name.
//
// This will return an error if it is unable to parse a param.Value. This will
// ErrNoSuchField if no field with the given name is present. It will return
// ErrManyFields if more than one field with the given name is found.
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

// getKeywordsList will return keywords for all header fields with the given
// name or return an error.
func (h *Header) getKeywordsList(name string) ([]string, error) {
	bs, err := h.GetAll(name)
	if err != nil {
		return nil, err
	}

	allKs := make([]string, 0, len(bs)*2)
	for _, b := range bs {
		ks := strings.Split(b, ",")
		for _, k := range ks {
			nextK := strings.TrimSpace(k)
			if nextK != "" {
				allKs = append(allKs, nextK)
			}
		}
	}

	h.setValue(name, allKs)

	return allKs, nil
}

// GetKeywordsList will return a list of strings representing all the keywords
// set on the named header. These are formatted as for the Keywords header, but
// this is the generic method that allows for treating other headers as
// Keywords. There can be zero or more Keywords headers. Each header is, then,
// a comma-separated list of Keywords. This will collect those values from all
// the headers with the given name and return them.
//
// This method will return nil with ErrNoSuchField if the named field does not
// exist.
func (h *Header) GetKeywordsList(name string) ([]string, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getKeywordsList(name)
	}

	ks, isStringSlice := v.([]string)
	if !isStringSlice {
		return h.getKeywordsList(name)
	}

	return ks, nil
}

// getAll fetches all the header field bodies for fields with the given
// name or returns an error if there are no such fields.
func (h *Header) getAll(name string) ([]string, error) {
	fs := h.GetAllFieldsNamed(name)
	if len(fs) == 0 {
		return nil, ErrNoSuchField
	}

	bs := make([]string, len(fs))
	for i, f := range fs {
		bs[i] = f.Body()
	}

	h.setValue(name, bs)

	return bs, nil
}

// GetAll fetches all the header field bodies for fields with the given
// name and returns them as a slice of strings.
//
// It returns nil with ErrNoSuchField if no field with the given name is set on
// the header.
func (h *Header) GetAll(name string) ([]string, error) {
	v, found := h.getValue(name)
	if !found {
		return h.getAll(name)
	}

	ss, isStringSlice := v.([]string)
	if !isStringSlice {
		return h.getAll(name)
	}

	return ss, nil
}

// SetAll replaces all the header fields with the given name with the
// bodies given. After a successful completion of this method, the field with
// the given name will occur exactly len(bodies) times in the header. If the
// field is already present in the header, existing fields will have their
// bodies replaced with the new values. Any new fields will be appended to the
// end of the header.
func (h *Header) SetAll(name string, bodies ...string) {
	ixs := h.GetIndexesNamed(name)

	for i, b := range bodies {
		if i < len(ixs) {
			// Replace existing Comments
			f := h.GetField(ixs[i])
			f.SetBody(b)
			continue
		}

		// Append more Comments
		h.InsertBeforeField(h.Len(), name, b)
	}

	if len(ixs) > len(bodies) {
		// Delete extra Comments
		for i := len(ixs) - 1; i >= len(bodies); i-- {
			_ = h.DeleteField(ixs[i])
		}
	}
}

// SetKeywordsList will replace all Keywords headers currently set in the
// header with one Keywords header with all the given keywords separated by
// a comma.
func (h *Header) SetKeywordsList(name string, keywords ...string) {
	h.setValue(name, keywords)
	bodyStr := strings.Join(keywords, ", ")
	h.Set(name, bodyStr)
}

// Set will replace all existing header fields with the given name with a single
// header field with the given name and body. If the field already exists on the
// header, then the first occurrence will be replaced with this value and any
// other values will be deleted. If the field does not exist, it will be
// appended to the end of the header.
//
// The procedure for replacing above is used to fall the Set* methods that
// replace all fields with a single field.
func (h *Header) Set(name, body string) {
	// Check for existing fields
	ixs := h.GetIndexesNamed(name)

	// if none, insert the new field and we're done
	if len(ixs) == 0 {
		h.InsertBeforeField(h.Len(), name, body)
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

// SetTime will replace all existing header fields with the given name with a
// single header field with the given name and time. The time will be formatted
// via time.RFC1123Z.
func (h *Header) SetTime(name string, body time.Time) {
	h.setValue(name, body)
	bodyStr := body.Format(time.RFC1123Z)
	h.Set(name, bodyStr)
}

// SetAddressList will replace all existing header fields with the given name
// with a single header containing the given addr.AddressList.
func (h *Header) SetAddressList(name string, body ...addr.Address) {
	h.setValue(name, body)
	bodyStr := addr.AddressList(body).String()
	h.Set(name, bodyStr)
}

// SetAllAddressLists will replace all existing header fields with a new set
// of header fields from the given slice of addr.AddressList.
func (h *Header) SetAllAddressLists(name string, bodies ...addr.AddressList) {
	h.setValue(name, bodies)
	strs := make([]string, len(bodies))
	for i, body := range bodies {
		strs[i] = body.String()
	}
	h.SetAll(name, strs...)
}

// SetParamValue will replace all existing header fields with the given name
// with a single param.Value header containing the given param.Value.
func (h *Header) SetParamValue(name string, body *param.Value) {
	h.setValue(name, body)
	bodyStr := body.String()
	h.Set(name, bodyStr)
}

// getParamValueValue reads the primary value of the param.Value header or
// returns an error.
func (h *Header) getParamValueValue(name string) (string, error) {
	pv, err := h.GetParamValue(name)
	if err != nil {
		return "", err
	}

	return pv.Value(), nil
}

// setParamValueValue sets the primary value of the param.Value header.
func (h *Header) setParamValueValue(name, v string) {
	// Before we start, let's make sure we cannot get an ErrManyFields first...
	ixs := h.GetIndexesNamed(name)
	for i := len(ixs) - 1; i > 0; i++ {
		_ = h.DeleteField(ixs[i])
	}

	// Then, pull the param.Value
	pv, err := h.GetParamValue(name)
	if err != nil {
		// we got an error, just overwrite the whole header
		pv = param.New(v)
	} else {
		// preserve everything else and update
		pv = param.Modify(pv, param.Change(v))
	}

	// and replace
	h.SetParamValue(name, pv)
}

// getParamValueParam gets a parameter value of the param.Value header or
// returns an error.
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

// setParamValueParam sets a parameter value of the param.Value header or
// returns an error. The header must already exist before calling this
// method.
func (h *Header) setParamValueParam(name, p, v string) error {
	pv, err := h.GetParamValue(name)
	if err != nil {
		return err
	}

	newPv := param.Modify(pv, param.Set(p, v))
	h.SetParamValue(name, newPv)

	return nil
}

// GetContentType returns the Content-type header as a param.Value.
//
// It returns nil and ErrNoSuchField if the field is not set on the header. It
// returns nil and ErrManyFields if the field is set more than once on the
// header. It will return nil and an error if there is a problem parsing the
// param.Value.
func (h *Header) GetContentType() (*param.Value, error) {
	return h.GetParamValue(ContentType)
}

// SetContentType replaces the Content-type with the given param.Value.
func (h *Header) SetContentType(v *param.Value) {
	h.SetParamValue(ContentType, v)
}

// GetMediaType returns the MIME type set in the Content-type header (other
// parameters will not be returned).
//
// It returns nil and ErrNoSuchField if the field is not set on the header. It
// returns nil and ErrManyFields fi the field is set more than once on the
// header. It will return nil and an error if there is a problem parsing the
// media type information out of the header.
func (h *Header) GetMediaType() (string, error) {
	return h.getParamValueValue(ContentType)
}

// SetMediaType replaces the MIME type on the Content-type header, creating it
// if it has not been set yet. If the Content-type header already exists, any
// other parameters already set will be preserved. If this header is set
// multiple times (in violation of RFC 5322), it will remove all but the first
// instance and replace the media type of the first instance.
func (h *Header) SetMediaType(mt string) {
	h.setParamValueValue(ContentType, mt)
}

// GetCharset gets the charset from the Content-type header field.
//
// This method returns an empty string with ErrNoSuchField if no field is
// present in the header. This method returns an empty string with
// ErrNoSuchFieldParameter if the field is present, but the parameter is not set
// on the field. This method returns an empty string with ErrManyFields if
// the field is set more than once on the header. This method returns an empty
// string and an error if the parameter values cannot be parsed out of the
// field for some reason.
func (h *Header) GetCharset() (string, error) {
	return h.getParamValueParam(ContentType, param.Charset)
}

// SetCharset sets the charset on the Content-type header.
//
// This method fails with a ErrNoSuchField if the field is not set on the
// header. This method fails with an error if the parameter values cannot be
// parsed out of the field for some reason.
func (h *Header) SetCharset(c string) error {
	return h.setParamValueParam(ContentType, param.Charset, c)
}

// GetBoundary gets the boundary from the Content-type header field.
//
// This method returns an empty string with ErrNoSuchField if no field is
// present in the header. This method returns an empty string with
// ErrNoSuchFieldParameter if the field is present, but the parameter is not set
// on the field. This method returns an empty string with ErrManyFields if
// the field is set more than once on the header. This method returns an empty
// string and an error if the parameter values cannot be parsed out of the
// field for some reason.
func (h *Header) GetBoundary() (string, error) {
	return h.getParamValueParam(ContentType, param.Boundary)
}

// SetBoundary sets the boundary on the Content-type header.
//
// This method fails with a ErrNoSuchField if the field is not set on the
// header. This method fails with an error if the parameter values cannot be
// parsed out of the field for some reason.
func (h *Header) SetBoundary(b string) error {
	return h.setParamValueParam(ContentType, param.Boundary, b)
}

// GetContentDisposition returns the Content-disposition header as a
// param.Value.
//
// It returns nil and ErrNoSuchField if the field is not set on the header. It
// returns nil and ErrManyFields if the field is set more than once on the
// header. It will return nil and an error if there is a problem parsing the
// param.Value.
func (h *Header) GetContentDisposition() (*param.Value, error) {
	return h.GetParamValue(ContentDisposition)
}

// SetContentDisposition sets the Content-disposition to a new value from a
// param.Value.
func (h *Header) SetContentDisposition(v *param.Value) {
	h.SetParamValue(ContentDisposition, v)
}

// GetPresentation returns the primary value of the Content-disposition
// header, describing what the function of this part of the message is.
//
// It returns nil and ErrNoSuchField if the field is not set on the header. It
// returns nil and ErrManyFields fi the field is set more than once on the
// header. It will return nil and an error if there is a problem parsing the
// presentation information out of the header.
func (h *Header) GetPresentation() (string, error) {
	return h.getParamValueValue(ContentDisposition)
}

// SetPresentation sets the disposition value of the Content-disposition
// header field. If the Content-disposition header already exists, any other
// parameters already set will be preserved. If this header is set multiple
// times (in violation of RFC 5322), it will remove all but the first instance
// and replace the presentation of the first instance.
func (h *Header) SetPresentation(d string) {
	h.setParamValueValue(ContentDisposition, d)
}

// GetFilename gets the filename parameter of the Content-disposition header.
//
// This method returns an empty string with ErrNoSuchField if no field is
// present in the header. This method returns an empty string with
// ErrNoSuchFieldParameter if the field is present, but the parameter is not set
// on the field. This method returns an empty string with ErrManyFields if
// the field is set more than once on the header. This method returns an empty
// string and an error if the parameter values cannot be parsed out of the
// field for some reason.
func (h *Header) GetFilename() (string, error) {
	return h.getParamValueParam(ContentDisposition, param.Filename)
}

// SetFilename sets the filename parameter of the Content-disposition header.
//
// This method fails with a ErrNoSuchField if the field is not set on the
// header. This method fails with an error if the parameter values cannot be
// parsed out of the field for some reason.
func (h *Header) SetFilename(f string) error {
	return h.setParamValueParam(ContentDisposition, param.Filename, f)
}

// GetDate retrieves the Date header as a time.Time value.
//
// It will return an error if it is unable to parse the time value from the Date
// header. It will return the zero value and ErrNoSuchField if the header does
// not exist. It will return the zero value and ErrManyFields if more than one
// Date field is set on the header.
func (h *Header) GetDate() (time.Time, error) {
	return h.GetTime(Date)
}

// SetDate updates the Date header from the given time.Time value.
func (h *Header) SetDate(d time.Time) {
	h.SetTime(Date, d)
}

// GetSubject returns the value of the Subject header field.
//
// If Subject is not set in the header, it will return an empty string with
// ErrNoSuchField. If there are multiple Subject headers, it will return
// ErrManyFields.
func (h *Header) GetSubject() (string, error) {
	return h.Get(Subject)
}

// SetSubject replaces the Subject header field.
func (h *Header) SetSubject(s string) {
	h.Set(Subject, s)
}

// setAddress allows the setting of an address field either from a string or
// from an address list or fails with an error.
func (h *Header) setAddress(n string, as []any) error {
	var al addr.AddressList
	for _, a := range as {
		switch v := a.(type) {
		case string:
			var err error
			add, err := addr.ParseEmailAddress(v)
			if err != nil {
				return err
			}
			al = append(al, add)
		case addr.Address:
			al = append(al, v)
		default:
			return ErrWrongAddressType
		}
	}
	h.SetAddressList(n, al...)
	return nil
}

// GetTo returns the To address field as an addr.AddressList.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetTo() (addr.AddressList, error) {
	return h.GetAddressList(To)
}

// SetTo sets the To address field with either an addr.AddressList or a string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetTo(a ...any) error {
	return h.setAddress(To, a)
}

// GetCc returns the Cc address field as an addr.AddressList.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetCc() (addr.AddressList, error) {
	return h.GetAddressList(Cc)
}

// SetCc sets the Cc address field with either an addr.AddressList or a string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetCc(a ...any) error {
	return h.setAddress(Cc, a)
}

// GetBcc returns the Bcc address field as an addr.AddressList.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetBcc() (addr.AddressList, error) {
	return h.GetAddressList(Bcc)
}

// SetBcc sets the Bcc address field with either an addr.AddressList or a
// string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetBcc(a ...any) error {
	return h.setAddress(Bcc, a)
}

// GetFrom returns the From address field as an addr.AddressList.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetFrom() (addr.AddressList, error) {
	return h.GetAddressList(From)
}

// SetFrom sets the From address field with either an addr.AddressList or a
// string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetFrom(a ...any) error {
	return h.setAddress(From, a)
}

// GetReplyTo returns the ReplyTo address field as an addr.AddressList.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetReplyTo() (addr.AddressList, error) {
	return h.GetAddressList(ReplyTo)
}

// SetReplyTo sets the ReplyTo address field with either an addr.AddressList or
// a string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetReplyTo(a ...any) error {
	return h.setAddress(ReplyTo, a)
}

// GetKeywords returns all the keywords set on all the Keywords fields.
//
// This method will return nil with ErrNoSuchField if the Keywords field does
// not exist.
func (h *Header) GetKeywords() ([]string, error) {
	return h.GetKeywordsList(Keywords)
}

// SetKeywords sets keywords on the Keywords header.
func (h *Header) SetKeywords(ks ...string) {
	h.SetKeywordsList(Keywords, ks...)
}

// GetComments returns the content of the Comments header fields.
func (h *Header) GetComments() ([]string, error) {
	return h.GetAll(Comments)
}

// SetComments replaces all Comments fields with the given bodies.
func (h *Header) SetComments(cs ...string) {
	h.SetAll(Comments, cs...)
}

// GetReferences returns the message ID in the References header, if any.
//
// If References is not set in the header, it will return an empty string with
// ErrNoSuchField. If there are multiple References headers, it will return
// ErrManyFields.
func (h *Header) GetReferences() (string, error) {
	return h.Get(References)
}

// SetReferences sets the message ID to store in the References header.
func (h *Header) SetReferences(ref string) {
	h.Set(References, ref)
}

// GetInReplyTo returns the message ID in the In-reply-to header, if any.
//
// If In-reply-to is not set in the header, it will return an empty string with
// ErrNoSuchField. If there are multiple In-reply-to headers, it will return
// ErrManyFields.
func (h *Header) GetInReplyTo() (string, error) {
	return h.Get(InReplyTo)
}

// SetInReplyTo returns the message ID in the In-reply-to header.
func (h *Header) SetInReplyTo(ref string) {
	h.Set(InReplyTo, ref)
}

// GetMessageID returns the Message ID found in the Message-id header, if any.
//
// If Message-id is not set in the header, it will return an empty string with
// ErrNoSuchField. If there are multiple Message-id headers, it will return
// ErrManyFields.
func (h *Header) GetMessageID() (string, error) {
	return h.Get(MessageID)
}

// SetMessageID sets the Message-ID header of the message header.
func (h *Header) SetMessageID(ref string) {
	h.Set(MessageID, ref)
}

// GetSender returns the address list in the Sender header, if any.
//
// It will return nil and ErrNoSuchField if the field is not set on the header.
// It will return ErrManyFields if the field is set more than once on the
// header.
func (h *Header) GetSender() (addr.AddressList, error) {
	return h.GetAddressList(Sender)
}

// SetSender sets the Sender address field with either an addr.AddressList or
// a string.
//
// It will fail with an error returned if something other than those types is
// provided or if the given string fails to strictly parse.
func (h *Header) SetSender(a ...any) error {
	return h.setAddress(Sender, a)
}

// GetTransferEncoding returns the content of the Content-transfer-encoding
// header.
//
// It will return ErrNoSuchField if the header is not set. it will return
// ErrManyFields if the field is set more than once.
func (h *Header) GetTransferEncoding() (string, error) {
	return h.Get(ContentTransferEncoding)
}

// SetTransferEncoding replaces the Content-transfer-encoding with the given
// value.
func (h *Header) SetTransferEncoding(b string) {
	h.Set(ContentTransferEncoding, b)
}

// TODO Add support for resent blocks

// TODO Add support for trace fields (Return-Path and Received)

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
			switch {
			case c == '(':
				nestLevel++
				if nestLevel == 1 {
					continue
				} else {
					comment.WriteRune(c)
				}
			case c == ')':
				nestLevel--
				switch {
				case nestLevel == 0:
					continue
				case nestLevel < 0:
					nestLevel = 0
					clean.WriteRune(c)
				default:
					comment.WriteRune(c)
				}
			case nestLevel > 0:
				comment.WriteRune(c)
			default:
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
		switch {
		case len(parts) == 0:
			email = ""
		case len(parts) > 1:
			dn = strings.Join(parts[:len(parts)-1], " ")
			email = parts[len(parts)-1]
		default:
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
