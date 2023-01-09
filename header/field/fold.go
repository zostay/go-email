package field

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

const (
	DefaultFoldIndent          = " "  // indent placed before folded lines
	DefaultPreferredFoldLength = 80   // we prefer headers and 7bit/8bit bodies lines shorter than this
	DefaultForcedFoldLength    = 1000 // we forceably break headers and 7bit/8bit bodies lines longer than this

	DoNotFold = -1 // we prefer not to fold at all
)

var (
	// DefaultFoldEncoding creates a new FoldEncoding using default settings. This
	// is the recommended way to create a FoldEncoding.
	DefaultFoldEncoding = &FoldEncoding{
		DefaultFoldIndent,
		DefaultPreferredFoldLength,
		DefaultForcedFoldLength,
	}

	// DoNotFoldEncoding is a FoldEncoding that doesn't perform folding.
	DoNotFoldEncoding = &FoldEncoding{
		DefaultFoldIndent,
		DoNotFold,
		DoNotFold,
	}
)

var (
	// ErrFoldIndentSpace is returned by NewFoldEncoding when a non-space/non-tab
	// character is put in the foldIndent setting.
	ErrFoldIndentSpace = errors.New("fold indent may only contains spaces and tabs")

	// ErrFoldIndentTooShort is returned by NewFoldEncoding when the foldIndent
	// is empty.
	ErrFoldIndentTooShort = errors.New("fold indent must contain at least one space or tab")

	// ErrFoldIndentTooLong is returned by NewFoldEncoding when the foldIndent
	// setting is equal to or longer than the preferredFoldLength.
	ErrFoldIndentTooLong = errors.New("fold indent must be shorter than the preferred fold length")

	// ErrFoldLengthTooLong is returned by NewFoldEncoding when the
	// preferredFoldLength is longer than the forcedFoldLength.
	ErrFoldLengthTooLong = errors.New("preferred fold length must be no longer than the forced fold length")

	// ErrFoldLengthTooShort is returned by NewFoldEncoding when the
	// forcedFoldLength is shorter than 3 bytes long.
	ErrFoldLengthTooShort = errors.New("preferred fold length and forced fold length cannot be too short")

	// ErrDoNotFold is returned by NewFoldEncoding when the preferredFoldLength
	// or forcedFoldLength are set to DoNotFold (-1), but both are not set that
	// way. You must set both to DoNotFold to prevent folding or neither to
	// DoNotFold.
	ErrDoNotFold = errors.New("preferred fold length and forced fold length must both be -1 if either are -1")
)

// Break is basically identical to header.Break, but with a focus on bytes.
type Break []byte

// FoldEncoding provides the tooling for folding email message headers.
type FoldEncoding struct {
	foldIndent          string
	preferredFoldLength int
	forcedFoldLength    int
}

// NewFoldEncoding creates a new FoldEncoding with the given settings. The
// foldIndent must be a string, filled with one or more space or tab characters,
// and it must be shorter than the preferredFoldLength. The preferredFoldLength
// must be equal to or less than forcedFoldLength. if any of the given inputs do
// not meet these requirements, an error will be returned.
//
// The fold encoding does not do anything special to ensure that no folding
// occurs before the colon even though that would be incorrect. It relies on the
// assumption that the fold lengths chosen will be wider than the longest field
// name. That should be enough to guarantee that field names never get folded.
func NewFoldEncoding(
	foldIndent string,
	preferredFoldLength,
	forcedFoldLength int,
) (*FoldEncoding, error) {
	if ix := strings.IndexFunc(foldIndent, func(c rune) bool { return !isSpace(c) }); ix >= 0 {
		return nil, ErrFoldIndentSpace
	}

	if len(foldIndent) < 1 {
		return nil, ErrFoldIndentTooShort
	}

	if (preferredFoldLength == DoNotFold && forcedFoldLength != DoNotFold) ||
		(forcedFoldLength == DoNotFold && preferredFoldLength != DoNotFold) {
		return nil, ErrDoNotFold
	}

	if preferredFoldLength != DoNotFold {
		// if we aren't folding, we don't have to worry about these
		if len(foldIndent) >= preferredFoldLength {
			return nil, ErrFoldIndentTooLong
		}

		if preferredFoldLength > forcedFoldLength {
			return nil, ErrFoldLengthTooLong
		}

		// This is WAY too short. Probably 80 is a good general minimum and
		// maybe as short as 50 or 60 for extreme cases, but less than that is
		// probably a bad idea. However, I'm not the too short line police, so
		// we just stop it at the point something in the code will choke and die
		// on.
		if preferredFoldLength < 3 || forcedFoldLength < 3 {
			return nil, ErrFoldLengthTooShort
		}
	}

	return &FoldEncoding{foldIndent, preferredFoldLength, forcedFoldLength}, nil
}

