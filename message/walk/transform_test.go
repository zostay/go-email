package walk_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/header"
	"github.com/zostay/go-email/v2/message/walk"
)

const complexMsgBase64 = `To: sterling@example.com
From: sterling@example.com
Subject: Hello World
Content-type: multipart/mixed; boundary=__boundary-one__

--__boundary-one__
Content-type: multipart/alternate; boundary=__boundary-two__

--__boundary-two__
Content-type: text/html
Content-transfer-encoding: base64

SGVsbG8gV29ybGQh
--__boundary-two__
Content-type: text/plain
Content-transfer-encoding: base64

SGVsbG8gV29ybGQh
--__boundary-two__--
--__boundary-one__
Content-type: application/pdf
Content-disposition: attachment; filename=micro.pdf
Content-transfer-encoding: base64

JVBERi0xLgp0cmFpbGVyPDwvUm9vdDw8L1BhZ2VzPDwvS2lkc1s8PC9NZWRpYUJveFswIDAgMyAz
XT4+XT4+Pj4+Pg==
--__boundary-one__
Content-type: application/image
Content-disposition: attachment; filename=att-1.gif
Content-transfer-encoding: base64

R0lGODlhDAAMAPcAAAAAAAgICBAQEBgYGCkpKTExMTk5OUpKSoyMjJSUlJycnKWlpbW1tc7O
zufn5+/v7/f39///////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////
/////////////////////////////////ywAAAAADAAMAAAIXwAjRICQwIAAAQYUQBAYwUEB
AAACEIBYwMHAhxARNIAIoAAEBBAPOICwkSMCjBAXlKQYgCMABSsjtuQI02UAlC9jFgBJMyYC
CCgRMODoseFElx0tCvxYIEAAAwkWRggIADs=
--__boundary-one__--`

func TestAndTransform_base64(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(complexMsg))
	assert.NoError(t, err)

	// turn every part into a lone opaque, so we get them all at the top
	tm, err := walk.AndTransform(func(part message.Part, parents []message.Part, state []any) (any, error) {
		var buf *message.Buffer
		var err error
		if part.IsMultipart() {
			buf = message.NewBlankBuffer(part)
		} else {
			buf, err = message.NewBuffer(part)
			if err != nil {
				return nil, err
			}

			te, err := buf.GetTransferEncoding()
			if te == "" || errors.Is(err, header.ErrNoSuchField) {
				buf.SetTransferEncoding("base64")
				buf.SetEncoded(false) // cause encoding on WriteTo()
			} else if err != nil {
				return nil, err
			}
		}

		if len(state) > 0 {
			state[len(state)-1].(*message.Buffer).Add(buf)
		}

		return buf, nil
	}, m)

	assert.NoError(t, err)
	assert.IsType(t, &message.Buffer{}, tm)

	buf := &bytes.Buffer{}
	_, err = tm.(message.Part).WriteTo(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte(complexMsgBase64), buf.Bytes())
}
