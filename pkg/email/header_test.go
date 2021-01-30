package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderParse(t *testing.T) {
	t.Parallel()

	const headerStr = `Foo: 1
Foo: 2
Foo: 3
Bar: 3
Baz: 1
`

	h, err := ParseHeaderLB([]byte(headerStr), []byte(LinuxLineBreak))
	assert.NoError(t, err)
	assert.NotNil(t, h)

	assert.Equal(t, "1", h.HeaderGet("foo"))

	b, err := h.HeaderGetN("foo", 0)
	assert.NoError(t, err)
	assert.Equal(t, "1", b)

	b, err = h.HeaderGetN("foo", 1)
	assert.NoError(t, err)
	assert.Equal(t, "2", b)

	b, err = h.HeaderGetN("foo", 2)
	assert.NoError(t, err)
	assert.Equal(t, "3", b)

	b, err = h.HeaderGetN("foo", 3)
	assert.Error(t, err)
	assert.Equal(t, "", b)

	b, err = h.HeaderGetN("foo", -1)
	assert.NoError(t, err)
	assert.Equal(t, "3", b)

	b, err = h.HeaderGetN("foo", -3)
	assert.NoError(t, err)
	assert.Equal(t, "1", b)

	b, err = h.HeaderGetN("foo", -4)
	assert.Error(t, err)
	assert.Equal(t, "", b)

	assert.Equal(t, []string{"1", "2", "3"}, h.HeaderGetAll("Foo"))
}
