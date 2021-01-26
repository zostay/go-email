package encoding

import (
	_ "golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/zostay/go-email/pkg/email/mime"
)

func init() {
	mime.CharsetEncoder = CharsetEncoder
	mime.CharsetDecoder = CharsetDecoder
}

func CharsetEncoder(m *mime.Message, s string) ([]byte, error) {
	e, err := ianaindex.MIME.Encoding(m.Charset())
	if err != nil {
		return nil, err
	}

	es, err := e.NewEncoder().String(s)
	if err != nil {
		return nil, err
	}

	return []byte(es), nil
}

func CharsetDecoder(m *mime.Message, b []byte) (string, error) {
	e, err := ianaindex.MIME.Encoding(m.Charset())
	if err != nil {
		return "", err
	}

	eb, err := e.NewDecoder().Bytes(b)
	if err != nil {
		return "", err
	}

	return string(eb), nil
}
