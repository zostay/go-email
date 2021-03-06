package simple

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	h, err := ParseHeaderLB([]byte(headerStr), []byte(email.LF))
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
