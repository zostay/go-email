package field

import (
	"bytes"
)

// BadStartError is returned when the header begins with junk text that does not
// appear to be a header. This text is preserved in the error object.
type BadStartError struct {
	BadStart []byte // the text skipped at the start of header
}

// Error returns the error message.
func (err *BadStartError) Error() string {
	return "header starts with text that does not appear to be a header"
}

// Line represents the unparsed content for a complete header field line.
type Line []byte

// Lines represents the unparsed content for zero or more header field
// lines.
type Lines []Line

// ParseLines splits the given input into lines according to the rules we use to
// determine how to break header fields up inside a header. The input bytes are
// expected to include only the header. It will parse the whole input as if all
// of it belongs to the header. It returns the input as Lines, which are
// [][]byte, ready to feed into the header field parser.
//
// This method does not follow RFC 5322 precisely. It will accept input that
// would be rejected by the specification as part of the effort this module
// library makes in attempting to be liberal in what it accepts, but strict in
// what it generates.
//
// If the first line (or lines) of input start with spaces or contain no colons,
// these lines will be skipped in the Lines returned. However, a BadStartError
// will be returned.
//
// From then on, this will start a new field on any line that does not start
// with a space and contains a colon. After the first such line is encountered,
// any line after that will be considered a continuation if it starts with a
// space or does not contain a colon. (And for the purposes of this method's
// documentation, we consider a space character or tab character to be a space,
// but all other characters are treated as non-spaces. This is in keeping with
// RFC 5322.)
func ParseLines(m, lb []byte) (Lines, error) {
	h := make(Lines, 0, len(m)/80)
	var err *BadStartError
	for _, line := range bytes.SplitAfter(m, lb) {
		if len(line) == 0 {
			break
		}
		if line[0] == '\t' || line[0] == ' ' || !bytes.Contains(line, []byte(":")) {
			// Start with a continuation? Weird, uh...
			if len(h) == 0 {
				if err != nil {
					err.BadStart = append(err.BadStart, line...)
				} else {
					err = &BadStartError{line}
				}
				continue
			}

			h[len(h)-1] = append(h[len(h)-1], line...)
		} else {
			h = append(h, line)
		}
	}

	if err != nil {
		return h, err
	} else {
		return h, nil
	}
}

// Parse will take a single header field line, including any folded continuation
// lines. This will then construct a header field object.
func Parse(f Line, lb []byte) *Field {
	rawField := bytes.TrimRight(f, string(lb))

	off := 1
	ix := bytes.Index(rawField, []byte{':'})
	if ix < 0 {
		ix = len(rawField)
		off = 0
	}

	// it's fine to use the default fold encoding because unfold is not affected
	// by choices made when folding
	name := string(DefaultFoldEncoding.Unfold(rawField[:ix]))
	body := string(bytes.TrimSpace(DefaultFoldEncoding.Unfold(rawField[ix+off:])))
	decBody, err := Decode(body)
	if err == nil {
		body = decBody
	}

	return &Field{
		Base: Base{name, body},
		Raw:  &Raw{rawField, ix},
	}
}
