package field_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/header/field"
)

func TestEncode(t *testing.T) {
	s := field.Encode("⚀⚁⚂⚃⚄⚅")
	assert.Equal(t, "=?utf-8?b?4pqA4pqB4pqC4pqD4pqE4pqF?=", s)
}

func TestDecode(t *testing.T) {
	s, err := field.Decode("=?utf-8?b?4pqA4pqB4pqC4pqD4pqE4pqF?=")
	assert.NoError(t, err)
	assert.Equal(t, "⚀⚁⚂⚃⚄⚅", s)
}
