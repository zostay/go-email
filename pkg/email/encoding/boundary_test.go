package encoding

import (
	"math/rand"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var nonAlphaNumericMatch = regexp.MustCompile(`[^a-zA-Z0-9]`)

func TestGenerateBoundary(t *testing.T) {
	t.Parallel()

	b := GenerateBoundary()
	assert.Len(t, b, 30)
	assert.False(t, nonAlphaNumericMatch.MatchString(b))
}

func TestGenerateSafeBoundary(t *testing.T) {
	// Do not test in parallel. This might be okay, but I don't trust that
	// setting the global seed does not have consequences elsewhere.

	rand.Seed(42)
	b := GenerateBoundary()

	rand.Seed(42)
	nb := GenerateSafeBoundary(b)
	assert.Len(t, nb, 30)
	assert.False(t, nonAlphaNumericMatch.MatchString(nb))
	assert.NotEqual(t, b, nb)
}
