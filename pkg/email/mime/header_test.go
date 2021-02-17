package mime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/pkg/email"
)

func TestWordEncodingHeader(t *testing.T) {
	h, err := NewHeader(email.LF)
	assert.NoError(t, err)
	err = h.HeaderSet("To", "\"Name ☺\" <user@host>")
	assert.NoError(t, err)
	assert.Equal(t, "To: =?utf-8?b?Ik5hbWUg4pi6IiA8dXNlckBob3N0Pg==?=\n", h.String())

	s := h.String() + "\nTest Messge\n"
	m, err := Parse([]byte(s))
	assert.NoError(t, err)
	assert.Equal(t, "\"Name ☺\" <user@host>", m.HeaderGet("To"))
}
