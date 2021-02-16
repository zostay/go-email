package mime

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Encoder represents the character encoding function used by the mime package
// to transform data supplied in native unicode format to be written out in the
// character encoding indicated by the charset of the message.
//
// The encoder should attempt to clean up and only output text that is valid in
// the target encoding. If no encoding is present, then us-ascii should be
// assumed.
//
// If the target charset is not supported, bytes should be returned as nil and an
// error should be returned.
type Encoder func(charset, s string) ([]byte, error)

// Decoder represents the character decoding function used by the mime package
// for transforming parsed data supplied in arbitrary text encodings. This will
// be decoded into native unicode.
//
// The decoder should only permit a valid transformation from the source format
// into unicode. Any byte present in the input that is invalid for the source
// character encoding should be replaced with the unicode.ReplacementChar.
//
// If the source charset is not supported, bytes should be returned as nil and
// an error should be returned.
type Decoder func(charset string, b []byte) (string, error)

var (
	// CharsetEncoder is the Encoder used for outputting unicode strings as
	// bytes in the output format. You may replace this with a custom encoder if
	// you like or to make use of an encoder that is able to handle a wide
	// variety of encodings, you can import the encoding package:
	//  import _ "github.com/zostay/go-email/pkg/encoding"
	CharsetEncoder Encoder = DefaultCharsetEncoder

	// CharsetDecoder is the Decoder used for transforming input characters into
	// unicode for use in the decoded fields of MIME messages. You may replace
	// this with a customer decoder you prefer or to make use of a decoder that
	// supports a broad range of encodings, you can import the encoding package:
	//  import _ "github.com/zostay/go-email/pkg/encoding"
	CharsetDecoder Decoder = DefaultCharsetDecoder
)

// DefaultCharsetEncoder is the default encoder. It is able to handle us-ascii
// and utf-8 only. Anything else will result in an error.
//
// When outputting us-ascii, ios-8859-1 (a.k.a. latin1), any utf-8 character
// present that does not fit in us-ascii will be replaced with "\x1a", which is
// the ASCII SUB character.
func DefaultCharsetEncoder(charset, s string) ([]byte, error) {
	switch strings.ToLower(charset) {
	case "us-ascii", "":
		var buf bytes.Buffer
		for _, c := range s {
			if c > unicode.MaxASCII {
				buf.WriteRune('\x1a') // ASCII substitution char
			} else {
				buf.WriteRune(c)
			}
		}
		return buf.Bytes(), nil
	case "iso-8859-1", "latin1", "utf-8":
		return []byte(s), nil
	default:
		return nil, fmt.Errorf("unsupported byte encoding %q", charset)
	}
}

// DefaultCharsetDecoder is the default decoder. It is able to handle us-ascii,
// iso-8859-1 (a.k.a. latin1), and utf-8 only. Anything else will result in an
// error.
//
// When us-ascii is input, any NUL or 8-bit character (i.e., bytes greater than
// 0x7f) will be translated into unicode.ReplacementChar.
//
// When utf-8 is input, the bytes will be read in and transformed into runes
// such that only valid unicode bytes will be permitted in. Errors will be
// brought in as unicode.ReplacementChar.
func DefaultCharsetDecoder(charset string, b []byte) (string, error) {
	switch strings.ToLower(charset) {
	case "us-ascii", "":
		var s strings.Builder
		for _, c := range b {
			if c > unicode.MaxASCII {
				s.WriteRune(unicode.ReplacementChar)
			} else {
				s.WriteByte(c)
			}
		}
		return s.String(), nil
	case "iso-8859-1", "latin1":
		return string(b), nil
	case "utf-8":
		var s strings.Builder
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)
			s.WriteRune(r)
			b = b[size:]
		}
		return s.String(), nil
	default:
		return "", fmt.Errorf("unsupported byte encoding %q", charset)
	}
}

// CharsetDecoderToCharsetReader transforms a CharsetDecoder defined here into the
// interface used by mime.WordDecoder.
func CharsetDecoderToCharsetReader(decode func(string, []byte) (string, error)) func(string, io.Reader) (io.Reader, error) {
	return func(charset string, r io.Reader) (io.Reader, error) {
		bs, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		s, err := decode(charset, bs)
		if err != nil {
			return nil, err
		}

		return strings.NewReader(s), nil
	}
}
