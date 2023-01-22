package transfer_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/transfer"
)

const asisString = `1234567890-=
~!@#$%^&*()_+
qwertyuiop[]\
QWERTYUIOP{}|
asdfghjkl;'
ASDFGHJKL:"
zxcvbnm,./
ZXCVBNM<>?
 
` + "\x80\x90\xa0\xb0\xc0\xd0\xe0\xf0\xff\r\n\t\b"

func TestNewAsIsDecoder(t *testing.T) {
	t.Parallel()

	r := strings.NewReader(asisString)
	ad := transfer.NewAsIsDecoder(r)
	db, err := io.ReadAll(ad)
	assert.NoError(t, err)
	assert.Equal(t, []byte(asisString), db)
}

func TestNewAsIsEncoder(t *testing.T) {
	t.Parallel()

	w := &bytes.Buffer{}
	ae := transfer.NewAsIsEncoder(w)
	n, err := ae.Write([]byte(asisString))
	assert.Equal(t, len(asisString), n)
	assert.NoError(t, err)
	assert.Equal(t, []byte(asisString), w.Bytes())
}
