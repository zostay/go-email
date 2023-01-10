package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultipart(t *testing.T) {
	t.Parallel()

	buf, expect, err := makeMultipart()
	assert.NoError(t, err)

	m, err := buf.Multipart()
	assert.NoError(t, err)

	assert.Equal(t, &m.Header, m.GetHeader())

	ps := m.GetParts()
	assert.NoError(t, err)
	assert.Len(t, ps, 1)

	r := m.GetReader()
	assert.Nil(t, r)

	assert.True(t, m.IsMultipart())
	assert.False(t, m.IsEncoded())

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, out.String())
}
