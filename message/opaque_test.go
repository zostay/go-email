package message_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/v2/message"
)

func TestOpaque(t *testing.T) {
	t.Parallel()

	buf, expect, err := makeSimple()
	assert.NoError(t, err)

	m := buf.Opaque()

	assert.Equal(t, &m.Header, m.GetHeader())

	ps := m.GetParts()
	assert.Nil(t, ps)

	r := m.GetReader()
	assert.NotNil(t, r)

	assert.False(t, m.IsMultipart())
	assert.False(t, m.IsEncoded())

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, out.String())
}

func makeSimpleWithEncoding() (*message.Buffer, string, string, error) {
	buf := &message.Buffer{}

	buf.SetSubject("test simple")
	buf.SetTransferEncoding("quoted-printable")
	buf.SetMediaType("text/plain")

	const (
		expect = `Subject: test simple
Content-transfer-encoding: quoted-printable
Content-type: text/plain

`
		encoded = "I =E2=9D=A4 email!\r\n"
		decoded = "I ❤ email!\n"
	)

	_, err := fmt.Fprint(buf, decoded)

	return buf, expect + encoded, expect + decoded, err
}

func TestOpaque_TransferEncodingEncoded(t *testing.T) {
	t.Parallel()

	buf, expectEnc, expectDec, err := makeSimpleWithEncoding()
	assert.NoError(t, err)

	m := buf.Opaque()

	assert.Equal(t, &m.Header, m.GetHeader())

	ps := m.GetParts()
	assert.Nil(t, ps)

	r := m.GetReader()
	assert.NotNil(t, r)

	assert.False(t, m.IsMultipart())
	assert.False(t, m.IsEncoded())

	// TODO the discrepancy between bytes written and n returned from WriteTo feels like a bug

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expectDec)), n)
	assert.NoError(t, err)
	assert.Equal(t, expectEnc, out.String())
}

func TestOpaque_TransferEncodingDecoded(t *testing.T) {
	t.Parallel()

	buf, _, expectDec, err := makeSimpleWithEncoding()
	assert.NoError(t, err)

	// This is actually wrong since the data created by makeSimpleWithEncoding
	// is not encoded. However, we just want to test that no encoding is
	// performed if we all OpaqueAlreadyEncoded.
	m := buf.OpaqueAlreadyEncoded()

	assert.Equal(t, &m.Header, m.GetHeader())

	ps := m.GetParts()
	assert.Nil(t, ps)

	r := m.GetReader()
	assert.NotNil(t, r)

	assert.False(t, m.IsMultipart())
	assert.True(t, m.IsEncoded())

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expectDec)), n)
	assert.NoError(t, err)
	assert.Equal(t, expectDec, out.String())
}

func TestAttachmentFile(t *testing.T) {
	t.Parallel()

	af, err := message.AttachmentFile("../test/data/att-1.gif", "image/gif", "base64")
	require.NoError(t, err)

	mt, err := af.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "image/gif", mt)

	p, err := af.GetPresentation()
	assert.NoError(t, err)
	assert.Equal(t, "attachment", p)

	fn, err := af.GetFilename()
	assert.NoError(t, err)
	assert.Equal(t, "att-1.gif", fn)

	const headerPart = `Content-type: image/gif
Content-disposition: attachment; filename=att-1.gif
Content-transfer-encoding: base64

`
	const attPart = `R0lGODlhDAAMAPcAAAAAAAgICBAQEBgYGCkpKTExMTk5OUpKSoyMjJSUlJycnKWlpbW1tc7Ozufn
5+/v7/f39///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////ywAAAAADAAMAAAIXwAjRICQ
wIAAAQYUQBAYwUEBAAACEIBYwMHAhxARNIAIoAAEBBAPOICwkSMCjBAXlKQYgCMABSsjtuQI02UA
lC9jFgBJMyYCCCgRMODoseFElx0tCvxYIEAAAwkWRggIADs=`

	buf := &bytes.Buffer{}
	n, err := af.WriteTo(buf)
	assert.Equal(t, int64(len(headerPart)+890), n)
	assert.NoError(t, err)
	assert.Equal(t, []byte(headerPart+attPart), buf.Bytes())
}
