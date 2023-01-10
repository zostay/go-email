package message_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/zostay/go-email/v2/message"
)

func ExampleOpaque_WriteTo() {
	buf := bytes.NewBufferString("Hello World")
	msg := &message.Opaque{Reader: buf}
	msg.SetSubject("A message to nowhere")
	_, _ = msg.WriteTo(os.Stdout)
}

func ExampleBuffer_opaque_buffer() {
	buf := &message.Buffer{}
	buf.SetSubject("Some spam for you inbox")
	_, _ = fmt.Fprintln(buf, "Hello World!")
	msg, err := buf.Opaque()
	if err != nil {
		panic(err)
	}
	_, _ = msg.WriteTo(os.Stdout)
}

func ExampleBuffer_multipart_buffer() {
	mm := &message.Buffer{}
	mm.SetSubject("Fancy message")
	mm.SetMediaType("multipart/mixed")

	altPart := &message.Buffer{}
	mm.SetMediaType("multipart/alternative")

	txtPart := &message.Buffer{}
	txtPart.SetMediaType("text/plain")
	txtPart.SetPresentation("attachment")
	_, _ = fmt.Fprintln(txtPart, "Hello *World*!")

	htmlPart := &message.Buffer{}
	htmlPart.SetMediaType("text/html")
	txtPart.SetPresentation("attachment")
	_, _ = fmt.Fprintln(htmlPart, "Hello <b>World</b>!")

	txtMsg, _ := txtPart.Opaque()
	htmlMsg, _ := htmlPart.Opaque()
	_ = altPart.Add(txtMsg, htmlMsg)

	imgAttach := &message.Buffer{}
	imgAttach.SetMediaType("image/jpeg")
	imgAttach.SetPresentation("attachment")
	_ = imgAttach.SetFilename("image.jpg")
	img, _ := os.Open("image.jpg")
	_, _ = io.Copy(imgAttach, img)

	altMsg, _ := altPart.Multipart()
	imgMsg, _ := imgAttach.Opaque()
	_ = mm.Add(altMsg, imgMsg)

	msg, err := mm.Multipart()
	if err != nil {
		panic(err)
	}
	_, _ = msg.WriteTo(os.Stdout)
}
