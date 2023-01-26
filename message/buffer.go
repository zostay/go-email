package message

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/zostay/go-email/v2/message/header"
)

const (
	// DefaultMultipartContentType is the Content-type to use with a multipart
	// message when no explicit Content-type header has been set.
	DefaultMultipartContentType = "multipart/mixed"
)

type BufferMode int

const (
	// ModeUnset indicates that the Buffer has not yet been modified.
	ModeUnset BufferMode = iota

	// ModeSingle indicates that the Buffer has been used as an io.Writer.
	ModeSingle

	// ModeMultipart indicates that the Buffer has had the parts manipulated.
	ModeMultipart
)

// ModeOpaque was the original name for ModeSingle.
//
// Deprecated: Use ModeSingle instead.
const ModeOpaque BufferMode = ModeSingle

var (
	// ErrPartsBuffer is returned by Write() if that method is called after
	// calling the Add() method.
	ErrPartsBuffer = errors.New("message buffer is in parts mode")

	// ErrOpaqueBuffer is returned by Add() if that method is called after
	// calling the Write() method.
	ErrOpaqueBuffer = errors.New("message buffer is in opaque mode")

	// ErrModeUnset is returned by Opaque() and Multipart() when they are called
	// before anything has been written to the current buffer.
	ErrModeUnset = errors.New("no message has been built")

	// ErrParsesAsNotMultipart is returned by Multipart() when the Buffer is in
	// ModeSingle and the message is not at all a *Multipart message.
	ErrParsesAsNotMultipart = errors.New("cannot parse non-multipart message as multipart")
)

// Buffer provides tools for constructing email messages. It can operate in
// either of two modes, depending on how you want to construct your message.
//
// * Opaque mode. When you use the Buffer as an io.Writer by calling the Write()
// method, you have chosen to treat the email message as a collection of bytes.
//
// * Multipart mode. When you use the Buffer to manipulate the parts of the
// message, such as calling the Add() method, you have chosen to treat the email
// message as a collection of sub-parts.
//
// You may not use a Buffer in both modes. If you call the Write() method first,
// then any subsequent call to the Add() method will return ErrOpaqueBuffer. If
// you call the Add() method first, then any call to the Write() method will
// result in ErrMultipartBuffer being returned.
//
// The BufferMode may be checked using the Mode() method.
//
// Whatever the mode is, you may call either Opaque() or Multipart() to get the
// constructed message at the end. However, there are some caveats, so be sure
// to about them in the documentation of those methods.
type Buffer struct {
	header.Header
	parts []Part
	buf   *bytes.Buffer
}

// Mode returns a constant that indicates what mode the Buffer is in. Until a
// modification method is called, this will return ModeUnset. Once a
// modification method is called, it will return ModeSingle if the Buffer has
// been used as an io.Writer or ModeMultipart if parts have been added to the
// Buffer.
func (b *Buffer) Mode() BufferMode {
	if b.parts != nil {
		return ModeMultipart
	} else if b.buf != nil {
		return ModeSingle
	}
	return ModeUnset
}

// SetMultipart sets the Mode of the buffer to ModeMultipart. This is useful
// during message transformation or when you want to pre-allocate the capacity
// of the internal slice used to hold parts. When calling this method, you need
// to pass the expected capacity of the multipart message. This will panic if
// the mode is already ModeSingle.
func (b *Buffer) SetMultipart(capacity int) {
	err := b.initParts(capacity)
	if err != nil {
		panic(err)
	}
}

// SetSingle sets the Mode of the buffer to ModeSingle. This is useful during
// message transformation, especially if the message content is to be empty.
// This will panic if the mode is already ModeMultipart.
func (b *Buffer) SetSingle() {
	err := b.initBuffer()
	if err != nil {
		panic(err)
	}
}

func (b *Buffer) initBuffer() error {
	if b.parts != nil {
		return ErrPartsBuffer
	}
	if b.buf == nil {
		b.buf = &bytes.Buffer{}
	}
	return nil
}

func (b *Buffer) initParts(capacity int) error {
	if capacity == 0 {
		capacity = 10
	}
	if b.buf != nil {
		return ErrOpaqueBuffer
	}
	if b.parts == nil {
		b.parts = make([]Part, 0, capacity)
	}
	return nil
}

// Add will add one or more parts to the message. It will panic if you attempt
// to call this function after already calling Write() or using this object as
// an io.Writer.
func (b *Buffer) Add(msgs ...Part) {
	if err := b.initParts(0); err != nil {
		panic(err)
	}
	b.parts = append(b.parts, msgs...)
}

// Write implements io.Writer so you can write the message to this buffer. This
// will panic if you attempt to call this method or use this object as an
// io.Writer after calling Add.
func (b *Buffer) Write(p []byte) (int, error) {
	if err := b.initBuffer(); err != nil {
		panic(err)
	}
	return b.buf.Write(p)
}

func (b *Buffer) prepareForMultipartOutput() {
	if _, err := b.GetMediaType(); errors.Is(err, header.ErrNoSuchField) {
		b.SetMediaType(DefaultMultipartContentType)
	}

	if _, err := b.GetBoundary(); errors.Is(err, header.ErrNoSuchFieldParameter) {
		_ = b.SetBoundary(GenerateBoundary())
	}
}

