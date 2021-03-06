// Package simple provides tools for basic email parsing. This primarily means
// splitting the header from the body, and parsing the header in full. It does
// not do anything special with the body.
//
// This is primarily useful in cases where you want to work with message
// headers. You can read and modify these headers and roundtrip the file without
// making any changes to the body.
//
// If you need to work with the parts of the body, you probably ought to
// consider using the MIME parser in "github.com/zostay/go-email/pkg/email/mime"
// instead.
package simple

import (
	"bytes"
)

// Parse will turn the given string into an email message.
func Parse(m []byte) (*Message, error) {
	pos, crlf := SplitHeadFromBody(m)

	var head, body []byte
	if pos > -1 {
		head = m[0 : pos+len(crlf)]
		body = m[2*len(crlf)+pos:]
	} else {
		head = m
		body = []byte{}
	}

	h, err := ParseHeaderLB(head, crlf)
	return NewMessage(h, body), err
}

// SplitHeadFromBody will detect the index of the split between the message
// header and the message body as well as the line break the email is using. It
// returns both.
func SplitHeadFromBody(m []byte) (int, []byte) {
	var splits = [][]byte{
		[]byte("\x0d\x0a\x0d\x0a"), // \r\n\r\n
		[]byte("\x0a\x0d\x0a\x0d"), // \n\r\n\r, extremely unlikely, possibly never
		[]byte("\x0a\x0a"),         // \n\n
		[]byte("\x0d\x0d"),         // \r\r
	}

	// Find the split between header/body
	var (
		epos  int
		ecrlf []byte
	)
	for _, s := range splits {
		if pos := bytes.Index(m, s); pos > -1 {
			if ecrlf == nil || pos < epos {
				epos = pos
				ecrlf = s[0 : len(s)/2]
			}
		}
	}

	if ecrlf != nil {
		return epos, ecrlf
	}

	// Assume the entire message is the header, but we still need to figure out
	// what the line breaks are...
	for _, s := range splits {
		crlf := s[0 : len(s)/2]
		if bytes.Contains(m, crlf) {
			return -1, crlf
		}
	}

	// And fallback to...
	return -1, []byte("\x0d")
}
