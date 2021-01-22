package mime

import (
	"strings"

	"github.com/zostay/go-email/pkg/email"
)

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
	m.parent.rebuildBody()
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
