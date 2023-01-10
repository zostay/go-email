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

	"github.com/zostay/go-email/v2/header"
	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/param"
	"github.com/zostay/go-email/v2/transfer"
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

	_, err = m.GetHeader().GetParamValue("badly-formatted-type")
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

	m.GetHeader().SetAddressList("To", people...)

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

func TestHeader_Get(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "A", "b")
	h.InsertBeforeField(1, "C", "d")
	h.InsertBeforeField(2, "E", "f")
	h.InsertBeforeField(3, "E", "g")

	b, err := h.Get("A")
	assert.NoError(t, err)
	assert.Equal(t, "b", b)

	b, err = h.Get("C")
	assert.NoError(t, err)
	assert.Equal(t, "d", b)

	_, err = h.Get("E")
	assert.ErrorIs(t, err, header.ErrManyFields)
}

func TestHeader_GetTime(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Second)
	h := &header.Header{}
	h.InsertBeforeField(0, "X-Date", "2010-10-10 10:10:10-0600")
	h.InsertBeforeField(1, "Date", now.Format(time.RFC1123Z))
	h.InsertBeforeField(2, "Not-Date", "blah")
	h.InsertBeforeField(3, "Dup", "")
	h.InsertBeforeField(4, "Dup", "")

	d, err := h.GetTime("x-date")
	assert.NoError(t, err)
	assert.Equal(t, "2010-10-10 10:10:10 -0600 -0600", d.String())

	d, err = h.GetTime("DATE")
	assert.NoError(t, err)
	assert.Equal(t, now.String(), d.String())

	_, err = h.GetTime("Not-date")
	assert.Error(t, err)

	_, err = h.GetTime("Nothing-burger")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	_, err = h.GetTime("Dup")
	assert.ErrorIs(t, err, header.ErrManyFields)
}

func TestHeader_GetAddressList(t *testing.T) {
	t.Parallel()

	const (
		sterlingStr = "sterling@example.com"
		steveStr    = `"Steve Steverson" <steve@example.com>`
		stanStr     = `"Stan Stanson" <stan@example.com>`
		stuStr      = `"Stu Stuson" <stu@example.com>`
	)

	h := &header.Header{}
	h.InsertBeforeField(0, "From", sterlingStr)
	h.InsertBeforeField(1, "To", steveStr)
	h.InsertBeforeField(2, "Cc", strings.Join([]string{stanStr, stuStr}, ", "))
	h.InsertBeforeField(3, "Not-Addr", "blah")
	h.InsertBeforeField(4, "Dup", "")
	h.InsertBeforeField(5, "Dup", "")

	sterling, err := addr.ParseEmailAddrSpec(sterlingStr)
	assert.NoError(t, err)

	steve, err := addr.ParseEmailMailbox(steveStr)
	assert.NoError(t, err)

	stan, err := addr.ParseEmailMailbox(stanStr)
	assert.NoError(t, err)

	stu, err := addr.ParseEmailMailbox(stuStr)
	assert.NoError(t, err)

	al, err := h.GetAddressList("From")
	assert.NoError(t, err)
	assert.Equal(t, addr.AddressList{sterling}, al)

	al, err = h.GetAddressList("To")
	assert.NoError(t, err)
	assert.Equal(t, addr.AddressList{steve}, al)

	al, err = h.GetAddressList("cc")
	assert.NoError(t, err)
	assert.Equal(t, addr.AddressList{stan, stu}, al)

	blah, err := addr.NewMailboxParsed("",
		addr.NewAddrSpecParsed("blah", "", "blah"),
		"", "blah",
	)
	assert.NoError(t, err)

	al, err = h.GetAddressList("not-addr")
	assert.NoError(t, err)
	assert.Equal(t, addr.AddressList{blah}, al)

	_, err = h.GetAddressList("NOPE")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	_, err = h.GetAddressList("dup")
	assert.ErrorIs(t, err, header.ErrManyFields)
}

