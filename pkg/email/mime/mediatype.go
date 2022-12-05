package mime

import (
	"errors"
	"mime"
	"strings"
)

// MediaType represents a Content-type or Content-disposition header.  This
// object is intended to be read-only.
type MediaType struct {
	mediaType string            // the content-type itself
	params    map[string]string // additional content-type parameters, like charset, boundary, etc.
	typeSplit int               // index position of the /
}

// ParseMediaType parses a structured media type header using
// mime.ParseMediaType and returns a pointer to a MediaType object. It returns
// an error if there's a problem parsing it.
func ParseMediaType(v string) (*MediaType, error) {
	mt, ps, err := mime.ParseMediaType(v)
	if err != nil {
		return nil, err
	}

	split := strings.IndexRune(mt, '/')
	mediaType := MediaType{mt, ps, split}
	return &mediaType, nil
}

// NewMediaType creates a new media type from the given values. Returns nil and
// an error if an odd number of parameters is given.
func NewMediaType(mt string, ps ...string) (*MediaType, error) {
	split := strings.IndexRune(mt, '/')
	if len(ps)%2 != 0 {
		return nil, errors.New("odd number of parameters when constructing media type")
	}

	mps := make(map[string]string, len(ps))
	var n string
	for i, p := range ps {
		if i%2 == 0 {
			n = p
		} else {
			mps[n] = p
		}
	}

	return &MediaType{mt, mps, split}, nil
}

// NewMediaTypeMap creates a new media type from the given values, using a map
// argument rather than a list of paired strings.
func NewMediaTypeMap(mt string, ps map[string]string) *MediaType {
	split := strings.IndexRune(mt, '/')
	return &MediaType{mt, ps, split}
}

// MediaType returns the media type word. For Content-Type headers, this is the
// MIME Type. For Content-Disposition headers, this will name the type for the
// content, such as "attachment" or "inline".
func (mt *MediaType) MediaType() string { return mt.mediaType }

// Type returns the type of the media type. This is only relevant for the
// Content-type header. Returns an empty string if it doesn't appear to be a
// MIME type.
func (mt *MediaType) Type() string {
	if mt.typeSplit >= 0 {
		return mt.mediaType[:mt.typeSplit]
	} else {
		return ""
	}
}

// Subtype returns the subtype of the media type. This is only relevant for the
// Content-type header. Returns an empty string if it doesn't appear to be a
// MIME type.
func (mt *MediaType) Subtype() string {
	if mt.typeSplit >= 0 {
		return mt.mediaType[mt.typeSplit+1:]
	} else {
		return ""
	}
}

// Parameters returns the map of structured parameters to the media type.
func (mt *MediaType) Parameters() map[string]string {
	return mt.params
}

// Parameter returns the value of the named parameter or the empty string.
func (mt *MediaType) Parameter(n string) string {
	return mt.params[n]
}

// Charset is a short name for
//
//	mt.Parameter("charset")
//
// This is useful for Content-type headers.
func (mt *MediaType) Charset() string {
	return mt.params[CTCharset]
}

// Boundary is a short name for
//
//	mt.Parameter("boundary")
//
// This is useful for Content-type headers.
func (mt *MediaType) Boundary() string {
	return mt.params[CTBoundary]
}

// Filename is a short name for
//
//	mt.Parameter("filename")
//
// This is useful for Content-disposition headers.
func (mt *MediaType) Filename() string {
	return mt.params[CDFilename]
}

// String returnes the formatted representation of the media type object.
func (mt *MediaType) String() string {
	return mime.FormatMediaType(mt.mediaType, mt.params)
}
