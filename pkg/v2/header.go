package email

import (
	"errors"
	"time"

	"github.com/zostay/go-addr/pkg/addr"

	"github.com/zostay/go-email/pkg/email/v2/param"
)

var (
	// ErrIndexOutOfRange signals that the field index was out of range on an
	// operation that does not allow out of range indexing.
	ErrIndexOutOfRange = errors.New("index is out of range")
)

// ImmutableHeader is used to represent an email header. An email header is essentially
// an array of fields. Each field has a name and a value. Each body value has
// additional semantics. Field names are not unique and many are expected to be
// repeated in an email message header. As such, operations that deal with
// fields by name need to be used with care.
type ImmutableHeader interface {
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

	// GetIndexesNamed returns the indexes of all fields with the given name.
	GetIndexesNamed(string) []int
}

// MutableHeader defines an email message header that may be modified.
type MutableHeader interface {
	ImmutableHeader
	WithMutableBreak

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

type BasicHeader = MutableHeader

var (
	// ErrNoSuchHeader is returned by FancyHeader methods when the operation
	// being performed failed because the header named does not exist.
	ErrNoSuchHeader = errors.New("not such header")

	// ErrNoSuchHeaderParameter is returned by FancyHeader methods when the
	// operation being performed failed because the header exists, but a
	// sub-field of the header does not exist.
	ErrNoSuchHeaderParameter = errors.New("no such header parameter")

	// ErrManyHeaders is returned by FancyHeader methods when teh operation
	// being performed failed because the there are multiple fields with the
	// given name.
	ErrManyHeaders = errors.New("many headers found")

	// ErrWrongAddressType is returned by address setting methods that accept
	// either a string or an addr.AddressList when soething other than those
	// types is provided.
	ErrWrongAddressType = errors.New("incorrect address type during write")
)


// FancyHeaderInt defines a mutable email message header that provides a number of
// additional convenience methods.
type FancyHeaderInt interface {
	MutableHeader

	// Get retrieve the body value of the named header as a string. It will
	// return an error instead if no such header exists or if there are multiple
	// headers found with that name.
	Get(string) (string, error)

	// GetTime retrieves the body value of the named header as a time.Time. It
	// will return an error if no such header exists or if more than one such
	// header exists or if the value cannot be parsed into a time value.
	GetTime(string) (time.Time, error)

	// GetAddressList retrieves the body value of the named header as an
	// addr.AddressList. It returns an error if there is no such header for the
	// given name, if too many headers with the given name exist, or if there is
	// some problem parsing the address list.
	GetAddressList(string) (addr.AddressList, error)

	// GetParamValue retrieves the body value of the named header as a
	// param.Value. It returns an error if there are too few or too many headers
	// or if the header cannot be parsed as a ParamValue.
	GetParamValue(string) (*param.Value, error)

	// Set updates the header to the given value. This will replace all headers
	// with the given name (or append if no such field exists yet) with a single
	// header with the given name and value.
	Set(string, string)

	// SetTime updates the header to the given time value. This will replace all
	// headers with the given name (or append a new header if no such field yet
	// exists) with a single header with the given name and time value.
	SetTime(string, time.Time)

	// SetAddressList updates the header to the given addr.AddressList value.
	// This will replace all headers (or append a new field if none exists) with
	// the given name with a single header with the given name and address list.
	SetAddressList(string, addr.AddressList)

	// SetParamValue updates the header to the given param.Value. This will
	// replace all headers with the given name (or append a new field if no such
	// field exists in the header yet) with a single header with the given name
	// and parameterized value.
	SetParamValue(string, *param.Value)

	// GetContentType returns the MIME type set in the Content-type header.
	// Returns an error if there are too few or too many headers named Content-type
	// or if that header cannot be parsed as a Content-type header.
	GetContentType() (string, error)

	// SetContentType sets the MIME type for the message to the given value by
	// setting the Content-type header. May return an error if there's a problem
	// validating the MIME type.
	SetContentType(string) error

	// GetCharset returns the character set defined on the Content-type header.
	// It will return an error if the Content-type header is not present, if
	// the charset parameter is not set, or if there is a problem decoding the
	// header.
	GetCharset() (string, error)

	// SetCharset sets the charset parameter on the Content-type header. This
	// will fail with an error if the Content-type header is not already
	// present. The Content-type parameter must be set before calling
	// SetCharset().
	SetCharset(string) error

	// GetBoundary returns the boundary defined on the Content-type header.
	// It will return an error if the Content-type header is not present,
	// if the boundary parameter is not set, or if the Content-type header
	// cannot be decoded.
	GetBoundary() (string, error)

	// SetBoundary sets the boundary sets the boundary parameter on the
	// Content-type header. The operation fails with an error if the
	// Content-type header is not already present. You must set the Content-type
	// header before setting the boundary.
	SetBoundary(string) error

	// GetContentDisposition returns the primary value of the
	// Content-disposition header. It returns an error if it is not present or
	// cannot be decoded.
	GetContentDisposition() (string, error)

	// SetContentDisposition sets the primary value of the Content-disposition
	// header.
	SetContentDisposition(string) error

	// GetFilename returns the filename parameter of the Content-disposition
	// header. This method returns an error if the Content-disposition header is
	// not set, if the filename parameter is not set, or if there is an error
	// decoding the Content-disposition header.
	GetFilename() (string, error)

	// SetFilename sets the filename parameter of the Content-disposition
	// header. This will fail with an error if the Content-disposition field
	// does not yet exist as part of the header.
	SetFilename(string) error
}
