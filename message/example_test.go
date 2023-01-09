package message_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/zostay/go-email/v2/message"
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
	mm.SetMediaType("multipart/mixed")

	altPart := &message.Buffer{}
	mm.SetMediaType("multipart/alternative")

	txtPart := &message.Buffer{}
	txtPart.SetMediaType("text/plain")
	txtPart.SetPresentation("attachment")
	fmt.Fprintln(txtPart, "Hello *World*!")

	htmlPart := &message.Buffer{}
	htmlPart.SetMediaType("text/html")
	txtPart.SetPresentation("attachment")
	fmt.Fprintln(htmlPart, "Hello <b>World</b>!")

	txtMsg, _ := txtPart.Opaque()
	htmlMsg, _ := htmlPart.Opaque()
	altPart.Add(txtMsg, htmlMsg)

	imgAttach := &message.Buffer{}
	imgAttach.SetMediaType("image/jpeg")
	imgAttach.SetPresentation("attachment")
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
