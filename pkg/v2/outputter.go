package email

// Outputter objects are those which provide a String() and Bytes() method.
// String and Bytes should return identical byte strings as different types.
type Outputter interface {
	// String is identical to the fmt.Stringer version.
	String() string

	// Bytes is the bytes version of fmt.Stringer.
	Bytes() []byte
}
