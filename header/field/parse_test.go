package field_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/v2/header/field"
)

func TestParseLines(t *testing.T) {
	t.Parallel()

	// basic parse, no folding
	input := []byte("a:\nb:\nc:\nd:\n")
	lb := field.Break("\n")
	lines, err := field.ParseLines(input, lb)
	assert.NoError(t, err)
	assert.Equal(t, field.Lines{
		[]byte("a:\n"),
		[]byte("b:\n"),
		[]byte("c:\n"),
		[]byte("d:\n"),
	}, lines)

	// folding parse
	input = []byte("a:b\n b\n b\nb:\nc:\nd:\n\teeee\n")
	lines, err = field.ParseLines(input, lb)
	assert.NoError(t, err)
	assert.Equal(t, field.Lines{
		[]byte("a:b\n b\n b\n"),
		[]byte("b:\n"),
		[]byte("c:\n"),
		[]byte("d:\n\teeee\n"),
	}, lines)

	// folding parse, with start junk
	input = []byte(" start:\njunk\na:b\n b\n b\nb:\nc:\nd:\n\teeee\n")
	lines, err = field.ParseLines(input, lb)
	badStart := &field.BadStartError{
		BadStart: []byte(" start:\njunk\n"),
	}
	assert.ErrorAs(t, err, &badStart)
	assert.Equal(t, field.Lines{
		[]byte("a:b\n b\n b\n"),
		[]byte("b:\n"),
		[]byte("c:\n"),
		[]byte("d:\n\teeee\n"),
	}, lines)
}

func TestParse(t *testing.T) {
	t.Parallel()

	f := field.Parse([]byte("Subject: test\n"), []byte{'\n'})
	require.NotNil(t, f)
	require.NotNil(t, f.Raw)
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "test", f.Body())
	assert.Equal(t, "Subject", f.Raw.Name())
	assert.Equal(t, " test", f.Raw.Body())
	assert.Equal(t, "Subject: test", f.Raw.String())

	f = field.Parse([]byte("Subject: =?utf-8?b?4pmg4pmj4pml4pmm?=\r\n"), []byte{'\r', '\n'})
	require.NotNil(t, f)
	require.NotNil(t, f.Raw)
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "♠♣♥♦", f.Body())
	assert.Equal(t, "Subject", f.Raw.Name())
	assert.Equal(t, " =?utf-8?b?4pmg4pmj4pml4pmm?=", f.Raw.Body())
	assert.Equal(t, "Subject: =?utf-8?b?4pmg4pmj4pml4pmm?=", f.Raw.String())

	f = field.Parse([]byte("Subject"), []byte{'\n'})
	require.NotNil(t, f)
	require.NotNil(t, f.Raw)
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "", f.Body())
	assert.Equal(t, "Subject", f.Raw.Name())
	assert.Equal(t, "", f.Raw.Body())
	assert.Equal(t, "Subject", f.Raw.String())
}
