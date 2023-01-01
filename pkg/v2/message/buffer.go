package message

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/zostay/go-email/pkg/v2/header"
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

	// ModeOpaque indicates that the Buffer has been used as an io.Writer.
	ModeOpaque

	// ModeMultipart indicates that the Buffer has had the parts manipulated.
	ModeMultipart
)

var (
	// ErrPartsBuffer is returned by Write() if that method is called after
	// calling Add()
	ErrPartsBuffer = errors.New("message buffer is in parts mode")

	// ErrOpaqueBuffer is returned by Add() if that method is called after
	// calling Write()
	ErrOpaqueBuffer = errors.New("message buffer is in opaque mode")

	// ErrModeUnset is returned by Opaque() and Multipart() when they are called
	// before anything has been written to the current buffer.
	ErrModeUnset = errors.New("no message has been built")

	// ErrParsesAsNotMultipart is returned by Multipart() when the Buffer is in
	// ModeOpaque and the message is not at all a *Multipart message.
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
// modification method is called, it will return ModeOpaque if the Buffer has
// been used as an io.Writer or ModeMultipart if parts have been added to the
// Buffer.
func (b *Buffer) Mode() BufferMode {
	if b.parts != nil {
		return ModeMultipart
	} else if b.buf != nil {
		return ModeOpaque
	}
	return ModeUnset
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

func (b *Buffer) initParts() error {
	if b.buf != nil {
		return ErrOpaqueBuffer
	}
	if b.parts == nil {
		b.parts = make([]Part, 0, 10)
	}
	return nil
}

// Add will add one or more parts to the message.
func (b *Buffer) Add(msgs ...Part) error {
	if err := b.initParts(); err != nil {
		return err
	}
	for _, msg := range msgs {
		b.parts = append(b.parts, msg)
	}
	return nil
}

// Write implements io.Writer so you can write the message to this buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	if err := b.initBuffer(); err != nil {
		return 0, err
	}
	return b.buf.Write(p)
}

func (b *Buffer) prepareForMultipartOutput() {
	if _, err := b.GetContentType(); errors.Is(err, header.ErrNoSuchField) {
		b.SetContentType(DefaultMultipartContentType)
	}

	if _, err := b.GetBoundary(); errors.Is(err, header.ErrNoSuchFieldParameter) {
		_ = b.SetBoundary(GenerateBoundary())
	}
}

// Opaque will return an Opaque message based upon the content written to the
// Buffer. The behavior of this method depends on which mode the Buffer is in.
// This method fails with an error if there's a problem.
//
// If the BufferMode is ModeUnset, nil will be returned with the ErrModeUnset
// error.
//
// If the BufferMode is ModeOpaque, the Header and the bytes written to the
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
func (b *Buffer) Opaque() (*Opaque, error) {
	switch b.Mode() {
	case ModeUnset:
		return nil, ErrModeUnset
	case ModeOpaque:
		return &Opaque{
			Header: b.Header,
			Reader: b.buf,
		}, nil
	case ModeMultipart:
		b.prepareForMultipartOutput()
		boundary, _ := b.GetBoundary()

		buf := &bytes.Buffer{}
		if len(b.parts) > 0 {
			for _, part := range b.parts {
				_, _ = fmt.Fprintf(buf, "--%s%s", boundary, b.Break())
				_, _ = part.WriteTo(buf)
			}
			_, _ = fmt.Fprintf(buf, "--%s%s", boundary, b.Break())
		}

		return &Opaque{
			Header: b.Header,
			Reader: buf,
		}, nil
	}

	return nil, errors.New("unknown mode")
}

// Multipart will return a Multipart message based upon the content written to
// the Buffer. This method will fail with an error if there's a problem. The
// behavior of this method depends on which mode the Buffer is in when called.
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
// If the BufferMode is ModeUnset, nil will be returned with the ErrModeUnset
// error.
//
// If the BufferMode is ModeOpaque, the bytes that have been written to the
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
func (b *Buffer) Multipart() (*Multipart, error) {
	b.prepareForMultipartOutput()
	switch b.Mode() {
	case ModeUnset:
		return nil, ErrModeUnset
	case ModeOpaque:
		msg := &Opaque{b.Header, b.buf}
		gmsg, err := Parse(msg, WithoutRecursion())
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
	}
	return nil, errors.New("unknown mode")
}
