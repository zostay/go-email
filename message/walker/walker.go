package walker

import (
	"github.com/zostay/go-email/v2/message"
)

// PartWalker  is a function that can be processed for each part of a message.
type PartWalker func(depth, i int, part message.Part) error

// Walk performs a depth first search for all the parts of a message starting
// with the message itself. It calls the PartWalker for each part of the
// message. If the PartWalker returns an error, then processing stops
// immediately and the error is returned.
func (w PartWalker) Walk(msg message.Generic) error {
	type part struct {
		depth int
		i     int
		part  message.Generic
	}

	openStack := make([]part, 0, 10)

	pushStack := func(depth int, msg message.Generic) {
		if parts, err := msg.GetParts(); err == nil {
			for i := len(parts) - 1; i >= 0; i-- {
				p := parts[i]
				openStack = append(openStack, part{depth, i, p})
			}
		}
	}

	popStack := func() part {
		end := len(openStack) - 1
		p := openStack[end]
		openStack = openStack[:end]
		return p
	}

	openStack = append(openStack, part{0, 0, msg})
	for len(openStack) > 0 {
		p := popStack()
		if err := w(p.depth, p.i, p.part); err != nil {
			return err
		}
		pushStack(p.depth+1, p.part)
	}

	return nil
}

// WalkOpaque will call the PartWalker function for each Opaque message using a
// depth first traversal. It will terminate the walk immediately if the
// PartWalker returns an error and will return the error.
func (w PartWalker) WalkOpaque(msg message.Generic) error {
	var opw PartWalker = func(depth, i int, part message.Part) error {
		if !part.IsMultipart() {
			if err := w(depth, i, part); err != nil {
				return err
			}
		}
		return nil
	}
	return opw.Walk(msg)
}

// WalkMultipart will call the PartWalker function for each Multiline message
// using a depth first traversal. It will terminate the walk immediately if the
// PartWalker returns an error and will return that error.
func (w PartWalker) WalkMultipart(msg message.Generic) error {
	var mlw PartWalker = func(depth, i int, part message.Part) error {
		if part.IsMultipart() {
			if err := w(depth, i, part); err != nil {
				return err
			}
		}
		return nil
	}
	return mlw.Walk(msg)
}
