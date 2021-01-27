package simple

import (
	"bytes"

	"github.com/zostay/go-email/pkg/email"
)

// Parse will turn the given string into an email message.
func Parse(m []byte) (*email.Message, error) {
	pos, crlf := SplitHeadFromBody(m)

	var head, body []byte
	if pos > -1 {
		head = m[0:pos]
		body = m[2*len(crlf)+pos:]
	} else {
		head = m
		body = []byte{}
	}

	h, err := email.ParseHeaderLB(head, crlf)
	return email.NewMessage(h, body), err
}

// SplitHeadFromBody will detect the index of the split between the message
// header and the message body as well as the line break the email is using. It
// returns both.
func SplitHeadFromBody(m []byte) (int, []byte) {
	var splits = [][]byte{
		[]byte("\x0a\x0d\x0a\x0d"), // \r\n\r\n
		[]byte("\x0d\x0a\x0d\x0a"), // \n\r\n\r, extremely unlikely, possibly never
		[]byte("\x0d\x0d"),         // \n\n
		[]byte("\x0a\x0a"),         // \r\r
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

	// Assume the entire message is the header
	for _, s := range splits {
		crlf := s[0 : len(s)/2]
		if bytes.Index(m, s) > -1 {
			return -1, crlf
		}
	}

	// And fallback to...
	return -1, []byte("\x0d")
}
