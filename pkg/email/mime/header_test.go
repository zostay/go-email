package mime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/pkg/email"
)

func TestWordEncodingHeader(t *testing.T) {
	t.Parallel()

	h, err := NewHeader(email.LF)
	assert.NoError(t, err)
	err = h.HeaderSet("To", "\"Name ☺\" <user@host>")
	assert.NoError(t, err)
	assert.Equal(t, "To: =?utf-8?b?Ik5hbWUg4pi6IiA8dXNlckBob3N0Pg==?=\n", h.String())

	s := h.String() + "\nTest Message\n"
	m, err := Parse([]byte(s))
	assert.NoError(t, err)
	assert.Equal(t, "\"Name ☺\" <user@host>", m.HeaderGet("To"))

	const finalResult = `To: =?utf-8?b?Ik5hbWUg4pi6IiA8dXNlckBob3N0Pg==?=

Test Message
`

	assert.Equal(t, finalResult, m.String())
}

func TestParseWordEncodedHeader(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: =?utf-8?Q?Andrew=2C_you=27ve_got_Smart_Matches=E2=84=A2=21?=
Mime-Version: 1.0

Hello`

	m, err := Parse([]byte(headerStr))
	assert.NoError(t, err)

	// Round-tripping works with Subject
	assert.Equal(t, headerStr, m.String())
}

func TestBlankRecipients(t *testing.T) {
	t.Parallel()

	const headerStr = `To: 

Hello`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	al, err := m.HeaderGetAddressList("To")
	require.NoError(t, err)
	assert.Len(t, al, 0)
}

func TestAllAddressLists(t *testing.T) {
	t.Parallel()

	const headerStr = `Delivered-To: one@example.com
Delivered-To: two@example.com, three@example.com

Hello`

	emails := []string{
		"one@example.com",
		"two@example.com",
		"three@example.com",
	}

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	// I should get addresses from all Delivered-To headers
	addrs, err := m.HeaderGetAllAddressLists("Delivered-To")
	require.NoError(t, err)
	for i, addr := range addrs {
		assert.Equal(t, emails[i], addr.Address())
	}
}

func TestHeader_HeaderGetMediaType(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-Type: text/plain; charset=UTF-8
Badly-formatted-type: x-text:foo; charset=UTF-8

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	mt, err := m.HeaderGetMediaType("Content-type")
	require.NoError(t, err)

	assert.Equal(t, "text/plain; charset=UTF-8", mt.String())
	assert.Equal(t, "text/plain", mt.MediaType())
	assert.Equal(t, "UTF-8", mt.Charset())
	assert.Equal(t, "text", mt.Type())
	assert.Equal(t, "plain", mt.Subtype())

	// missing header is no error, but no value either
	mt, err = m.HeaderGetMediaType("Some-other-type")
	assert.NoError(t, err)
	assert.Nil(t, mt)

	mt, err = m.HeaderGetMediaType("badly-formatted-type")
	assert.Error(t, err)
}

func TestHeader_HeaderSetMediaType(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test
`
	h, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	mt, err := NewMediaType("text/html")
	require.NoError(t, err)

	h.HeaderSetMediaType("Content-type", mt)

	const afterHeaderStr = `Subject: test
Content-type: text/html

`
	assert.Equal(t, afterHeaderStr, h.String())
}

func TestHeader_HeaderContentType(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-Type: text/plain; charset=UTF-8

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	ct := m.HeaderContentType()
	assert.Equal(t, "text/plain", ct)

	ctType := m.HeaderContentTypeType()
	assert.Equal(t, "text", ctType)

	ctSubtype := m.HeaderContentTypeSubtype()
	assert.Equal(t, "plain", ctSubtype)

	ctCharset := m.HeaderContentTypeCharset()
	assert.Equal(t, "UTF-8", ctCharset)
}

func TestHeader_HeaderSetContentType(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	err = m.HeaderSetContentType("text/html")
	assert.NoError(t, err)

	err = m.HeaderSetContentTypeCharset("latin1")
	assert.NoError(t, err)

	err = m.HeaderSetContentTypeBoundary("abc123")
	assert.NoError(t, err)

	const afterHeaderStr = `Subject: test
Content-type: text/html; boundary=abc123; charset=latin1

`

	assert.Equal(t, afterHeaderStr, m.String())

	err = m.HeaderSetContentType("x-text/mshtml")
	require.NoError(t, err)

	const afterHeaderStr2 = `Subject: test
Content-type: x-text/mshtml; boundary=abc123; charset=latin1

`

	assert.Equal(t, afterHeaderStr2, m.String())
}

func TestHeader_HeaderContentDisposition(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-disposition: attachment; filename=something.jpg

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	ct := m.HeaderContentDisposition()
	assert.Equal(t, "attachment", ct)

	ctType := m.HeaderContentDispositionFilename()
	assert.Equal(t, "something.jpg", ctType)
}

func TestHeader_HeaderSetContentDisposition(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	err = m.HeaderSetContentDisposition("inline")
	assert.NoError(t, err)

	err = m.HeaderSetContentDispositionFilename("foo.txt")
	assert.NoError(t, err)

	const afterHeaderStr = `Subject: test
Content-disposition: inline; filename=foo.txt

`

	assert.Equal(t, afterHeaderStr, m.String())

	err = m.HeaderSetContentDisposition("attachment")
	assert.NoError(t, err)

	const afterHeaderStr2 = `Subject: test
Content-disposition: attachment; filename=foo.txt

`

	assert.Equal(t, afterHeaderStr2, m.String())
}

func TestHeader_HeaderSetAddressList(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	people, err := addr.ParseEmailAddressList("sterling@example.com, steve@example.com, bob@example.com")
	require.NoError(t, err)

	m.HeaderSetAddressList("To", people)

	const afterHeaderStr = `Subject: test
To: sterling@example.com, steve@example.com, bob@example.com

`

	assert.Equal(t, afterHeaderStr, m.String())
}

func TestHeader_HeaderGetDate(t *testing.T) {
	t.Parallel()

	const headerStr = `Date: Mon, 05 Dec 2022 16:46:38Z

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	d, err := m.HeaderGetDate()
	assert.NoError(t, err)

	assert.Equal(t, time.Date(2022, time.December, 5, 16, 46, 38, 0, time.UTC), d)
}

func TestHeader_HeaderSetDate(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: testing

`

	m, err := Parse([]byte(headerStr))
	require.NoError(t, err)

	m.HeaderSetDate(time.Date(2022, time.December, 5, 16, 46, 38, 0, time.UTC))

	const afterHeaderStr = `Subject: testing
Date: Mon, 05 Dec 2022 16:46:38 +0000

`

	assert.Equal(t, afterHeaderStr, m.String())
}

type special struct{}

func (special) String() string { return "I'm a little teapot" }

func TestNewHeader(t *testing.T) {
	t.Parallel()

	m, err := NewHeader(email.LF,
		"To", "sterling@example.com",
		"From", "steve@example.com",
		"Subject", "sup?",
		"Date", time.Date(2022, time.December, 5, 17, 9, 53, 0, time.UTC),
		"X-Foo", []byte{'a', 'b', 'c'},
		"X-Special", special{},
	)

	const headerStr = `To: sterling@example.com
From: steve@example.com
Subject: sup?
Date: Mon, 05 Dec 2022 17:09:53 +0000
X-Foo: abc
X-Special: I'm a little teapot
`

	require.NoError(t, err)
	assert.Equal(t, headerStr, m.String())
}
