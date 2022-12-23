package email

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
