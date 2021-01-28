package simple

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/pkg/email"
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
	joseyFold           = readTestFile("josey-fold")
	joseyNoFold         = readTestFile("josey-nofold")
	badlyFolded         = readTestFile("badly-folded")
	badlyFoldedNoIndent = readTestFile("badly-folded-noindent")
)

func TestBasic(t *testing.T) {
	t.Parallel()

	assert.NotZero(t, len(joseyNoFold))

	mail, err := Parse(joseyNoFold)
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

	assert.Equal(t, string(joseyNoFold), mail.String())

	const (
		pu = "Previously-Unknown"
		ws = "wonderful species"
	)
	err = mail.HeaderSet(pu, ws)
	assert.NoError(t, err)
	assert.Equal(t, ws, mail.HeaderGet(pu))
}

func TestNastyNewline(t *testing.T) {
	t.Parallel()

	const nasty = "Subject: test\n\rTo: foo\n\r\n\rfoo\n\r"

	mail, err := Parse([]byte(nasty))
	assert.NoError(t, err)

	pos, crlf := SplitHeadFromBody([]byte(nasty))

	assert.Equal(t, 22, pos)
	assert.Equal(t, []byte("\n\r"), crlf)

	assert.Equal(t, nasty, mail.String())
}

func TestBadlyFolded(t *testing.T) {
	t.Parallel()

	m1, err := Parse(badlyFolded)
	assert.NoError(t, err)

	m2, err := Parse([]byte(m1.String()))
	assert.NoError(t, err)

	assert.Equal(t, "CMU Sieve 2.2", m2.HeaderGet("X-Sieve"))
}

func TestBadlyFoldedNoIndent(t *testing.T) {
	t.Parallel()

	m, err := Parse(badlyFoldedNoIndent)
	assert.NoError(t, err)

	assert.Equal(t, "Bar", m.HeaderGet("Bar"))
	assert.Equal(t, "This header is badly folded because even though it goes onto the second line, it has no indent.", m.HeaderGet("Badly-Folded"))
	assert.Equal(t, "Foo", m.HeaderGet("Foo"))
}

func TestFolding(t *testing.T) {
	t.Parallel()

	m, err := Parse(joseyFold)
	assert.NoError(t, err)

	const refs = "<200211120937.JAA28130@xoneweb.opengroup.org>  <1021112125524.ZM7503@skye.rdg.opengroup.org>  <3DD221BB.13116D47@sun.com>"
	assert.Equal(t, refs, m.HeaderGet("References"))
	assert.Equal(t, refs, m.HeaderGet("reFerEnceS"))

	var recvd = []string{
		"from mailman.opengroup.org ([192.153.166.9]) by deep-dark-truthful-mirror.pad with smtp (Exim 3.36 #1 (Debian)) id 18Buh5-0006Zr-00 for <posix@simon-cozens.org>; Wed, 13 Nov 2002 10:24:23 +0000",
		"(qmail 1679 invoked by uid 503); 13 Nov 2002 10:10:49 -0000",
	}
	assert.Equal(t, recvd, m.HeaderGetAll("Received"))
}

func TestFoldingHeaderFormatting(t *testing.T) {
	t.Parallel()

	const text = `Fold-1: 1
 2 3
Fold-2: 0
 1 2

Body
`

	m, err := Parse([]byte(text))
	assert.NoError(t, err)
	assert.Equal(t, "0 1 2", m.HeaderGet("Fold-2"))
}

func TestFoldingLongLine(t *testing.T) {
	t.Parallel()

	const (
		to   = "to@example.com"
		from = "from@example.com"
	)

	subject := strings.Repeat("A ", 50)

	var h email.Header
	h.HeaderSet("To", to)
	h.HeaderSet("From", from)
	h.HeaderSet("Subject", subject)

	assert.NotEqual(t, subject, h.HeaderGetField("Subject").String())
}