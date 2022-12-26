package message

import (
	"errors"
	"fmt"

	"github.com/zostay/go-email/pkg/email/v2/header"
	"github.com/zostay/go-email/pkg/email/v2/mime"
)

const (
	DefaultMultipartContentType = "multipart/mixed"
)

// MimeBuffer provides tools for constructing complete MIME messages. It is
// built up a piece at a time by adding sub-messages to the buffer via the Add()
// method. You are responsible for ensuring that each of the parts are correct.
type MimeBuffer struct {
	header.Header
	parts  []*Message
}

func (mb *MimeBuffer) initParts() {
	if mb.parts == nil {
		mb.parts = make([]*Message, 0, 10)
	}
}

// Add will add one or more parts to the message.
func (mb *MimeBuffer) Add(msgs ...*Message) {
	mb.initParts()
	for _, msg := range msgs {
		mb.parts = append(mb.parts, msg)
	}
}

// Message will return a MIME multipart-style message based upon the parts that
// have been added to the buffer.
//
// You should set the Content-Type header yourself to one of the multipart/*
// types (e.g., multipart/alternative if you aare providing text and HTML forms
// of the same message or multipart/mixed if you are providing attachments).
//
// If you do not provide that header yourself, this method will set it to
// DefaultMultipartContentType, which may not be what you want.
//
// It will also check to see if the Content-type boundary is set and set it to
// something random using mime.GenerateBound() automatically. This boundary
// will then be used to join the parts together.
func (mb *MimeBuffer) Message() *Message {
	if _, err := mb.GetContentType(); errors.Is(err, header.ErrNoSuchField) {
		mb.SetContentType(DefaultMultipartContentType)
	}

	if _, err := mb.GetBoundary(); errors.Is(err, header.ErrNoSuchFieldParameter) {
		mb.SetBoundary(mime.GenerateBoundary())
	}

	boundary, _ := mb.GetBoundary()

	buf := &Buffer{Header: mb.Header}
	if len(mb.parts) == 0 {
		return buf.Message()
	}

	for _, part := range mb.parts {
		fmt.Fprintf(buf, "--%s%s", boundary, mb.Break())
		part.WriteTo(buf)
	}
	fmt.Fprintf(buf, "--%s%s", boundary, mb.Break())

	return buf.Message()
}
