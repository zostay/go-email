package walker

import (
	"github.com/zostay/go-email/v2/message"
)

// Parts is a function that can be processed for each part of a message.
//
// Deprecated: Use walk.Processor with walk.AndProcess() or walk.Transformer
// with walk.AndTransform() instead.
type Parts func(depth, i int, part message.Part) error

// Walk performs a depth first search for all the parts of a message starting
// with the message itself. It calls the Parts for each part of the
// message. If the Parts function returns an error, then processing stops
// immediately and the error is returned.
//
// Deprecated: Use walk.AndProcess() instead.
func (w Parts) Walk(msg message.Generic) error {
	type part struct {
		depth int
		i     int
		part  message.Generic
	}

	openStack := make([]part, 0, 10)

	pushStack := func(depth int, msg message.Generic) {
		parts := msg.GetParts()
		for i := len(parts) - 1; i >= 0; i-- {
			p := parts[i]
			openStack = append(openStack, part{depth, i, p})
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

// WalkOpaque will call the Parts function for each Opaque message using a
// depth first traversal. It will terminate the walk immediately if the
// Parts function returns an error and will return the error.
//
// Deprecated: Use walk.AndProcessOpaque() instead.
func (w Parts) WalkOpaque(msg message.Generic) error {
	var opw Parts = func(depth, i int, part message.Part) error {
		if !part.IsMultipart() {
			if err := w(depth, i, part); err != nil {
				return err
			}
		}
		return nil
	}
	return opw.Walk(msg)
}

// WalkMultipart will call the Parts function for each Multiline message using a
// depth first traversal. It will terminate the walk immediately if the Parts
// function returns an error and will return that error.
//
// Deprecated: Use walk.AndProcessMultipart() instead.
func (w Parts) WalkMultipart(msg message.Generic) error {
	var mlw Parts = func(depth, i int, part message.Part) error {
		if part.IsMultipart() {
			if err := w(depth, i, part); err != nil {
				return err
			}
		}
		return nil
	}
	return mlw.Walk(msg)
}
