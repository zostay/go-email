package mime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestParseWordEncodedHeader(t *testing.T) {
	const headerStr = `Subject: =?utf-8?Q?Andrew=2C_you=27ve_got_Smart_Matches=E2=84=A2=21?=
Mime-Version: 1.0

Hello`

	m, err := Parse([]byte(headerStr))
	assert.NoError(t, err)

	// Roundtripping works with Subject
	assert.Equal(t, headerStr, m.String())
}

func TestBlankRecipients(t *testing.T) {
	const headerStr = `To: 

Hello`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	al, err := m.HeaderGetAddressList("To")
	require.NoError(t, err)
	assert.Len(t, al, 0)
}
