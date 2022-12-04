package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var content = []byte("Test email.")

func TestNewBody(t *testing.T) {
	b := NewBody(content)
	require.NotNil(t, b)

	assert.Equal(t, "Test email.", b.ContentString())
	assert.Equal(t, []byte("Test email."), b.Content())

	assert.Equal(t, "Test email.", b.String())
	assert.Equal(t, []byte("Test email."), b.Bytes())

	b.SetContent([]byte("Another test."))
	assert.Equal(t, "Another test.", b.ContentString())
	assert.Equal(t, []byte("Another test."), b.Content())

	b.SetContentString("And more test.")
	assert.Equal(t, "And more test.", b.ContentString())
	assert.Equal(t, []byte("And more test."), b.Content())
}
