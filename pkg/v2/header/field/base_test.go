package field_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/pkg/v2/header/field"
)

func TestBase(t *testing.T) {
	b := &field.Base{}

	b.SetName("subject")
	b.SetBody("test")

	assert.Equal(t, "subject", b.Name())
	assert.Equal(t, "test", b.Body())
	assert.Equal(t, "subject: test", b.String())
	assert.Equal(t, []byte("subject: test"), b.Bytes())

	b.SetBody("☺")

	assert.Equal(t, "☺", b.Body())
	assert.Equal(t, "subject: =?utf-8?b?4pi6?=", b.String())
}
