package email

import (
	"bytes"
	"unicode"
)

const (
	FoldIndent          = " "
	PreferredFoldLength = 80
	ForcedFoldLength    = 1000
)

// UnfoldValue will take a folded header line from an email and unfold it for
// reading. This gives you the proper header body value.
func UnfoldValue(f, lb []byte) []byte {
	var uf bytes.Buffer
	folds := bytes.Split(f, lb)
	needsSpace := false
	trim := false
	for _, fold := range folds {
		if trim {
			fold = bytes.TrimLeft(fold, " \t")
		}

		if needsSpace {
			uf.WriteRune(' ')
		}
		uf.Write(fold)

		if len(fold) > 0 {
			needsSpace = unicode.IsPrint(rune(fold[len(fold)-1]))
		}
		trim = true
	}

	return uf.Bytes()
}

func isSpace(c rune) bool { return c == ' ' || c == '\t' }

// FoldValue will take an unfolded or perhaps partially folded value from an
// email and fold it. It will make sure that every fold line is properly
// indented, try to break lines on appropriate spaces, and force long lines to
// be broken before the maximum line length.
func FoldValue(f, lb []byte) []byte {
	if len(f) < PreferredFoldLength {
		return f
	}

	var out bytes.Buffer
	foldSpace := false
	writeFold := func(f []byte, end int) []byte {
		if foldSpace {
			out.WriteString(FoldIndent)
		}
		out.Write(f[:end])
		out.Write(lb)
		f = f[end:]
		foldSpace = true

		return f
	}

	lines := bytes.Split(f, lb)
	for _, line := range lines {
		// Will we be forced to fold?
		fforced := len(line) > ForcedFoldLength-2

	FoldingSingle:
		for len(line) > 0 {
			// Do we need to fold lines?
			fneed := len(line) > PreferredFoldLength-2
			if !fneed {
				line = writeFold(line, len(line))
				continue FoldingSingle
			}

			// best case, we find a space in the first 78 chars, break there
			if ix := bytes.LastIndexFunc(line[0:PreferredFoldLength-2], isSpace); ix > -1 {
				line = writeFold(line, ix)
				continue FoldingSingle
			}

			// barring that, try to find a space after the 78 char mark
			if ix := bytes.IndexFunc(line, isSpace); ix > -1 && ix < ForcedFoldLength-2 {
				line = writeFold(line, ix)
				continue FoldingSingle
			}

			// but if it's really long with no space, force a break at 78
			if fforced {
				line = writeFold(line, PreferredFoldLength-2)
				continue FoldingSingle
			}

			// We're not forced to fold this line. Allow it to be longer than we
			// prefer.
			line = writeFold(line, len(line))
		}
	}

	return out.Bytes()
}
