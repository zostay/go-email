package mime

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
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

func TestBasic7bitEncoding(t *testing.T) {
	t.Parallel()

	const emailFmt = `Content-Transfer-Encoding: %s
Content-Type: text/plain

Hello World!
I like you!
`

	encodes := []string{"7bit", "7-bit"}
	for _, encode := range encodes {
		emailMsg := fmt.Sprintf(emailFmt, encode)
		m, err := Parse([]byte(emailMsg))
		assert.NoError(t, err)

		assert.Equal(t, "Hello World!\nI like you!\n", m.ContentString())

		c, err := m.ContentUnicode()
		assert.NoError(t, err)
		assert.Equal(t, "Hello World!\nI like you!\n", c)
		assert.Equal(t, encode, m.HeaderGet("Content-Transfer-Encoding"))
	}
}

func TestSetEncoding(t *testing.T) {
	const emailMsg = `Content-Transfer-Encoding: 7bit
Content-Type: text/plain

Hello World!
I like you!
`

	m, err := Parse([]byte(emailMsg))
	assert.NoError(t, err)

	assert.Equal(t, "Hello World!\nI like you!\n", m.ContentString())

	c, err := m.ContentUnicode()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!\nI like you!\n", c)

	err = m.SetContentTransferEncoding("base64")
	assert.NoError(t, err)
	assert.Equal(t,
		"SGVsbG8gV29ybGQhCkkgbGlrZSB5b3UhCg==",
		m.ContentString(),
	)

	c, err = m.ContentUnicode()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!\nI like you!\n", c)

	err = m.SetContentTransferEncoding("binary")
	assert.NoError(t, err)
	assert.Equal(t,
		"Hello World!\nI like you!\n",
		m.ContentString(),
	)

	c, err = m.ContentUnicode()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!\nI like you!\n", c)

	err = m.SetContentTransferEncoding("quoted-printable")
	assert.NoError(t, err)
	assert.Equal(t,
		"Hello World!\r\nI like you!\r\n",
		m.ContentString(),
	)

	c, err = m.ContentUnicode()
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!\r\nI like you!\r\n", c)

	err = m.SetContentUnicode(strings.Repeat("Long line! ", 100))
	assert.NoError(t, err)
	assert.Equal(t,
		strings.ReplaceAll(`Long line! Long line! Long line! Long line! Long line! Long line! Long line=
! Long line! Long line! Long line! Long line! Long line! Long line! Long li=
ne! Long line! Long line! Long line! Long line! Long line! Long line! Long =
line! Long line! Long line! Long line! Long line! Long line! Long line! Lon=
g line! Long line! Long line! Long line! Long line! Long line! Long line! L=
ong line! Long line! Long line! Long line! Long line! Long line! Long line!=
 Long line! Long line! Long line! Long line! Long line! Long line! Long lin=
e! Long line! Long line! Long line! Long line! Long line! Long line! Long l=
ine! Long line! Long line! Long line! Long line! Long line! Long line! Long=
 line! Long line! Long line! Long line! Long line! Long line! Long line! Lo=
ng line! Long line! Long line! Long line! Long line! Long line! Long line! =
Long line! Long line! Long line! Long line! Long line! Long line! Long line=
! Long line! Long line! Long line! Long line! Long line! Long line! Long li=
ne! Long line! Long line! Long line! Long line! Long line! Long line! Long =
line! Long line! Long line! Long line! Long line!=20`, "\n", "\r\n"),
		m.ContentString(),
	)

	c, err = m.ContentUnicode()
	assert.NoError(t, err)
	assert.Equal(t, strings.Repeat("Long line! ", 100), c)
}

func TestParseMultipart(t *testing.T) {
	t.Parallel()

	const emailMsg = `Subject: hello
Content-Type: multipart/mixed; boundary="0"

Prelude

--0
Content-Type: text/plain

This is plain text.
--0--

Postlude
`

	m, err := Parse([]byte(emailMsg))
	assert.NoError(t, err)

	assert.Equal(t, emailMsg, m.String())

	assert.Equal(t, []byte("Prelude\n"), m.Preamble)
	assert.Equal(t, []byte("\n--0--\n\nPostlude\n"), m.Epilogue)

	if assert.Equal(t, 1, len(m.Parts)) {
		p := m.Parts[0]

		assert.Equal(t, "Content-Type: text/plain\n\nThis is plain text.", p.String())

		assert.Equal(t, 0, len(p.Parts))
		assert.Equal(t, "This is plain text.", p.ContentString())

		c, err := p.ContentUnicode()
		assert.NoError(t, err)
		assert.Equal(t, "This is plain text.", c)
	}
}
