package simple

import (
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

// Parse will turn the given string into an email message.
func Parse(m string) (*email.Message, error) {
	pos, crlf := SplitHeadFromBody(m)

	var head, body string
	if pos > -1 {
		head = m[0 : pos-len(crlf)]
		body = m[pos:]
	} else {
		head = m
		body = ""
	}

	h, err := email.ParseHeaderLB(head, crlf)
	return &email.Message{
		Header: h,
		Body:   strings.NewReader(body),
	}, err
}

// SplitHeadFromBody will detect the index of the split between the message
// header and the message body as well as the line break the email is using. It
// returns both.
func SplitHeadFromBody(m string) (int, string) {
	var splits = []string{
		"\x0a\x0d\x0a\x0d", // \r\n\r\n
		"\x0d\x0a\x0d\x0a", // \n\r\n\r, extremely unlikely, possibly never
		"\x0d\x0d",         // \n\n
		"\x0a\x0a",         // \r\r
	}

	// Find the split between header/body
	for _, s := range splits {
		if pos := strings.Index(m, s); pos > -1 {
			crlf := s[0 : len(s)/2]
			return pos, crlf
		}
	}

	// Assume the entire message is the header
	for _, s := range splits {
		crlf := s[0 : len(s)/2]
		if strings.Index(m, s) > -1 {
			return -1, crlf
		}
	}

	// And fallback to...
	return -1, "\x0d"
}
