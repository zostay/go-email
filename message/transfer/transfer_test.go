package transfer_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message/header"
	"github.com/zostay/go-email/v2/message/transfer"
)

const dec = `1 Timothy 6:10 - For the love of money is a root of all kinds of evils. It is through this craving that some have wandered away from the faith and pierced themselves with many pangs.`
const enc = `MSBUaW1vdGh5IDY6MTAgLSBGb3IgdGhlIGxvdmUgb2YgbW9uZXkgaXMgYSByb290IG9mIGFsbCBr
aW5kcyBvZiBldmlscy4gSXQgaXMgdGhyb3VnaCB0aGlzIGNyYXZpbmcgdGhhdCBzb21lIGhhdmUg
d2FuZGVyZWQgYXdheSBmcm9tIHRoZSBmYWl0aCBhbmQgcGllcmNlZCB0aGVtc2VsdmVzIHdpdGgg
bWFueSBwYW5ncy4=`

func TestApplyTransferDecoding(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetTransferEncoding(transfer.Base64)

	r := strings.NewReader(enc)
	tdr := transfer.ApplyTransferDecoding(h, r)
	tdb, err := io.ReadAll(tdr)
	assert.NoError(t, err)
	assert.Equal(t, []byte(dec), tdb)
}

func TestApplyTransferEncoding(t *testing.T) {
	t.Parallel()

	h := &header.Header{}
	h.SetTransferEncoding(transfer.Base64)

	w := &bytes.Buffer{}
	tdwc := transfer.ApplyTransferEncoding(h, w)
	n, err := tdwc.Write([]byte(dec))
	assert.Equal(t, len(dec), n)
	assert.NoError(t, err)

	err = tdwc.Close()
	assert.NoError(t, err)

	assert.Equal(t, []byte(enc), w.Bytes())
}