func TestHeader_GetAllAddressLists(t *testing.T) {
	t.Parallel()

	const (
		sterlingStr = "sterling@example.com"
		steveStr    = `"Steve Steverson" <steve@example.com>`
		stanStr     = `"Stan Stanson" <stan@example.com>`
		stuStr      = `"Stu Stuson" <stu@example.com>`
	)

	h := &header.Header{}
	h.InsertBeforeField(0, "Delivered-To", sterlingStr)
	h.InsertBeforeField(1, "Delivered-To", steveStr)
	h.InsertBeforeField(2, "Delivered-To", strings.Join([]string{stanStr, stuStr}, ", "))

	sterling, err := addr.ParseEmailAddrSpec(sterlingStr)
	assert.NoError(t, err)

	steve, err := addr.ParseEmailMailbox(steveStr)
	assert.NoError(t, err)

	stan, err := addr.ParseEmailMailbox(stanStr)
	assert.NoError(t, err)

	stu, err := addr.ParseEmailMailbox(stuStr)
	assert.NoError(t, err)

	als, err := h.GetAllAddressLists("DELIVered-tO")
	assert.NoError(t, err)
	assert.Equal(t, []addr.AddressList{
		{sterling},
		{steve},
		{stan, stu},
	}, als)

	_, err = h.GetAllAddressLists("not-a-thing")
	assert.ErrorIs(t, err, header.ErrNoSuchField)
}

func TestHeader_GetParamValue(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "mime-thing", "media/type; charset=transylvanian1; bob=upanddown")
	h.InsertBeforeField(1, "anothering", "xyx")
	h.InsertBeforeField(2, "dup", "")
	h.InsertBeforeField(3, "dup", "")

	pv, err := h.GetParamValue("mime-thing")
	assert.NoError(t, err)
	assert.Equal(t, "media/type", pv.MediaType())
	assert.Equal(t, "transylvanian1", pv.Charset())
	assert.Equal(t, "upanddown", pv.Parameter("bob"))

	pv, err = h.GetParamValue("anothering")
	assert.NoError(t, err)
	assert.Equal(t, "xyx", pv.Value())

	_, err = h.GetParamValue("zip")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	_, err = h.GetParamValue("dup")
	assert.ErrorIs(t, err, header.ErrManyFields)
}

func TestHeader_GetKeywordsList(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "word-things", "one, two, three, four, five")
	h.InsertBeforeField(1, "word-things", `\six, \seven, \eight, \nine, \ten`)
	h.InsertBeforeField(2, "word-things", "")
	h.InsertBeforeField(2, "word-things", "eleven twelve")

	ks, err := h.GetKeywordsList("WORD-THINGS")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"one", "two", "three", "four", "five",
		`\six`, `\seven`, `\eight`, `\nine`, `\ten`,
		"eleven twelve",
	}, ks)

	_, err = h.GetKeywordsList("null")
	assert.ErrorIs(t, err, header.ErrNoSuchField)
}

func TestHeader_GetAll(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "One", "two")
	h.InsertBeforeField(1, "One", "five")
	h.InsertBeforeField(2, "Three", "four")
	h.InsertBeforeField(3, "One", "")

	bs, err := h.GetAll("One")
	assert.NoError(t, err)
	assert.Equal(t, []string{"two", "five", ""}, bs)

	bs, err = h.GetAll("Three")
	assert.NoError(t, err)
	assert.Equal(t, []string{"four"}, bs)

	_, err = h.GetAll("Six")
	assert.ErrorIs(t, err, header.ErrNoSuchField)
}

