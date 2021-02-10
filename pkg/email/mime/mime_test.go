package mime

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readTestFile(n string) []byte {
	gopath := os.Getenv("GOPATH")
	apppath := path.Join(gopath, "src/github.com/zostay/go-email")
	tdpath := path.Join(apppath, "test")
	tfpath := path.Join(tdpath, n)
	data, _ := ioutil.ReadFile(tfpath)
	return data
}

var (
	mail1   = readTestFile("mail-1")
	att1Gif = readTestFile("att-1.gif")
)

func TestBasic(t *testing.T) {
	m, err := Parse(mail1)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	assert.Empty(t, m.Parts)
	assert.NotEmpty(t, m.Content())

	bb, err := m.ContentBinary()
	assert.NoError(t, err)

	enc := base64.StdEncoding
	binaryWant := make([]byte, enc.DecodedLen(len(m.Content())))
	n, err := enc.Decode(binaryWant, m.Content())
	assert.NoError(t, err)
	binaryWant = binaryWant[:n]
	assert.Equal(t, binaryWant, bb)

	assert.Equal(t, "1.gif", m.HeaderContentDispositionFilename())
	assert.Equal(t, att1Gif, bb)

	assert.Equal(t, []string{"one", "two", "three"}, m.HeaderGetAll("X-MultiHeader"))
}
