package message_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/pkg/v2/message"
)

const simpleMessage = `To: sterling@example.com
From: steve@example.com
Subject: A basic test of round-tripping

More testing is needed.
`

const multipartMessage = `To: steve@example.com
From: sterling@example.com
Subject: Re: A basic test of round-tripping
Content-type: multipart/alternative; boundary=abcdefghijklm

--abcdefghijklm
Content-type: text/html

<strong>I disagree!</strong> This amount of testing is <em>adequate</em>.
--abcdefghijklm
Content-type: text/plain

*I disagree!* This amount of testing is _adequate_.
--abcdefghijklm--
`

func TestParse_OpaqueRoundTrip(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(simpleMessage))
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(simpleMessage)), n)
	assert.NoError(t, err)
	assert.Equal(t, simpleMessage, buf.String())
}

func TestParse_MultipartRoundTrip(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(multipartMessage))
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, int64(len(multipartMessage)), n)
	assert.NoError(t, err)
	assert.Equal(t, multipartMessage, buf.String())
}
