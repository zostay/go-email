package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderFieldIndex(t *testing.T) {
	t.Parallel()

	h := NewHeader(LF,
		NewHeaderField("Received", "a", []byte(LF)),
		NewHeaderField("Foo", "1", []byte(LF)),
		NewHeaderField("From", "b", []byte(LF)),
		NewHeaderField("Foo", "2", []byte(LF)),
		NewHeaderField("To", "c", []byte(LF)),
		NewHeaderField("Foo", "3", []byte(LF)),
		NewHeaderField("Subject", "d", []byte(LF)),
	)

	assert.Equal(t, 1, h.HeaderFieldIndex("Foo", 0, false))
	assert.Equal(t, 3, h.HeaderFieldIndex("Foo", 1, false))
	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", 2, false))
	assert.Equal(t, -1, h.HeaderFieldIndex("Foo", 3, false))

	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", -1, false))
	assert.Equal(t, 3, h.HeaderFieldIndex("Foo", -2, false))
	assert.Equal(t, 1, h.HeaderFieldIndex("Foo", -3, false))
	assert.Equal(t, -1, h.HeaderFieldIndex("Foo", -4, false))

	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", 3, true))
	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", -1, true))
}
