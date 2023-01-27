package walk

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/zostay/go-email/v2/message"
)

var (
	// ErrSkip may be returned by a Transformer callback to signal that the part
	// should be skipped entirely.
	ErrSkip = errors.New("skip part")

	// ErrCopy may be returned by a Transformer callback to signal that the part
	// should be copied as is.
	ErrCopy = errors.New("clone part")

	// ErrNilNil is returned by AndTransform when a Transformer callback returns
	// no parts and provides no error.
	ErrNilNil = errors.New("no parts and no error")
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
// The Transformer is given the part to transform and the ancestry of the part.
// If len(parents) is zero, then this is the top-level part. The parents are the
// original parents of the given original part, not the transformed parents.
//
// The transform must return zero or more message.Buffer objects or an
// error. If message.Buffer objects are returned, each object must return a
// message.BufferMode other than message.ModeUnset (i.e., either
// message.ModeOpaque or message.ModeMultipart).
//
// If an error is returned, it will result in AndTransform() failing with that
// error.
type Transformer func(part message.Part, parents []message.Part) ([]message.Buffer, error)

// AndTransform will perform a transformation on the given message. This will
// process the message as is (i.e., if there's a message.Opaque part whose bytes
// describe sub-parts, that message.Opaque part will be processed was as a
// single item). The transformation is performed in depth-first order.
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
) ([]message.Part, error) {
	parents := make([]message.Part, 0, 10)
	return andTransform(transformer, msg, parents)
}

func andTransform(
	transformer Transformer,
	msg message.Part,
	parents []message.Part,
) ([]message.Part, error) {
	switch msg.IsMultipart() {
	case true:
		tbufs, err := transformer(msg, parents)
		if err != nil {
			return nil, err
		}

		hasMultipart := false
		for _, tbuf := range tbufs {
			hasMultipart = hasMultipart || tbuf.Mode() == message.ModeMultipart
			if hasMultipart {
				break
			}
		}

		if hasMultipart {

		}
	}
	return nil, fmt.Errorf("not implemented")
}

// func handleTransformationErrors(
// 	origPart message.Part,
// 	newParts []message.Part,
// 	err error,
// ) ([]message.Part, error) {
// 	var bte *BadTransformationError
// 	if newParts == nil && err == nil {
// 		return nil, &BadTransformationError{ErrNilNil, "Transformer error"}
// 	} else if newParts != nil && err == != nil {
// 		return nil, &BadTransformationError{err, "Transformer incorrectly returned error and parts"}
// 	} else if newParts != nil {
// 		return newParts, nil
// 	} else if errors.Is(err, ErrSkip) {
// 		return nil, ErrSkip
// 	} else if errors.Is(err, ErrCopy) {
// 		return copyPart(origPart)
// 	} else {
// 		return nil, err
// 	}
// }

// TransCopyPart provides a handy utility for copying an original part through
// to make a transformed part with no changes. This is intended for use with
// defining a Transformer, so this doesn't exactly copy a part.
//
// A non-multipart part will be copied into message.Opaque by cloning its
// headers and then copying its io.Reader into a buffer attached to the returned
// part.
//
// A multipart part will result in the creation of a new message.Multipart with
// headers cloned from the original. But it will be empty.
func TransCopyPart(orig message.Part) (message.Generic, error) {
	var copy message.Generic
	header := orig.GetHeader().Clone()
	if orig.IsMultipart() {
		copy = &message.Multipart{
			Header: *header,
		}
	} else {
		buf := &bytes.Buffer{}
		_, err := io.Copy(buf, orig.GetReader())
		if err != nil {
			return nil, err
		}

		copy = &message.Opaque{
			Header: *header,
			Reader: buf,
		}
	}

	return copy, nil
}
