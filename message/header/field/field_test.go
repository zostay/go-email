package field_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/header/field"
)

func TestNew(t *testing.T) {
	t.Parallel()

	f := field.New("Subject", "testing")

	assert.Equal(t, "Subject: testing", f.String())
	assert.Equal(t, []byte("Subject: testing"), f.Bytes())
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "testing", f.Body())

	f.SetName("X-Subject")
	assert.Equal(t, "X-Subject: testing", f.String())
	assert.Equal(t, []byte("X-Subject: testing"), f.Bytes())
	assert.Equal(t, "X-Subject", f.Name())
	assert.Equal(t, "testing", f.Body())

	f.SetBody("foo bar baz")
	assert.Equal(t, "X-Subject: foo bar baz", f.String())
	assert.Equal(t, []byte("X-Subject: foo bar baz"), f.Bytes())
	assert.Equal(t, "X-Subject", f.Name())
	assert.Equal(t, "foo bar baz", f.Body())

	f.SetRaw([]byte("sUBJECT: TESTING"))
	assert.Equal(t, "sUBJECT: TESTING", f.String())
	assert.Equal(t, []byte("sUBJECT: TESTING"), f.Bytes())
	assert.Equal(t, "X-Subject", f.Name())
	assert.Equal(t, "foo bar baz", f.Body())

	f.SetName("Subject")
	assert.Equal(t, "Subject: foo bar baz", f.String())
	assert.Equal(t, []byte("Subject: foo bar baz"), f.Bytes())
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "foo bar baz", f.Body())

	f.SetRaw([]byte("oopsie poopsie"))
	assert.Equal(t, "oopsie poopsie", f.String())
	assert.Equal(t, []byte("oopsie poopsie"), f.Bytes())
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "foo bar baz", f.Body())
}
