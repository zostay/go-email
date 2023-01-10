package message

import (
	"errors"
	"fmt"
	"io"

	"github.com/zostay/go-email/v2/header"
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

// Part is an interface define the parts of a Multipart. Each Part is
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

	// IsEncoded will return true if this Part will return the original bytes
	// from the associated io.Reader returned from GetReader(). If it returns
	// false, then the bytes returned from that io.Reader will have had any
	// Content-transfer-encoding decoded first. This does not indicate whether
	// any Content-transfer-encoding header is present or whether the encoding
	// made any changes to the bytes (e.g., the Content-transfer-encoding is set
	// to 7bit, we keep the data as-is and no special decoding is performed by
	// default).
	//
	// This method must return false if IsMultipart() returns true. As transfer
	// encodings cannot be applied to parts with sub-parts, this method makes
	// no sense in that circumstance anyway.
	IsEncoded() bool

	// GetHeader is available on all Part objects.
	GetHeader() *header.Header

	// GetReader provides the content of the message, but only if IsMultipart()
	// returns false. This will return nil and error if IsMultipart() would
	// return true because this is a leaf part. The error that should be
	// returned is ErrMultipart.
	GetReader() (io.Reader, error)

	// GetParts provides the content of a multipart message with sub-parts. This
	// should only be called when IsMultipart() returns true. If you call this
	// when IsMultipart() would return false, a nil and an error will be
	// returned. The error that should be returned is ErrNotMultipart.
	GetParts() ([]Part, error)
}

// Generic is just an alias for Part, which is intended to convey
// additional semantics:
//
// 1. The message returned is not necessarily a sub-part of a message.
//
// 2. The returned message is guaranteed to either be a *Opaque or a
// *Multipart. Therefore, it is safe to use this in a type-switch
// and only look for either of those two objects.
type Generic = Part

// Multipart is a multipart MIME message. When building these methods the MIME
// type set in the Content-type header should always start with multipart/*.
type Multipart struct {
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

// WriteTo writes the Opaque header and parts to the destination io.Writer.
// This method will fail with an error if the given message does not have a
// Content-type boundary parameter set. May return an error on an IO error as
// well.
//
// This may only be safely called one time because it will consume all the bytes
// from all the io.Reader objects associated with all the given Opaque objects
// within.
func (mm *Multipart) WriteTo(w io.Writer) (int64, error) {
	boundary, err := mm.GetBoundary()
	if err != nil {
		return 0, err
	}

	br := mm.Break()

	hn, err := mm.Header.WriteTo(w)
	if err != nil {
		return hn, err
	}

	n := hn
	if len(mm.parts) > 0 {
		first := true
		for _, part := range mm.parts {
			if !first {
				bn, err := fmt.Fprint(w, br)
				n += int64(bn)
				if err != nil {
					return n, err
				}
			} else {
				first = false
			}

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

		bn, err := fmt.Fprintf(w, "%s--%s--%s", br, boundary, br)
		n += int64(bn)
		return n, err
	}

	return n, nil
}

// IsMultipart always returns true.
func (mm *Multipart) IsMultipart() bool {
	return true
}

// IsEncoded always returns false.
func (mm *Multipart) IsEncoded() bool {
	return false
}

// GetHeader returns the header for the message.
func (mm *Multipart) GetHeader() *header.Header {
	return &mm.Header
}

// GetReader always returns nil and ErrMultipart.
func (mm *Multipart) GetReader() (io.Reader, error) {
	return nil, ErrMultipart
}

// GetParts returns the sub-parts of this message or nil if there aren't any.
func (mm *Multipart) GetParts() ([]Part, error) {
	return mm.parts, nil
}
