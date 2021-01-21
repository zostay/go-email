package email

import (
	"io"
)

type Message struct {
	Header *Header
	Body   io.Reader
}