// Unfold will take a folded header line from an email and unfold it for
// reading. This gives you the proper header body value.
func (vf *FoldEncoding) Unfold(f []byte) []byte {
	uf := make([]byte, 0, len(f))
	for _, b := range f {
		if !isCRLF(rune(b)) {
			uf = append(uf, b)
		}
	}
	return uf
}

func isCRLF(c rune) bool     { return c == '\r' || c == '\n' }
func isSpace(c rune) bool    { return c == ' ' || c == '\t' }
func isNonSpace(c rune) bool { return c != ' ' && c != '\t' }

// Fold will take an unfolded or perhaps partially folded value from an
// email and fold it. It will make sure that every fold line is properly
// indented, try to break lines on appropriate spaces, and force long lines to
// be broken before the maximum line length.
//
// Writes the folded output to the given io.Writer and returns the number of
// bytes written and returns an error if there's an error writing the data.
func (vf *FoldEncoding) Fold(out io.Writer, f []byte, lb Break) (int64, error) {
	total := int64(0)
	continuingLine := false
	writeFold := func(f []byte, end int) ([]byte, error) {
		// only indent if there's no space already present at the break
		if continuingLine && !isSpace(rune(f[0])) {
			n, err := out.Write([]byte(vf.foldIndent))
			total += int64(n)
			if err != nil {
				return nil, err
			}
		}
		n, err := out.Write(f[:end])
		total += int64(n)
		if err != nil {
			return nil, err
		}

		n, err = out.Write(lb)
		total += int64(n)
		if err != nil {
			return nil, err
		}

		f = f[end:]
		continuingLine = true

		return bytes.TrimLeft(f, " \t"), nil
	}

	if len(f) < vf.preferredFoldLength || vf.preferredFoldLength == DoNotFold {
		_, err := writeFold(f, len(f))
		return total, err
	}

	// NOTE: We just assume the lb will be 2 chars.

	lines := bytes.Split(f, lb)
	for _, line := range lines {
	FoldingSingle:
		for len(line) > 0 {
			var err error

			// Will we be forced to fold?
			fforced := len(line) > vf.forcedFoldLength-2

			// Do we need to fold lines?
			fneed := len(line) > vf.preferredFoldLength-2
			if !fneed {
				line, err = writeFold(line, len(line))
				if err != nil {
					return total, err
				}
				continue FoldingSingle
			}

			var firstChar int
			if continuingLine {
				// if we're past the first line, the first non-space is the first char
				firstChar = bytes.IndexFunc(line, isNonSpace)
			} else {
				// if we're on the first line, the first none space after the colon is the first char
				colon := bytes.IndexRune(line, ':')
				firstChar = bytes.IndexFunc(line[colon+1:], isNonSpace)
				if firstChar >= 0 {
					firstChar += colon + 1
				}
			}

			if firstChar < -1 {
				// TODO Consider if this is really necessary or if it should result in an error insteat
				firstChar = 0
			}

			// best case, we find a space in the first n-2 chars, break there
			if ix := bytes.LastIndexFunc(line[firstChar:vf.preferredFoldLength-2], isSpace); ix >= 0 {
				line, err = writeFold(line, ix+firstChar)
				if err != nil {
					return total, err
				}
				continue FoldingSingle
			}

			// barring that, try to find a space after the n-2 char mark
			if ix := bytes.IndexFunc(line[firstChar:], isSpace); ix >= 0 && ix < vf.forcedFoldLength-2 {
				line, err = writeFold(line, ix+firstChar)
				if err != nil {
					return total, err
				}
				continue FoldingSingle
			}

			// but if it's really long with no space, force a break at n-2
			if fforced {
				line, err = writeFold(line, vf.preferredFoldLength-2)
				if err != nil {
					return total, err
				}
				continue FoldingSingle
			}

			// We're not forced to fold this line. Allow it to be longer than we
			// prefer.
			line, err = writeFold(line, len(line))
			if err != nil {
				return total, err
			}
		}
	}

	return total, nil
}
