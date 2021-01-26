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
	TransferDecoders map[string]TransferDecoder
)

type TransferDecoderFunc func([]byte) ([]byte, error)
type TransferDecoder struct {
	To   TransferDecoderFunc
	From TransferDecoderFunc
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

func SelectTransferDecoder(cte string) (TransferDecoder, error) {
	if td, ok := TransferDecoders[strings.ToLower(cte)]; ok {
		return td, nil
	}
	return TransferDecoders[""], fmt.Errorf("incorrect Content-transfer-encoding header value %q", cte)
}

func DecodeAsIs(b []byte) ([]byte, error) {
	return b, nil
}

func DecodeToQuotedPrintable(b []byte) ([]byte, error) {
	var w bytes.Buffer
	qpw := quotedprintable.NewWriter(&w)
	_, err := qpw.Write(b)
	return w.Bytes(), err
}

func DecodeFromQuotedPrintable(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	qpr := quotedprintable.NewReader(r)
	return ioutil.ReadAll(qpr)
}

func DecodeToBase64(b []byte) ([]byte, error) {
	var w bytes.Buffer
	b64w := base64.NewEncoder(base64.StdEncoding, &w)
	_, err := b64w.Write(b)
	b64w.Close()
	return w.Bytes(), err
}

func DecodeFromBase64(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	b64r := base64.NewDecoder(base64.StdEncoding, r)
	return ioutil.ReadAll(b64r)
}
