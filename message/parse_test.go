package message_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/header"
	"github.com/zostay/go-email/v2/message/header/field"
	"github.com/zostay/go-email/v2/message/transfer"
)

func TestParse_WithBadlyFolded(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/badly-folded")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}

func TestParse_WithBadlyFoldedNoIndent(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/badly-folded-noindent")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	bf, err := m.GetHeader().Get("Badly-Folded")
	assert.NoError(t, err)
	assert.Equal(t, "This header is badly folded because even though it goes onto thesecond line, it has no indent.", bf)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}

func TestParse_WithJoseyFold(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/josey-fold")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}

func TestParse_WithJoseyNoFold(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/josey-nofold")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}

func TestParse_WithJunkInHeader(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/junk-in-header")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	const expectedJunk = "linden boulevard represent, represent\n"

	m, err := message.Parse(src)
	var badStartErr *field.BadStartError
	assert.ErrorAs(t, err, &badStartErr)
	assert.Equal(t, []byte(expectedJunk), badStartErr.BadStart)
	require.NotNil(t, m)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)-len(expectedJunk)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes[len(expectedJunk):], buf.Bytes())
}

func TestParse_WithMail1_HeaderBody(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/mail-1")
	assert.NoError(t, err)

	m, err := message.Parse(src, message.DecodeTransferEncoding())
	assert.NoError(t, err)

	om, isOpaque := m.(*message.Opaque)
	require.True(t, isOpaque)

	// the colon keeps it from being seen as a bad start and is expected
	// behavior
	weirdHeader, err := om.Get("from simon@simon-cozens.org tue apr 15 20")
	assert.NoError(t, err)
	assert.Equal(t, "24:47 2003", weirdHeader)

	xmh, err := om.GetAll("x-multiheader")
	assert.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, xmh)

	expectFrom, err := addr.ParseEmailAddressList("Simon Cozens <simon@simon-cozens.org>")
	assert.NoError(t, err)

	from, err := om.GetFrom()
	assert.NoError(t, err)
	assert.Equal(t, expectFrom, from)

	to, err := om.Get(header.To)
	assert.NoError(t, err)
	assert.Equal(t, "test", to)

	expectToAddr, err := addr.NewMailboxParsed("", addr.NewAddrSpecParsed("test", "", "test"), "", "test")
	assert.NoError(t, err)

	// flexible parsing of addr
	toAddr, err := om.GetTo()
	assert.NoError(t, err)
	assert.Equal(t, addr.AddressList{expectToAddr}, toAddr)

	expectBcc, err := addr.ParseEmailAddressList("simon@twingle.net")
	assert.NoError(t, err)

	bcc, err := om.GetBcc()
	assert.NoError(t, err)
	assert.Equal(t, expectBcc, bcc)

	mv, err := om.Get("mime-version")
	assert.NoError(t, err)
	assert.Equal(t, mv, "1.0")

	mt, err := om.GetMediaType()
	assert.NoError(t, err)
	assert.Equal(t, "image/gif", mt)

	p, err := om.GetPresentation()
	assert.NoError(t, err)
	assert.Equal(t, "attachment", p)

	fn, err := om.GetFilename()
	assert.NoError(t, err)
	assert.Equal(t, "1.gif", fn)

	te, err := om.GetTransferEncoding()
	assert.NoError(t, err)
	assert.Equal(t, transfer.Base64, te)

	xos, err := om.Get("x-operating-system")
	assert.NoError(t, err)
	assert.Equal(t, "Linux deep-dark-truthful-mirror 2.4.9", xos)

	xp, err := om.Get("x-pom")
	assert.NoError(t, err)
	assert.Equal(t, "The Moon is Waxing Gibbous (98% of Full)", xp)

	xa, err := om.Get("x-addresses")
	assert.NoError(t, err)
	assert.Equal(t, "The simon@cozens.net address is deprecated due to being broken. simon@brecon.co.uk still works, but simon-cozens.org or netthink.co.uk are preferred.", xa)

	xmFcc, err := om.Get("x-mutt-fcc")
	assert.NoError(t, err)
	assert.Equal(t, "=outbox-200304", xmFcc)

	status, err := om.Get("status")
	assert.NoError(t, err)
	assert.Equal(t, "RO", status)

	cl, err := om.Get("content-length")
	assert.NoError(t, err)
	assert.Equal(t, "1205", cl)

	lines, err := om.Get("lines")
	assert.NoError(t, err)
	assert.Equal(t, "17", lines)

	expectBin, err := os.ReadFile("../test/data/att-1.gif")
	assert.NoError(t, err)

	content, err := io.ReadAll(om)
	assert.NoError(t, err)
	assert.Equal(t, expectBin, content)
}

func TestParse_WithMail1_RoundTrip(t *testing.T) {
	t.Parallel()

	src, err := os.Open("../test/data/mail-1")
	assert.NoError(t, err)

	srcBytes, err := io.ReadAll(src)
	assert.NoError(t, err)

	_, err = src.Seek(0, 0)
	assert.NoError(t, err)

	m, err := message.Parse(src)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	n, err := m.WriteTo(buf)
	assert.Equal(t, n, int64(len(srcBytes)))
	assert.NoError(t, err)

	assert.Equal(t, srcBytes, buf.Bytes())
}
