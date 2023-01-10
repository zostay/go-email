package header_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/header"
)

func TestBreak_Bytes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []byte{}, header.Meh.Bytes())
	assert.Equal(t, []byte{0x0d, 0x0a}, header.CRLF.Bytes())
	assert.Equal(t, []byte{0x0a}, header.LF.Bytes())
	assert.Equal(t, []byte{0x0d}, header.CR.Bytes())
	assert.Equal(t, []byte{0x0a, 0x0d}, header.LFCR.Bytes())
}

func TestBreak_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", header.Meh.String())
	assert.Equal(t, "\r\n", header.CRLF.String())
	assert.Equal(t, "\n", header.LF.String())
	assert.Equal(t, "\r", header.CR.String())
	assert.Equal(t, "\n\r", header.LFCR.String())
}
