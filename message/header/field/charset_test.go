package field_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/zostay/go-email/v2/message/header/encoding"
	"github.com/zostay/go-email/v2/message/header/field"
)

// Εν αρχη ητο ο Λογος, και ο Λογος ητο παρα τω Θεω, και Θεος ητο ο Λογος.

var greekText = []byte{
	0xc5, 0xed, 0x20, 0xe1, 0xf1, 0xf7, 0xe7, 0x20, 0xe7, 0xf4, 0xef, 0x20,
	0xef, 0x20, 0xcb, 0xef, 0xe3, 0xef, 0xf2, 0x2c, 0x20, 0xea, 0xe1, 0xe9,
	0x20, 0xef, 0x20, 0xcb, 0xef, 0xe3, 0xef, 0xf2, 0x20, 0xe7, 0xf4, 0xef,
	0x20, 0xf0, 0xe1, 0xf1, 0xe1, 0x20, 0xf4, 0xf9, 0x20, 0xc8, 0xe5, 0xf9,
	0x2c, 0x20, 0xea, 0xe1, 0xe9, 0x20, 0xc8, 0xe5, 0xef, 0xf2, 0x20, 0xe7,
	0xf4, 0xef, 0x20, 0xef, 0x20, 0xcb, 0xef, 0xe3, 0xef, 0xf2, 0x2e,
}

var unicodeText = []byte{
	0xce, 0x95, 0xce, 0xbd, 0x20, 0xce, 0xb1, 0xcf, 0x81, 0xcf, 0x87, 0xce,
	0xb7, 0x20, 0xce, 0xb7, 0xcf, 0x84, 0xce, 0xbf, 0x20, 0xce, 0xbf, 0x20,
	0xce, 0x9b, 0xce, 0xbf, 0xce, 0xb3, 0xce, 0xbf, 0xcf, 0x82, 0x2c, 0x20,
	0xce, 0xba, 0xce, 0xb1, 0xce, 0xb9, 0x20, 0xce, 0xbf, 0x20, 0xce, 0x9b,
	0xce, 0xbf, 0xce, 0xb3, 0xce, 0xbf, 0xcf, 0x82, 0x20, 0xce, 0xb7, 0xcf,
	0x84, 0xce, 0xbf, 0x20, 0xcf, 0x80, 0xce, 0xb1, 0xcf, 0x81, 0xce, 0xb1,
	0x20, 0xcf, 0x84, 0xcf, 0x89, 0x20, 0xce, 0x98, 0xce, 0xb5, 0xcf, 0x89,
	0x2c, 0x20, 0xce, 0xba, 0xce, 0xb1, 0xce, 0xb9, 0x20, 0xce, 0x98, 0xce,
	0xb5, 0xce, 0xbf, 0xcf, 0x82, 0x20, 0xce, 0xb7, 0xcf, 0x84, 0xce, 0xbf,
	0x20, 0xce, 0xbf, 0x20, 0xce, 0x9b, 0xce, 0xbf, 0xce, 0xb3, 0xce, 0xbf,
	0xcf, 0x82, 0x2e,
}

// Contains one illegal char to test out the encoding.
var asciiText = "In the beginning was the Word, and the Word was with God, and the Word was God. 🖊"
var asciiTextEnc = "In the beginning was the Word, and the Word was with God, and the Word was God. \x1a"
var asciiTextDec = "In the beginning was the Word, and the Word was with God, and the Word was God. \xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd\xef\xbf\xbd"

func TestDefaultCharsetDecoder(t *testing.T) {
	t.Parallel()

	_, err := field.DefaultCharsetDecoder("greek", greekText)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported byte encoding")

	// unicode to unicode decoding is supported, but not very exciting
	dec, err := field.DefaultCharsetDecoder("utf-8", unicodeText)
	assert.NoError(t, err)
	assert.Equal(t, unicodeText, []byte(dec))

	// ascii to unicode decoding is also supported, but only slightly more exciting
	dec, err = field.DefaultCharsetDecoder("", []byte(asciiText))
	assert.NoError(t, err)
	assert.Equal(t, []byte(asciiTextDec), []byte(dec))
}

func TestCharsetDecoder(t *testing.T) {
	t.Parallel()

	// testing the full version from header/encoding
	dec, err := field.CharsetDecoder("greek", greekText)
	assert.NoError(t, err)
	assert.Equal(t, unicodeText, []byte(dec))
}

func TestDefaultCharsetEncoder(t *testing.T) {
	t.Parallel()

	_, err := field.DefaultCharsetEncoder("greek", string(unicodeText))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported byte encoding")

	// unicode to unicode encoding is supported, but not very exciting
	enc, err := field.DefaultCharsetEncoder("utf-8", string(unicodeText))
	assert.NoError(t, err)
	assert.Equal(t, unicodeText, enc)

	// unicode to ascii encoding is also supported, but only slightly more exciting
	enc, err = field.DefaultCharsetEncoder("", asciiText)
	assert.NoError(t, err)
	assert.Equal(t, []byte(asciiTextEnc), enc)
}

func TestCharsetEncoder(t *testing.T) {
	t.Parallel()

	// testing the full version from header/encoding
	enc, err := field.CharsetEncoder("greek", string(unicodeText))
	assert.NoError(t, err)
	assert.Equal(t, greekText, enc)
}

func TestCharsetDecoderToCharsetReader(t *testing.T) {
	t.Parallel()

	cr := field.CharsetDecoderToCharsetReader(field.CharsetDecoder)
	in := bytes.NewReader(greekText)

	out, err := cr("greek", in)
	assert.NoError(t, err)
	dec, err := io.ReadAll(out)
	assert.NoError(t, err)
	assert.Equal(t, unicodeText, dec)
}
