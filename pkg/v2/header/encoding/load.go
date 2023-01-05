// Package encoding provides a replacement encoder and decoder for use with
// mime.CharsetEncoder and mime.CharsetDecoder. This loads all the encodings
// provided with:
//
// * golang.org/x/text/encoding/ianaindex
//
// This will make the size of your compiled binaries considerably larger. But it
// will also give your code the ability to encode and decode pretty much any
// character set it might encounter in the wild wild world of email.
package encoding

import (
	"fmt"

	_ "golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/zostay/go-email/pkg/v2/header/field"
)

func init() {
	field.CharsetEncoder = CharsetEncoder
	field.CharsetDecoder = CharsetDecoder
}

// CharsetEncoder provides a replacement encoder for mime.CharsetEncoder, which
// can encode a wide range of rare and unusual character sets.
func CharsetEncoder(charset, s string) ([]byte, error) {
	e, err := ianaindex.MIME.Encoding(charset)
	if err != nil {
		return nil, err
	}

	es, err := e.NewEncoder().String(s)
	if err != nil {
		return nil, err
	}

	return []byte(es), nil
}

// CharsetDecoder provides a replacement decoder for mime.CharsetDecoder, which
// can decode a wide range of rare and unusual character sets.
func CharsetDecoder(charset string, b []byte) (string, error) {
	e, err := ianaindex.MIME.Encoding(charset)
	if err != nil {
		return "", err
	}

	if e == nil {
		return "", fmt.Errorf("no encoding found for charset %q", charset)
	}

	eb, err := e.NewDecoder().Bytes(b)
	if err != nil {
		return "", err
	}

	return string(eb), nil
}
