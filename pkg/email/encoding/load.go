package encoding

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/zostay/go-email/pkg/email/mime"
)

func init() {
	mime.CharsetEncoder = CharsetEncoder
	mime.CharsetDecoder = CharsetDecoder
}

func CharsetEncoder(m *mime.Message, b []byte) (string, error) {
	e, err := ianaindex.MIME.Encoding(m.Charset())
	if m.EncodingCheck() {
		if err != nil {
			return "", err
		}

		eb, err := e.NewEncoder().Bytes(b)
		if err != nil {
			return "", err
		}

		return string(eb), nil
	}

	// This is VERY naughty.
	if err != nil {
		return string(b), nil
	}

	enc := encoding.ReplaceUnsupported(e.NewEncoder())
	eb, _ := enc.Bytes(b)
	return string(eb), nil
}

func CharsetDecoder(m *mime.Message, s string) ([]byte, error) {
	e, err := ianaindex.MIME.Encoding(m.Charset())
	if err != nil {
		return nil, err
	}

	es, err := e.NewDecoder().String(s)
	if err != nil {
		return nil, err
	}

	return []byte(es), nil
}
