package message

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/zostay/go-email/v2/internal/scanner"
	"github.com/zostay/go-email/v2/message/header"
	"github.com/zostay/go-email/v2/message/transfer"
)

// Constants related to Parse() options.
const (
	// DefaultMaxMultipartDepth is the default depth the parser will recurse
	// into a message.
	DefaultMaxMultipartDepth = 10

	// DefaultChunkSize the default size of chunks to read from the input while
	// splitting the message into header and body. Defaults to 16K, though this
	// could change at any time.
	DefaultChunkSize = 16_384

	// DefaultMaxHeaderLength is the default maximum byte length to scan before
	// giving up on finding the end of the header.
	DefaultMaxHeaderLength = bufio.MaxScanTokenSize

	// DefaultMaxPartLength is the default maximum byte length to scan before
	// given up on scanning a message part at any given level.
	DefaultMaxPartLength = bufio.MaxScanTokenSize
)

// Errors that occur during parsing.
var (
	// ErrNoBoundary is returned by Parse when the boundary parameter is not set
	// on the Content-type field of the message header.
	ErrNoBoundary = errors.New("the boundary parameter is missing from Content-type")

	// ErrLargeHeader is returned by Parse when the header is longer than the
	// configured WithMaxHeaderLength option (or the default,
	// DefaultMaxHeaderLength).
	ErrLargeHeader = errors.New("the header exceeds the maximum parse length")

	// ErrLargePart is returned by Parse when  apart is longer than the configured
	// WithMaxPartLength option (or the default, DefaultMaxPartLength).
	ErrLargePart = errors.New("a message part exceeds the maximum parse length")
)

var splits = [][]byte{
	[]byte("\x0d\x0a\x0d\x0a"), // \r\n\r\n
	[]byte("\x0a\x0d\x0a\x0d"), // \n\r\n\r, extremely unlikely, possibly never
	[]byte("\x0a\x0a"),         // \n\n
	[]byte("\x0d\x0d"),         // \r\r
}

type parser struct {
	maxHeaderLen int
	maxPartLen   int
	maxDepth     int
	chunkSize    int
	decode       bool
}

func (pr *parser) clone() *parser {
	p := *pr
	return &p
}

var defaultParser = &parser{
	maxHeaderLen: DefaultMaxHeaderLength,
	maxPartLen:   DefaultMaxPartLength,
	maxDepth:     DefaultMaxMultipartDepth,
	chunkSize:    DefaultChunkSize,
	decode:       false,
}

// ParseOption refers to options that may be passed to the Parse function to
// modify how the parser works.
type ParseOption func(pr *parser)

// WithMaxHeaderLength is a ParseOption that sets the maximum size the buffer is
// allowed to reach before parsing exits with an ErrLargeHeader error. During
// parsing, the io.Reader will be read from a chunk at a time until the end of
// the header is found. This setting prevents bad input from resulting in an out
// of memory error. Setting this to a value less than or equal to 0 will result
// in there being no maximum length. The default value is
// DefaultMaxHeaderLength.
func WithMaxHeaderLength(n int) ParseOption {
	return func(pr *parser) { pr.maxHeaderLen = n }
}

// WithMaxPartLength is a ParseOption that sets the maximum size the buffer is
// allowed to reach while scanning for message parts at any level. The parts are
// parsed out at each level of depth separately, so this must be large enough to
// accommodate the largest part at the top level being parsed. If the part gets
// too large, Parse will fail with an ErrLargePart error. There is, at this
// time, no way to disable this limit.
func WithMaxPartLength(n int) ParseOption {
	return func(pr *parser) { pr.maxPartLen = n }
}

// DecodeTransferEncoding is a ParseOption that enables the decoding of
// Content-transfer-encoding. By default, Content-transfer-encoding will not be
// decoded, which allows for safer round-tripping of messages. However, if you
// want to display or process the message body, you will want to enable this.
func DecodeTransferEncoding() ParseOption {
	return func(pr *parser) { pr.decode = true }
}

