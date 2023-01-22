package message

import (
	"io"
	"os"
	"path/filepath"

	"github.com/zostay/go-email/v2/message/header"
	"github.com/zostay/go-email/v2/message/transfer"
)

// Opaque is the base-level email message interface. It is simply a header
// and a message body, very similar to the net/mail message implementation.
type Opaque struct {
	// Header will contain the header of the message. A top-level message must
	// have several headers to be correct. A message part should have one or
	// more headers as well.
	header.Header

	// Reader will contain the body content of the message. If the content is
	// zero bytes long, then Reader should be set to nil.
	io.Reader

	// encoded tracks whether the body has had the content-transfer-encoding
	// still encoded or not...
	//
	// - parsing leaves encoding in place by default (unless
	// DecodeTransferEncoding() option is specified)
	//
	// - creating an opaque with a buffer will leave this false unless the
	// object is constructed using OpaqueAlreadyEncoded
	encoded bool
}

// WriteTo writes the Opaque header and body to the destination
// io.Writer.
//
// If the bytes head in io.Reader have had the Content-transfer-encoding decoded
// (e.g., the message was parsed with the DecodeTransferEncoding() option or was
// created via a Buffer), then this will encode the data as it is being written.
//
// This can only be safely called once as it will consume the io.Reader.
func (m *Opaque) WriteTo(w io.Writer) (int64, error) {
	var tw io.WriteCloser
	if !m.encoded {
		tw = transfer.ApplyTransferEncoding(&m.Header, w)
		defer func() { _ = tw.Close() }()
	}

	total, err := m.Header.WriteTo(w)
	if err != nil {
		return total, err
	}

	if tw != nil {
		w = tw
	}

	if m.Reader != nil {
		bn, err := io.Copy(w, m.Reader)
		total += bn
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// IsMultipart always returns false.
func (m *Opaque) IsMultipart() bool {
	return false
}

// IsEncoded returns true if the Content-transfer-encoding has not been decoded
// for the bytes returned by the associated io.Reader. It will return false if
// that decoding has been performed.
//
// Be aware that a false value here does not mean any actually changes to the
// bytes have been made. If the Content-type of this message is a "multipart/*"
// type, then any Content-transfer-encoding is ignored. If the
// Content-transfer-encoding is set to something like "8bit", the transfer
// encoding returns the bytes as-is and no transformation of the data is
// performed anyway.
//
// However, if this returns true, then reading the data from io.Reader will
// return exactly the same bytes as would be written via WriteTo().
func (m *Opaque) IsEncoded() bool {
	return m.encoded
}

// GetHeader returns the header for the message.
func (m *Opaque) GetHeader() *header.Header {
	return &m.Header
}

// GetReader returns the reader containing the body of the message.
//
// If IsEncoded() returns false, the data returned by reading this io.Reader
// may differ from the data that would be written via WriteTo(). This is
// because the data here will have been decoded, but WriteTo() will encode the
// data anew as it writes.
func (m *Opaque) GetReader() io.Reader {
	return m.Reader
}

// GetParts always returns nil and ErrNotMultipart.
func (m *Opaque) GetParts() []Part {
	return nil
}

// AttachmentFile is a constructor that will create an Opaque from the given
// filename and MIME type. This will read the given file path from the disk,
// make that filename the name of an attachment, and return it. It will return
// an error if there's a problem reading the file from the disk.
//
// The last argument is optional and is the transfer encoding to use. Use
// transfer.None if you do not want to set a transfer encoding.
func AttachmentFile(fn, mt, te string) (*Opaque, error) {
	m := &Opaque{}
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	m.Reader = f
	m.SetMediaType(mt)
	_ = m.SetFilename(filepath.Base(fn))

	m.SetPresentation("attachment")
	_ = m.SetFilename(filepath.Base(fn))

	if te != transfer.None {
		m.SetTransferEncoding(te)
	}

	return m, nil
}
