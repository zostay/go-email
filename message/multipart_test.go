package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
)

func TestMultipart(t *testing.T) {
	buf, expect, err := makeMultipart()
	assert.NoError(t, err)

	m, err := buf.Multipart()
	assert.NoError(t, err)

	assert.Equal(t, &m.Header, m.GetHeader())

	ps, err := m.GetParts()
	assert.NoError(t, err)
	assert.Len(t, ps, 1)

	_, err = m.GetReader()
	assert.Error(t, err, message.ErrMultipart)

	assert.True(t, m.IsMultipart())

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, out.String())
}