// WithChunkSize is a ParseOption that controls how many bytes to read at a time
// while parsing an email message. The default chunk size is DefaultChunkSize.
func WithChunkSize(chunkSize int) ParseOption {
	return func(pr *parser) { pr.chunkSize = chunkSize }
}

// WithMaxDepth is a ParseOption that controls how deep the parser will go in
// recursively parsing a multipart message. This is set to DefaultMaxDepth by
// default.
func WithMaxDepth(maxDepth int) ParseOption {
	return func(pr *parser) { pr.maxDepth = maxDepth }
}

// WithoutMultipart is a ParseOption that will not allow parsing of any
// multipart messages. The message returned from Parse() will always be *Opaque.
//
// You should use this option if all you are interested in is the top-level
// headers. For large email messages, use of this option can grant extreme
// improvements to memory performance. This is because this option prevents any
// multipart processing, which means the header will be read, parsed, and stored
// in memory. However, only a single chunk of the body will have been read. The
// rest of the input io.Reader is left unread.
func WithoutMultipart() ParseOption {
	return func(pr *parser) { pr.maxDepth = 0 }
}

// WithoutRecursion is a ParseOption that will only allow a single level of
// multipart parsing.
func WithoutRecursion() ParseOption {
	return func(pr *parser) { pr.maxDepth = 1 }
}

// WithUnlimitedRecursion is a ParseOption that will allow the parser to parse
// sub-parts of any depth.
func WithUnlimitedRecursion() ParseOption {
	return func(pr *parser) { pr.maxDepth = -1 }
}

// searchForSplit looks for a header/body split. Returns -1, nil if none is
// found. If the header/body split is found, it returns the location of the
// split (including the split newlines) and the line break to use with the
// header as a slice of bytes.
func searchForSplit(buf []byte, subpart bool) (pos int, crlf []byte) {
	if subpart {
		// if the header is empty, the first char might be a line break, indicating
		// an empty header, right? It happens.
		for _, s := range splits {
			if testPos := bytes.Index(buf, s[0:len(s)/2]); testPos == 0 {
				pos = testPos + len(s)/2
				crlf = s[0 : len(s)/2]
				return
			}
		}
	}

	// Find the split between header/body
	pos = -1
	for _, s := range splits {
		if testPos := bytes.Index(buf, s); testPos > -1 {
			pos = testPos + len(s)
			crlf = s[0 : len(s)/2]
			return
		}
	}
	return
}

// splitHeadFromBody will pull the header off the front of the given input
// splitHeadFromBody will detect the index of the split between the message
// header and the message body as well as the line break the email is using. It
// returns both.
func (pr *parser) splitHeadFromBody(r io.Reader, subpart bool) ([]byte, []byte, io.Reader, error) {
	p := make([]byte, pr.chunkSize)
	buf := &bytes.Buffer{}
	searched := 0
	for {
		// read in some bytes
		n, err := r.Read(p)

		// check to see if the header is too long
		if pr.maxHeaderLen > 0 && n+buf.Len() > pr.maxHeaderLen {
			return nil, nil, nil, ErrLargeHeader
		}

		isEof := false
		if errors.Is(err, io.EOF) {
			isEof = true
		} else if err != nil {
			return nil, nil, nil, err
		}

		// add that to our buffer
		_, err = buf.Write(p[:n])
		if err != nil {
			return nil, nil, nil, err
		}

		// check the tail of the buffer for end of header
		pos, crlf := searchForSplit(buf.Bytes()[searched:], subpart)
		if pos >= 0 {
			pos += searched
			// we found the split, header is bytes up to the split
			hdr := make([]byte, pos)
			for hdrRead, n := 0, 0; hdrRead < pos; hdrRead += n {
				n, err = buf.Read(hdr[hdrRead:])
				if err != nil {
					return nil, nil, nil, err
				}
			}

			// the rest is the body
			var body io.Reader
			if _, isBytesReader := r.(*bytes.Reader); isBytesReader {
				// We treat bytes.Reader special because this is what we use
				// internally to parse each part of a multipart message. This
				// will pull the data out of the bytes.Reader and attach it to
				// the end of the byte.Buffer we've been building.
				_, err = buf.ReadFrom(r)
				if err != nil {
					return nil, nil, nil, err
				}
				// Without this, the header bytes will still be in the buffer.
				// This will cause those bytes to be discarded, which will
				// improve memory performance somewhat.
				body = bytes.NewReader(buf.Bytes())
			} else {
				// If it's something else, we will leave the remainder unread
				// as we must be reading an original input io.Reader. By not
				// consuming it, we can improve the memory performance of
				// Opaque message.
				body = &remainder{buf.Bytes(), r}
			}
			return hdr, crlf, body, nil
		}

		// No split found and EOF? Let's break out and then we'll process as if
		// the entire message is just header.
		if isEof {
			break
		}

		// The last 3 bytes might be the prefix to the split point
		searched = buf.Len() - 3
		if searched < 0 {
			searched = 0
		}
	}

	// If we're here, we were unable to find a header/body split. We will just
	// assume the message is all header, no body. Let's see if we can find what
	// to use as a break.
	for _, s := range splits {
		crlf := s[0 : len(s)/2]
		if bytes.Contains(buf.Bytes(), crlf) {
			return buf.Bytes(), crlf, nil, nil
		}
	}

	// Or the ultimate fallback is...
	return buf.Bytes(), []byte("\x0d"), nil, nil
}

