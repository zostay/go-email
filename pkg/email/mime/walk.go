package mime

import (
	"container/list"
)

// PartWalker is the callback used to iterate through parts in WalkParts and
// WalkSingleParts.
//
// The first argument is the depth level of the part being passed, which can be
// helpful at understanding where you are in the document.  The top-level
// message is depth 0, sub-parts of it is depth 1, sub-parts of those are level
// 2, etc.
//
// The second argument is the index within that level. At level 0 there will
// only ever be a single index, 0. Sub-parts are indexed starting from 0.
//
// The third argument is a pointer to the Message found.
//
// The function should return an error to immediately stop the process of
// iteration.
type PartWalker func(depth, i int, part *Message) error

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

	type part struct {
		depth int
		i     int
		part  *Message
	}

	pushParts := func(depth int, parts []*Message) {
		for i, p := range parts {
			ms.PushBack(&part{depth, i, p})
		}
	}

	shiftPart := func() *part {
		if ms.Len() == 0 {
			return nil
		}

		e := ms.Front()
		ms.Remove(e)
		return e.Value.(*part)
	}

	pushParts(0, []*Message{m})
	for {
		p := shiftPart()
		if p == nil {
			return nil
		}

		if len(p.part.Parts) > 0 {
			pushParts(p.depth+1, p.part.Parts)
		}

		err := pw(p.depth, p.i, p.part)
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
	return m.WalkParts(func(depth, i int, m *Message) error {
		if len(m.Parts) > 0 {
			return nil
		}

		return pw(depth, i, m)
	})
}
