package email

// Header is used to represent an email header. An email header is essentially
// an array of fields. Each field has a name and a value. Each body value has
// additional semantics. Field names are not unique and many are expected to be
// repeated in an email message header. As such, operations that deal with
// fields by name need to be used with care.
type Header interface {
	WithBreak
	Outputter

	// GetField returns the first HeaderField for the given name or nil if no
	// such field is currently set.
	GetField(string) HeaderField

	// GetFieldN returns a HeaderField with the given name. If the number given
	// is 0, it will return the first. Setting the number to 1 will return the
	// second and so on. If no such header exists, this returns nil.
	GetFieldN(string, int) HeaderField

	// GetAllFields returns a slice of HeaderField objects containing all the
	// headers with the given name. If no fields exist with that given name, nil
	// is returned.
	GetAllFields(string) []HeaderField
}
