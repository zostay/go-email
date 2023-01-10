package message_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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

func Example_readme_synopsis_1() {
	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	// WithoutMultipart() means we want the top level message only.
	m, err := message.Parse(msg, message.WithoutMultipart())
	if err != nil {
		panic(err)
	}

	out := &message.Buffer{}
	out.Header = *m.GetHeader()
	_, err = io.Copy(out, m.GetReader())
	if err != nil {
		panic(err)
	}

	changed := false
	if kws, err := m.GetHeader().GetKeywords(); err == nil && len(kws) > 0 {
		for _, kw := range kws {
			if kw == "Snuffle" {
				outKws := make([]string, len(kws)+1)
				outKws[len(kws)] = "Upagus"
				out.SetKeywords(outKws...)
				changed = true
			}
		}
	}

	if changed {
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

	var saveAttachments func(message.Generic)
	saveAttachments = func(m message.Generic) {
		if m.IsMultipart() {
			for _, p := range m.GetParts() {
				saveAttachments(p)
			}
			return
		}

		h := m.GetHeader()

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
			_, err = io.Copy(outMsg, m.GetReader())
			if err != nil {
				panic(err)
			}
		}
	}

	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	m, err := message.Parse(msg)
	if err != nil {
		panic(err)
	}

	saveAttachments(m)
}
