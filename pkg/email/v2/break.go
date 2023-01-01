package email

// Break represents the linebreak to use when working with an email.
type Break string

// Constants for use when selecting a line break to use with a new header. If
// you don't know what to pick, choose CRLF.
const (
	Meh  Break = ""         // Sometimes it doesn't matter
	CRLF Break = "\x0d\x0a" // \r\n - Network linebreak
	LF   Break = "\x0a"     // \n - Unix/Linux/BSD linebreak
	CR   Break = "\x0d"     // \r - Commodores/old Macs linebreak
	LFCR Break = "\x0a\x0d" // \n\r - for weirdos
)

// String returns the break as a string.
func (b Break) String() string {
	return string(b)
}

// Bytes returns the break as a slice of bytes.
func (b Break) Bytes() []byte {
	return []byte(b)
}

// WithBreak is an object that has associated line breaks.
type WithBreak interface {
	// Break returns the line break to use with this object.
	Break() Break
}

// WithMutableBreak is an object whose associated line breaks can be changed.
// Changing the line break must immediately change the entire format of the
// object to use those line breaks.
type WithMutableBreak interface {
	WithBreak

	// SetBreak sets the line break to the newly given value.
	SetBreak(Break)
}