func TestHeader_SetAll(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "A", "b")
	h.InsertBeforeField(1, "C", "d")
	h.InsertBeforeField(2, "E", "f")
	h.InsertBeforeField(3, "E", "g")

	h.SetAll("A", "one", "two")
	h.SetAll("B", "three", "four")
	h.SetAll("C", "five", "six")
	h.SetAll("E", "seven")

	const expect = `A: one
C: five
E: seven
A: two
B: three
B: four
C: six

`

	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(56), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())

	h.SetAll("C")
	_, err = h.GetAll("C")
	assert.ErrorIs(t, err, header.ErrNoSuchField)
}

func TestHeader_SetKeywordsList(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "A", "b")
	h.InsertBeforeField(1, "C", "d")
	h.InsertBeforeField(2, "C", "e")
	h.InsertBeforeField(3, "F", "g")

	h.SetKeywordsList("C", "one", "two", "three")

	b, err := h.Get("C")
	assert.NoError(t, err)
	assert.Equal(t, "one, two, three", b)

	h.SetKeywordsList("C")

	b, err = h.Get("C")
	assert.NoError(t, err)
	assert.Equal(t, "", b)
}

func TestHeader_Set(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "A", "b")
	h.InsertBeforeField(1, "C", "d")
	h.InsertBeforeField(2, "E", "f")
	h.InsertBeforeField(3, "E", "g")

	h.Set("A", "one")
	h.Set("B", "two")
	h.Set("C", "three")
	h.Set("E", "four")

	const expect = `A: one
C: three
E: four
B: two

`
	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(32), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestHeader_Set2(t *testing.T) {
	// check the edge case when the deleted field is last
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "A", "b")
	h.InsertBeforeField(1, "C", "d")
	h.InsertBeforeField(2, "E", "f")
	h.InsertBeforeField(3, "E", "g")

	h.Set("A", "one")
	h.Set("C", "three")
	h.Set("E", "four")

	const expect = `A: one
C: three
E: four

`
	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(25), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestHeader_SetTime(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	now := time.Now().Truncate(time.Second)
	h.SetTime("X-Date", now)

	b, err := h.Get("X-Date")
	assert.NoError(t, err)
	assert.Equal(t, now.Format(time.RFC1123Z), b)
}

func TestHeader_SetAddressList(t *testing.T) {
	t.Parallel()

	sterling, err := addr.ParseEmailMailbox("sterling@example.com")
	assert.NoError(t, err)
	steve, err := addr.ParseEmailMailbox(`"Steve" <steve@example.com>`)
	assert.NoError(t, err)

	h := &header.Header{}
	h.SetAddressList("X-To", sterling, steve)

	b, err := h.Get("X-To")
	assert.NoError(t, err)
	assert.Equal(t, `sterling@example.com, Steve <steve@example.com>`, b)
}

func TestHeader_SetAllAddressLists(t *testing.T) {
	t.Parallel()

	const (
		sterlingStr = "sterling@example.com"
		steveStr    = `"Steve Steverson" <steve@example.com>`
		stanStr     = `"Stan Stanson" <stan@example.com>`
		stuStr      = `"Stu Stuson" <stu@example.com>`
	)

	sterling, err := addr.ParseEmailAddrSpec(sterlingStr)
	assert.NoError(t, err)

	steve, err := addr.ParseEmailMailbox(steveStr)
	assert.NoError(t, err)

	stan, err := addr.ParseEmailMailbox(stanStr)
	assert.NoError(t, err)

	stu, err := addr.ParseEmailMailbox(stuStr)
	assert.NoError(t, err)

	h := &header.Header{}
	h.SetAllAddressLists("X-Addr",
		addr.AddressList{sterling},
		addr.AddressList{steve, stan},
		addr.AddressList{stu},
	)

	bs, err := h.GetAll("X-Addr")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		sterlingStr,
		strings.Join([]string{steveStr, stanStr}, ", "),
		stuStr,
	}, bs)
}

func TestHeader_SetParamValue(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetParamValue("X-Type", param.New("image/jpeg", map[string]string{
		"boundary": "testboundary",
		"charset":  "latin1",
	}))

	b, err := h.Get("X-Type")
	assert.NoError(t, err)
	assert.Equal(t, "image/jpeg; boundary=testboundary; charset=latin1", b)
}

