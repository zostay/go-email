package field

import (
	"mime"
	"strings"
)

// Encode transforms a single header field body by looking for any characters
// allowed for header encoding and turning them into encode body values using
// word encoder. It will always output q-type encoding using UTF-8 as the
// character set.
func Encode(body string) string {
	return mime.QEncoding.Encode("UTF-8", body)
}

// Decode transforms a single header field body and looks for MIME word encoded field
// values. When they are found, these are decoded into native unicode.
func Decode(body string) (string, error) {
	dec := &mime.WordDecoder{
		CharsetReader: CharsetDecoderToCharsetReader(CharsetDecoder),
	}

	if strings.Contains(body, "=?") {
		return dec.DecodeHeader(body)
	}

	return body, nil
}
