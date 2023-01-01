package mime

import (
	"math/rand"
	"strings"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GenerateBoundary will generate a random MIME boundary that is probably unique
// in most circumstances.
func GenerateBoundary() string {
	s := make([]rune, 30)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// GenerateSafeBoundary will generate a random MIME boundary that is guaranteed
// to be safe with the given corpus of data. Use this when you want to generate
// a boundary for a known set of parts:
//
//	boundary := encoding.GenerateSafeBoundary(strings.Join("", parts))
//
// using this is likely to be total overkill, but in case you're paranoid.
func GenerateSafeBoundary(contents string) string {
	for {
		boundary := GenerateBoundary()
		if !strings.Contains(contents, boundary) {
			return boundary
		}
	}
}
