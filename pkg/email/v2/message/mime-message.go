package message

import (
	"errors"
	"fmt"
	"io"

	"github.com/zostay/go-email/pkg/email/v2/header"
)

var (
	// ErrNotMultipart is the error that should be returned from GetParts() when
	// that method is called on a Part that is not a branch Part with
	// sub-parts.
	ErrNotMultipart = errors.New("not a multipart message")

	// ErrMultipart is the error that should be returend from GetReader() when
	// that method is called on a Part that is not a leaf Part with a
	// body to be read.
	ErrMultipart = errors.New("this is a multipart message")
)

// Part is an interface define the parts of a MultipartMessage. Each Part is
// either a branch or a leaf.
//
// A branch Part is one that has sub-parts. In this case, the IsMultipart()
// method will return true. The GetParts() method is available, but the
// GetReader() must not be called.
//
// A leaf Part is one that contains content. In this case, the IsMultipart()
// method will return false. The GetParts() method must not be called on a leaf
// Part. However, the GetReader() method will return a reader for reading
// the content of the part.
//
// It should be noted that it is possible for a Part to contain content that
// is a multipart MIME message when IsMultipart() returns false. If the
// sub-parts have been serialized such that the parts are not provided
// separately. This is perfectly legal.
type Part interface {
	io.WriterTo

	// IsMultipart will return true if this Part is a branch with nested
	// parts. You may call the GetParts() method to process the parts only if
	// this returns true. If it returns false, this Part is a leaf and it
	// has no sub-parts. You may call GetReader() only when this method returns
	// false.
	//
	// It is okay to skip call this and just call the GetReader() or GetParts()
	// methods directly so long as you check for the errors they may return.
	IsMultipart() bool

	// GetHeader is available on all Part objects.
	GetHeader() *header.Header

	// GetReader provides the content of the message, but only if IsMultipart()
	// returns false. This will return nil and error if IsMultipart() would
	// return true because this is a leaf part. The error that should be
	// returned is ErrMultipart.
	GetReader() (io.Reader, error)

	// GetParts provides teh content of a multipart message with sub-parts. This
	// should only be called when IsMultipart() returns true. If you call this
	// when IsMultipart() would return false, a nil and an error will be
	// returned. The error that should be returned is ErrNotMultipart.
	GetParts() ([]Part, error)
}

// GenericMessage is just an alias for Part, which is intended to convey the
// meaning that the message returned is not necessary a sub-part of a parent
// message.
//
// Where this type is used, we guarantee the following semantics: the object
// returned is either a *Message or a *MultipartMessage (or nil, in case of an
// error).
type GenericMessage = Part

// MultipartMessage is a multipart MIME message. When building these methods the MIME
// type set in the Content-type header should always start with multipart/*.
type MultipartMessage struct {
	// Header is the header for the message.
	header.Header

	// prefix and suffix are here so can do a byte-for-byte round trip in case
	// there are extra bytes before the first boundary that don't look like a
	// part or after the last boundary that don't look like a part (as in, just
	// whitespace)
	//
	// Some special semantics:
	//
	// * if prefix is nil, then the email consists of no internal boundaries
	// (though, it may have a final boundary). When round-tripping, no initial
	// boundary will be output.
	//
	// * if suffix is nil, then the email lacks a final boundary. When
	// round-tripping, no final boundary will be output.
	prefix, suffix []byte

	// parts holds this layer's parts
	parts []Part
}

func (mm *MultipartMessage) initParts() {
	if mm.parts == nil {
		mm.parts = make([]Part, 0, 10)
	}
}

// WriteTo writes the Message header and parts to the destination io.Writer.
// This method will fail with an error if the given message does not have a
// Content-type boundary parameter set. May return an error on an IO error as
// well.
func (mm *MultipartMessage) WriteTo(w io.Writer) (int64, error) {
	boundary, err := mm.GetBoundary()
	if err != nil {
		return 0, err
	}

	br := mm.Break()

	hb := mm.Header.Bytes()
	hn, err := w.Write(hb)
	if err != nil {
		return int64(hn), err
	}

	n := int64(hn)
	if len(mm.parts) > 0 {
		for _, part := range mm.parts {
			bn, err := fmt.Fprintf(w, "--%s%s", boundary, br)
			n += int64(bn)
			if err != nil {
				return n, err
			}

			pn, err := part.WriteTo(w)
			n += pn
			if err != nil {
				return n, err
			}
		}

		bn, err := fmt.Fprintf(w, "--%s%s", boundary, br)
		n += int64(bn)
		return n, err
	}

	return n, nil
}

// IsMultipart always returns true.
func (mm *MultipartMessage) IsMultipart() bool {
	return true
}

// GetHeader returns the header for the message.
func (mm *MultipartMessage) GetHeader() *header.Header {
	return &mm.Header
}

// GetReader always returns nil and ErrMultipart.
func (mm *MultipartMessage) GetReader() (io.Reader, error) {
	return nil, ErrMultipart
}

// GetParts returns the sub-parts of this message or nil if there aren't any.
func (mm *MultipartMessage) GetParts() ([]Part, error) {
	return mm.parts, nil
}
