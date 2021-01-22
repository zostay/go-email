package mime

import (
	"github.com/zostay/go-email/pkg/email"
)

type ContentType struct {
	MediaType string
	Params    map[string]string
}

type Message struct {
	email.Message
	EncodingCheck bool
	Depth         int
	ContentType   *ContentType
	Preamble      []byte
	Parts         []*Part
	Epilogue      []byte
}

type Part struct {
	Message
	Prefix []byte
}

func (m *Message) RawContentType() string {
	return m.Get("Content-type")
}
