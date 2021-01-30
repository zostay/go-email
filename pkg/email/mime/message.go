package mime

import (
	"errors"
	"mime"
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

// ContentType represents a parsed Content-type header.
type contentType struct {
	mediaType string            // the content-type itself
	params    map[string]string // additional content-type parameters, like charset, boundary, etc.
}

// Message represents a MIME message.
type Message struct {
	email.Message     // basic message within
	MaxDepth      int // maximum depth permitted for subparts during parsing
	prefix        []byte
	boundary      string
	Preamble      []byte     // preamble before MIME parts
	Parts         []*Message // the MIME sub-parts
	Epilogue      []byte     // epilogue after MIME parts
}

// UpdateBody will reconstruct the basic message whenever the higher level
// elements are adjusted, preserving the original byte-for-byte as much as
// possible.
//
// Whenever changes are made to the sub-parts, header, or other parts of the
// body, this method must be called prior to writing the message to output. This
// will recursively call UpdateBody on all sub-parts.
//
// When this method is called, it will also check to see if the boundary between
// parts has changed. If so, it will update the boundaries between MIME parts.
// This can trigger an error if the boundary is missing or contains a space. In
// that case an error will be returned.
//
// If an error occurs, the body will not have been updated. It is possible that
// some of the sub-parts will have had their body updated.
func (m *Message) UpdateBody() error {
	// Ruh-roh, we need to rewrite the boundary
	if len(m.Parts) > 0 && m.boundary != m.Boundary() {
		nb := m.Boundary()
		if len(nb) == 0 {
			return errors.New("no boundary set")
		} else if strings.Contains(nb, " \t") {
			return errors.New("boundary contains whitesace")
		}

		lb := string(m.Break())
		for _, p := range m.Parts {
			p.prefix = []byte(lb + "--" + nb + lb)
		}

		m.boundary = nb
	}

	var a strings.Builder
	a.Write(m.prefix)
	a.Write(m.Preamble)
	for _, p := range m.Parts {
		err := p.UpdateBody()
		if err != nil {
			return err
		}
		a.Write(p.Body())
	}
	a.Write(m.Epilogue)

	m.SetBodyString(a.String())

	return nil
}

// RawContentType is shorthand for
//  m.Get("Content-type")
func (m *Message) RawContentType() string {
	return m.HeaderGet("Content-type")
}

func (m *Message) structuredContentType() (*contentType, error) {
	const CTCK = "github.com/zostay/go-email/pkg/email/mime.ContentType"

	// header set
	hf := m.HeaderGetField("Content-type")
	if hf == nil {
		return nil, nil
	}

	// parsed content type cached on header?
	cti := hf.CacheGet(CTCK)
	var ct *contentType
	if ct, ok := cti.(*contentType); ok {
		return ct, nil
	}

	// still nothing? parse the content type
	if ct == nil {
		mt, ps, err := mime.ParseMediaType(hf.Body())
		if err != nil {
			return nil, err
		}

		ct = &contentType{mt, ps}
		hf.CacheSet(CTCK, ct)
	}

	return ct, nil
}

// ContentType retrieves only the media-type of the Content-type header (i.e.,
// the parameters are stripped.)
func (m *Message) ContentType() string {
	ct, _ := m.structuredContentType()
	return ct.mediaType
}

// Charset retrieves the character set on the Content-type header or an empty
// string.
func (m *Message) Charset() string {
	ct, _ := m.structuredContentType()
	return ct.params["charset"]
}

// Boundary is the boundary set on the Content-type header for multipart
// messages.
func (m *Message) Boundary() string {
	ct, _ := m.structuredContentType()
	return ct.params["boundary"]
}

// BodyUnicode is for retrieving a MIME single part body after having the
// transfer encoding decoded and any charsets decoded into Go's native unicode
// handling. If the message is multipart, it returns an empty string with no
// error. If there is an error decoding the transfer encoding or converting to
// unicode, an empty string is returned with an error.
func (m *Message) BodyUnicode() (string, error) {
	bb, err := m.BodyBinary()
	if err != nil {
		return "", err
	}

	bs, err := CharsetDecoder(m, bb)
	if err != nil {
		return "", err
	}

	return bs, nil
}

// BodyBinary is for retrieving a MIME single part body after having the
// transfer encoding decoded. No charset handling will be performed. If this is
// a multipart body, a nil slice is returned with a nil error. If an error
// occurs decoding the transfer encoding, a nil slice is returned with an
// error.
func (m *Message) BodyBinary() ([]byte, error) {
	if len(m.Parts) > 0 {
		return nil, errors.New("cannot treat multipart MIME message as single part")
	}

	cte := m.HeaderGet("Content-transfer-encoding")
	td, _ := SelectTransferDecoder(cte)
	decode := td.From
	return decode(m.Body())
}

// SetBodyUnicode replaces the MIME message body with the given unicode string.
// This method performs actions based on the current state of the Content-type
// and Content-transfer-encoding headers. You must set those as desired before
// calling this method.
//
// The given string will be encoded from Go's native unicode into the
// destination charset, as specified by the Content-type header. After this, the
// Content-transfer-encoding will be applied to transform the body to that
// encoding (if any).
//
// If the body was previously a multipart message, this will also clear the
// Preamble, Parts, and Epilogue.
//
// This method returns an error and won't make any changes to the message if an
// error occurs either with the transfer encoding or the character set encoding.
func (m *Message) SetBodyUnicode(s string) error {
	if len(m.Parts) > 0 {
		return errors.New("cannot treat multipart MIME message as single part")
	}

	eb, err := CharsetEncoder(m, s)
	if err != nil {
		return err
	}

	cte := m.HeaderGet("Content-transfer-encoding")
	td, _ := SelectTransferDecoder(cte)
	decode := td.To

	bb, err := decode(eb)
	if err != nil {
		return err
	}

	m.SetBody(bb)
	return nil
}

// SetBodyBinary replaces the MIME message body with the given bytes. This
// method performs actions based on teh current state of the
// Content-transfer-encoding header. You must set that header as desired before
// calling this method.
//
// The given bytes will be transformed according to Content-transfer-encoding
// header (if any).
//
// If the body was previously a multipart message, this will also clear the
// Preamble, Parts, and Epilog.
//
// This method returns ane error and won't make any changes to the message if an
// error occurs with the transfer encoding.
func (m *Message) SetBodyBinary(b []byte) error {
	if len(m.Parts) > 0 {
		return errors.New("cannot treat multipart MIME message as single part")
	}

	cte := m.HeaderGet("Content-transfer-encoding")
	td, _ := SelectTransferDecoder(cte)
	decode := td.To

	b, err := decode(b)
	if err != nil {
		return err
	}

	m.SetBody(b)
	return nil
}
