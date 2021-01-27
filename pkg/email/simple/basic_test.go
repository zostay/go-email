package simple

import (
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

var fc = readTestFile("josey-nofold")

func TestBasic(t *testing.T) {
	t.Parallel()

	assert.NotZero(t, len(fc))

	mail, err := Parse(fc)
	if !assert.NoError(t, err) && assert.NotNil(t, mail) {
		return
	}

	oldFrom := mail.HeaderGet("From")
	assert.Equal(t,
		"Andrew Josey <ajosey@rdg.opengroup.org>",
		mail.HeaderGet("From"))

	const newFrom = "Simon Cozens <simon@cpan.org>"
	err = mail.HeaderSet("From", newFrom)
	assert.NoError(t, err)
	assert.Equal(t, newFrom, mail.HeaderGet("From"))

	err = mail.HeaderSet("From", oldFrom)
	assert.NoError(t, err)

	assert.Equal(t, "", mail.HeaderGet("Bogus"))

	assert.Contains(t, mail.BodyString(), "Austin Group Chair")

	oldBody := mail.Body()

	const hi = "Hi there!\n"
	mail.SetBodyString(hi)
	assert.Equal(t, hi, mail.BodyString())

	mail.SetBody(oldBody)

	assert.Equal(t, string(fc), mail.String())

	const (
		pu = "Previously-Unknown"
		ws = "wonderful species"
	)
	err = mail.HeaderSet(pu, ws)
	assert.NoError(t, err)
	assert.Equal(t, ws, mail.HeaderGet(pu))
}

func TestNastyNewline(t *testing.T) {
	const nasty = "Subject: test\n\rTo: foo\n\r\n\rfoo\n\r"

	mail, err := Parse([]byte(nasty))
	assert.NoError(t, err)

	pos, crlf := SplitHeadFromBody([]byte(nasty))

	assert.Equal(t, 22, pos)
	assert.Equal(t, []byte("\n\r"), crlf)

	assert.Equal(t, nasty, mail.String())
}
