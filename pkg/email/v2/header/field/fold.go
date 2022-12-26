package field

import (
	"bytes"
	"errors"
	"strings"
)

const (
	DefaultFoldIndent          = " "  // indent placed before folded lines
	DefaultPreferredFoldLength = 80   // we prefer headers and 7bit/8bit bodies lines shorter than this
	DefaultForcedFoldLength    = 1000 // we forceably break headers and 7bit/8bit bodies lines longer than this
)

var (
	// DefaultFoldEncoding creates a new FoldEncoding using default settings. This
	// is the recommended way to create a FoldEncoding.
	DefaultFoldEncoding = &FoldEncoding{
		DefaultFoldIndent,
		DefaultPreferredFoldLength,
		DefaultForcedFoldLength,
	}
)

var (
	// ErrFoldIndentSpace is returned by NewFoldEncoding when a non-space/non-tab
	// character is put in the foldIndent setting.
	ErrFoldIndentSpace = errors.New("fold indent may only contains spaces and tabs")

	// ErrFoldIndentTooLong is returned by NewFoldEncoding when the foldIndent
	// setting is equal to or longer than the preferredFoldLength.
	ErrFoldIndentTooLong = errors.New("fold indent must be shorter than the preferred fold length")

	// ErrFoldLengthTooLong is returned by NewFoldEncoding when the
	// preferredFoldLength is longer than the forcedFoldLength.
	ErrFoldLengthTooLong = errors.New("preferred fold length must be no longer than the forced fold length")

	// ErrFoldLengthTooShort is returned by NewFoldEncoding when the
	// forcedFoldLength is shorter than 3 bytes long.
	ErrFoldLengthTooShort = errors.New("forced fold length cannot be too short")
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
// foldIndent must be a string made up of at least one or more space or tab
// characters and it must be shorter than the preferredFoldLength. The
// preferredFoldLength must be equal to or less than forcedFoldLength. if
// any of the given inputs do not meet these requirements, an error will be
// returned.
func NewFoldEncoding(
	foldIndent string,
	preferredFoldLength,
	forcedFoldLength int,
) (*FoldEncoding, error) {
	if ix := strings.IndexFunc(foldIndent, func(c rune) bool { return !isSpace(c) }); ix >= 0 {
		return nil, ErrFoldIndentSpace
	}

	if len(foldIndent) >= preferredFoldLength {
		return nil, ErrFoldIndentTooLong
	}

	if preferredFoldLength > forcedFoldLength {
		return nil, ErrFoldLengthTooLong
	}

	if forcedFoldLength < 3 {
		return nil, ErrFoldLengthTooShort
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

func isCRLF(c rune) bool  { return c == '\r' || c == '\n' }
func isSpace(c rune) bool { return c == ' ' || c == '\t' }

// Fold will take an unfolded or perhaps partially folded value from an
// email and fold it. It will make sure that every fold line is properly
// indented, try to break lines on appropriate spaces, and force long lines to
// be broken before the maximum line length.
func (vf *FoldEncoding) Fold(f []byte, lb Break) []byte {
	if len(f) < vf.preferredFoldLength {
		return f
	}

	var out bytes.Buffer
	foldSpace := false
	writeFold := func(f []byte, end int) []byte {
		// only indent if there's no space already present at the break
		if foldSpace && !isSpace(rune(f[0])) {
			out.WriteString(vf.foldIndent)
		}
		out.Write(f[:end])
		out.Write(lb)
		f = f[end:]
		foldSpace = true

		return bytes.TrimLeft(f, " \t")
	}

	lines := bytes.Split(f, lb)
	for _, line := range lines {
		// Will we be forced to fold?
		fforced := len(line) > vf.forcedFoldLength-2

	FoldingSingle:
		for len(line) > 0 {

			// Do we need to fold lines?
			fneed := len(line) > vf.preferredFoldLength-2
			if !fneed {
				line = writeFold(line, len(line))
				continue FoldingSingle
			}

			// best case, we find a space in the first 78 chars, break there
			if ix := bytes.LastIndexFunc(line[0:vf.preferredFoldLength-2], isSpace); ix > -1 {
				line = writeFold(line, ix)
				continue FoldingSingle
			}

			// barring that, try to find a space after the 78 char mark
			if ix := bytes.IndexFunc(line, isSpace); ix > -1 && ix < vf.forcedFoldLength-2 {
				line = writeFold(line, ix)
				continue FoldingSingle
			}

			// but if it's really long with no space, force a break at 78
			if fforced {
				line = writeFold(line, vf.preferredFoldLength-2)
				continue FoldingSingle
			}

			// We're not forced to fold this line. Allow it to be longer than we
			// prefer.
			line = writeFold(line, len(line))
		}
	}

	return out.Bytes()
}