func TestHeader_GetContentType(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetContentType()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "content-type", "application/json; charset=utf-8")

	pv, err := h.GetContentType()
	assert.NoError(t, err)
	assert.Equal(t, "application/json", pv.MediaType())
	assert.Equal(t, "utf-8", pv.Charset())
}

func TestHeader_SetContentType(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetContentType(param.New("text/plain", map[string]string{"boundary": "abc123"}))

	b, err := h.Get(header.ContentType)
	assert.NoError(t, err)
	assert.Equal(t, "text/plain; boundary=abc123", b)
}

func TestHeader_GetMediaType(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetContentType()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "content-type", "application/json; charset=utf-8")

	mt, err := h.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "application/json", mt)
}

func TestHeader_GetCharset(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, header.ContentType, "something; charset=greek7")

	c, err := h.GetCharset()
	assert.NoError(t, err)
	assert.Equal(t, "greek7", c)
}

func TestHeader_SetCharset(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	err := h.SetCharset("something")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.SetMediaType("something")
	err = h.SetCharset("something")
	assert.NoError(t, err)

	b, err := h.Get(header.ContentType)
	assert.NoError(t, err)
	assert.Equal(t, "something; charset=something", b)
}

func TestHeader_SetBoundary(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	err := h.SetBoundary("something")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.SetMediaType("something")
	err = h.SetBoundary("something")
	assert.NoError(t, err)

	b, err := h.Get(header.ContentType)
	assert.NoError(t, err)
	assert.Equal(t, "something; boundary=something", b)
}

func TestHeader_GetContentDisposition(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetContentDisposition()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "content-disposition", "inline; filename=uh")

	pv, err := h.GetContentDisposition()
	assert.NoError(t, err)
	assert.Equal(t, "inline", pv.MediaType())
	assert.Equal(t, "uh", pv.Filename())
}

func TestHeader_SetContentDisposition(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetContentDisposition(param.New("attachment", map[string]string{"filename": "abc123"}))

	b, err := h.Get(header.ContentDisposition)
	assert.NoError(t, err)
	assert.Equal(t, "attachment; filename=abc123", b)
}

func TestHeader_GetPresentation(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetContentType()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "content-disposition", "attachment; filename=foo.json")

	mt, err := h.GetPresentation()
	assert.NoError(t, err)
	assert.Equal(t, "attachment", mt)
}

func TestHeader_GetFilename(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetFilename()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.SetPresentation("something")

	_, err = h.GetFilename()
	assert.ErrorIs(t, err, header.ErrNoSuchFieldParameter)

	err = h.SetFilename("else")
	assert.NoError(t, err)

	f, err := h.GetFilename()
	assert.NoError(t, err)
	assert.Equal(t, "else", f)
}

func TestHeader_SetFilename(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	err := h.SetFilename("something")
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.SetPresentation("something")
	err = h.SetFilename("something")
	assert.NoError(t, err)

	b, err := h.Get(header.ContentDisposition)
	assert.NoError(t, err)
	assert.Equal(t, "something; filename=something", b)
}

func TestHeader_GetDate(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	now := time.Now().Truncate(time.Second)
	h.InsertBeforeField(0, "Date", now.Format(time.RFC1123Z))

	d, err := h.GetDate()
	assert.NoError(t, err)
	assert.Equal(t, now, d)
}

func TestHeader_GetSubject(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.InsertBeforeField(0, "SUBJEct", "this is a test")

	s, err := h.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "this is a test", s)
}

func TestHeader_SetSubject(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetSubject("woo boo too")

	b, err := h.Get(header.Subject)
	assert.NoError(t, err)
	assert.Equal(t, "woo boo too", b)
}

