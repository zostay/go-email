package email

import "errors"

var (
	// ErrIndexOutOfRange signals that the field index was out of range on an
	// operation that does not allow out of range indexing.
	ErrIndexOutOfRange = errors.New("index is out of range")
)

// Header is used to represent an email header. An email header is essentially
// an array of fields. Each field has a name and a value. Each body value has
// additional semantics. Field names are not unique and many are expected to be
// repeated in an email message header. As such, operations that deal with
// fields by name need to be used with care.
type Header interface {
	WithBreak
	Outputter

	// GetField returns the nth field (0-indexed) from the header. Returns nil
	// if the given index is greater than or equal to Size().
	GetField(int) HeaderField

	// Size returns the length of of the header.
	Size() int

	// ListFields returns all header fields in the header.
	ListFields() []HeaderField

	// GetFieldNamed returns a HeaderField with the given name. If the number given
	// is 0, it will return the first. Setting the number to 1 will return the
	// second and so on. If no such header exists, this returns nil.
	GetFieldNamed(string, int) HeaderField

	// GetAllFieldsNamed returns a slice of HeaderField objects containing all the
	// headers with the given name. If no fields exist with that given name, nil
	// is returned.
	GetAllFieldsNamed(string) []HeaderField
}

// MutableHeader defines an email message header that may be modified.
type MutableHeader interface {
	Header

	// InsertBeforeField will insert a field before the nth index in the header
	// (0-indexed). If n is 0 or negative, the new field will become the very
	// last. If the n is greater than or equal to Size(), the header field will
	// be inserted at the end of the header.
	InsertBeforeField(int, string, string)

	// ClearFields will empty the header.
	ClearFields()

	// DeleteField will delete the nth header field (0-indexed) from the header.
	// If the n give is out of bounds an error will be returned.
	DeleteField(int) error
}
