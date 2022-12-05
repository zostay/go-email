package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderField(t *testing.T) {
	t.Parallel()

	hf := NewHeaderField("ONE ", "two", []byte(LF))
	require.NotNil(t, hf)

	assert.Equal(t, "one", hf.Match())
	assert.Equal(t, "ONE ", hf.Name())
	assert.Equal(t, "two", hf.Body())
}

func TestNewHeaderFieldParsed(t *testing.T) {
	t.Parallel()

	hf := NewHeaderFieldParsed("ONE ", "two", []byte("ONE :  two\n"))
	require.NotNil(t, hf)

	assert.Equal(t, "one", hf.Match())
	assert.Equal(t, "ONE ", hf.Name())
	assert.Equal(t, "two", hf.Body())
	assert.Equal(t, []byte(" two\n"), hf.RawBody())
	assert.Equal(t, []byte("ONE :  two\n"), hf.Original())
	assert.Equal(t, "ONE :  two\n", hf.String())
	assert.Equal(t, []byte("ONE :  two\n"), hf.Bytes())

	hf.SetName("One", []byte(LF))
	assert.Equal(t, []byte("One:  two\n"), hf.Original())

	hf.SetNameNoFold("ONE")
	assert.Equal(t, []byte("ONE:  two\n"), hf.Original())

	hf.SetBody("three", []byte(LF))
	assert.Equal(t, "three", hf.Body())
	assert.Equal(t, []byte("ONE: three\n"), hf.Original())

	hf.SetBody("this is a long string that should be folded by the set body method because it is too long to stay beyond the preferred length", []byte(LF))
	assert.Equal(t,
		"this is a long string that should be folded by the set body method because it is too long to stay beyond the preferred length",
		hf.Body())
	assert.Equal(t,
		[]byte("ONE: this is a long string that should be folded by the set body method\n because it is too long to stay beyond the preferred length\n"),
		hf.Original())

	hf.SetName("This is a longer key so it will change where the header is folded", []byte(LF))
	assert.Equal(t,
		"This is a longer key so it will change where the header is folded",
		hf.Name())
	assert.Equal(t,
		[]byte("This is a longer key so it will change where the header is folded: this is a\n long string that should be folded by the set body method\n because it is too long to stay beyond the preferred length\n"),
		hf.Original())

	hf.SetBodyNoFold("This is a long string but we'll be stopping the folding because we don't include any, though we should", []byte(LF))
	assert.Equal(t,
		"This is a long string but we'll be stopping the folding because we don't include any, though we should",
		hf.Body())
	assert.Equal(t,
		[]byte("This is a longer key so it will change where the header is folded: This is a long string but we'll be stopping the folding because we don't include any, though we should\n"),
		hf.Original())
}

func TestMakeHeaderFieldMatch(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "one", MakeHeaderFieldMatch("ONE"))
	assert.Equal(t, "one", MakeHeaderFieldMatch("   \toNE\r\n"))
	assert.Equal(t, "one", MakeHeaderFieldMatch("one"))
}

func TestHeader_Cache(t *testing.T) {
	t.Parallel()

	hf := NewHeaderField("One", "two", []byte(LF))
	require.NotNil(t, hf)

	type testthing struct{}

	assert.Nil(t, hf.CacheGet("test"))
	hf.CacheSet("test", testthing{})

	assert.Equal(t, testthing{}, hf.CacheGet("test"))
	assert.Nil(t, hf.CacheGet("test2"))
}

func TestHeaderField_SetBodyEncoded(t *testing.T) {
	t.Parallel()

	hf := NewHeaderField("One", "two", []byte(LF))
	hf.SetBodyEncoded("three", []byte("four"), []byte(LF))

	assert.Equal(t, "three", hf.Body())
	assert.Equal(t, []byte("One: four\n"), hf.Original())

	hf.SetBodyEncoded(
		"this is a long string that should be folded by the set body method because it is too long to stay beyond the preferred length",
		[]byte("this is a long string that should be folded by the set body method because it is too long to stay beyond the preferred length"),
		[]byte(LF),
	)
	assert.Equal(t,
		"this is a long string that should be folded by the set body method because it is too long to stay beyond the preferred length",
		hf.Body())
	assert.Equal(t,
		[]byte("One: this is a long string that should be folded by the set body method\n because it is too long to stay beyond the preferred length\n"),
		hf.Original())

	hf.SetBodyEncodedNoFold(
		"long string that could get folded if folding were to be employed in writing out the value",
		[]byte("long string that could get foldedXif folding were to be employed in writing out the value\n"),
	)
	assert.Equal(t,
		"long string that could get folded if folding were to be employed in writing out the value",
		hf.Body())
	assert.Equal(t,
		[]byte("One: long string that could get foldedXif folding were to be employed in writing out the value\n"),
		hf.Original())
}
