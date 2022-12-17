package simple

// HeaderField implements a mutable field that tracks each header up two
// two times. It combines a HeaderFieldFolded with a HeaderFieldOriginal. Each
// may be modified independently by calling their methods directly. This must be
// done cautiously as it allows for two completely different values to be
// store, which is the intention. Therefore, you can store one value in a
// natural form for local work, but in a fully encoded form for writing out to
// the rendered email message.
type HeaderField struct {
	HeaderFieldBase

	// Original may be nil. If not nil, then String() and Bytes() for the
	// HeaderFieldEncoded object will defer to here, but the Name(), SetName(),
	// Body(), and SetBody() methods will not.
	Original *HeaderFieldOriginal
}

// NewHeaderField will construct a HeaderField. The original is optional. If the
// original is provided and there is an error in parsing it, this function will
// return an error.
func NewHeaderField(
	name,
	body string,
	original []byte,
) (*HeaderField, error) {
	f := &HeaderField{
		HeaderFieldBase: HeaderFieldBase{name, body},
		Original:        nil,
	}

	if original != nil {
		var err error
		f.Original, err = ParseOriginal(original)
		if err != nil {
			return f, err
		}
	}

	return f, nil
}

// String returns the original String() if Original is not nil. It returns the
// HeaderFieldFolded.String() otherwise.
func (f *HeaderField) String() string {
	if f.Original != nil {
		return f.Original.String()
	}
	return f.HeaderFieldBase.String()
}

// Bytes returns the original Bytes() if Original is not nil. It returns the
// HeaderFieldFolded.Bytes() otherwise.
func (f *HeaderField) Bytes() []byte {
	if f.Original != nil {
		return f.Original.Bytes()
	}
	return f.HeaderFieldBase.Bytes()
}
