package walk

import (
	"github.com/zostay/go-email/v2/message"
)

// Transformer is a tool for aiding with advanced transformation of email. It is
// the callback passed to walk.AndTransform().
//
// Each time it is called, the part parameter will refer to the part that is
// being walked for transformation.
//
// The parents parameter will be set to the ancestors of the part. If parents
// has zero length, then the part being transformed is the top-most part being
// transformed (i.e., the one passed to walk.AndTransform() for transformation).
// The elements of parents will always be message.Part objects that return true
// from IsMultipart().
//
// The state parameter will be an array that is the same length as parents. Each
// element will be the stateInit value that was returned when the Transformer
// was called on that parent part. If you are transforming the message into a
// new complex message, you can use this to hold *message.Buffer or some other
// object to manipulate further with sub-part transformations.
//
// The stateInit value returned is used for two possible purposes. As mentioned
// when describing state above, it will be used to initialize the index of state
// that corresponds to the index of parents when the Transformer is called on a
// multipart part. When parents is zero length, it will be used as the return
// value from the call to walk.AndTransform().
//
// If err is set to a non-nil value on return, the walk.AndTransform() function
// will fail immediately with an error.
type Transformer func(
	part message.Part,
	parents []message.Part,
	state []any,
) (stateInit any, err error)

// AndTransform will walk each of the parts of the given message.Part in
// depth-first order. The provided Transformer callback will be called for each
// part.
//
// The return value from the top-most call to Transformer will be the result
// returned from this function.
func AndTransform(
	transformer Transformer,
	msg message.Part,
) (any, error) {
	parents := make([]message.Part, 0, 10)
	state := make([]any, 0, 10)
	return andTransform(transformer, msg, parents, state)
}

func andTransform(
	transformer Transformer,
	part message.Part,
	parents []message.Part,
	state []any,
) (result any, err error) {
	result, err = transformer(part, parents, state)
	if err != nil {
		return
	}

	if part.IsMultipart() {
		parents = append(parents, part)
		state = append(state, result)
		for _, subPart := range part.GetParts() {
			_, err = andTransform(transformer, subPart, parents, state)
			if err != nil {
				return
			}
		}
	}

	return
}
