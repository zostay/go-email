package header

import (
	"github.com/zostay/go-email/pkg/v2/header/field"
)

// Parse will parse the given slice of bytes into an email header using the
// given line break string. It will assume the entire string given represents
// the header to be parsed.
func Parse(m []byte, lb Break) (*Header, error) {
	lines, err := field.ParseLines(m, lb.Bytes())
	if err != nil {
		return nil, err
	}

	fields := make([]*field.Field, len(lines))
	for i, line := range lines {
		fields[i] = field.Parse(line, lb.Bytes())
	}

	h := &Header{
		Base: Base{
			lbr:    lb,
			vf:     field.DefaultFoldEncoding,
			fields: fields,
		},
		valueCache: nil,
	}

	return h, nil
}
