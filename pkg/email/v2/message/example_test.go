package message_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/zostay/go-email/pkg/email/v2/message"
)

func ExampleMessage_WriteTo() {
	buf := bytes.NewBufferString("Hello World")
	msg := &message.Message{Reader: buf}
	msg.SetSubject("A message to nowhere")
	msg.WriteTo(os.Stdout)
}

func ExampleBuffer() {
	buf := &message.Buffer{}
	buf.SetSubject("Some spam for you inbox")
	fmt.Fprintln(buf, "Hello World!")
	msg := buf.Message()
	msg.WriteTo(os.Stdout)
}

func ExampleMimeBuffer() {
	mm := &message.MimeBuffer{}
	mm.SetSubject("Fancy message")
	mm.SetContentType("multipart/mixed")

	altPart := &message.MimeBuffer{}
	mm.SetContentType("multipart/alternative")

	txtPart := &message.Buffer{}
	txtPart.SetContentType("text/plain")
	txtPart.SetContentDisposition("attachment")
	fmt.Fprintln(txtPart, "Hello *World*!")

	htmlPart := &message.Buffer{}
	htmlPart.SetContentType("text/html")
	txtPart.SetContentDisposition("attachment")
	fmt.Fprintln(htmlPart, "Hello <b>World</b>!")

	altPart.Add(txtPart.Message(), htmlPart.Message())

	imgAttach := &message.Buffer{}
	imgAttach.SetContentType("image/jpeg")
	imgAttach.SetContentDisposition("attachment")
	imgAttach.SetFilename("image.jpg")
	img, _ := os.Open("image.jpg")
	io.Copy(imgAttach, img)

	mm.Add(altPart.Message(), imgAttach.Message())

	msg := mm.Message()
	msg.WriteTo(os.Stdout)
}
