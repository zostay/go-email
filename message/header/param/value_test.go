package param_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/header/param"
)

func TestParse(t *testing.T) {
	t.Parallel()

	_, err := param.Parse("test:plain")
	assert.Error(t, err)

	mt, err := param.Parse("text")
	assert.NoError(t, err)

	assert.Equal(t, "text", mt.MediaType())
	assert.Equal(t, "", mt.Type())
	assert.Equal(t, "", mt.Subtype())
	assert.Equal(t, "text", mt.Presentation())
	assert.Equal(t, "text", mt.Value())
	assert.Equal(t, map[string]string{}, mt.Parameters())

	mt, err = param.Parse("image/jpeg")
	assert.NoError(t, err)

	assert.Equal(t, "image/jpeg", mt.MediaType())
	assert.Equal(t, "image", mt.Type())
	assert.Equal(t, "jpeg", mt.Subtype())
	assert.Equal(t, map[string]string{}, mt.Parameters())

	mt, err = param.Parse("application/json; charset=UTF-8; foo=bar")
	assert.NoError(t, err)

	assert.Equal(t, "application/json", mt.MediaType())
	assert.Equal(t, "application", mt.Type())
	assert.Equal(t, "json", mt.Subtype())
	assert.Equal(t, map[string]string{
		"charset": "UTF-8",
		"foo":     "bar",
	}, mt.Parameters())
}

func TestNew(t *testing.T) {
	t.Parallel()

	mt := param.New("text/json", map[string]string{
		"charset": "trash",
	})

	assert.Equal(t, "text/json", mt.MediaType())
	assert.Equal(t, "text", mt.Type())
	assert.Equal(t, "json", mt.Subtype())
	assert.Equal(t, map[string]string{"charset": "trash"}, mt.Parameters())
}

func TestModify(t *testing.T) {
	t.Parallel()

	mt := param.New("text/json")
	assert.Equal(t, "text/json", mt.String())

	mt = param.Modify(mt,
		param.Set(param.Boundary, "abc123"),
		param.Change("application/json"),
	)
	assert.Equal(t, "application/json; boundary=abc123", mt.String())

	mt = param.Modify(mt,
		param.Change("text/x-json"),
		param.Set(param.Charset, "utf-8"),
		param.Delete(param.Boundary),
	)
	assert.Equal(t, "text/x-json; charset=utf-8", mt.String())
	assert.Equal(t, []byte("text/x-json; charset=utf-8"), mt.Bytes())
}

func TestValue_Parameter(t *testing.T) {
	t.Parallel()

	mt := param.New("text/plain", map[string]string{
		"boundary": "abc123",
		"charset":  "latin1",
		"blah":     "BLOOP",
	})

	assert.Equal(t, "abc123", mt.Parameter(param.Boundary))
	assert.Equal(t, "abc123", mt.Boundary())
	assert.Equal(t, "latin1", mt.Charset())
	assert.Equal(t, "latin1", mt.Parameter(param.Charset))
	assert.Equal(t, "BLOOP", mt.Parameter("blah"))
	assert.Equal(t, "", mt.Parameter(param.Filename))
	assert.Equal(t, "", mt.Filename())
}
