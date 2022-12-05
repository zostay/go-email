package mime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMediaType(t *testing.T) {
	t.Parallel()

	_, err := ParseMediaType("text:plain")
	assert.Error(t, err)

	mt, err := ParseMediaType("text")
	assert.NoError(t, err)

	assert.Equal(t, "text", mt.MediaType())
	assert.Equal(t, "", mt.Type())
	assert.Equal(t, "", mt.Subtype())
	assert.Equal(t, map[string]string{}, mt.Parameters())

	mt, err = ParseMediaType("image/jpeg")
	assert.NoError(t, err)

	assert.Equal(t, "image/jpeg", mt.MediaType())
	assert.Equal(t, "image", mt.Type())
	assert.Equal(t, "jpeg", mt.Subtype())
	assert.Equal(t, map[string]string{}, mt.Parameters())

	mt, err = ParseMediaType("application/json; charset=UTF-8; foo=bar")
	assert.NoError(t, err)

	assert.Equal(t, "application/json", mt.MediaType())
	assert.Equal(t, "application", mt.Type())
	assert.Equal(t, "json", mt.Subtype())
	assert.Equal(t, map[string]string{
		"charset": "UTF-8",
		"foo":     "bar",
	}, mt.Parameters())
}

func TestNewMediaType(t *testing.T) {
	t.Parallel()

	mt, err := NewMediaType("text/json", "charset", "trash")
	assert.NoError(t, err)

	assert.Equal(t, "text/json", mt.MediaType())
	assert.Equal(t, "text", mt.Type())
	assert.Equal(t, "json", mt.Subtype())
	assert.Equal(t, map[string]string{"charset": "trash"}, mt.Parameters())

	_, err = NewMediaType("text/json", "charset")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "odd number")
}

func TestNewMediaTypeMapa(t *testing.T) {
	t.Parallel()

	mt := NewMediaTypeMap("text/html", map[string]string{"zip": "zap"})

	assert.Equal(t, "text/html", mt.MediaType())
	assert.Equal(t, "text", mt.Type())
	assert.Equal(t, "html", mt.Subtype())
	assert.Equal(t, map[string]string{"zip": "zap"}, mt.Parameters())
	assert.Equal(t, "zap", mt.Parameter("zip"))
}
