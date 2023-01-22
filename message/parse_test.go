package message_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/header/field"
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

func TestParse_WithBadlyFoldedNoIndent(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/badly-folded-noindent")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	bf, err := m.GetHeader().Get("Badly-Folded")
	assert.NoError(t, err)
	assert.Equal(t, "This header is badly folded because even though it goes onto thesecond line, it has no indent.", bf)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}

func TestParse_WithJoseyFold(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/josey-fold")
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

func TestParse_WithJoseyNoFold(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/josey-nofold")
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

func TestParse_WithJunkInHeader(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/junk-in-header")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	const expectedJunk = "linden boulevard represent, represent\n"

	m, err := message.Parse(src)
	var badStartErr *field.BadStartError
	assert.ErrorAs(t, err, &badStartErr)
	assert.Equal(t, []byte(expectedJunk), badStartErr.BadStart)
	require.NotNil(t, m)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)-len(expectedJunk)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes[len(expectedJunk):], buf.Bytes())
}