func TestHeader_Get_BccCcToFromSenderReplyTo(t *testing.T) {
	t.Parallel()

	const sterling = `sterling@example.com`

	h := &header.Header{}
	h.InsertBeforeField(0, "to", sterling)
	h.InsertBeforeField(1, "cc", sterling)
	h.InsertBeforeField(2, "bcc", sterling)
	h.InsertBeforeField(3, "from", sterling)
	h.InsertBeforeField(4, "sender", sterling)
	h.InsertBeforeField(5, "reply-to", sterling)

	sterlingAddr, err := addr.ParseEmailAddrSpec(sterling)
	assert.NoError(t, err)

	sa := addr.AddressList{sterlingAddr}

	to, err := h.GetTo()
	assert.NoError(t, err)
	assert.Equal(t, sa, to)

	cc, err := h.GetCc()
	assert.NoError(t, err)
	assert.Equal(t, sa, cc)

	bcc, err := h.GetBcc()
	assert.NoError(t, err)
	assert.Equal(t, sa, bcc)

	from, err := h.GetFrom()
	assert.NoError(t, err)
	assert.Equal(t, sa, from)

	sender, err := h.GetSender()
	assert.NoError(t, err)
	assert.Equal(t, sa, sender)

	replyTo, err := h.GetReplyTo()
	assert.NoError(t, err)
	assert.Equal(t, sa, replyTo)
}

