package header_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/pkg/v2/header"
	"github.com/zostay/go-email/pkg/v2/message"
	"github.com/zostay/go-email/pkg/v2/param"
)

func TestWordEncodingHeader(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.Set("To", "\"Name ☺\" <user@host>")
	s := &bytes.Buffer{}
	_, _ = h.WriteTo(s)
	assert.Equal(t, "To: =?utf-8?b?Ik5hbWUg4pi6IiA8dXNlckBob3N0Pg==?=\n\n", s.String())

	s = &bytes.Buffer{}
	_, _ = h.WriteTo(s)
	_, _ = fmt.Fprintln(s, "Test Message")
	m, err := message.Parse(s)
	assert.NoError(t, err)
	to, err := m.GetHeader().Get("To")
	assert.NoError(t, err)
	assert.Equal(t, "\"Name ☺\" <user@host>", to)

	const finalResult = `To: =?utf-8?b?Ik5hbWUg4pi6IiA8dXNlckBob3N0Pg==?=

Test Message
`

	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, finalResult, buf.String())
}

func TestParseWordEncodedHeader(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: =?utf-8?Q?Andrew=2C_you=27ve_got_Smart_Matches=E2=84=A2=21?=
Mime-Version: 1.0

Hello`

	m, err := message.Parse(strings.NewReader(headerStr))
	assert.NoError(t, err)

	// Round-tripping works with Subject
	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, headerStr, buf.String())
}

func TestBlankRecipients(t *testing.T) {
	t.Parallel()

	const headerStr = `To: 

Hello`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	al, err := m.GetHeader().GetAddressList("To")
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

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	// I should get addresses from all Delivered-To headers
	addrs, err := m.GetHeader().GetAllAddressLists("Delivered-To")
	require.NoError(t, err)
	i := 0
	for _, als := range addrs {
		for _, al := range als {
			assert.Equal(t, emails[i], al.Address())
			i++
		}
	}
}

func TestHeader_HeaderGetMediaType(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-Type: text/plain; charset=UTF-8
Badly-formatted-type: x-text:foo; charset=UTF-8

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	mt, err := m.GetHeader().GetContentType()
	require.NoError(t, err)

	assert.Equal(t, "text/plain; charset=UTF-8", mt.String())
	assert.Equal(t, "text/plain", mt.MediaType())
	assert.Equal(t, "UTF-8", mt.Charset())
	assert.Equal(t, "text", mt.Type())
	assert.Equal(t, "plain", mt.Subtype())

	// missing header is no error, but no value either
	mt, err = m.GetHeader().GetParamValue("Some-other-type")
	assert.ErrorIs(t, err, header.ErrNoSuchField)
	assert.Nil(t, mt)

	mt, err = m.GetHeader().GetParamValue("badly-formatted-type")
	assert.Error(t, err)
}

func TestHeader_HeaderSetMediaType(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test
`
	h, err := header.Parse([]byte(headerStr), header.LF)
	require.NoError(t, err)

	mt := param.New("text/html")

	h.SetParamValue("Content-type", mt)

	const afterHeaderStr = `Subject: test
Content-type: text/html

`
	s := &bytes.Buffer{}
	_, _ = h.WriteTo(s)
	assert.Equal(t, afterHeaderStr, s.String())
}

func TestHeader_HeaderContentType(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-Type: text/plain; charset=UTF-8

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	ct, err := m.GetHeader().GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "text/plain", ct)

	mt, err := m.GetHeader().GetContentType()
	assert.NoError(t, err)

	ctType := mt.Type()
	assert.Equal(t, "text", ctType)

	ctSubtype := mt.Subtype()
	assert.Equal(t, "plain", ctSubtype)

	ctCharset := mt.Charset()
	assert.Equal(t, "UTF-8", ctCharset)
}

func TestHeader_SetMediaType(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	m.GetHeader().SetMediaType("text/html")

	err = m.GetHeader().SetCharset("latin1")
	assert.NoError(t, err)

	err = m.GetHeader().SetBoundary("abc123")
	assert.NoError(t, err)

	const afterHeaderStr = `Subject: test
Content-type: text/html; boundary=abc123; charset=latin1

`

	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr, buf.String())

	m.GetHeader().SetMediaType("x-text/mshtml")

	const afterHeaderStr2 = `Subject: test
Content-type: x-text/mshtml; boundary=abc123; charset=latin1

`

	buf = &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr2, buf.String())
}

func TestHeader_HeaderContentDisposition(t *testing.T) {
	t.Parallel()

	const headerStr = `Content-disposition: attachment; filename=something.jpg

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	ct, err := m.GetHeader().GetPresentation()
	assert.NoError(t, err)
	assert.Equal(t, "attachment", ct)

	ctType, err := m.GetHeader().GetFilename()
	assert.NoError(t, err)
	assert.Equal(t, "something.jpg", ctType)
}

func TestHeader_HeaderSetContentDisposition(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	m.GetHeader().SetPresentation("inline")

	err = m.GetHeader().SetFilename("foo.txt")
	assert.NoError(t, err)

	const afterHeaderStr = `Subject: test
Content-disposition: inline; filename=foo.txt

`

	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr, buf.String())

	m.GetHeader().SetPresentation("attachment")
	assert.NoError(t, err)

	const afterHeaderStr2 = `Subject: test
Content-disposition: attachment; filename=foo.txt

`

	buf = &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr2, buf.String())
}

func TestHeader_HeaderSetAddressList(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: test

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	people, err := addr.ParseEmailAddressList("sterling@example.com, steve@example.com, bob@example.com")
	require.NoError(t, err)

	m.GetHeader().SetAddressList("To", people)

	const afterHeaderStr = `Subject: test
To: sterling@example.com, steve@example.com, bob@example.com

`

	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr, buf.String())
}

func TestHeader_HeaderGetDate(t *testing.T) {
	t.Parallel()

	const headerStr = `Date: Mon, 05 Dec 2022 16:46:38Z

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	d, err := m.GetHeader().GetDate()
	assert.NoError(t, err)

	assert.Equal(t, time.Date(2022, time.December, 5, 16, 46, 38, 0, time.UTC), d)
}

func TestHeader_HeaderSetDate(t *testing.T) {
	t.Parallel()

	const headerStr = `Subject: testing

`

	m, err := message.Parse(strings.NewReader(headerStr))
	require.NoError(t, err)

	m.GetHeader().SetDate(time.Date(2022, time.December, 5, 16, 46, 38, 0, time.UTC))

	const afterHeaderStr = `Subject: testing
Date: Mon, 05 Dec 2022 16:46:38 +0000

`

	buf := &strings.Builder{}
	_, err = m.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, afterHeaderStr, buf.String())
}

func TestNewHeader(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	err := h.SetTo("sterling@example.com")
	assert.NoError(t, err)

	err = h.SetFrom("steve@example.com")
	assert.NoError(t, err)

	h.SetSubject("sup?")
	h.SetDate(time.Date(2022, time.December, 5, 17, 9, 53, 0, time.UTC))
	h.Set("X-Foo", "abc")

	const headerStr = `To: sterling@example.com
From: steve@example.com
Subject: sup?
Date: Mon, 05 Dec 2022 17:09:53 +0000
X-Foo: abc

`

	require.NoError(t, err)
	s := &bytes.Buffer{}
	_, _ = h.WriteTo(s)
	assert.Equal(t, headerStr, s.String())
}
