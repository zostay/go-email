package message_test

import (
	"bytes"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/transfer"
	"github.com/zostay/go-email/v2/message/walker"
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

	altPart.Add(txtPart.Opaque(), htmlPart.Opaque())

	imgAttach := &message.Buffer{}
	imgAttach.SetMediaType("image/jpeg")
	imgAttach.SetPresentation("attachment")
	_ = imgAttach.SetFilename("image.jpg")
	img, _ := os.Open("image.jpg")
	_, _ = io.Copy(imgAttach, img)

	mm.Add(altPart.Opaque(), imgAttach.Opaque())

	_, _ = mm.Opaque().WriteTo(os.Stdout)
}

func Example_readme_synopsis_1() {
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

func Example_readme_synopsis_2() {
	var fileCount = 0
	isUnsafeExt := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsDigit(c)
	}

	outputSafeFilename := func(fn string) string {
		safeExt := filepath.Ext(fn)
		if strings.IndexFunc(safeExt, isUnsafeExt) > -1 {
			safeExt = ".wasnotsafe" // check your input
		}
		fileCount++
		return fmt.Sprintf("%d.%s", fileCount, safeExt)
	}

	var saveAttachments walker.Parts = func(depth, i int, part message.Part) error {
		h := part.GetHeader()

		presentation, err := h.GetPresentation()
		if err != nil {
			panic(err)
		}

		fn, err := h.GetFilename()
		if err != nil {
			panic(err)
		}

		if presentation == "attachment" && fn != "" {
			of := outputSafeFilename(fn)
			outMsg, err := os.Create(of)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(outMsg, part.GetReader())
			if err != nil {
				panic(err)
			}
		}
	}

	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	// we want to decode the transfer encoding to make sure we get the original
	// binary values of the message contents when saving off attachments
	m, err := message.Parse(msg, message.DecodeTransferEncoding())
	if err != nil {
		panic(err)
	}

	_ = saveAttachments.WalkOpaque(m)
}

func Example_readme_synopsis_3() {
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
	main.SetTo("recruiter@example.com")
	main.SetFrom("me@example.com")
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
