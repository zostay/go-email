package mime

import (
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
	preamble      []byte
	parts         []*Part
	epilogue      []byte
}

type Part struct {
	Message
	Prefix []byte
}

func (m *Message) RawContentType() string {
	return m.Get("Content-type")
}