// parseOpaque turns a reader into an Opaque.
func (pr *parser) parseToOpaque(r io.Reader, subpart bool) (*Opaque, error) {
	hdr, crlf, body, err := pr.splitHeadFromBody(r, subpart)
	if err != nil {
		return nil, err
	}

	head, err := header.Parse(hdr, header.Break(crlf))
	if err != nil {
		return nil, err
	}

	if pr.decode {
		body = transfer.ApplyTransferDecoding(head, body)
	}

	return &Opaque{*head, body, !pr.decode}, nil
}

// Parse will consume input from the given reader and return a Generic message
// containing the parsed content. Parse will proceed in two or three phases.
//
// During the first phase, the given io.Reader will be read in chunks at a time,
// as defined by the WithChunkSize() option (or by the default,
// DefaultChunkSize). Each chunk will be checked for a double line break of some
// kind (e.g., "\r\n\r\n" or "\n\n" are the most common). Once found, that line
// break is used to determine what line break the message will use for breaking
// up the header into fields. The fields will be parsed from the accumulated
// header chunks using the bytes read in so far preceding the header break.
//
// The last part of the final chunk read and the remainder of the io.Reader will
// then make up the body content of an *Opaque message.
//
// If accumulated header chunks total larger than the WithMaxHeaderLength()
// option (or the default, DefaultMaxHeaderLength) while searching for the
// double line break, the Parse will fail with an error and return
// ErrLargeHeader. If this happens, the io.Reader may be in a partial read
// state.
//
// If the first phase completes successfully, the second phase will begin.
// During the second phase, the *Opaque message created during the first phase
// may be transformed into a *Multipart, if the message seems to be a multipart
// message. The way this will proceed is determined by the WithMaxDepth()
// related options and also the WithMaxPartLength() option.
//
// If the Content-type of the message is a multipart/* MIME type and the
// WithMaxDepth() option (or the default, DefaultMaxMultipartDepth) is less than
// or greater than 0, the body will be scanned to break it into parts according
// to the boundary parameter set on the Content-type. The parts must be smaller
// than the setting in WithMaxPartLength() option (or the default,
// DefaultMaxPartLength). If not, the parse will fail with ErrLargePart.
//
// These newly broken up parts will each go through the two phase parsing
// process themselves. This continues until either the deepest multipart sub-part is
// parsed or the maximum depth is reached.
//
// If the DecodeTransferEncoding() option is passed, a third phase of parsing
// will also be performed. The parts of the message that do not have sub-parts
// and have a Content-transfer-encoding header set, will be decoded.
//
// This third phase is not the default behavior because one of those goals of
// this library is to try and preserve the original bytes as is. However, decoding
// the transfer encoding and then re-encoding it again is very likely to modify
// the original message. The modification will be trivial, but it won't preserve
// the original message for round-trip modification with minimal changes.
//
// If the transfer encoding is decoded in this third phase, rendering the
// message with WriteTo() will perform new encoding of the data and write
// freshly encoded data to the destination writer. However, if you read the data
// using the io.Reader, you will receive un-encoded bytes.
//
// Errors at any point in the process may lead to a completely failed parse,
// especially those involving ErrLargeHeader or ErrLargePart. However, whenever
// possible, the partially parsed message object will be returned.
//
// The original io.Reader provided may or may not be completely read upon
// return. This is true whether an error has occurred or not. If you either read
// all the message body contents of all sub-parts or use the WriteTo() method on
// the returned message object, the io.Reader will be completely consumed.
func Parse(r io.Reader, opts ...ParseOption) (Generic, error) {
	pr := defaultParser.clone()
	for _, opt := range opts {
		opt(pr)
	}

	msg, err := pr.parseToOpaque(r, false)
	if err != nil {
		return msg, err
	}

	return pr.parse(msg, 0)
}

