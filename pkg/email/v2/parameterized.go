package email

// ParameterizedValue is a header value that contains sub-fields inside the
// parameter. The Content-Type and Content-Disposition headers are examples of
// ParameterizedValue headers.
type ParameterizedValue interface {
	// Value returns the primary value of the header as a string.
	Value() string

	// Parameters returns a map containing the parameters of the body value.
	Parameters() map[string]string

	// Parameter returns the value as a string for the named parameter.
	Parameter(string) string
}

// Disposition is a ParameterizedValue, but with special accessors for use
// with a Content-Disposition header.
type Disposition interface {
	ParameterizedValue

	// Filename returns the filename parameter.
	Filename() string
}

// MediaType is a ParameterizedValue, but with special accessors for use with a
// Content-type header.
type MediaType interface {
	ParameterizedValue

	// Charset returns the character set parameter.
	Charset() string

	// Boundary returns the boundary parameter.
	Boundary() string
}
