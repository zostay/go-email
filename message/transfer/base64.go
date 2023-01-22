package transfer

import (
	"encoding/base64"
	"io"
)

const defaultBase64LineLength = 76

var defaultBase64LineBreak = []byte{'\n'}

type newlineWriter struct {
	every int
	acc   int
	lbr   []byte
	w     io.Writer
}

func (nw *newlineWriter) Write(b []byte) (int, error) {
	ix, n := 0, 0
	for len(b[ix:])+nw.acc > nw.every {
		n := 0
		ln, err := nw.w.Write(b[ix : ix+(nw.every-nw.acc)])
		n += ln
		if err != nil {
			return n, err
		}

		_, err = nw.w.Write(nw.lbr)
		if err != nil {
			return n, err
		}

		ix += nw.every - nw.acc
		nw.acc = 0
	}

	ln, err := nw.w.Write(b[ix:])
	n += ln
	if err != nil {
		return n, err
	}

	nw.acc = len(b[ix:]) % nw.every

	return n, nil
}

func (nw *newlineWriter) Close() error {
	_, err := nw.w.Write(nw.lbr)
	if wc, isCloser := nw.w.(io.Closer); isCloser {
		err := wc.Close()
		if err != nil {
			return err
		}
	}
	return err
}

// NewBase64Encoder will translate all bytes written to the returned
// io.WriteCloser into base64 encoding and write those to the give io.Writer.
func NewBase64Encoder(w io.Writer) io.WriteCloser {
	return &writer{
		base64.NewEncoder(base64.StdEncoding, &newlineWriter{
			every: defaultBase64LineLength,
			lbr:   defaultBase64LineBreak,
			w:     w,
		}), true,
	}
}

// NewBase64Decoder will translate all bytes read from the given io.Reader as
// base64 and return the binary data to the returned io.Reader.
func NewBase64Decoder(r io.Reader) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, r)
}
