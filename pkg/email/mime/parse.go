package mime

import (
	"bytes"
	"errors"
	"mime"
	"regexp"
	"strings"

	"github.com/zostay/go-email/pkg/email/simple"
)

// ParseError is returned when one or more errors occur while parsing an email
// message. It collects all the errors and returns them as a group.
type ParseError struct {
	Errs []error // the list of errors that occurred during parsing
}

// Error returns the list of errors encounted while parsing an email message.
func (err *ParseError) Error() string {
	errs := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errs[i] = e.Error()
	}
	return "error parsing MIME message: " + strings.Join(errs, ", ")
}

const (
	// MaxMultipartDepth is the default depth the parser will recurse into a
	// message.
	DefaultMaxMultipartDepth = 10
)

type option func(*Message)

// WithMaxDepth sets the maximum depth the parse is allowed to descend
// recursively within subparts. This value is saved as part of the object and
// future calls to FillParts will obey it. If this option is not passed, it will
// use DefaultMaxMultipartDepth.
func WithMaxDepth(d int) option {
	return func(m *Message) { m.MaxDepth = d }
}

// Parse parses the given bytes as an email message and returns the message
// object. As much of the message as can be parsed will be returned even if an
// error is returned.
//
// Options may be passed to modify the construction and parsing of the object.
func Parse(m []byte, o ...option) (*Message, error) {
	var mm *Message

	msg, err := simple.Parse(m)
	if msg != nil {
		h := Header{msg.Header}
		mm = &Message{
			Header:   h,
			Body:     msg.Body,
			MaxDepth: DefaultMaxMultipartDepth,
		}
	}

	for _, opt := range o {
		opt(mm)
	}

	if err != nil {
		return mm, err
	}

	derr := mm.DecodeHeader()
	ferr := mm.FillParts()
	var pderr, pferr *ParseError
	if errors.As(derr, &pderr) && errors.As(ferr, &pferr) {
		errs := append(pderr.Errs, pferr.Errs...)
		return mm, &ParseError{errs}
	} else if ferr != nil {
		return mm, derr
	} else if derr != nil {
		return mm, ferr
	}

	return mm, nil
}

// DecodeHeader scans through the headers and looks for MIME word encoded field
// values. When they are found, these are decoded into native unicode.
func (m *Message) DecodeHeader() error {
	dec := &mime.WordDecoder{
		CharsetReader: CharsetDecoderToCharsetReader(CharsetDecoder),
	}
	errs := make([]error, 0)
	for _, hf := range m.Fields {
		if strings.Contains(hf.Body(), "=?") {
			dv, err := dec.Decode(hf.Body())
			if err != nil {
				errs = append(errs, err)
			}

			hf.SetBodyEncoded(dv, []byte(hf.Body()), m.Break())
		}
	}

	if len(errs) > 0 {
		return &ParseError{errs}
	} else {
		return nil
	}
}

// FillParts performs the work of parsing the message body into preamble,
// sub-parts, and epilogue.
func (m *Message) FillParts() error {
	m.Preamble = nil
	m.Parts = nil
	m.Epilogue = nil

	mtt := m.HeaderContentTypeType()
	if mtt == "multipart" || mtt == "message" {
		return m.fillPartsMultiPart()
	} else {
		return m.fillPartsSinglePart()
	}

}

func (m *Message) boundaries(body []byte, boundary string) []int {
	lbq := regexp.QuoteMeta(string(m.Break()))
	bq := regexp.QuoteMeta(boundary)
	bmre := regexp.MustCompile("(?:^|" + lbq + ")--" + bq + "(?:--)?\\s*(?:" + lbq + "|$)")

	matches := bmre.FindAllIndex(body, -1)
	res := make([]int, len(matches))
	for i, m := range matches {
		res[i] = m[0]
	}

	return res
}

// finalBoundary checks to see if this is a final boundary formatted like
//  // --boundary--
// In that case, it returns true. Otherwise, it returns false.
//
// This assumes that the body given is the start of a boundary, so it doesn't
// verify anything but the last part.
func (m *Message) finalBoundary(body []byte, boundary string) bool {
	lbq := regexp.QuoteMeta(string(m.Break()))
	bq := regexp.QuoteMeta(boundary)
	cmre := regexp.MustCompile("^(?:" + lbq + ")?--" + bq + "--\\s*(?:" + lbq + "|$)")
	return cmre.Match(body)
}

func (m *Message) fillPartsMultiPart() error {
	boundary := m.HeaderContentTypeBoundary()

	if m.MaxDepth <= 0 {
		return errors.New("message is nested too deeply")
	}

	// No boundary set, so it's not multipart
	if boundary == "" {
		return m.fillPartsSinglePart()
	}

	boundaries := m.boundaries(m.Content(), boundary)

	// There are no boundaries found, so it's not multipart. Treat it as single
	// part anyway.
	if len(boundaries) == 0 {
		return m.fillPartsSinglePart()
	}

	m.boundary = boundary

	bits := make([][]byte, 0, len(boundaries))
	lb := -1
	for i, b := range boundaries {
		if lb == -1 {
			// Anything before the first boundary is the preamble. This is not a
			// MIME, but extra text to be ignored by the reader. We keep it
			// around for the purpose of round-tripping.
			if b > 0 {
				m.Preamble = m.Content()[0:b]
			}
			lb = b
			continue
		}

		bits = append(bits, m.Content()[lb:b])

		if i == len(boundaries)-1 {
			if m.finalBoundary(m.Content()[b:], boundary) {
				// Anything after the last boundary is the epilogue. This is
				// also not a MIME part and we also keep it around for
				// round-tripping.
				m.Epilogue = m.Content()[b:]
				break
			} else {
				// This is badly formatted, but whatever. We did not find a
				// final boundary, so the last boundary appears to be a part
				// instead so keep it as one.
				bits = append(bits, m.Content()[b:])
			}
		}
	}

	errs := make([]error, 0)
	parts := make([]*Message, len(bits))
	for i, bit := range bits {
		lbr := len(m.Break())
		bend := bytes.Index(bit[lbr:], m.Break()) + lbr*2
		prefix := bit[:bend]
		postBoundary := bit[bend:]
		pm, err := Parse(postBoundary,
			WithMaxDepth(m.MaxDepth-1),
		)
		pm.prefix = prefix
		errs = append(errs, err)
		parts[i] = pm
	}

	m.Parts = parts

	if len(errs) > 0 {
		return &ParseError{errs}
	}

	return nil
}

func (m *Message) fillPartsSinglePart() error {
	m.Parts = []*Message{}
	return nil
}