// parse implements the Parse methods.
func (pr *parser) parse(msg *Opaque, depth int) (Generic, error) {
	// we're too deep: stop here and just return the original
	if pr.maxDepth >= 0 && depth >= pr.maxDepth {
		return msg, nil
	}

	// lookup the Content-type header
	pv, err := msg.GetParamValue(header.ContentType)
	if err != nil {
		return msg, nil
	}

	// if this is not a multipart, don't parse it
	if pv.Type() != "multipart" && pv.Type() != "message" {
		return msg, nil
	}

	// if the boundary is missing, don't parse it and return an error
	if pv.Boundary() == "" {
		return msg, ErrNoBoundary
	}

	// The initial boundaries are like --boundary and final boundary is like
	// --boundary-- and these must be on their own line. This means that every
	// boundary but the very first must begin with a newline, but the first
	// might not have one. We search without a newline until the first boundary
	// is found, then prefix it with the newline for subsequent searches. When
	// dealing with the first, we will also look for the newline before the
	// found boundary to ensure the prefix is captured correctly.
	//
	// Newline handling is nuanced in order to preserve the original message for
	// round-tripping. The newline before the start boundary (if any) belongs to
	// the prefix. The newline after the final boundary (if any) belongs to the
	// suffix. The newlines before and after the middle boundaries belong to the
	// boundary and are not included with the part (because they have to be
	// there or message parsing does not work).
	sb := []byte(fmt.Sprintf("--%s%s", pv.Boundary(), msg.Break()))
	mb := []byte(fmt.Sprintf("%s--%s%s", msg.Break(), pv.Boundary(), msg.Break()))
	eb := []byte(fmt.Sprintf("%s--%s--%s", msg.Break(), pv.Boundary(), msg.Break()))
	fb := []byte(fmt.Sprintf("%s--%s--", msg.Break(), pv.Boundary()))

	const (
		modeStart = iota
		modeMiddle
		modeEnd
	)

	// This scanner split function splits on any email message boundary. It
	// returns the parts as tokens, but the prefix and suffix, it captures
	// itself in the prefix/suffix vars.
	sc := bufio.NewScanner(msg.Reader)
	sc.Buffer(make([]byte, pr.chunkSize), pr.maxPartLen)
	var prefix, suffix []byte
	mode := modeStart
	awaitingPrefix := true
	sc.Split(
		scanner.MakeSplitFuncExitByAdvance( // bufio.SplitFunc sucks
			func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				switch mode {
				case modeStart:
					// looking for an empty prefix
					if atEOF || len(data) >= len(sb) {
						if bytes.Equal(data[:len(sb)], sb) {
							// initial string is the boundary, so we have an empty
							// prefix
							prefix = []byte{}
							awaitingPrefix = false
							advance = len(sb)
						}
						// else, no zero-length prefix

						// either way, move on to modeMiddle
						mode = modeMiddle
						err = scanner.ErrContinue
					}
					// else, we don't have enough data to know if we've got a
					// zero-length prefix yet or not.

				case modeMiddle:
					// we are now looking for parts or possibly the prefix if it is
					// not a zero byte prefix
					if ix := bytes.Index(data, mb); ix >= 0 {
						// we found a \n--boundary\n string:
						// |-> advance past the boundary for the next token
						// |-> if awaitingPrefix, capture prefix
						// |-> if not awaitingPrefix, return token
						advance = ix + len(mb)
						if awaitingPrefix {
							// this is the first boundary, so the input so far is
							// the prefix
							ps := data[:ix+1]
							prefix = make([]byte, len(ps))
							copy(prefix, ps)
							awaitingPrefix = false
						} else {
							// this is a subsequent boundary, so the input is a part
							token = data[:ix]
						}
					} else if atEOF {
						// we didn't find a regular boundary, but we're at EOF, so
						// it's time to search for the final boundary
						mode = modeEnd
						err = scanner.ErrContinue
					}
					// else, we aren't at EOF, so there's more input and we may
					// yet find more interior boundaries to split on
				case modeEnd:
					// If we get here and we are still awaitingPrefix, this message
					// is badly formatted. We have no initial boundary at all. We
					// record that by setting prefix to nil so that when
					// round-tripping, the initial prefix will be missing. Then, we
					// treat the data before the final boundary as if it is the
					// message.
					if awaitingPrefix {
						prefix = nil
					}

					// if we are here, we know that atEOF is true
					if ix := bytes.Index(data, eb); ix >= 0 {
						// we found the end \n--boundary--\n string:
						// |-> capture the suffix, which is everything after the
						// |   boundary (including the line ending, which is why
						// |   we use len(fb) here and not len(eb), that is
						// |   deliberate.
						// |-> capture the token to return as the final part
						token = data[:ix]
						ss := data[ix+len(fb):]
						suffix = make([]byte, len(ss))
						copy(suffix, ss)
					} else if ix := bytes.Index(data, fb); ix == len(data)-len(fb) {
						// we found the final \n--boundary-- string at the actual
						// end of input (no final line break)
						// |-> there's no suffix, not even a newline
						// |-> capture the token to return as the final part
						token = data[:ix]
						suffix = []byte{}
					} else {
						// bummer, we have no final boundary, so we'll just treat
						// the rest of the data as the final part and record that
						// we have no suffix (when round-tripping, the final
						// boundary will still be omitted).
						token = data
						suffix = nil
					}
					// either way, we're done
					err = bufio.ErrFinalToken
				default:
					// never happens, right?
					panic("unexpected parser state")
				}
				return
			},
		),
	)

	// This function will recover the original message if we get an error
	// parsing a sub-part.
	parts := make([][]byte, 0, 10)
	originalMessage := func() (*Opaque, error) {
		// finish accumulating the parts and find the suffix (if any)
		for sc.Scan() {
			part := sc.Bytes()
			parts = append(parts, part)
		}

		if err := sc.Err(); err != nil {
			if errors.Is(err, bufio.ErrTooLong) {
				return nil, ErrLargePart
			} else {
				// TODO Can this ever happen? If so, how should we handle it?
				return nil, err
			}
		}

		r := &bytes.Buffer{}
		if prefix != nil {
			r.Write(prefix)
			r.Write(sb)
		}
		r.Write(bytes.Join(parts, mb))
		if suffix != nil {
			r.Write(eb)
			r.Write(suffix)
		}

		return &Opaque{
			Header: msg.Header,
			Reader: r,
		}, nil
	}

	// All returned tokens are parts
	msgParts := make([]Generic, 0, 10)
	for sc.Scan() {
		part := sc.Bytes()
		parts = append(parts, part)

		// parse each part as a simple message first
		opMsg, err := pr.parseToOpaque(bytes.NewReader(part), true)
		if err != nil {
			orig, _ := originalMessage()
			return orig, err
		}

		msg, err := pr.parse(opMsg, depth-1)
		if err != nil {
			orig, _ := originalMessage()
			return orig, err
		}

		msgParts = append(msgParts, msg)
	}

	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, ErrLargePart
		} else {
			// TODO Can this ever happen and, if so, how should we handle it?
			orig, _ := originalMessage()
			return orig, err
		}
	}

	return &Multipart{
		Header: msg.Header,
		prefix: prefix,
		suffix: suffix,
		parts:  msgParts,
	}, nil
}
