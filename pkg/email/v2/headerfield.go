package email

import (
	"time"

	"github.com/zostay/go-addr/pkg/addr"
)

// HeaderField represents an individual field in a header.
type HeaderField interface {
	// Name returns the header field name.
	Name() string

	// Body returns the header field body.
	Body() string
}

// MutableHeaderField is a mutable header field in a header.
type MutableHeaderField interface {
	HeaderField

	// SetName modifies the name field.
	SetName(string)

	// SetBody sets the body to given value.
	SetBody(string)
}

// DateHeaderField marks the body as capable of returning a Date.
type DateHeaderField interface {
	// IsBodyDate returns true if the header field body is available as a
	// time.Time.
	IsBodyDate() bool

	// BodyDate returns the header field body as a time.Time.
	BodyDate() time.Time
}

// AddressListHeaderField marks the body as capable of returning an
// addr.AddressList.
type AddressListHeaderField interface {
	// IsBodyAddressList returns true if the header field body is available as
	// an addr.AddressList.
	IsBodyAddressList() bool

	// BodyAddressList returns the header field body as an addr.AddressList.
	BodyAddressList() addr.AddressList
}

// ParameterizedHeaderField marks the body as capable of returning a ParameterizedValue.
type ParameterizedHeaderField interface {
	// IsBodyParameterized returns true if the header field body is available as
	// ParameterizedValue.
	IsBodyParameterized() bool

	// BodyParameterized returns the header field body as a ParameterizedValue.
	BodyParameterized() ParameterizedValue
}

// MediaTypeHeaderField marks the body as capable of returning a MediaType.
type MediaTypeHeaderField interface {
	// IsBodyMediaType returns true if the header field body is available as
	// a MediaType.
	IsBodyMediaType() bool

	// BodyMediaType returns the header field body as a MediaType.
	BodyMediaType() MediaType
}

// DispositionHeaderField marks the body as capable of returning a Disposition.
type DispositionHeaderField interface {
	// IsBodyDisposition returns true if the header field body is available as
	// a Disposition.
	IsBodyDisposition() bool

	// BodyDisposition returns the header field body as a Disposition.
	BodyDisposition() Disposition
}
