package mime

import (
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/pkg/email"
	"github.com/zostay/go-email/pkg/email/simple"
)

const (
	alck = "github.com/zostay/go-email/pkg/email/mime.AddressList"
	mtck = "github.com/zostay/go-email/pkg/email/mime.MediaType"
	dtck = "github.com/zostay/go-email/pkg/email/mime.Date"
)

// Header provides tools built on simple.Header to work with MIME headers.
type Header struct {
	simple.Header
}

// NewHeader will build a new MIME header. The arguments in the latter part are
// provided as name/value pairs. The names must be provided as strings. The
// values may be provided as string, []byte, time.Time, MediaType, addr.Mailbox,
// addr.Address, addr.MailboxList, addr.AddressList, or a fmt.Stringer.
//
// If one of the header arguments comes in as an unexpected object or with an
// odd length, this will return an error.
//
// If one of the header arguments includes a character illegal for use in a
// header, this will return an error.
//
// On success, it will return a constructed header object.
func NewHeader(lb string, hs ...interface{}) (*Header, error) {
	if len(hs)%2 != 0 {
		return nil, errors.New("header field name provided with no body value")
	}

	sh, err := simple.NewHeader(lb)
	if err != nil {
		return nil, err
	}
	sh.SetBody = defaultBodySetter
	h := Header{*sh}

	var n string
	for i, v := range hs {
		if i%2 == 0 {
			if name, ok := v.(string); ok {
				n = name
			} else {
				return nil, errors.New("header field name is not a string")
			}
		} else {
			err := h.HeaderSet(n, v)
			if err != nil {
				return nil, err
			}
		}
	}

	return &h, nil
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

func defaultBodySetter(hf *email.HeaderField, v interface{}, lb []byte) error {
	var sb string
	switch b := v.(type) {
	case string:
		sb = b
	case []byte:
		sb = string(b)
	case time.Time:
		sb = b.Format(time.RFC1123Z)
	case fmt.Stringer:
		sb = b.String()
	default:
		return errors.New("unsupported value type set on header field body")
	}

	var bb []byte
	if strings.IndexFunc(sb, forbiddenBodyChars) > -1 {
		bb = []byte(mime.BEncoding.Encode("utf-8", sb))
	}

	hf.SetBodyEncoded(sb, bb, lb)
	return nil
}

func (h *Header) structuredMediaType(n string) (*MediaType, error) {
	// header set
	hf := h.HeaderGetField(n)
	if hf == nil {
		return nil, nil
	}

	// parsed content type cached on header?
	mti := hf.CacheGet(mtck)
	var mt *MediaType
	if mt, ok := mti.(*MediaType); ok {
		return mt, nil
	}

	// still nothing? parse the content type
	if mt == nil {
		var err error
		mt, err = ParseMediaType(hf.Body())
		if err != nil {
			return nil, err
		}

		hf.CacheSet(mtck, mt)
	}

	return mt, nil
}

// HeaderGetMediaType retrieves a MediaType object for the named header. Returns
// an error if the given header cannot be parsed into a MediaType object.
// Returns nil and no error if the header is not set.
func (h *Header) HeaderGetMediaType(n string) (*MediaType, error) {
	mt, err := h.structuredMediaType(n)
	if err != nil {
		return nil, err
	}
	return mt, nil
}

// HeaderSetMediaType sets the named header to the given MediaType object.
func (h *Header) HeaderSetMediaType(n string, mt *MediaType) error {
	hf := h.HeaderGetField(n)
	if hf != nil {
		if err := h.SetBody(hf, mt, h.Break()); err != nil {
			return err
		}
	} else {
		hf = email.NewHeaderField(n, "", h.Break())
		if err := h.SetBody(hf, mt, h.Break()); err != nil {
			return err
		}
		h.Fields = append(h.Fields, hf)
	}
	hf.CacheSet(mtck, mt)
	return nil
}

// HeaderContentType retrieves only the full MIME type set in the Content-type
// header.
func (h *Header) HeaderContentType() string {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		return ct.mediaType
	}
	return ""
}

