package walk_test

import (
	"os"

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
