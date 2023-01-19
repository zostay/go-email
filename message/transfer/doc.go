// Package transfer contains utilities related to encoding and decoding transfer
// encodings, which interpret the Content-transfer-encoding header to apply
// certain 8bit to 7bit encodings. If a Content-transfer-encoding is present,
// only the values of quoted-printable and base64 will actually result in
// changes to the document being encoded or decoded. Other settings such as
// binary, 7bit, or 8bit will result in the bytes being left as-is.
//
// For the sake of this module, the term "decoded" means that the content has
// been transformed from the named Content-transfer-encoding to the charset
// encoded form. Meanwhile, "encoded" means that the content has been
// transformed from the charset encoding to the named Content-transfer-encoding.
package transfer