// HeaderSetContentType allows you to modify just the media-type of the
// Content-type header. The parameters will remain unchanged.
//
// If the existing Content-type header cannot be parsed for some reason, setting
// this value will replace the entire value with this MIME-type.
func (h *Header) HeaderSetContentType(mt string) error {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		nct := NewMediaTypeMap(mt, ct.params)
		hf := h.HeaderGetField("Content-type")
		if hf != nil {
			if err := h.SetBody(hf, nct, h.Break()); err != nil {
				return err
			}
			hf.CacheSet(mtck, nct)
		} else if err := h.HeaderSetMediaType("Content-type", nct); err != nil {
			return err
		}
	} else if err := h.HeaderSet("Content-type", mt); err != nil {
		return err
	}

	return nil
}

// HeaderContentTypeType retrieves the first part, the type, of the MIME type set in
// the Content-Type header.
func (h *Header) HeaderContentTypeType() string {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		return ct.Type()
	}
	return ""
}

// HeaderContentTypeSubtype retrieves the second part, the subtype, of the MIME type
// set in the Content-type header.
func (h *Header) HeaderContentTypeSubtype() string {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		return ct.Subtype()
	}
	return ""
}

// HeaderContentTypeCharset retrieves the character set on the Content-type
// header or an empty string.
func (h *Header) HeaderContentTypeCharset() string {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		return ct.Charset()
	}
	return ""
}

func (h *Header) structuredParameterUpdate(n, pn, pv string) error {
	ct, err := h.structuredMediaType(n)
	if err != nil {
		return err
	}
	if ct == nil {
		return fmt.Errorf("cannot set %q when %q header field is not yet set", pn, n)
	}

	ct.params[pn] = pv
	nct := NewMediaTypeMap(ct.mediaType, ct.params)
	hf := h.HeaderGetField(n)
	if hf != nil {
		if err := h.SetBody(hf, nct, h.Break()); err != nil {
			return err
		}
	} else {
		hf = email.NewHeaderField(n, nct.String(), h.Break())
		h.Fields = append(h.Fields, hf)
	}
	hf.CacheSet(mtck, nct)

	return nil

}

// HeaderSetContentTypeCharset modifies the charset on the Content-type header.
//
// If no Content-type header has been set yet or the value set cannot be
// parsed, this returns an error.
func (h *Header) HeaderSetContentTypeCharset(cs string) error {
	return h.structuredParameterUpdate("Content-type", "charset", cs)
}

// HeaderContentTypeBoundary is the boundary set on the Content-type header for
// multipart messages.
func (h *Header) HeaderContentTypeBoundary() string {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		return ct.Boundary()
	}
	return ""
}

// HeaderSetContentTypeBoundary updates the boundary string set in the
// Content-type header.
//
// If no Content-type header has been set yet or the value set cannot be
// parsed, this returns an error.
//
// Beware that this does not update the boundaries used in any associated
// message body, so if there are existing boundaries, you need to update those
// separately.
func (h *Header) HeaderSetContentTypeBoundary(b string) error {
	return h.structuredParameterUpdate("Content-type", "boundary", b)
}

// HeaderContentDisposition is the value of the Content-dispotion header value.
func (h *Header) HeaderContentDisposition() string {
	cd, _ := h.structuredMediaType("Content-disposition")
	if cd != nil {
		return cd.mediaType
	}
	return ""
}

// HeaderSetContentDisposition allows you to modify just the media-type of the
// Content-disposition header. The parameters will remain unchanged.
//
// If the existing Content-disposition header cannot be parsed for some reason, setting
// this value will replace the entire value with this media-type.
func (h *Header) HeaderSetContentDisposition(mt string) error {
	cd, _ := h.structuredMediaType("Content-disposition")
	if cd != nil {
		ncd := NewMediaTypeMap(mt, cd.params)
		hf := h.HeaderGetField("Content-disposition")
		if hf != nil {
			if err := h.SetBody(hf, ncd, h.Break()); err != nil {
				return err
			}
		} else {
			hf = email.NewHeaderField("Content-disposition", ncd.String(), h.Break())
		}
		hf.CacheSet(mtck, ncd)
	} else {
		var err = h.HeaderSet("Content-disposition", mt)
		if err != nil {
			return err
		}
	}
	return nil
}

