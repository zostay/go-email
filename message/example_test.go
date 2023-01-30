package message_test

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"strings"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/transfer"
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
	msg := buf.Opaque()
	_, _ = msg.WriteTo(os.Stdout)
}

func ExampleBuffer_multipart_buffer() {
	mm := &message.Buffer{}
	mm.SetSubject("Fancy message")
	mm.SetMediaType("multipart/mixed")

	txtPart := &message.Buffer{}
	txtPart.SetMediaType("text/plain")
	_, _ = fmt.Fprintln(txtPart, "Hello *World*!")

	htmlPart := &message.Buffer{}
	htmlPart.SetMediaType("text/html")
	_, _ = fmt.Fprintln(htmlPart, "Hello <b>World</b>!")

	mm.Add(message.MultipartAlternative(txtPart.Opaque(), htmlPart.Opaque()))

	imgAttach, _ := message.AttachmentFile(
		"image.jpg",
		"image/jpeg",
		transfer.Base64,
	)
	mm.Add(imgAttach)

	_, _ = mm.Opaque().WriteTo(os.Stdout)
}

func Example_rewrite_keywords() {
	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	// WithoutMultipart() means we want the top level headers only.
	m, err := message.Parse(msg, message.WithoutMultipart())
	if err != nil {
		panic(err)
	}

	// update the keywords of the new message
	if kws, err := m.GetHeader().GetKeywords(); err == nil && len(kws) > 0 {
		for _, kw := range kws {
			if kw == "Snuffle" {
				out := &message.Buffer{}
				out.Header = *m.GetHeader() // copy the original header
				content := m.GetReader()
				_, err = io.Copy(out, content) // copy the original message body
				if err != nil {
					panic(err)
				}

				// add Upagus to Keywords
				outKws := make([]string, len(kws)+1)
				outKws[len(kws)] = "Upagus"
				out.SetKeywords(outKws...)

				outMsg, err := os.Create("output.msg")
				if err != nil {
					panic(err)
				}

				_, err = out.WriteTo(outMsg)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func Example_create_message() {
	// Build a part that will be the attached document
	resume, _ := message.AttachmentFile(
		"resume.pdf",
		"application/pdf",
		transfer.Base64,
	)

	// Build a part that will contain the message content as text
	text := &message.Buffer{}
	text.SetMediaType("text/plain")
	_, _ = fmt.Fprintln(text, "You will find my awesome resume attached.")

	// Build a part that will contain the message content as HTML
	html := &message.Buffer{}
	html.SetMediaType("text/html")
	_, _ = fmt.Fprintln(html, "You will find my <strong>awesome</strong> resume attached.")

	// Build the top-level message from the parts.
	main := &message.Buffer{}
	main.SetSubject("My resume")
	_ = main.SetTo("recruiter@example.com")
	_ = main.SetFrom("me@example.com")
	main.SetMediaType("multipart/mixed")
	main.Add(
		message.MultipartAlternative(html.Opaque(), text.Opaque()),
		resume,
	)
	mainMsg := main.Opaque()

	// send the message via SMTP
	c, err := smtp.Dial("smtp.example.com:25")
	if err != nil {
		panic(err)
	}

	_ = c.Hello("me")
	_ = c.Mail("me@example.com")
	_ = c.Rcpt("recruiter@example.com")
	w, err := c.Data()
	if err != nil {
		panic(err)
	}
	_, _ = mainMsg.WriteTo(w)
	_ = w.Close()
	_ = c.Quit()
}

func ExampleParse() {
	r := strings.NewReader("Subject: test\n\nThis is a test.")
	msg, err := message.Parse(r)
	if err != nil {
		panic(err)
	}
	// snip end

	_, _ = msg.WriteTo(os.Stdout)
}

func ExampleParse_options() {
	var r io.Reader
	// snip start

	// This will only parse to the 5th layer deep.
	m, _ := message.Parse(r, message.WithMaxDepth(5)) //nolint:ineffassign

	// This will not parse even the first layer.
	// This always returns an *message.Opaque object.
	m, _ = message.Parse(r, message.WithoutMultipart()) //nolint:ineffassign
	// ^^^ same as WithMaxDepth(0)

	// This will parse the first layer, but no further. If the message is a
	// multipart message it will be *message.Multipart but all sub-parts are
	// guaranteed to be *message.Opaque. Otherwise, it may return *message.Opaque.
	m, _ = message.Parse(r, message.WithoutRecursion()) //nolint:ineffassign
	// ^^^ same as WithMaxDepth(1)

	// Or you can turn off all limits and get everything...
	m, _ = message.Parse(r, message.WithUnlimitedRecursion())
	// ^^^ ame as WithMaxDepth(-1)
	// snip end

	_, _ = m.WriteTo(os.Stdout)
}
