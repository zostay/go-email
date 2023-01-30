package walk_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/walk"
)

func ExampleAndTransform() {
	msgText, err := os.Open("message.txt")
	if err != nil {
		panic(err)
	}
	defer msgText.Close()

	msg, err := message.Parse(msgText)
	if err != nil {
		panic(err)
	}

	// Strip PDFs from a message.
	tmsg, err := walk.AndTransform(
		func(part message.Part, parents []message.Part, state []any) (any, error) {
			if part.IsMultipart() {
				buf := message.NewBlankBuffer(part)
				return buf, nil
			}

			mt, err := part.GetHeader().GetMediaType()
			if err != nil {
				return nil, err
			}

			if mt == "application/pdf" {
				return nil, nil
			}

			buf, err := message.NewBuffer(part)
			if err != nil {
				return nil, err
			}

			state[len(state)-1].(*message.Buffer).Add(buf)

			return buf, nil
		},
		msg)
	if err != nil {
		panic(err)
	}

	tw, err := os.Create("outmessage.txt")
	if err != nil {
		panic(err)
	}
	defer tw.Close()

	_, err = tmsg.(message.Part).WriteTo(tw)
	if err != nil {
		panic(err)
	}
}

var fileCount = 0

func IsUnsafeExt(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsDigit(c)
}

func OutputSafeFilename(fn string) string {
	safeExt := filepath.Ext(fn)
	if strings.IndexFunc(safeExt, IsUnsafeExt) > -1 {
		safeExt = ".wasnotsafe" // check your input
	}
	fileCount++
	return fmt.Sprintf("%d.%s", fileCount, safeExt)
}

func ExampleAndProcessOpaque() {
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

	// Write out every attachment as a local file.
	err = walk.AndProcessOpaque(func(part message.Part, _ []message.Part) error {
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
			of := OutputSafeFilename(fn)
			outMsg, err := os.Create(of)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(outMsg, part.GetReader())
			if err != nil {
				panic(err)
			}
		}

		return nil
	}, m)
	if err != nil {
		panic(err)
	}
}

func ExampleAndProcess_flatten() {
	var m message.Generic
	bufs := make([]*message.Buffer, 0, 10)
	err := walk.AndProcess(
		func(part message.Part, _ []message.Part) error {
			var (
				buf *message.Buffer
				err error
			)
			if part.IsMultipart() {
				buf = message.NewBlankBuffer(part)
			} else {
				buf, err = message.NewBuffer(part)
				if err != nil {
					return err
				}
			}
			bufs = append(bufs, buf)
			return nil
		}, m)
	if err != nil {
		panic(err)
	}
}

func ExampleAndProcess_mark_evil() {
	var m message.Generic
	tm, err := walk.AndTransform(
		func(part message.Part, parents []message.Part, state []any) (any, error) {
			buf := message.NewBlankBuffer(part)
			if !part.IsMultipart() {
				mt, _ := part.GetHeader().GetMediaType()
				if mt == "text/plain" {
					_, _ = fmt.Fprint(buf, "This content is evil.\n\n")
				} else if mt == "text/html" {
					_, _ = fmt.Fprint(buf, "This content is evil.<br><br>")
				}

				_, err := io.Copy(buf, part.GetReader())
				if err != nil {
					return nil, err
				}
			}

			pbuf := state[len(state)-1].(*message.Buffer)
			pbuf.Add(buf)

			return buf, nil
		}, m)

	if err != nil {
		panic(err)
	}

	_, err = tm.(message.Part).WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
