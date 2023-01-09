package header_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/go-email/v2/header"
	"github.com/zostay/go-email/v2/header/field"
)

func TestBase(t *testing.T) {
	t.Parallel()

	// these must safely initialize internal state which is set to zero values
	// first constructed, so let's make sure it all works and nothing panics
	testFuncs := []func(*header.Base){
		func(b *header.Base) { assert.Equal(t, field.DefaultFoldEncoding, b.FoldEncoding()) },
		func(b *header.Base) {
			newVf, err := field.NewFoldEncoding("\t", 72, 1000)
			require.NoError(t, err)
			b.SetFoldEncoding(newVf)
			vf := b.FoldEncoding()
			assert.Equal(t, newVf, vf)
		},
		func(b *header.Base) { assert.Equal(t, header.LF, b.Break()) },
		func(b *header.Base) {
			b.SetBreak(header.CRLF)
			assert.Equal(t, header.CRLF, b.Break())
		},
		func(b *header.Base) { assert.Nil(t, b.GetField(0)) },
		func(b *header.Base) { assert.Equal(t, 0, b.Len()) },
		func(b *header.Base) { assert.Nil(t, b.GetFieldNamed(header.Subject, 0)) },
		func(b *header.Base) { assert.Empty(t, b.GetAllFieldsNamed(header.Subject)) },
		func(b *header.Base) { assert.Empty(t, b.GetIndexesNamed(header.Subject)) },
		func(b *header.Base) { assert.Empty(t, b.ListFields()) },
		func(b *header.Base) {
			buf := &bytes.Buffer{}
			n, err := b.WriteTo(buf)
			assert.Zero(t, n)
			assert.NoError(t, err)
			assert.Empty(t, buf.String())
		},
		func(b *header.Base) {
			b.InsertBeforeField(0, "Subject", "testing")
		},
		func(b *header.Base) {
			b.ClearFields()
		},
		func(b *header.Base) {
			err := b.DeleteField(0)
			assert.ErrorIs(t, err, header.ErrIndexOutOfRange)
		},
	}
	for _, testFunc := range testFuncs {
		b := &header.Base{}
		assert.NotPanics(t, func() { testFunc(b) })
	}
}

func TestBase_GetField(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	f := b.GetField(0)
	assert.Equal(t, field.New("A", "b"), f)

	f = b.GetField(1)
	assert.Equal(t, field.New("C", "d"), f)

	f = b.GetField(2)
	assert.Equal(t, field.New("E", "f"), f)

	f = b.GetField(3)
	assert.Equal(t, field.New("E", "g"), f)

	f = b.GetField(4)
	assert.Nil(t, f)

	assert.Equal(t, 4, b.Len())
}

func TestBase_GetFieldNamed(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	f := b.GetFieldNamed("A", 0)
	assert.Equal(t, field.New("A", "b"), f)

	f = b.GetFieldNamed("B", 1)
	assert.Nil(t, f)

	f = b.GetFieldNamed("C", 0)
	assert.Equal(t, field.New("C", "d"), f)

	f = b.GetFieldNamed("C", 1)
	assert.Nil(t, f)

	f = b.GetFieldNamed("E", 0)
	assert.Equal(t, field.New("E", "f"), f)

	f = b.GetFieldNamed("E", 1)
	assert.Equal(t, field.New("E", "g"), f)

	f = b.GetFieldNamed("E", 2)
	assert.Nil(t, f)

	f = b.GetFieldNamed("H", 0)
	assert.Nil(t, f)

	assert.Equal(t, 4, b.Len())
}

func TestBase_GetAllFieldsNamed(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	fs := b.GetAllFieldsNamed("A")
	assert.Equal(t, []*field.Field{field.New("A", "b")}, fs)

	fs = b.GetAllFieldsNamed("B")
	assert.Len(t, fs, 0)

	fs = b.GetAllFieldsNamed("C")
	assert.Equal(t, []*field.Field{field.New("C", "d")}, fs)

	fs = b.GetAllFieldsNamed("E")
	assert.Equal(t, []*field.Field{
		field.New("E", "f"),
		field.New("E", "g"),
	}, fs)

	fs = b.GetAllFieldsNamed("H")
	assert.Len(t, fs, 0)

	assert.Equal(t, 4, b.Len())
}

func TestBase_GetIndexesNamed(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	is := b.GetIndexesNamed("A")
	assert.Equal(t, []int{0}, is)

	is = b.GetIndexesNamed("B")
	assert.Len(t, is, 0)

	is = b.GetIndexesNamed("C")
	assert.Equal(t, []int{1}, is)

	is = b.GetIndexesNamed("E")
	assert.Equal(t, []int{2, 3}, is)

	is = b.GetIndexesNamed("H")
	assert.Len(t, is, 0)

	assert.Equal(t, 4, b.Len())
}

func TestBase_ListFields(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	assert.Equal(t, []*field.Field{
		field.New("A", "b"),
		field.New("C", "d"),
		field.New("E", "f"),
		field.New("E", "g"),
	}, b.ListFields())
}

func TestBase_WriteTo(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	const expect = `A: b
C: d
E: f
E: g

`

	buf := &bytes.Buffer{}
	n, err := b.WriteTo(buf)
	assert.Equal(t, int64(21), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestBase_InsertBeforeField(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	b.InsertBeforeField(0, "H", "i")
	b.InsertBeforeField(4, "J", "k")
	b.InsertBeforeField(20, "L", "m")
	b.InsertBeforeField(4, "N", "o")
	b.InsertBeforeField(4, "P", "q")

	const expect = `H: i
A: b
C: d
E: f
P: q
N: o
J: k
E: g
L: m

`

	buf := &bytes.Buffer{}
	n, err := b.WriteTo(buf)
	assert.Equal(t, int64(46), n)
	assert.NoError(t, err)
	assert.Equal(t, expect, buf.String())
}

func TestBase_ClearFields(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")
	assert.Equal(t, 4, b.Len())

	b.ClearFields()
	assert.Equal(t, 0, b.Len())
}

func TestBase_DeleteField(t *testing.T) {
	t.Parallel()

	b := &header.Base{}
	assert.Equal(t, 0, b.Len())

	b.InsertBeforeField(0, "A", "b")
	b.InsertBeforeField(1, "C", "d")
	b.InsertBeforeField(2, "E", "f")
	b.InsertBeforeField(3, "E", "g")

	assert.Equal(t, 4, b.Len())
	assert.Equal(t, field.New("E", "f"), b.GetField(2))

	err := b.DeleteField(4)
	assert.ErrorIs(t, err, header.ErrIndexOutOfRange)

	err = b.DeleteField(2)
	assert.NoError(t, err)
	assert.Equal(t, 3, b.Len())

	assert.Equal(t, field.New("E", "g"), b.GetField(2))

	err = b.DeleteField(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, b.Len())

	assert.Nil(t, b.GetField(2))

	assert.Equal(t, field.New("A", "b"), b.GetField(0))

	err = b.DeleteField(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.Len())

	assert.Equal(t, field.New("C", "d"), b.GetField(0))

	err = b.DeleteField(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, b.Len())

	assert.Nil(t, b.GetField(0))
}