// HeaderContentDispositionFilename is the filename set in the
// Content-disposition header, if set.  Otherwise, it returns an empty string.
func (m *Message) HeaderContentDispositionFilename() string {
	cd, _ := m.structuredMediaType("Content-disposition")
	if cd != nil {
		return cd.Filename()
	}
	return ""
}

// HeaderSetContentDispositionFilename updates the filename string set in the
// Content-disposition header.
//
// If no Content-disposition header has been set yet or the value set cannot be
// parsed, this returns an error.
func (h *Header) HeaderSetContentDispositionFilename(fn string) error {
	return h.structuredParameterUpdate("Content-disposition", "filename", fn)
}

// HeaderGetAddressList returns addresses for a header. If the header is not set or
// empty, it will return nil and no error. If the header has a value, but cannot
// be parsed as an address list, it will return nil and an error. If the header
// can be parsed as an email list, the email addresses will be returned.
//
// This only returns the addresses for the first occurence of a header, as the
// email address headers are only permitted a single time in email.
func (h *Header) HeaderGetAddressList(n string) (addr.AddressList, error) {
	hf := h.HeaderGetField(n)
	if hf == nil {
		return nil, nil
	}

	if addrs := hf.CacheGet(alck); addrs != nil {
		return addrs.(addr.AddressList), nil
	}

	addrs, err := addr.ParseEmailAddressList(string(hf.RawBody()))
	if err != nil {
		addrs = parseEmailAddressList(hf.Body())
	}

	hf.CacheSet(alck, addrs)
	return addrs, nil
}

// HeaderGetAllAddressLists handles address headers that have multiple header
// entries, such as Delivered-To. This will return an address list for all the
// headers as a single AddressList.
func (h *Header) HeaderGetAllAddressLists(n string) (addr.AddressList, error) {
	hfs := h.HeaderGetAllFields(n)
	allAddrs := make(addr.AddressList, 0)
	for _, hf := range hfs {
		if addrs := hf.CacheGet(alck); addrs != nil {
			allAddrs = append(allAddrs, addrs.(addr.AddressList)...)
		}

		addrs, err := addr.ParseEmailAddressList(string(hf.RawBody()))
		if err != nil {
			addrs = parseEmailAddressList(hf.Body())
		}

		hf.CacheSet(alck, addrs)
		allAddrs = append(allAddrs, addrs...)
	}

	return allAddrs, nil
}

// parseEmailAddressList is a fallback method for email address parsing. The
// parser in github.com/zostay/go-addr is a strict parser, which is useful for
// getting good accurate parsing of email addresses, but especially for
// validating data entry. However, when working with the mess that is the web
// when you want to get something useful, even if its technically wrong, well,
// this method an be used to clean up the mess.
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

// HeaderSetAddressList will update an address list header with the given
// address list.
func (h *Header) HeaderSetAddressList(n string, addrs addr.AddressList) {
	as := addrs.String()
	_ = h.HeaderSetAll(n, as)

	hf := h.HeaderGetField(n)
	hf.CacheSet(alck, addrs)
}

// HeaderGetDate parses and returns the date in the email. This will read the header
// named "Date". As this header is always required, it will return the time.Time
// zero value and an error if this method is called and no value is present. If
// the date header is present, it will returned the parsed value or an error if
// the date cannot be parsed.
func (h *Header) HeaderGetDate() (time.Time, error) {
	hf := h.HeaderGetField("Date")
	if hf == nil {
		return time.Time{}, nil
	}

	if date := hf.CacheGet(dtck); date != nil {
		return date.(time.Time), nil
	}

	date, err := mail.ParseDate(hf.Body())
	if err != nil {
		date, err = dateparse.ParseAny(hf.Body())
		if err != nil {
			return date, fmt.Errorf("unable to parse \"Date\" header value %q: %w", hf.Body(), err)
		}
	}

	hf.CacheSet(dtck, date)
	return date, nil
}

// HeaderSetDate takes a time.Time input and sets the header field named "Date"
// to an RFC5322 formatted date from that input. This uses the built-in RFC1123Z
// format.
func (h *Header) HeaderSetDate(d time.Time) {
	df := d.Format(time.RFC1123Z)
	_ = h.HeaderSetAll("Date", df)

	hf := h.HeaderGetField("Date")
	hf.CacheSet(dtck, d)
}