// Opaque will return an Opaque message based upon the content written to the
// Buffer. The behavior of this method depends on which mode the Buffer is in.
//
// This method will panic if the BufferMode is ModeUnset.
//
// If the BufferMode is ModeSingle, the Header and the bytes written to the
// internal buffer will be returned in the *Opaque. Opaque will return a MIME
//
// If the BufferMode is ModeMultipart, the parts will be serialized into a byte
// buffer and that will be attached with the Header to the returned *Opaque
// object.
//
// In this last case, you should set the Content-Type header yourself to one of
// the multipart/* types (e.g., multipart/alternative if you are providing text
// and HTML forms of the same message or multipart/mixed if you are providing
// attachments).
//
// If you do not provide that header yourself, this method will set it to
// DefaultMultipartContentType, which may not be what you want.
//
// It will also check to see if the Content-type boundary is set and set it to
// something random using mime.GenerateBound() automatically. This boundary
// will then be used when joining the parts together during serialization.
//
// After this method is called, the Buffer should be disposed of and no longer
// used.
func (b *Buffer) Opaque() *Opaque {
	switch b.Mode() {
	case ModeSingle:
		return &Opaque{
			Header: b.Header,
			Reader: b.buf,
		}
	case ModeMultipart:
		b.prepareForMultipartOutput()
		boundary, _ := b.GetBoundary()

		buf := &bytes.Buffer{}
		if len(b.parts) > 0 {
			for _, part := range b.parts {
				_, _ = fmt.Fprintf(buf, "--%s%s", boundary, b.Break())
				_, _ = part.WriteTo(buf)
				_, _ = fmt.Fprint(buf, b.Break())
			}
			_, _ = fmt.Fprintf(buf, "--%s--", boundary)
		}

		return &Opaque{
			Header: b.Header,
			Reader: buf,
		}
	case ModeUnset:
		panic(ErrModeUnset)
	}
	panic("unknown error")
}

// OpaqueAlreadyEncoded works just like Opaque(), but marks the object as
// already having the Content-transfer-encoding applied. Use this when you write
// a message in encoded form.
//
// NOTE: This does not perform any encoding! If you want the output to be
// automatically encoded, you actually want to call Opaque() and then WriteTo()
// on the returned object will perform encoding. This method is for indicating
// that you have already performed the required encoding.
//
// After this method is called, the Buffer should be disposed of and no longer
// used.
func (b *Buffer) OpaqueAlreadyEncoded() *Opaque {
	msg := b.Opaque()
	if msg != nil {
		msg.encoded = true
	}
	return msg
}

// Multipart will return a Multipart message based upon the content written to
// the Buffer. This method will fail with an error if there's a problem. The
// behavior of this method depends on which mode the Buffer is in when called.
//
// If you are just generating a message to output to a file or network socket
// (e.g., an SMTP connection), you probably do not want to call this method.
// Calling Opaque is generally preferable in that case. However, this is
// provided in cases when you really do want a Multipart for some reason.
//
// Whenever you plan on calling this method, you should set the Content-Type
// header yourself to one of the multipart/* types (e.g., multipart/alternative
// if you are providing text and HTML forms of the same message or
// multipart/mixed if you are providing attachments).
//
// If you do not provide that header yourself, this method will set it to
// DefaultMultipartContentType, which may not be what you want.
//
// It will also check to see if the Content-type boundary is set and set it to
// something random using mime.GenerateBound() automatically. This boundary
// will then be used when joining the parts together during serialization.
//
// If the BufferMode is ModeUnset, this method will panic.
//
// If the BufferMode is ModeSingle, the bytes that have been written to the
// buffer must be parsed in order to generate the returned *Multipart. In that
// case, the function will create an *Opaque and use the Parse() function to try
// to generate a *Multipart. The Parse() function will be called with the
// WithoutRecursion() option. Errors from Parse() will be passed back from this
// method. If Parse() returns nil or an *Opaque, this method will return nil.
// Otherwise, the *Multipart from Parse() will be returned. If an *Opaque is
// returned from Parse() without an error, this method will return nil with
// ErrParsesAsNotMultipart.
//
// If the BufferMode is ModeMultipart, the Header and collected parts will be
// returned in the returned *Multipart.
//
// After this method is called, the Buffer should be disposed of and no longer
// used.
func (b *Buffer) Multipart() (*Multipart, error) {
	b.prepareForMultipartOutput()
	switch b.Mode() {
	case ModeSingle:
		msg := &Opaque{b.Header, b.buf, false}
		pr := defaultParser.clone()
		WithoutRecursion()(pr)
		gmsg, err := pr.parse(msg, 0)
		switch vmsg := gmsg.(type) {
		case *Opaque:
			if err != nil {
				return nil, err
			}
			return nil, ErrParsesAsNotMultipart
		case *Multipart:
			return vmsg, err
		}
		return nil, errors.New("generic message came back as something other than Opaque or Multipart")
	case ModeMultipart:
		return &Multipart{
			Header: b.Header,
			prefix: []byte{},
			suffix: []byte{},
			parts:  b.parts,
		}, nil
	case ModeUnset:
		panic(ErrModeUnset)
	}
	panic("unknown error")
}
