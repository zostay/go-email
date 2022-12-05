package simple

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/pkg/email"
)

func TestHeaderParse(t *testing.T) {
	t.Parallel()

	const headerStr = `Foo: 1
Foo: 2
Foo: 3
Bar: 3
Baz: 1
`

	h, err := ParseHeader([]byte(headerStr))
	assert.NoError(t, err)
	assert.NotNil(t, h)

	assert.Equal(t, "1", h.HeaderGet("foo"))

	b, err := h.HeaderGetN("foo", 0)
	assert.NoError(t, err)
	assert.Equal(t, "1", b)

	b, err = h.HeaderGetN("foo", 1)
	assert.NoError(t, err)
	assert.Equal(t, "2", b)

	b, err = h.HeaderGetN("foo", 2)
	assert.NoError(t, err)
	assert.Equal(t, "3", b)

	b, err = h.HeaderGetN("foo", 3)
	assert.Error(t, err)
	assert.Equal(t, "", b)

	b, err = h.HeaderGetN("foo", -1)
	assert.NoError(t, err)
	assert.Equal(t, "3", b)

	b, err = h.HeaderGetN("foo", -3)
	assert.NoError(t, err)
	assert.Equal(t, "1", b)

	b, err = h.HeaderGetN("foo", -4)
	assert.Error(t, err)
	assert.Equal(t, "", b)

	assert.Equal(t, []string{"1", "2", "3"}, h.HeaderGetAll("Foo"))
}

func TestRename(t *testing.T) {
	t.Parallel()

	const basic = `Foo: F1
fOO: F2
bar: B1
FoO: F3
Baz: Z1
BAR: B2
`

	m, err := ParseHeaderLB([]byte(basic), []byte(email.LF))
	assert.NoError(t, err)

	err = m.HeaderRename("foo", "zap")
	assert.NoError(t, err)

	err = m.HeaderRename("foo", "zep")
	assert.NoError(t, err)

	err = m.HeaderRename("foo", "zip")
	assert.NoError(t, err)

	err = m.HeaderRename("foo", "zop")
	assert.Error(t, err)

	const basicRenamed = `zap: F1
zep: F2
bar: B1
zip: F3
Baz: Z1
BAR: B2
`

	assert.Equal(t, basicRenamed, m.String())
}

func TestRenameAll(t *testing.T) {
	t.Parallel()

	const basic = `Foo: F1
fOO: F2
bar: B1
FoO: F3
Baz: Z1
BAR: B2
`

	m, err := ParseHeaderLB([]byte(basic), []byte(email.LF))
	assert.NoError(t, err)

	err = m.HeaderSetAll("Bar", "B1A", "B2A")
	assert.NoError(t, err)

	err = m.HeaderRenameAll("Foo", "XYZ")
	assert.NoError(t, err)

	err = m.HeaderRenameAll("XYZ", "ZZZ")
	assert.NoError(t, err)

	err = m.HeaderRenameAll("Bar", "AAA")
	assert.NoError(t, err)

	const basicRenamed = `ZZZ: F1
ZZZ: F2
AAA: B1A
ZZZ: F3
Baz: Z1
AAA: B2A
`

	assert.Equal(t, basicRenamed, m.String())
}

func TestReanameN(t *testing.T) {
	t.Parallel()

	const emailText = `Foo: F1
fOO: F2
bar: B1
FoO: F3
Baz: Z1
FOO: F4
BAR: B2
`

	m, err := ParseHeaderLB([]byte(emailText), []byte(email.LF))
	assert.NoError(t, err)

	err = m.HeaderRenameN("Foo", "XYZ", -1)
	assert.NoError(t, err)
	assert.Equal(t, "F4", m.HeaderGet("XYZ"))

	err = m.HeaderRenameN("Foo", "XYZ", 3)
	assert.Error(t, err) // too high

	err = m.HeaderRenameN("Foo", "XYZ", -4)
	assert.Error(t, err) // too low

	err = m.HeaderRenameN("Foo", "Two", 2)
	assert.NoError(t, err)

	err = m.HeaderRenameN("Foo", "One", 1)
	assert.NoError(t, err)

	err = m.HeaderRenameN("Foo", "Zero", 0)
	assert.NoError(t, err)

	const wantHeader = `Zero: F1
One: F2
bar: B1
Two: F3
Baz: Z1
XYZ: F4
BAR: B2
`

	assert.Equal(t, wantHeader, m.String())
}

