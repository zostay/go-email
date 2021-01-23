package mime

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/zostay/go-email/pkg/email"
)

type Encoder func(m *Message, b []byte) (string, error)
type Decoder func(m *Message, s string) ([]byte, error)

var (
	CharsetEncoder Encoder = DefaultCharsetEncoder
	CharsetDecoder Decoder = DefaultCharsetDecoder
)

func DefaultCharsetEncoder(m *Message, b []byte) (string, error) {
	charset := m.Charset()
	if m.EncodingCheck() {
		switch strings.ToLower(charset) {
		case "us-ascii", "":
			for _, c := range b {
				if c > unicode.MaxASCII {
					return "", fmt.Errorf("invalid %q encoding", charset)
				}
			}
			return string(b), nil
		case "utf-8":
			if utf8.Valid(b) {
				return string(b), nil
			}
			return "", fmt.Errorf("invalid %q encoding", charset)
		default:
			return "", fmt.Errorf("unsupported byte encoding %q", charset)
		}
	} else {
		return string(b), nil
	}
}

func DefaultCharsetDecoder(m *Message, s string) ([]byte, error) {
	charset := m.Charset()
	if m.EncodingCheck() {
		switch strings.ToLower(charset) {
		case "us-ascii", "":
			for _, c := range s {
				if c > 127 {
					return nil, fmt.Errorf("invalid %q encoding", charset)
				}
			}
			return []byte(s), nil
		case "utf-8":
			if utf8.ValidString(s) {
				return []byte(s), nil
			}
			return nil, fmt.Errorf("invalid %q encoding", charset)
		default:
			return nil, fmt.Errorf("unsupported byte encoding %q", charset)
		}
	} else {
		return []byte(s), nil
	}
}

type ContentType struct {
	mediaType string
	params    map[string]string
}

type Message struct {
	email.Message
	encodingCheck bool
	depth         int
	contentType   *ContentType
	prefix        []byte
	preamble      []byte
	parts         []*Message
	epilogue      []byte
	parent        *Message
}

func (m *Message) rebuildBody() {
	var a strings.Builder
	a.Write(m.prefix)
	a.Write(m.preamble)
	for _, p := range m.parts {
		p.rebuildBody()
		a.Write(p.Body())
	}
	a.Write(m.epilogue)
	m.SetBodyString(a.String())
	if m.parent != nil {
		m.parent.rebuildBody()
	}
}

func (m *Message) RawContentType() string {
	return m.Get("Content-type")
}

func (m *Message) ContentType() string {
	return m.contentType.mediaType
}

func (m *Message) Charset() string {
	return m.contentType.params["charset"]
}

func (m *Message) Premble() string {
	return string(m.preamble)
}

func (m *Message) SetPreamble(p string) {
	m.preamble = []byte(p)
	m.rebuildBody()
}

func (m *Message) Epilogue() string {
	return string(m.epilogue)
}

func (m *Message) SetEpilogue(e string) {
	m.epilogue = []byte(e)
	m.rebuildBody()
}

func (m *Message) PartsLen() int          { return len(m.parts) }
func (m *Message) GetPart(i int) *Message { return m.parts[i] }

func (m *Message) Parent() *Message { return m.parent }

func (m *Message) SetBody(b []byte) error {
	m.preamble = nil
	m.parts = nil
	m.epilogue = nil
	m.Message.SetBody(b)

	return m.FillParts()
}

func (m *Message) SetBodyString(s string) error {
	m.preamble = nil
	m.parts = nil
	m.epilogue = nil
	m.Message.SetBodyString(s)

	return m.FillParts()
}

func (m *Message) EncodingCheck() bool { return m.encodingCheck }

func (m *Message) SetEncodingCheck(ec bool) { m.encodingCheck = ec }

// func (m *Message) BodyString() string {
// }
