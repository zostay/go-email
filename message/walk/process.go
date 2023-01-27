package walk

import "github.com/zostay/go-email/v2/message"

// Processor is a callback that can be passed to the AndProcess() function to
// do any kind of generic processing of a message and its sub-parts.
//
// The Processor is given a part to transform and the ancestry of the part. If
// len(parents) is zero, then this is the top-level part (i.e., the top-level
// part that AndProcess() was called upon, which might not be the root message).
//
// The Processor may return an error to cause message.AndProcess() to terminate
// immediately and return that error.
type Processor func(part message.Part, parents []message.Part) error

// AndProcess will walk the message parts tree of a message (or a part of a
// message) and call the given Processor function for each part found. It will
// terminate once all parts have been processed and return nil. If the Processor
// function returns an error, it will terminate early and return that error.
func AndProcess(
	processor Processor,
	msg message.Part,
) error {
	parents := make([]message.Part, 0, 10)
	return andProcess(processor, msg, parents)
}

func andProcess(
	processor Processor,
	part message.Part,
	parents []message.Part,
) error {
	err := processor(part, parents)
	if err != nil {
		return err
	}

	if part.IsMultipart() {
		parents = append(parents, part)
		for _, subPart := range part.GetParts() {
			err := andProcess(processor, subPart, parents)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
