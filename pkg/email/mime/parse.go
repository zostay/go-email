package mime

import (
	"bytes"
	"fmt"
	"mime"
	"regexp"
	"strings"

	"github.com/zostay/go-email/pkg/email/simple"
)

type ParseError struct {
	Errs []error
}

func (err *ParseError) Error() string {
	errs := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errs[i] = e.Error()
	}
	return "error parsing MIME message: " + strings.Join(errs, ", ")
}

type option func(*Message)

func WithEncodingCheck(ec bool) option {
	return func(m *Message) { m.EncodingCheck = ec }
}
func withDepth(d int) option {
	return func(m *Message) { m.Depth = d }
}

func Parse(m []byte, o ...option) (*Message, error) {
	var mm *Message

	msg, err := simple.Parse(m)
	if msg != nil {
		mm = &Message{
			Message:       *msg,
			Depth:         0,
			EncodingCheck: false,
		}
	}

	for _, opt := range o {
		opt(mm)
	}

	if err != nil {
		return mm, err
	}

	ct, ps, err := mime.ParseMediaType(mm.RawContentType())
	if err != nil {
		return mm, err
	}

	mm.ContentType = &ContentType{
		MediaType: ct,
		Params:    ps,
	}

	err = mm.FillParts()
	if err != nil {
		return mm, err
	}

	return mm, nil
}

func (m *Message) FillParts() error {
	if strings.HasPrefix(m.ContentType.MediaType, "multipart/") ||
		strings.HasPrefix(m.ContentType.MediaType, "message/") {
		return m.FillPartsMultiPart()
	} else {
		return m.FillPartsSinglePart()
	}

}

const (
	MaxMultipartDepth = 10
)

func (m *Message) boundaries(body []byte, boundary string) []int {
	lbq := regexp.QuoteMeta(string(m.Break()))
	bq := regexp.QuoteMeta(boundary)
	bmre := regexp.MustCompile("(?:^|" + lbq + ")--" + bq + "\\s*(?:" + lbq + "|$)")

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

func (m *Message) FillPartsMultiPart() error {
	boundary := m.ContentType.Params["boundary"]

	if m.Depth > MaxMultipartDepth {
		return fmt.Errorf("message is more than %d deep in parts", MaxMultipartDepth)
	}

	// No boundary set, so it's not multipart
	if boundary == "" {
		return m.FillPartsSinglePart()
	}

	boundaries := m.boundaries(m.Body(), boundary)

	// There are no boundaries found, so it's not multipart
	if len(boundaries) == 0 {
		return m.FillPartsSinglePart()
	}

	bits := make([][]byte, 0, len(boundaries))
	lb := -1
	for i, b := range boundaries {
		if lb == -1 {
			// Anything before the first boundary is the preamble. This is not a
			// MIME, but extra text to be ignored by the reader. We keep it
			// around for the purpose of round-tripping.
			if b > 0 {
				m.Preamble = m.Body()[0:b]
			}
			lb = b
			continue
		}

		bits = append(bits, m.Body()[lb:b])

		if i == len(boundaries)-1 {
			if m.finalBoundary(m.Body(), boundary) {
				// Anything after the last boundary is the epilogue. This is
				// also not a MIME part and we also keep it around for
				// round-tripping.
				m.Epilogue = m.Body()[b:]
				break
			} else {
				// This is badly formatted, but whatever. We did not find a
				// final boundary, so the last boundary appears to be a part
				// instead so keep it as one.
				bits = append(bits, m.Body()[b:])
			}
		}
	}

	errs := make([]error, 0)
	parts := make([]*Part, len(bits))
	for i, bit := range bits {
		bend := bytes.Index(bit[2:], m.Break()) + 4
		prefix := bit[:bend]
		postBoundary := bit[bend:]
		m, err := Parse(postBoundary,
			WithEncodingCheck(m.EncodingCheck),
			withDepth(m.Depth+1),
		)
		errs = append(errs, err)
		parts[i] = &Part{*m, prefix}
	}

	if len(errs) > 0 {
		return &ParseError{errs}
	}

	return nil
}

func (m *Message) FillPartsSinglePart() error {
	m.Parts = []*Part{}
	return nil
}