func TestHeaderWrapping(t *testing.T) {
	t.Parallel()

	const emailText = `Foo: Wrapped
  Needlessly
Foo: Not wrapped, but may need wrapping if the field name becomes long.
Foo: Wrapped, and will generally need to be wrapped again, if the field
  name stays long.
`

	m, err := ParseHeaderLB([]byte(emailText), []byte(email.LF))
	assert.NoError(t, err)

	// round-tripping
	assert.Equal(t, emailText, m.String())

	err = m.HeaderRenameAll("Foo", "The-Field-Formerly-Known-As-Foo")
	assert.NoError(t, err)

	const rewrappedText = `The-Field-Formerly-Known-As-Foo: Wrapped
  Needlessly
The-Field-Formerly-Known-As-Foo: Not wrapped, but may need wrapping if the
 field name becomes long.
The-Field-Formerly-Known-As-Foo: Wrapped, and will generally need to be
 wrapped again, if the field
  name stays long.
`

	assert.Equal(t, rewrappedText, m.String())
}

func TestBadStartError(t *testing.T) {
	t.Parallel()

	err := &BadStartError{}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be a header")
}

func TestHeaderParseError(t *testing.T) {
	t.Parallel()

	err := &HeaderParseError{
		Errs: []error{
			fmt.Errorf("one"),
			fmt.Errorf("two"),
			fmt.Errorf("three"),
		},
	}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing email")
	assert.Contains(t, err.Error(), "one, two, three")
}

func TestNewHeader(t *testing.T) {
	t.Parallel()

	_, err := NewHeader(email.LF, "One")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no body value")

	h, err := NewHeader(email.LF, "One", "two", "Three", "four", "One", "five")
	assert.NoError(t, err)
	require.NotNil(t, h)

	assert.Len(t, h.Fields, 3)
	assert.Equal(t, []string{"One", "Three"}, h.HeaderNames())
	assert.Equal(t, []string{"two", "five"}, h.HeaderGetAll("One"))
	assert.Equal(t, []string{"four"}, h.HeaderGetAll("Three"))
}

func TestHeader_HeaderSetN(t *testing.T) {
	t.Parallel()

	const header = `One: two
Three: four
One: five
`

	h, err := ParseHeader([]byte(header))
	require.NoError(t, err)

	err = h.HeaderSetN("One", "six", 0)
	assert.NoError(t, err)

	err = h.HeaderSetN("One", "seven", 1)
	assert.NoError(t, err)

	err = h.HeaderSetN("One", "eight", 2)
	assert.Error(t, err)

	err = h.HeaderSetN("One", "nine", -2)
	assert.NoError(t, err)

	const headerSet = `One: nine
Three: four
One: seven
`

	assert.Equal(t, headerSet, h.String())
}

func TestHeader_HeaderAddN(t *testing.T) {
	t.Parallel()

	const basic = `One: two
Three: four
One: five
`

	h, err := ParseHeader([]byte(basic))
	require.NoError(t, err)

	err = h.HeaderAddN("One", "six", 0)
	assert.NoError(t, err)

	err = h.HeaderAddN("One", "seven", 3)
	assert.NoError(t, err)

	err = h.HeaderAddN("One", "eight", 1)
	assert.NoError(t, err)

	err = h.HeaderAddN("One", "nine", -2)
	assert.NoError(t, err)

	err = h.HeaderAddN("Foo", "ten", 0)
	assert.NoError(t, err)

	const basicAdded = `One: two
One: six
One: eight
Three: four
One: five
One: nine
One: seven
Foo: ten
`

	assert.Equal(t, basicAdded, h.String())
}

func TestHeader_HeaderAddBeforeN(t *testing.T) {
	t.Parallel()

	const basic = `One: two
Three: four
One: five
`

	h, err := ParseHeader([]byte(basic))
	require.NoError(t, err)

	err = h.HeaderAddBeforeN("One", "six", 0)
	assert.NoError(t, err)

	err = h.HeaderAddBeforeN("One", "seven", 3)
	assert.NoError(t, err)

	err = h.HeaderAddBeforeN("One", "eight", 1)
	assert.NoError(t, err)

	err = h.HeaderAddBeforeN("One", "nine", -2)
	assert.NoError(t, err)

	err = h.HeaderAddBeforeN("Foo", "ten", 0)
	assert.NoError(t, err)

	const basicAdded = `Foo: ten
One: six
One: eight
One: two
Three: four
One: nine
One: seven
One: five
`

	assert.Equal(t, basicAdded, h.String())
}

func TestHeader_HeaderDelete(t *testing.T) {
	t.Parallel()

	const basic = `One: two
Three: four
One: five
`

	h, err := ParseHeader([]byte(basic))
	require.NoError(t, err)

	err = h.HeaderDelete("One", 0)
	assert.NoError(t, err)

	const basicDeleteFirstOne = `Three: four
One: five
`

	assert.Equal(t, basicDeleteFirstOne, h.String())

	h, err = ParseHeader([]byte(basic))
	require.NoError(t, err)

	err = h.HeaderDelete("One", 1)
	assert.NoError(t, err)

	const basicDeleteSecondOne = `One: two
Three: four
`

	assert.Equal(t, basicDeleteSecondOne, h.String())

	h, err = ParseHeader([]byte(basic))
	require.NoError(t, err)

	err = h.HeaderDelete("One", 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete")
}
