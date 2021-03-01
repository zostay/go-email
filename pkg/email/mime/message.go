// Package mime provides tools for parsing and processing MIME email messages.
// If you are just looking to work with the primary headers, you might prefer to
// use the parse in "github.com/zostay/go-email/pkg/email/simple".
//
// This provides tools for reading the headers and bodies of subparts and the
// header and messages of those subparts and so on. This will also perform email
// decoding for base64, quotedprintable, and charset encodings.
package mime

import (
	"errors"
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

// Message represents a MIME message.
type Message struct {
	Header         // MIME email header within
	email.Body     // basic message body within
	MaxDepth   int // maximum depth permitted for subparts during parsing
	prefix     []byte
	boundary   string
	Preamble   []byte     // preamble before MIME parts
	Parts      []*Message // the MIME sub-parts
	Epilogue   []byte     // epilogue after MIME parts
}

// NewMessage will create a message with the selected boundary. The header and
// body will be empty, so it won't really be a legal message yet. This does not
// check to make sure the boundary is sane. You will also need to set this
// boundary on the Content-Type header if you want to actually have multiple
// parts. If you aren't going to have multiple parts, you can safely set the
// boundary to an empty string.
func NewMessage(boundary string) *Message {
	return &Message{
		boundary: boundary,
		Parts:    make([]*Message, 0),
	}
}

// InsertPart will attach a MIME message to this message at the specified point.
// Inserting at 0 will make it the first part. Using a negative index will make
// it the last. Using a value greater than or equal to the length of Parts will
// also insert it as the last.
func (m *Message) InsertPart(ix int, p *Message) {
	if ix < -1 || ix >= len(m.Parts) {
		m.Parts = append(m.Parts, p)
	}

	lb := m.Break()
	p.prefix = append(lb, []byte("--")...)
	p.prefix = append(p.prefix, []byte(m.boundary)...)
	p.prefix = append(p.prefix, lb...)

	bp := m.Parts[:ix]
	ap := m.Parts[ix:]
	m.Parts = append(bp, p)
	m.Parts = append(m.Parts, ap...)
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
	nb := m.HeaderContentTypeBoundary()
	if len(m.Parts) > 0 && m.boundary != nb {
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
		a.Write(p.Content())
	}
	a.Write(m.Epilogue)

	m.SetContentString(a.String())

	return nil
}

// ContentTransferEncoding is just an alias for:
//  cte := strings.ToLower(m.HeaderGet("Content-transfer-encoding"))
// It returns the transfer encoding used for the message body. For
// MIME-compliant messages, this should always be one of the following values:
//  7bit
//  8bit
//  binary
//  base64
//  quoted-printable
// Occassionally, you will see other oddball values, of course.
func (m *Message) ContentTransferEncoding() string {
	return strings.ToLower(m.HeaderGet("Content-transfer-encoding"))
}

// SetContentTransferEncoding updates the transfer encoding for the message and
// then it will rewrite the content to adhere to the new encoding. It willr
// return an error if there's a problem decoding or re-encoding the content
// onthe way. If an error is returned, the Content-transfer-encoding will have
// remain at its original value.
func (m *Message) SetContentTransferEncoding(cte string) error {
	oldCte := m.ContentTransferEncoding()
	if strings.ToLower(cte) == oldCte {
		return nil
	}

	decodedContent, err := m.ContentBinary()
	if err != nil {
		return err
	}

	_ = m.HeaderSet("Content-transfer-encoding", cte)

	err = m.SetContentBinary(decodedContent)
	if err != nil {
		_ = m.HeaderSet("Content-transfer-encoding", oldCte)
		return err
	}

	return nil
}

// ContentUnicode is for retrieving a MIME single part body after having the
// transfer encoding decoded and any charsets decoded into Go's native unicode
// handling. If the message is multipart, it returns an empty string with no
// error. If there is an error decoding the transfer encoding or converting to
// unicode, an empty string is returned with an error.
func (m *Message) ContentUnicode() (string, error) {
	bb, err := m.ContentBinary()
	if err != nil {
		return "", err
	}

	bs, err := CharsetDecoder(m.HeaderContentTypeCharset(), bb)
	if err != nil {
		return "", err
	}

	return bs, nil
}

// ContentBinary is for retrieving a MIME single part body after having the
// transfer encoding decoded. No charset handling will be performed. If this is
// a multipart body, a nil slice is returned with a nil error. If an error
// occurs decoding the transfer encoding, a nil slice is returned with an
// error.
func (m *Message) ContentBinary() ([]byte, error) {
	if len(m.Parts) > 0 {
		return nil, errors.New("cannot treat multipart MIME message as single part")
	}

	cte := m.HeaderGet("Content-transfer-encoding")
	td, _ := SelectTransferDecoder(cte)
	decode := td.From
	return decode(m.Content())
}

// SetContentUnicode replaces the MIME message body with the given unicode string.
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
func (m *Message) SetContentUnicode(s string) error {
	if len(m.Parts) > 0 {
		return errors.New("cannot treat multipart MIME message as single part")
	}

	eb, err := CharsetEncoder(m.HeaderContentTypeCharset(), s)
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

	m.SetContent(bb)
	return nil
}

// SetContentBinary replaces the MIME message body with the given bytes. This
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
func (m *Message) SetContentBinary(b []byte) error {
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

	m.SetContent(b)
	return nil
}

// String outputs the message as a string.
func (m *Message) String() string {
	return m.Header.String() + string(m.Break()) + m.Body.String()
}

// Bytes outputs the messages as a slice of bytes.
func (m *Message) Bytes() []byte {
	msg := append(m.Header.Bytes(), m.Break()...)
	return append(msg, m.Body.Bytes()...)
}
