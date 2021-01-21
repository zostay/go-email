package email

import (
	"strings"
	"unicode"
)

const (
	FoldIndent          = " "
	PreferredFoldLength = 80
	ForcedFoldLength    = 1000
)

func UnfoldValue(f, lb string) string {
	var uf strings.Builder
	folds := strings.Split(f, lb)
	needsSpace := false
	trim := false
	for _, fold := range folds {
		if trim {
			fold = strings.TrimLeft(fold, " \t")
		}

		if needsSpace {
			uf.WriteRune(' ')
		}
		uf.WriteString(fold)

		needsSpace = unicode.IsPrint(rune(fold[len(fold)-1]))
		trim = true
	}

	return uf.String()
}

func isSpace(c rune) bool { return c == ' ' || c == '\t' }

func FoldValue(f, lb string) string {
	if len(f) < PreferredFoldLength {
		return f
	}

	var out strings.Builder
	foldSpace := false
	writeFold := func(f *string, end int) {
		if foldSpace {
			out.WriteString(FoldIndent)
		}
		out.WriteString((*f)[:end])
		out.WriteString(lb)
		*f = (*f)[end:]
		foldSpace = true
	}

	lines := strings.Split(f, lb)
	for _, line := range lines {
		for len(line) > 0 {
			if ix := strings.LastIndexFunc(line[0:PreferredFoldLength-2], isSpace); ix > -1 {
				// best case, we find a space in the first 78 chars, break there
				writeFold(&line, ix)
			} else if ix := strings.IndexFunc(line, isSpace); ix > -1 && ix < ForcedFoldLength-2 {
				// barring that, try to find a space after the 78 char mark
				writeFold(&line, ix)
			} else if len(line) > PreferredFoldLength-2 {
				// but if it's really long with no space, force a break at 78
				writeFold(&line, PreferredFoldLength-2)
			} else {
				// write the last bit out
				writeFold(&line, len(line))
			}
		}
	}

	return out.String()
}
