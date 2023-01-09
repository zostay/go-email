package field

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRaw(t *testing.T) {
	f := &Raw{[]byte("Subject: test"), 7}
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, " test", f.Body())
	assert.Equal(t, []byte("Subject: test"), f.Bytes())
	assert.Equal(t, "Subject: test", f.String())

	f = &Raw{[]byte("Subject"), 7}
	assert.Equal(t, "Subject", f.Name())
	assert.Equal(t, "", f.Body())
	assert.Equal(t, []byte("Subject"), f.Bytes())
	assert.Equal(t, "Subject", f.String())
}
