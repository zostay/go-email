package message_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
)

func TestParse_WithBadlyFolded(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/badly-folded")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}
