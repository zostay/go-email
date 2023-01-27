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

	tmsgs, err := walk.AndTransform(
		func(part message.Part, parents []message.Part) ([]*message.Buffer, error) {
			mt, err := part.GetHeader().GetMessageID()
			if err != nil {
				return nil, err
			}

			if mt == "application/pdf" {
				return nil, nil
			}

			return walk.PartToBuffer(part)
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

	_, err = tmsgs[0].WriteTo(tw)
	if err != nil {
		panic(err)
	}
}

func ExampleAndProcessOpaque() {
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

		return nil
	}, m)
	if err != nil {
		panic(err)
	}
}
