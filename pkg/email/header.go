package email

import (
	"bytes"
	"fmt"
	"strings"
)

// Constants for use when selecting a line break to use with a new header. If
// you don't know what to pick, choose the StandardLineBreak.
const (
	CRLF = "\x0d\x0a"
	LF   = "\x0a"
	CR   = "\x0d"
	LFCR = "\x0a\x0d"
)

// Header represents an email message header. The header object stores enough
// detail that the original header can be recreated byte for byte for
// roundtripping.
type Header struct {
	Fields []*HeaderField // The list of fields
	lb     []byte         // The line break string to use
}

// NewHeader creates a new header.
func NewHeader(lb string, f ...*HeaderField) *Header {
	return &Header{f, []byte(lb)}
}

// Break returns the line break string associated with this header.
func (h *Header) Break() []byte {
	if h.lb == nil {
		return []byte(CRLF)
	} else {
		return h.lb
	}
}

// String will return the string representation of the header. If the header was
// parsed from an email header and not modified, this will output the original
// header, preserved byte-for-byte.
func (h *Header) String() string {
	var out strings.Builder
	for _, f := range h.Fields {
		out.WriteString(f.String())
	}
	return out.String()
}

// Bytes will return the byte representation of the header. If the header was
// parsed from an email header and not modified, this will output the original
// header, preserved byte-for-byte.
func (h *Header) Bytes() []byte {
	var out bytes.Buffer
	for _, f := range h.Fields {
		out.Write(f.Bytes())
	}
	return out.Bytes()
}

// HeaderGetField will find the first header field and return the header field object
// itself. It will return nil if no such header is present.
func (h *Header) HeaderGetField(n string) *HeaderField {
	if i := h.HeaderFieldIndex(n, 0, false); i > -1 {
		return h.Fields[i]
	}
	return nil
}

func (h *Header) HeaderFieldIndex(n string, ix int, fb bool) int {
	m := MakeHeaderFieldMatch(n)
	lasti := -1
	if ix < 0 {
		count := -1
		for i := len(h.Fields) - 1; i >= 0; i-- {
			f := h.Fields[i]
			if f.Match() == m {
				lasti = i
				if count == ix {
					return i
				}
				count--
			}
		}
	} else {
		count := 0
		for i, f := range h.Fields {
			if f.Match() == m {
				lasti = i
				if count == ix {
					return i
				}
				count++
			}
		}
	}

	if fb {
		return lasti
	} else {
		return -1
	}
}

// HeaderGetFieldN locates the (ix+1)th named header and returns the header
// field object. If no such header exists, the field is returned as nil and an
// error is returned.
//
// If ix is negative, it will return the (-ix)th header from the end.
func (h *Header) HeaderGetFieldN(n string, ix int) (*HeaderField, error) {
	if i := h.HeaderFieldIndex(n, ix, false); i > -1 {
		return h.Fields[i], nil
	}
	return nil, fmt.Errorf("unable to find index %d of header named %q", ix, n)
}

// HeaderGetAllFields will find all the header fields with a matching name and return
// the list of field objects. It will return any empty slice if no headers with
// this name are present.
func (h *Header) HeaderGetAllFields(n string) []*HeaderField {
	hfs := make([]*HeaderField, 0)
	m := MakeHeaderFieldMatch(n)
	for _, f := range h.Fields {
		if f.Match() == m {
			hfs = append(hfs, f)
		}
	}
	return hfs
}
