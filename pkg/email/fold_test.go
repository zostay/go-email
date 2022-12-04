package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnfoldValue(t *testing.T) {
	uv := UnfoldValue([]byte("folded\n line"))
	assert.Equal(t, []byte("folded line"), uv)
}

func TestFoldValue(t *testing.T) {
	fv := FoldValue([]byte("this is a very long value that will need to be folded after some number of characters because long lines aren't allowed in email"), []byte(CRLF))
	assert.Equal(t, []byte("this is a very long value that will need to be folded after some number of\r\n characters because long lines aren't allowed in email\r\n"), fv)

	fv = FoldValue([]byte("thisisaverylongvaluethatislackinginspacesforaverylongtimebutdoeseventuallyaddaspacesomewhereso itcanbeusedforfolding"), []byte(CRLF))
	assert.Equal(t, []byte("thisisaverylongvaluethatislackinginspacesforaverylongtimebutdoeseventuallyaddaspacesomewhereso\r\n itcanbeusedforfolding\r\n"), fv)

	fv = FoldValue([]byte(strings.Repeat("foldit", 200)), []byte(CRLF))
	assert.Equal(t, []byte(strings.Repeat(strings.Repeat("foldit", 13)+"\r\n ", 15)+"folditfolditfolditfolditfoldit\r\n"), fv)

	fv = FoldValue([]byte("thisisaverylongvaluethatislackinginspacesforaverylongtimebutnotsolongthatwehavetofolditsoleaveitasis"), []byte(CRLF))
	assert.Equal(t, []byte("thisisaverylongvaluethatislackinginspacesforaverylongtimebutnotsolongthatwehavetofolditsoleaveitasis\r\n"), fv)
}
