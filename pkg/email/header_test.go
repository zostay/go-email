package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader_HeaderFieldIndex(t *testing.T) {
	t.Parallel()

	h := NewHeader(LF,
		NewHeaderField("Received", "a", []byte(LF)),
		NewHeaderField("Foo", "1", []byte(LF)),
		NewHeaderField("From", "b", []byte(LF)),
		NewHeaderField("Foo", "2", []byte(LF)),
		NewHeaderField("To", "c", []byte(LF)),
		NewHeaderField("Foo", "3", []byte(LF)),
		NewHeaderField("Subject", "d", []byte(LF)),
	)

	assert.Equal(t, 1, h.HeaderFieldIndex("Foo", 0, false))
	assert.Equal(t, 3, h.HeaderFieldIndex("Foo", 1, false))
	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", 2, false))
	assert.Equal(t, -1, h.HeaderFieldIndex("Foo", 3, false))

	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", -1, false))
	assert.Equal(t, 3, h.HeaderFieldIndex("Foo", -2, false))
	assert.Equal(t, 1, h.HeaderFieldIndex("Foo", -3, false))
	assert.Equal(t, -1, h.HeaderFieldIndex("Foo", -4, false))

	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", 3, true))
	assert.Equal(t, 5, h.HeaderFieldIndex("Foo", -1, true))

	hf := h.HeaderGetField("foo")
	assert.Equal(t, "Foo", hf.Name())
	assert.Equal(t, "1", hf.Body())

	hf = h.HeaderGetField("Zip")
	assert.Nil(t, hf)

	hf, err := h.HeaderGetFieldN("fOO", 2)
	assert.NoError(t, err)
	assert.Equal(t, "Foo", hf.Name())
	assert.Equal(t, "3", hf.Body())

	_, err = h.HeaderGetFieldN("Foo", 3)
	assert.Error(t, err)
}

func TestHeader_Break(t *testing.T) {
	t.Parallel()

	h := NewHeader(CRLF)
	assert.Equal(t, []byte(CRLF), h.Break())

	h = NewHeader(CR)
	assert.Equal(t, []byte(CR), h.Break())

	h = NewHeader(LF)
	assert.Equal(t, []byte(LF), h.Break())
}

func TestHeader_Stringify(t *testing.T) {
	t.Parallel()

	h := NewHeader(LF,
		NewHeaderField("Received", "a", []byte(LF)),
		NewHeaderField("Foo", "1", []byte(LF)),
		NewHeaderField("From", "b", []byte(LF)),
		NewHeaderField("Foo", "2", []byte(LF)),
		NewHeaderField("To", "c", []byte(LF)),
		NewHeaderField("Foo", "3", []byte(LF)),
		NewHeaderField("Subject", "d", []byte(LF)),
	)

	expect := `Received: a
Foo: 1
From: b
Foo: 2
To: c
Foo: 3
Subject: d
`

	assert.Equal(t, expect, h.String())
	assert.Equal(t, []byte(expect), h.Bytes())
}

func TestHeader_HeaderGetAllFields(t *testing.T) {
	t.Parallel()

	h := NewHeader(LF,
		NewHeaderField("Received", "a", []byte(LF)),
		NewHeaderField("Foo", "1", []byte(LF)),
		NewHeaderField("From", "b", []byte(LF)),
		NewHeaderField("Foo", "2", []byte(LF)),
		NewHeaderField("To", "c", []byte(LF)),
		NewHeaderField("Foo", "3", []byte(LF)),
		NewHeaderField("Subject", "d", []byte(LF)),
	)

	hfs := h.HeaderGetAllFields("Zip")
	assert.Len(t, hfs, 0)

	hfs = h.HeaderGetAllFields("Foo")
	assert.Len(t, hfs, 3)
	assert.Equal(t, "Foo: 1\n", hfs[0].String())
	assert.Equal(t, "Foo: 2\n", hfs[1].String())
	assert.Equal(t, "Foo: 3\n", hfs[2].String())

	hfs = h.HeaderGetAllFields("Subject")
	assert.Len(t, hfs, 1)
	assert.Equal(t, "Subject: d\n", hfs[0].String())
}
