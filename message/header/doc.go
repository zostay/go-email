// Package header provides low-level and high-level tooling for dealing with
// email message headers. If you need low-level access, you want to deal with
// methods that work with field.Field objects. However, it is generally expected
// that devs will prefer the high-level methods which will try to keep your
// reading and manipulation of the header safe and strictly correct on output.
//
// The provided Parse() method will parse up headers in a flexible way that is
// built on top of field.Parse() to preserve message headers as-is for output.
package header
