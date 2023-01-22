package header

import (
	"errors"

	"github.com/zostay/go-email/v2/message/header/field"
)

// Parse will parse the given slice of bytes into an email header using the
// given line break string. It will assume the entire string given represents
// the header to be parsed.
//
// The parsed message will have field.DoNotFoldEncoding. This allows us the code
// to round-trip without modifying the original. Use SetFoldEncoding() if this
// is something you would like to change.
func Parse(m []byte, lb Break) (*Header, error) {
	lines, err := field.ParseLines(m, lb.Bytes())

	var badStartErr *field.BadStartError // recoverable
	var finalErr error
	if errors.As(err, &badStartErr) {
		finalErr = badStartErr
	} else if err != nil {
		return nil, err
	}

	fields := make([]*field.Field, len(lines))
	for i, line := range lines {
		fields[i] = field.Parse(line, lb.Bytes())
	}

	h := &Header{
		Base: Base{
			lbr:    lb,
			vf:     field.DoNotFoldEncoding,
			fields: fields,
		},
		valueCache: nil,
	}

	return h, finalErr
}
