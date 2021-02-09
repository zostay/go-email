package mime

import (
	"errors"
	"fmt"
	"time"

	"github.com/zostay/go-addr/pkg/addr"
	"github.com/zostay/go-email/pkg/email"
	"github.com/zostay/go-email/pkg/email/simple"
)

const mtck = "github.com/zostay/go-email/pkg/email/mime.mediaType"

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
			err := h.headerSet(n, v)
			if err != nil {
				return nil, err
			}
		}
	}

	return &h, nil
}

func (h *Header) headerSet(n string, v interface{}) (err error) {
	switch b := v.(type) {
	case string:
		err = h.HeaderSet(n, b)
	case []byte:
		err = h.HeaderSet(n, string(b))
	case time.Time:
		if n == "Date" {
			h.HeaderSetDate(b)
		} else {
			err = h.HeaderSet(n, b.Format(time.RFC1123Z))
		}
	case *MediaType:
		err = h.HeaderSetMediaType(n, b)
	case *addr.Mailbox:
		h.HeaderSetAddressList(n, addr.AddressList{b})
	case addr.Address:
		h.HeaderSetAddressList(n, addr.AddressList{b})
	case addr.MailboxList:
		h.HeaderSetAddressList(n, b.AddressList())
	case addr.AddressList:
		h.HeaderSetAddressList(n, b)
	case fmt.Stringer:
		err = h.HeaderSet(n, b.String())
	default:
		return errors.New("unsupported value type set on header field body")
	}

	return
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
	var err error
	if hf != nil {
		err = hf.SetBody(mt.String(), h.Break())
		if err != nil {
			return err
		}
	} else {
		hf, err = email.NewHeaderField(n, mt.String(), h.Break())
		if err != nil {
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
func (h *Header) HeaderSetContentType(mt string) (err error) {
	ct, _ := h.structuredMediaType("Content-type")
	if ct != nil {
		nct := NewMediaTypeMap(mt, ct.params)
		hf := h.HeaderGetField("Content-type")
		if hf != nil {
			err = hf.SetBody(nct.String(), h.Break())
			if err != nil {
				return
			}
			hf.CacheSet(mtck, nct)
		} else {
			err = h.HeaderSetMediaType("Content-type", nct)
			if err != nil {
				return
			}
		}
	} else {
		err = h.HeaderSet("Content-type", mt)
	}

	return
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
		err = hf.SetBody(nct.String(), h.Break())
		if err != nil {
			return err
		}
	} else {
		hf, err = email.NewHeaderField(n, nct.String(), h.Break())
		if err != nil {
			return err
		}

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
func (m *Message) HeaderContentDisposition() string {
	cd, _ := m.structuredMediaType("Content-disposition")
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
		var err error
		if hf != nil {
			err = hf.SetBody(ncd.String(), h.Break())
			if err != nil {
				return err
			}
		} else {
			hf, err = email.NewHeaderField("Content-disposition", ncd.String(), h.Break())
			if err != nil {
				return err
			}
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
