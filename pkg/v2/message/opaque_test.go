package message_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/pkg/v2/message"
)

func TestOpaque(t *testing.T) {
	buf, expect, err := makeSimple()
	assert.NoError(t, err)

	m, err := buf.Opaque()
	assert.NoError(t, err)

	assert.Equal(t, &m.Header, m.GetHeader())

	_, err = m.GetParts()
	assert.Error(t, err, message.ErrNotMultipart)

	_, err = m.GetReader()
	assert.NoError(t, err)

	assert.False(t, m.IsMultipart())

	out := &bytes.Buffer{}
	n, err := m.WriteTo(out)
	assert.Equal(t, int64(len(expect)), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, out.String())
}
