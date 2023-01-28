package walk_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/walk"
)

// special thanks to plinth:
// https://stackoverflow.com/questions/17279712/what-is-the-smallest-possible-valid-pdf
// (micro-PDF pulled from that link 2023-01-28)
const complexMsg = `To: sterling@example.com
From: sterling@example.com
Subject: Hello World
Content-type: multipart/mixed; boundary=__boundary-one__

--__boundary-one__
Content-type: multipart/alternate; boundary=__boundary-two__

--__boundary-two__
Content-type: text/html

Hello World!
--__boundary-two__
Content-type: text/plain

Hello World!
--__boundary-two__--
--__boundary-one__
Content-type: application/pdf
Content-disposition: attachment; filename=micro.pdf

%PDF-1.
trailer<</Root<</Pages<</Kids[<</MediaBox[0 0 3 3]>>]>>>>>>
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

func TestAndProcess(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(complexMsg))
	assert.NoError(t, err)

	counts := make([]int, 10)
	err = walk.AndProcess(
		func(part message.Part, parents []message.Part) error {
			count := counts[len(parents)]
			switch {
			case len(parents) == 0 && count == 0:
				assert.True(t, part.IsMultipart())
				assert.Len(t, part.GetParts(), 3)

				s, err := part.GetHeader().GetSubject()
				assert.NoError(t, err)
				assert.Equal(t, "Hello World", s)
			case len(parents) == 1 && count == 0:
				assert.True(t, part.IsMultipart())
				assert.Len(t, part.GetParts(), 2)
			case len(parents) == 1 && count == 1:
				assert.False(t, part.IsMultipart())

				fn, err := part.GetHeader().GetFilename()
				assert.NoError(t, err)
				assert.Equal(t, "micro.pdf", fn)
			case len(parents) == 1 && count == 2:
				assert.False(t, part.IsMultipart())

				fn, err := part.GetHeader().GetFilename()
				assert.NoError(t, err)
				assert.Equal(t, "att-1.gif", fn)
			case len(parents) == 2 && count == 0:
				assert.False(t, part.IsMultipart())

				mt, err := part.GetHeader().GetMediaType()
				assert.NoError(t, err)
				assert.Equal(t, "text/html", mt)
			case len(parents) == 2 && count == 1:
				assert.False(t, part.IsMultipart())

				mt, err := part.GetHeader().GetMediaType()
				assert.NoError(t, err)
				assert.Equal(t, "text/plain", mt)
			default:
				assert.Fail(t, "Unexpected part processed")
			}

			counts[len(parents)]++
			return nil
		}, m,
	)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 3, 2, 0, 0, 0, 0, 0, 0, 0}, counts)
}

func TestAndProcessOpaque(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(complexMsg))
	assert.NoError(t, err)

	counts := make([]int, 10)
	err = walk.AndProcessOpaque(
		func(part message.Part, parents []message.Part) error {
			count := counts[len(parents)]
			switch {
			case len(parents) == 1 && count == 0:
				assert.False(t, part.IsMultipart())

				fn, err := part.GetHeader().GetFilename()
				assert.NoError(t, err)
				assert.Equal(t, "micro.pdf", fn)
			case len(parents) == 1 && count == 1:
				assert.False(t, part.IsMultipart())

				fn, err := part.GetHeader().GetFilename()
				assert.NoError(t, err)
				assert.Equal(t, "att-1.gif", fn)
			case len(parents) == 2 && count == 0:
				assert.False(t, part.IsMultipart())

				mt, err := part.GetHeader().GetMediaType()
				assert.NoError(t, err)
				assert.Equal(t, "text/html", mt)
			case len(parents) == 2 && count == 1:
				assert.False(t, part.IsMultipart())

				mt, err := part.GetHeader().GetMediaType()
				assert.NoError(t, err)
				assert.Equal(t, "text/plain", mt)
			default:
				assert.Fail(t, "Unexpected part processed")
			}

			counts[len(parents)]++
			return nil
		}, m,
	)

	assert.NoError(t, err)
	assert.Equal(t, []int{0, 2, 2, 0, 0, 0, 0, 0, 0, 0}, counts)
}

func TestAndProcessMultipart(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(complexMsg))
	assert.NoError(t, err)

	counts := make([]int, 10)
	err = walk.AndProcessMultipart(
		func(part message.Part, parents []message.Part) error {
			count := counts[len(parents)]
			switch {
			case len(parents) == 0 && count == 0:
				assert.True(t, part.IsMultipart())
				assert.Len(t, part.GetParts(), 3)

				s, err := part.GetHeader().GetSubject()
				assert.NoError(t, err)
				assert.Equal(t, "Hello World", s)
			case len(parents) == 1 && count == 0:
				assert.True(t, part.IsMultipart())
				assert.Len(t, part.GetParts(), 2)
			default:
				assert.Fail(t, "Unexpected part processed")
			}

			counts[len(parents)]++
			return nil
		}, m,
	)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0}, counts)
}

type testError struct{}

func (testError) Error() string { return "I'm a little teapot." }

func TestAndProcess_Error(t *testing.T) {
	t.Parallel()

	m, err := message.Parse(strings.NewReader(complexMsg))
	assert.NoError(t, err)

	runs := 0
	err = walk.AndProcess(
		func(part message.Part, parents []message.Part) error {
			runs++
			return testError{}
		},
		m,
	)

	assert.ErrorIs(t, err, testError{})
	assert.Equal(t, 1, runs)
}
