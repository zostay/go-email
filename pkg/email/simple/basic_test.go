package simple

import (
	"fmt"
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
	junkInHeader        = readTestFile("junk-in-header")
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
	assert.Equal(t, "This header is badly folded because even though it goes onto thesecond line, it has no indent.", m.HeaderGet("Badly-Folded"))
	assert.Equal(t, "Foo", m.HeaderGet("Foo"))
}

func TestFolding(t *testing.T) {
	t.Parallel()

	m, err := Parse(joseyFold)
	assert.NoError(t, err)

	const refs = "<200211120937.JAA28130@xoneweb.opengroup.org> \t<1021112125524.ZM7503@skye.rdg.opengroup.org> \t<3DD221BB.13116D47@sun.com>"
	assert.Equal(t, refs, m.HeaderGet("References"))
	assert.Equal(t, refs, m.HeaderGet("reFerEnceS"))

	var recvd = []string{
		"from mailman.opengroup.org ([192.153.166.9])\tby deep-dark-truthful-mirror.pad with smtp (Exim 3.36 #1 (Debian))\tid 18Buh5-0006Zr-00\tfor <posix@simon-cozens.org>; Wed, 13 Nov 2002 10:24:23 +0000",
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

	h := email.NewHeader(email.LF)
	err := h.HeaderSet("To", to)
	assert.NoError(t, err)

	err = h.HeaderSet("From", from)
	assert.NoError(t, err)

	err = h.HeaderSet("Subject", subject)
	assert.NoError(t, err)

	assert.NotEqual(t, subject, h.HeaderGetField("Subject").String())

	assert.Equal(t, "To: to@example.com\nFrom: from@example.com\nSubject: A A A A A A A A A A A A A A A A A A A A A A A A A A A A A A A A A A\n A A A A A A A A A A A A A A A A \n", h.String())
}

func TestHeaderCase(t *testing.T) {
	t.Parallel()

	m, err := Parse([]byte("Foo-Bar: Baz\n\ntest\n"))
	assert.NoError(t, err)

	err = m.HeaderSet("Foo-bar", "quux")
	assert.NoError(t, err)
	assert.Equal(t, "Foo-Bar: quux\n", m.HeaderGetField("FOO-BAR").String())
}

func TestHeaderJunk(t *testing.T) {
	t.Parallel()

	m, err := Parse(junkInHeader)
	var hpErr *email.HeaderParseError
	if assert.ErrorAs(t, err, &hpErr) {
		assert.Equal(t, 1, len(hpErr.Errs))
		var bsErr *email.BadStartError
		if assert.ErrorAs(t, hpErr.Errs[0], &bsErr) {
			assert.Contains(t, string(bsErr.BadStart), "linden")
		}
	}

	assert.NotContains(t, m.String(), "linden")
}

const mylb = email.LF

func myParseField(f string) *email.HeaderField {
	hf, _ := email.ParseHeaderField([]byte(f+mylb), []byte(mylb))
	hf.Match() // make sure the match is cached
	return hf
}

func myParseFieldNew(f string) *email.HeaderField {
	hf, _ := email.ParseHeaderField([]byte(f+mylb), []byte(mylb))
	return hf
}

func TestHeaderMany(t *testing.T) {
	t.Parallel()

	const emailText = `Alpha: this header comes first
Bravo: this header comes second
Alpha: this header comes third

The body is irrelevant.
`

	m, err := Parse([]byte(emailText))
	assert.NoError(t, err)

	assert.Equal(t, []string{
		"this header comes first",
		"this header comes third",
	}, m.HeaderGetAll("alpha"))

	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: this header comes first"),
		myParseField("Bravo: this header comes second"),
		myParseField("Alpha: this header comes third"),
	}, m.HeaderFields())

	assert.Equal(t, []string{"Alpha", "Bravo"}, m.HeaderNames())

	err = m.HeaderSetAll("Alpha", "header one", "header three")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: header one"),
		myParseField("Bravo: this header comes second"),
		myParseField("Alpha: header three"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("Alpha", "h1", "h3", "h4")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: h1"),
		myParseField("Bravo: this header comes second"),
		myParseField("Alpha: h3"),
		myParseFieldNew("Alpha: h4"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("alpha", "one is the loneliest header")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: one is the loneliest header"),
		myParseField("Bravo: this header comes second"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("Gamma", "gammalon")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: one is the loneliest header"),
		myParseField("Bravo: this header comes second"),
		myParseFieldNew("Gamma: gammalon"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("alpha", "header one", "header omega")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: header one"),
		myParseField("Bravo: this header comes second"),
		myParseField("Gamma: gammalon"),
		myParseFieldNew("alpha: header omega"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("bravo")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: header one"),
		myParseField("Gamma: gammalon"),
		myParseField("alpha: header omega"),
	}, m.HeaderFields())

	err = m.HeaderSetAll("Omega")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseField("Alpha: header one"),
		myParseField("Gamma: gammalon"),
		myParseField("alpha: header omega"),
	}, m.HeaderFields())
}