func TestHeader_Set_ToCcBccFromSenderReplyTo_String(t *testing.T) {
	t.Parallel()

	{
		// parse error test
		const broken = `example.com`

		h := &header.Header{}

		err := h.SetTo(broken)
		assert.Error(t, err)

		err = h.SetCc(broken)
		assert.Error(t, err)

		err = h.SetBcc(broken)
		assert.Error(t, err)

		err = h.SetFrom(broken)
		assert.Error(t, err)

		err = h.SetSender(broken)
		assert.Error(t, err)

		err = h.SetReplyTo(broken)
		assert.Error(t, err)

		assert.Equal(t, 0, h.Len())
	}

	{
		// bad input error test
		h := &header.Header{}

		err := h.SetTo(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		err = h.SetCc(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		err = h.SetBcc(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		err = h.SetFrom(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		err = h.SetSender(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		err = h.SetReplyTo(42)
		assert.ErrorIs(t, err, header.ErrWrongAddressType)

		assert.Equal(t, 0, h.Len())
	}

	{
		// str ok input
		h := &header.Header{}

		const str = `sterling@example.com`
		ad, err := addr.ParseEmailAddrSpec(str)
		assert.NoError(t, err)

		add := addr.AddressList{ad}

		err = h.SetTo(str)
		assert.NoError(t, err)
		b, err := h.GetTo()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		err = h.SetCc(str)
		assert.NoError(t, err)
		b, err = h.GetCc()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		err = h.SetBcc(str)
		assert.NoError(t, err)
		b, err = h.GetBcc()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		err = h.SetFrom(str)
		assert.NoError(t, err)
		b, err = h.GetFrom()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		err = h.SetSender(str)
		assert.NoError(t, err)
		b, err = h.GetSender()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		err = h.SetReplyTo(str)
		assert.NoError(t, err)
		b, err = h.GetReplyTo()
		assert.NoError(t, err)
		assert.Equal(t, add, b)

		const expect = `To: sterling@example.com
Cc: sterling@example.com
Bcc: sterling@example.com
From: sterling@example.com
Sender: sterling@example.com
Reply-to: sterling@example.com

`

		buf := &bytes.Buffer{}
		n, err := h.WriteTo(buf)
		assert.Equal(t, int64(164), n)
		assert.NoError(t, err)
		assert.Equal(t, expect, buf.String())
	}

	{
		// addr ok input
		h := &header.Header{}

		const str = `sterling@example.com`
		ad, err := addr.ParseEmailAddrSpec(str)
		assert.NoError(t, err)

		err = h.SetTo(ad)
		assert.NoError(t, err)
		b, err := h.Get(header.To)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		err = h.SetCc(ad)
		assert.NoError(t, err)
		b, err = h.Get(header.Cc)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		err = h.SetBcc(ad)
		assert.NoError(t, err)
		b, err = h.Get(header.Bcc)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		err = h.SetFrom(ad)
		assert.NoError(t, err)
		b, err = h.Get(header.From)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		err = h.SetSender(ad)
		assert.NoError(t, err)
		b, err = h.Get(header.Sender)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		err = h.SetReplyTo(ad)
		assert.NoError(t, err)
		b, err = h.Get(header.ReplyTo)
		assert.NoError(t, err)
		assert.Equal(t, str, b)

		const expect = `To: sterling@example.com
Cc: sterling@example.com
Bcc: sterling@example.com
From: sterling@example.com
Sender: sterling@example.com
Reply-to: sterling@example.com

`

		buf := &bytes.Buffer{}
		n, err := h.WriteTo(buf)
		assert.Equal(t, int64(164), n)
		assert.NoError(t, err)
		assert.Equal(t, expect, buf.String())
	}
}

func TestHeader_GetKeywords(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetKeywords()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "Keywords", "a, b, c")
	h.InsertBeforeField(1, "Keywords", "d")
	h.InsertBeforeField(3, "Keywords", "e, f")

	ks, err := h.GetKeywords()
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, ks)
	assert.NoError(t, err)
}

func TestHeader_SetKeywords(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	h.InsertBeforeField(0, "Keywords", "a, b, c")
	h.InsertBeforeField(1, "Keywords", "d")
	h.InsertBeforeField(3, "Keywords", "e, f")

	h.SetKeywords("one", "two", "three")

	const expect = `Keywords: one, two, three

`

	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(27), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestHeader_GetComments(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	h.InsertBeforeField(0, "Comments", "a")
	h.InsertBeforeField(1, "Comments", "b")
	h.InsertBeforeField(2, "Comments", "c")

	cs, err := h.GetComments()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, cs)
}

func TestHeader_SetComments(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	h.SetComments("one", "two", "three")

	const expect = `Comments: one
Comments: two
Comments: three

`

	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, int64(45), n)
	assert.Equal(t, expect, buf.String())
}

func TestHeader_Get_ReferencesInReplyToMessageID(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	_, err := h.GetReferences()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	_, err = h.GetInReplyTo()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	_, err = h.GetMessageID()
	assert.ErrorIs(t, err, header.ErrNoSuchField)

	h.InsertBeforeField(0, "References", "abc123")
	h.InsertBeforeField(1, "In-Reply-To", "def456")
	h.InsertBeforeField(2, "Message-id", "ghi789")

	b, err := h.GetReferences()
	assert.NoError(t, err)
	assert.Equal(t, "abc123", b)

	b, err = h.GetInReplyTo()
	assert.NoError(t, err)
	assert.Equal(t, "def456", b)

	b, err = h.GetMessageID()
	assert.NoError(t, err)
	assert.Equal(t, "ghi789", b)
}

func TestHeader_Set_ReferencesInReplyToMessageID(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	h.SetMessageID("foo")
	h.SetReferences("bar")
	h.SetInReplyTo("baz")

	const expect = `Message-id: foo
References: bar
In-reply-to: baz

`

	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(50), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestHeader_GetTransferEncoding(t *testing.T) {
	t.Parallel()

	h := &header.Header{}

	h.InsertBeforeField(0, header.ContentTransferEncoding, transfer.Bit7)

	cte, err := h.GetTransferEncoding()
	assert.NoError(t, err)
	assert.Equal(t, transfer.Bit7, cte)
}

func TestHeader_SetTransferEncoding(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetTransferEncoding(transfer.QuotedPrintable)

	const expect = `Content-transfer-encoding: quoted-printable

`

	buf := &bytes.Buffer{}
	n, err := h.WriteTo(buf)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}
