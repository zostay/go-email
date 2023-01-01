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
	msg := &message.Opaque{Reader: buf}
	msg.SetSubject("A message to nowhere")
	msg.WriteTo(os.Stdout)
}

func ExampleOpaqueBuffer() {
	buf := &message.Buffer{}
	buf.SetSubject("Some spam for you inbox")
	fmt.Fprintln(buf, "Hello World!")
	msg, err := buf.Opaque()
	if err != nil {
		panic(err)
	}
	msg.WriteTo(os.Stdout)
}

func ExampleMultipartBuffer() {
	mm := &message.Buffer{}
	mm.SetSubject("Fancy message")
	mm.SetContentType("multipart/mixed")

	altPart := &message.Buffer{}
	mm.SetContentType("multipart/alternative")

	txtPart := &message.Buffer{}
	txtPart.SetContentType("text/plain")
	txtPart.SetContentDisposition("attachment")
	fmt.Fprintln(txtPart, "Hello *World*!")

	htmlPart := &message.Buffer{}
	htmlPart.SetContentType("text/html")
	txtPart.SetContentDisposition("attachment")
	fmt.Fprintln(htmlPart, "Hello <b>World</b>!")

	txtMsg, _ := txtPart.Opaque()
	htmlMsg, _ := htmlPart.Opaque()
	altPart.Add(txtMsg, htmlMsg)

	imgAttach := &message.Buffer{}
	imgAttach.SetContentType("image/jpeg")
	imgAttach.SetContentDisposition("attachment")
	imgAttach.SetFilename("image.jpg")
	img, _ := os.Open("image.jpg")
	io.Copy(imgAttach, img)

	altMsg, _ := altPart.Multipart()
	imgMsg, _ := imgAttach.Opaque()
	mm.Add(altMsg, imgMsg)

	msg, err := mm.Multipart()
	if err != nil {
		panic(err)
	}
	msg.WriteTo(os.Stdout)
}