func TestHeaderNames(t *testing.T) {
	t.Parallel()

	m, err := Parse([]byte(`From: casey@geeknest.com
To: drain@example.com
Subject: Message in a bottle
`))
	assert.NoError(t, err)
	assert.Equal(t, m.HeaderNames(), []string{"From", "To", "Subject"})

	m, err = Parse([]byte(`From: casey@geeknest.com
To: drain@example.com
Subject: Message in a bottle
subject: second subject!

HELP!
`))
	assert.NoError(t, err)
	assert.Equal(t, m.HeaderNames(), []string{"From", "To", "Subject"})

	m, err = Parse([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, m.HeaderNames(), []string{})
}

func TestHeaderFields(t *testing.T) {
	t.Parallel()

	const emailText = `From: casey@geeknest.example.com
X-Your-Face: your face is your face
To: drain@example.com
X-Your-Face: your face is my face
X-Your-Face: from california
Reply-To: xyzzy@plugh.example.net
X-Your-Face: to the new york islface
Subject: Message in a bottle

HELP!
`
	m, err := Parse([]byte(emailText))
	assert.NoError(t, err)
	assert.NotNil(t, m)

	assert.Equal(t, []*email.HeaderField{
		myParseFieldNew("From: casey@geeknest.example.com"),
		myParseFieldNew("X-Your-Face: your face is your face"),
		myParseFieldNew("To: drain@example.com"),
		myParseFieldNew("X-Your-Face: your face is my face"),
		myParseFieldNew("X-Your-Face: from california"),
		myParseFieldNew("Reply-To: xyzzy@plugh.example.net"),
		myParseFieldNew("X-Your-Face: to the new york islface"),
		myParseFieldNew("Subject: Message in a bottle"),
	}, m.HeaderFields())
}

func TestHeaderAddBefore(t *testing.T) {
	const emailText = `Alpha: this header comes first
Bravo: this header comes second
Alpha: this header comes third

The body is irrelevant.
`

	m, err := Parse([]byte(emailText))
	assert.NoError(t, err)

	err = m.HeaderAddBefore("Alpha", "this header comes firstest")
	assert.NoError(t, err)
	assert.Equal(t, []*email.HeaderField{
		myParseFieldNew("Alpha: this header comes firstest"),
		myParseFieldNew("Alpha: this header comes first"),
		myParseFieldNew("Bravo: this header comes second"),
		myParseFieldNew("Alpha: this header comes third"),
	}, m.HeaderFields())

	err = m.HeaderAddBefore("Zero", "and 0+1th")
	assert.NoError(t, err)

	err = m.HeaderAddBefore("Zero", "this header comes zeroth")
	assert.NoError(t, err)

	assert.Equal(t, []*email.HeaderField{
		myParseFieldNew("Zero: this header comes zeroth"),
		myParseFieldNew("Zero: and 0+1th"),
		myParseFieldNew("Alpha: this header comes firstest"),
		myParseFieldNew("Alpha: this header comes first"),
		myParseFieldNew("Bravo: this header comes second"),
		myParseFieldNew("Alpha: this header comes third"),
	}, m.HeaderFields())
}

func TestLineBreakDetection(t *testing.T) {
	crlf := []string{email.CR, email.CRLF, email.LF, email.LFCR}

	for _, lb := range crlf {
		emailText := fmt.Sprintf("Foo-Bar: Baz%s%stest%s", lb, lb, lb)

		m, err := Parse([]byte(emailText))
		assert.NoError(t, err)

		assert.Equal(t, []byte(lb), m.Break())

		emailText = fmt.Sprintf("Foo-Bar: Baz%s", lb)

		m, err = Parse([]byte(emailText))
		assert.NoError(t, err)

		assert.Equal(t, []byte(lb), m.Break())
	}
}
