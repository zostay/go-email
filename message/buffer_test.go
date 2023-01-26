package message_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
)

func makePart() *message.Opaque {
	op := &message.Opaque{
		Reader: strings.NewReader("Test message."),
	}
	op.SetMediaType("text/html")
	return op
}

func makeSimple() (*message.Buffer, string, error) {
	buf := &message.Buffer{}

	buf.SetSubject("test simple")
	buf.SetMediaType("text/plain")

	_, err := fmt.Fprintln(buf, "This is a simple message.")

	const expect = `Subject: test simple
Content-type: text/plain

This is a simple message.
`

	return buf, expect, err
}

func makeOpaqueMultipart() (*message.Buffer, string, error) { //nolint:unparam // this is a test
	const expect = `Subject: test multipart
Content-type: multipart/alternative; boundary=testing

--testing
Content-type: text/html

Test message.
--testing--`

	buf := &message.Buffer{}

	buf.SetSubject("test multipart")
	buf.SetMediaType("multipart/alternative")
	err := buf.SetBoundary("testing")
	if err != nil {
		return nil, expect, err
	}

	p := makePart()

	_, err = fmt.Fprintln(buf, "--testing")
	_, _ = p.WriteTo(buf)
	if err != nil {
		return nil, expect, err
	}
	_, _ = fmt.Fprintln(buf)
	_, _ = fmt.Fprint(buf, "--testing--")

	return buf, expect, nil
}

func makeMultipart() (*message.Buffer, string, error) { //nolint:unparam // this is a test
	const expect = `Subject: test multipart
Content-type: multipart/alternative; boundary=testing

--testing
Content-type: text/html

Test message.
--testing--`

	buf := &message.Buffer{}

	buf.SetSubject("test multipart")
	buf.SetMediaType("multipart/alternative")
	err := buf.SetBoundary("testing")
	if err != nil {
		return nil, expect, err
	}

	buf.Add(makePart())

	return buf, expect, nil
}

func TestBuffer_Add(t *testing.T) {
	t.Parallel()

	buf := &message.Buffer{}

	buf.SetSubject("test multipart")
	buf.SetMediaType("multipart/alternative")
	err := buf.SetBoundary("testing")
	assert.NoError(t, err)

	assert.Equal(t, message.ModeUnset, buf.Mode())

	buf.Add(makePart())

	assert.Equal(t, message.ModeMultipart, buf.Mode())

	assert.Panics(t, func() {
		_, _ = buf.Write([]byte{})
	})

	m, err := buf.Multipart()
	assert.NoError(t, err)

	const expected = `Subject: test multipart
Content-type: multipart/alternative; boundary=testing

--testing
Content-type: text/html

Test message.
--testing--`

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expected)), n)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}

func TestBuffer_Write(t *testing.T) {
	t.Parallel()

	buf := &message.Buffer{}

	assert.Equal(t, message.ModeUnset, buf.Mode())

	buf.SetSubject("test opaque")
	buf.SetMediaType("text/plain")

	n, err := fmt.Fprintln(buf, "This is a simple opaque message.")
	assert.Equal(t, 33, n)
	assert.NoError(t, err)

	assert.Equal(t, message.ModeSingle, buf.Mode())

	assert.Panics(t, func() {
		buf.Add(makePart())
	})

	m := buf.Opaque()

	const expected = `Subject: test opaque
Content-type: text/plain

This is a simple opaque message.
`

	out := &bytes.Buffer{}
	n64, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expected)), n64)
	assert.NoError(t, err)
	assert.Equal(t, expected, out.String())
}

func TestBuffer_Opaque_FromSimple(t *testing.T) {
	t.Parallel()

	s, expect, err := makeSimple()
	assert.NoError(t, err)

	m := s.Opaque()

	subj, err := m.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "test simple", subj)

	mt, err := m.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "text/plain", mt)

	assert.False(t, m.IsMultipart())

	ps := m.GetParts()
	assert.Nil(t, ps)

	r := m.GetReader()
	assert.NotNil(t, r)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestBuffer_Opaque_FromMultipart(t *testing.T) {
	t.Parallel()

	s, expect, err := makeMultipart()
	assert.NoError(t, err)

	m := s.Opaque()

	subj, err := m.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "test multipart", subj)

	mt, err := m.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "multipart/alternative", mt)

	// it may be constructed as multipart, but it's been returned opaque
	assert.False(t, m.IsMultipart())

	ps := m.GetParts()
	assert.Nil(t, ps)

	r := m.GetReader()
	assert.NotNil(t, r)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestBuffer_Multipart_FromSimple(t *testing.T) {
	t.Parallel()

	s, _, err := makeSimple()
	assert.NoError(t, err)

	_, err = s.Multipart()
	assert.ErrorIs(t, err, message.ErrParsesAsNotMultipart)
}

func TestBuffer_Multipart_FromMultipart(t *testing.T) {
	t.Parallel()

	s, expect, err := makeMultipart()
	assert.NoError(t, err)

	m, err := s.Multipart()
	assert.NoError(t, err)

	subj, err := m.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "test multipart", subj)

	mt, err := m.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "multipart/alternative", mt)

	assert.True(t, m.IsMultipart())

	p := m.GetParts()
	assert.Len(t, p, 1)

	r := m.GetReader()
	assert.Nil(t, r)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestBuffer_Multipart_FromOpaqueMultipart(t *testing.T) {
	t.Parallel()

	s, expect, err := makeOpaqueMultipart()
	assert.NoError(t, err)

	m, err := s.Multipart()
	assert.NoError(t, err)

	subj, err := m.GetSubject()
	assert.NoError(t, err)
	assert.Equal(t, "test multipart", subj)

	mt, err := m.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "multipart/alternative", mt)

	assert.True(t, m.IsMultipart())

	p := m.GetParts()
	assert.Len(t, p, 1)

	r := m.GetReader()
	assert.Nil(t, r)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}
