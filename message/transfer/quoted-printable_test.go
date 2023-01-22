package transfer_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/transfer"
)

// we only need to test that qp is being applied, not that the encoding is
// working correctly... we'll trust the golang core team to have done that
// already

var qpEnc = []byte("=3D>?")
var qpDec = []byte{0x3d, 0x3e, 0x3f}

func TestNewQuotedPrintableDecoder(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader(qpEnc)
	qpdr := transfer.NewQuotedPrintableDecoder(r)
	db, err := io.ReadAll(qpdr)
	assert.NoError(t, err)
	assert.Equal(t, qpDec, db)
}

func TestNewQuotedPrintableEncoder(t *testing.T) {
	t.Parallel()

	w := &bytes.Buffer{}
	qpewc := transfer.NewQuotedPrintableEncoder(w)
	n, err := qpewc.Write(qpDec)
	assert.Equal(t, len(qpDec), n)
	assert.NoError(t, err)

	err = qpewc.Close()
	assert.NoError(t, err)

	assert.Equal(t, qpEnc, w.Bytes())
}
