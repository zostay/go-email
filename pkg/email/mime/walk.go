package mime

import (
	"container/list"
)

type PartWalker func(part *Message) error

// WalkParts executes the given function for every part. This does a depth first
// descent into nested parts. If the given function returns an error, the
// descent will stop and exit with the error. This includes the top-level
// messages as well.
//
// If you just want the leaf parts (the single/non-multipart MIME bits that are
// the usual containers of text, HTML, or attachements), you should try
// WalkSingleParts.
func (m *Message) WalkParts(pw PartWalker) error {
	ms := list.New()

	pushParts := func(parts []*Message) {
		for _, p := range parts {
			ms.PushBack(p)
		}
	}

	shiftPart := func() *Message {
		if ms.Len() == 0 {
			return nil
		}

		e := ms.Front()
		ms.Remove(e)
		return e.Value.(*Message)
	}

	pushParts([]*Message{m})
	pushParts(m.Parts)
	for {
		p := shiftPart()
		if p == nil {
			return nil
		}

		if len(p.Parts) > 0 {
			pushParts(p.Parts)
		}

		err := pw(p)
		if err != nil {
			return err
		}
	}
}

// WalkSingleParts performs the same operation as WalkParts, but it only
// executes the given PartWalker when the part is a leaf, i.e., a single part
// with no sub-parts, i.e., not a multipart part. Got it? The parts are walked
// in the same order they appear in the message. If this is a simple message
// with now parts, the main message itself will be the part you see.
//
// As with WalkSingleParts, this will exit the function if the PartWalker
// returns an error without continuing the traversal.
func (m *Message) WalkSingleParts(pw PartWalker) error {
	return m.WalkParts(func(m *Message) error {
		if len(m.Parts) > 0 {
			return nil
		}

		return pw(m)
	})
}
