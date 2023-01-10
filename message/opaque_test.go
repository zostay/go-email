package message_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
)

func TestOpaque(t *testing.T) {
	t.Parallel()

	buf, expect, err := makeSimple()
	assert.NoError(t, err)

	m, err := buf.Opaque()
	assert.NoError(t, err)

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

	m, err := buf.Opaque()
	assert.NoError(t, err)

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
	m, err := buf.OpaqueAlreadyEncoded()
	assert.NoError(t, err)

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
