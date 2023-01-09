package field

// Raw is a email.Field implementation that presents the
// parsed Raw value. Objects of this type are immutable.
type Raw struct {
	field []byte // complete Raw field
	colon int    // the index of the colon
}

// String returns the Raw as a string.
func (f *Raw) String() string {
	return string(f.field)
}

// Bytes returns the Raw.
func (f *Raw) Bytes() []byte {
	return f.field
}

// Name returns the name part of the Raw. Please note that the value returned
// may be foleded.
func (f *Raw) Name() string {
	return string(f.field[:f.colon])
}

// Body returns the body part of the Raw as bytes. Please note that the value
// returned may be folded.
func (f *Raw) Body() string {
	off := 1
	if f.colon == len(f.field) {
		off = 0
	}
	return string(f.field[f.colon+off:])
}
