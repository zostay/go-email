package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	t.Parallel()

	m := NewMessage(
		NewHeader(LF,
			NewHeaderField("Subject", "test", []byte(LF))),
		[]byte("This is a message."),
	)

	expected := `Subject: test

This is a message.`

	assert.NotNil(t, m)
	assert.Equal(t, expected, m.String())
	assert.Equal(t, []byte(expected), m.Bytes())
}
