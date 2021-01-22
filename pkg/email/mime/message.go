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
	Preamble      string
	Parts         []*Part
	Epilogue      string
}

type Part struct {
	Message
	Prefix string
}

func (m *Message) RawContentType() string {
	return m.Header.Get("Content-type")
}
