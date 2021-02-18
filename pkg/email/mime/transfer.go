package mime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/quotedprintable"
	"strings"
)

var (
	// TransferDecoders contains a map of transfer decoder objects for handling
	// all standard transfer encodings. This will contain values for 7bit, 8bit,
	// binary, base64, and quoted-printable transfer encodings.
	TransferDecoders map[string]TransferDecoder
)

// TransferDecoderFunc represents a function that writes a MIME message body
// according to a given Content-Transfer-Encoding.
type TransferDecoderFunc func([]byte) ([]byte, error)

// TransferDecoder represents the pair of decoders for use when
// encoding/decoding the transfer encoding used for the body of a MIME message.
type TransferDecoder struct {
	To   TransferDecoderFunc // write unencoded bytes out in encoded form
	From TransferDecoderFunc // write encoded bytes out in unencoded form
}

func init() {
	TransferDecoders = map[string]TransferDecoder{
		"":                 {DecodeAsIs, DecodeAsIs},
		"7bit":             {DecodeAsIs, DecodeAsIs},
		"8bit":             {DecodeAsIs, DecodeAsIs},
		"binary":           {DecodeAsIs, DecodeAsIs},
		"quoted-printable": {DecodeToQuotedPrintable, DecodeFromQuotedPrintable},
		"base64":           {DecodeToBase64, DecodeFromBase64},
	}
}

// SelectTransferDecoder returns a transfer decoder to use for the given
// Content-Transfer-Encoding header value. Falls back to a basic as-is encoder
// if no encoding matches.
func SelectTransferDecoder(cte string) (TransferDecoder, error) {
	if td, ok := TransferDecoders[strings.ToLower(cte)]; ok {
		return td, nil
	}
	return TransferDecoders[""], fmt.Errorf("incorrect Content-transfer-encoding header value %q", cte)
}

// DecodeAsIs is the identity encoding used whenever no decoding is required for
// a particular transfer encoding.
func DecodeAsIs(b []byte) ([]byte, error) {
	return b, nil
}

// DecodeToQuotedPrintable is used to output quoted-printed encoded bytes from
// unencoded bytes.
func DecodeToQuotedPrintable(b []byte) ([]byte, error) {
	var w bytes.Buffer
	qpw := quotedprintable.NewWriter(&w)
	_, err := qpw.Write(b)
	qpw.Close()
	return w.Bytes(), err
}

// DecodeFromQuotedPrintable is used to output unencoded bytes from
// quoted-printable encoded bytes.
func DecodeFromQuotedPrintable(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	qpr := quotedprintable.NewReader(r)
	return ioutil.ReadAll(qpr)
}

// DecodeToBase64 is used to output MIME BASE-64 encoded bytes from unencoded
// bytes.
func DecodeToBase64(b []byte) ([]byte, error) {
	var w bytes.Buffer
	b64w := base64.NewEncoder(base64.StdEncoding, &w)
	_, err := b64w.Write(b)
	b64w.Close()
	return w.Bytes(), err
}

// DecodeFromBase64 is used to output unencoded bytes from MIME BASE-64 encoded
// bytes.
func DecodeFromBase64(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	b64r := base64.NewDecoder(base64.StdEncoding, r)
	return ioutil.ReadAll(b64r)
}
