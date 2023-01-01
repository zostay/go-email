package simple

import (
	"github.com/zostay/go-email/pkg/email/v2"
)

// HeaderFieldFolded implements a mutable header field with folding support.
// This is mostly identical to HeaderFieldBase, which it extends, except that
// when String() or Bytes() are called, the header value returned will be folded
// according to the settings in the ValueFolder set at construction time.
type HeaderFieldFolded struct {
	HeaderFieldBase

	vf *ValueFolder
}

// NewHeaderFieldFolded returns a new header field object for the given name and
// body.
func NewHeaderFieldFolded(
	lbr email.Break,
	vf *ValueFolder,
	name,
	body string,
) *HeaderFieldFolded {
	return &HeaderFieldFolded{
		HeaderFieldBase{lbr, name, body},
		vf,
	}
}

// refold refolds the value after the name or body are set.
func (f *HeaderFieldFolded) fold() []byte {
	// field equals "$name: $body$lbr"
	field := make([]byte, len(f.name)+len(f.body)+len(f.lbr)+2)
	copy(field, f.name)
	field[len(f.name)] = ':'
	field[len(f.name)+1] = ' '
	copy(field[len(f.name)+2:], f.body)
	copy(field[len(f.name)+len(f.body)+2:], f.lbr)

	return f.vf.Fold(field, f.lbr)
}

// String returns the folded version of the header field.
func (f *HeaderFieldFolded) String() string {
	return string(f.fold())
}

// Bytes returns the folded version of the header field.
func (f *HeaderFieldFolded) Bytes() []byte {
	return f.fold()
}

var _ email.MutableHeaderField = &HeaderFieldFolded{}
var _ email.WithMutableBreak = &HeaderFieldFolded{}
var _ email.Outputter = &HeaderFieldFolded{}
