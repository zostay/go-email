package walk

import (
	"fmt"
	"io"

	"github.com/zostay/go-email/v2/message"
)

// BadTransformationError is used when transformation needs to fail with an
// error.
type BadTransformationError struct {
	Cause   error
	Message string
}

// Error returns the error message describing the bad transformation.
func (b *BadTransformationError) Error() string {
	return fmt.Sprintf("%s: %v", b.Message, b.Cause)
}

// Unwrap returns the error that caused the bad transformation.
func (b *BadTransformationError) Unwrap() error {
	return b.Cause
}

// Transformer is a callback that can be passed to the AndTransform() function
// to transform a message and its sub-parts into a new message.
//
// The Transformer is given a part to transform and the ancestry of the part. If
// len(parents) is zero, then this is the top-level part (i.e., the top-level
// part that AndTransform() was called upon, which might not be root message).
// The parents are the original parents of the given original part, not the
// transformed parents. The Transformer should only work to transform the part
// given and not it's children. It will be called again to transform those after
// it transforms the parent.
//
// The Transformer function must return zero or more message.Buffer objects or
// an error.
//
// When message.Buffer objects are returned, each object must return a
// message.BufferMode other than message.ModeUnset (i.e., either
// message.ModeOpaque or message.ModeMultipart). If the message.BufferMode of a
// returned message.Buffer is message.ModeMultipart, it may be empty or have
// parts already present but those parts must also be message.Buffer objects and
// so on down the tree, or you may have unexpected results.
//
// If nothing is returned, transformation of children will continue on anyway.
// Any transformed children will be attached to the nearest transformed
// parent(s). If there is no such parent, they will be returned from
// AndTransform(). IF multiple message.Buffer objects are returned, then any
// children transformed will be attached to all of them that have a
// message.BufferMode of message.ModeMultipart.
//
// If an error is returned, it will result in AndTransform() failing with that
// error.
type Transformer func(part message.Part, parents []message.Part) ([]*message.Buffer, error)

// AndTransform will perform a transformation on the given message or message
// part. This will process the message as is (i.e., if there's a message.Opaque
// part whose bytes describe sub-parts, that message.Opaque part will be
// processed was as a single item). The transformation is performed in
// depth-first order. Each part in the message three will be transformed exactly
// once.
//
// The parents will be transformed before the children. If a multipart message
// part is transformed into an opaque part, then its children will not be
// transformed. If a parent is transformed into a multipart and then all of its
// children are skipped, it will also be skipped (empty multipart message parts
// will not be created by this transformation).
//
// The given Transformer is expected to return zero or more message.Buffer
// objects for each part transformed. Each returned message.Buffer must have its
// message.BufferMode set to something other than message.ModeUnset. This
// function will convert each returned message.Buffer to message.Opaque or
// message.Multipart based upon the message.BufferMode. If multiple parts are
// returned and len(parents) is non-zero, then multiple parts will be added to
// the parent to replace the single part in the transformed message. If multiple
// parts are returned and len(parents) is zero, then those parts will be
// returned as multiple messages from this function.
//
// If the Transformer returns an error, this function will immediately fail with
// that error.
//
// This function will return either a message or an error.
func AndTransform(
	transformer Transformer,
	msg message.Part,
) ([]*message.Buffer, error) {
	parents := make([]message.Part, 0, 10)
	return andTransform(transformer, msg, parents)
}

func andTransform(
	transformer Transformer,
	part message.Part,
	parents []message.Part,
) ([]*message.Buffer, error) {
	tParts, err := transformer(part, parents)
	if err != nil {
		return nil, err
	}

	if part.IsMultipart() {
		allTSubParts := make([]*message.Buffer, 0, len(part.GetParts()))
		parents = append(parents, part)
		for _, subPart := range part.GetParts() {
			tSubParts, err := andTransform(transformer, subPart, parents)
			if err != nil {
				return nil, err
			}

			allTSubParts = append(allTSubParts, tSubParts...)
		}

		if len(tParts) > 0 {
			for _, tPart := range tParts {
				tPart.AddBuffers(allTSubParts...)
			}
		} else {
			return allTSubParts, nil
		}
	}

	return tParts, nil
}

// PartToBuffer is a tool similar to message.BufferFrom(), with two
// important differences. First, the return value matches the expected return
// value of Transformer. Second, the message.BufferFrom() function copies an
// object and all of its children recursively. This does not copy andy children
// and only deals with the part itself if it's a multipart part.
func PartToBuffer(orig message.Part) ([]*message.Buffer, error) {
	cp := []*message.Buffer{
		{Header: *orig.GetHeader().Clone()},
	}

	if !orig.IsMultipart() {
		_, err := io.Copy(cp[0], orig.GetReader())
		if err != nil {
			return nil, err
		}
	}

	return cp, nil
}
