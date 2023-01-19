package walker_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/walker"
)

const msg = `X-Where: A
Content-type: multipart/mixed; boundary=aaaaaaa

--aaaaaaa
X-Where: B
Content-type: multipart/mixed; boundary=bbbbbbb

--bbbbbbb
X-Where: E
Content-type: text/plain

--bbbbbbb
X-Where: F
Content-type: text/plain

--bbbbbbb--
--aaaaaaa
X-Where: C
Content-type: multipart/mixed; boundary=ccccccc

--ccccccc
X-Where: G
Content-type: text/plain

--ccccccc
X-Where: H
Content-type: text/plain

--ccccccc--
--aaaaaaa
X-Where: D
Content-type: multipart/mixed; boundary=ddddddd

--ddddddd
X-Where: I
Content-type: text/plain

--ddddddd
X-Where: J
Content-type: text/plain

--ddddddd--
--aaaaaaa--
`

func TestPartWalker_Walk(t *testing.T) {
	t.Parallel()

	msg, err := message.Parse(strings.NewReader(msg))
	assert.NoError(t, err)

	expectOrder := []string{"A", "B", "E", "F", "C", "G", "H", "D", "I", "J"}
	expectDepth := []int{0, 1, 2, 2, 1, 2, 2, 1, 2, 2}
	expectIndex := []int{0, 0, 0, 1, 1, 0, 1, 2, 0, 1}
	i := 0
	var pw walker.Parts = func(depth, j int, msg message.Part) error {
		where, err := msg.GetHeader().Get("X-Where")
		assert.NoError(t, err)
		assert.Equal(t, expectOrder[i], where)
		assert.Equal(t, expectDepth[i], depth)
		assert.Equal(t, expectIndex[i], j)
		i++
		return nil
	}

	err = pw.Walk(msg)
	assert.NoError(t, err)
}

func TestPartWalker_WalkOpaque(t *testing.T) {
	t.Parallel()

	msg, err := message.Parse(strings.NewReader(msg))
	assert.NoError(t, err)

	expectOrder := []string{"E", "F", "G", "H", "I", "J"}
	expectDepth := []int{2, 2, 2, 2, 2, 2}
	expectIndex := []int{0, 1, 0, 1, 0, 1}
	i := 0
	var pw walker.Parts = func(depth, j int, msg message.Part) error {
		where, err := msg.GetHeader().Get("X-Where")
		assert.NoError(t, err)
		assert.Equal(t, expectOrder[i], where)
		assert.Equal(t, expectDepth[i], depth)
		assert.Equal(t, expectIndex[i], j)
		i++
		return nil
	}

	err = pw.WalkOpaque(msg)
	assert.NoError(t, err)
}

func TestPartWalker_WalkMultipart(t *testing.T) {
	t.Parallel()

	msg, err := message.Parse(strings.NewReader(msg))
	assert.NoError(t, err)

	expectOrder := []string{"A", "B", "C", "D"}
	expectDepth := []int{0, 1, 1, 1}
	expectIndex := []int{0, 0, 1, 2}
	i := 0
	var pw walker.Parts = func(depth, j int, msg message.Part) error {
		where, err := msg.GetHeader().Get("X-Where")
		assert.NoError(t, err)
		assert.Equal(t, expectOrder[i], where)
		assert.Equal(t, expectDepth[i], depth)
		assert.Equal(t, expectIndex[i], j)
		i++
		return nil
	}

	err = pw.WalkMultipart(msg)
	assert.NoError(t, err)
}
