package message

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/zostay/go-email/pkg/v2/header"
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
)

// Errors that occur during parsing.
var (
	// ErrNoBoundary is returned by Parse when the boundary parameter is not set
	// on the Content-type field of the message header.
	ErrNoBoundary = errors.New("the boundary parameter is missing from Content-type")
)

// ParseError is returned when one or more errors occur while parsing an email
// message. It collects all the errors and returns them as a group.
type ParseError struct {
	Errs []error // the list of errors that occurred during parsing
}

// Error returns the list of errors encountered while parsing an email message.
func (err *ParseError) Error() string {
	errs := make([]string, len(err.Errs))
	for i, e := range err.Errs {
		errs[i] = e.Error()
	}
	return "error parsing MIME message: " + strings.Join(errs, ", ")
}

var splits = [][]byte{
	[]byte("\x0d\x0a\x0d\x0a"), // \r\n\r\n
	[]byte("\x0a\x0d\x0a\x0d"), // \n\r\n\r, extremely unlikely, possibly never
	[]byte("\x0a\x0a"),         // \n\n
	[]byte("\x0d\x0d"),         // \r\r
}

type parser struct {
	maxDepth  int
	chunkSize int
}

func (pr *parser) clone() *parser {
	p := *pr
	return &p
}

var defaultParser = &parser{DefaultMaxMultipartDepth, DefaultChunkSize}

// ParseOption refers to options that may be passed to the Parse function to
// modify how the parser works.
type ParseOption func(pr *parser)

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
func searchForSplit(buf []byte) (pos int, crlf []byte) {
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
func (pr *parser) splitHeadFromBody(r io.Reader) ([]byte, []byte, io.Reader, error) {
	p := make([]byte, pr.chunkSize)
	buf := &bytes.Buffer{}
	searched := 0
	for {
		// read in some bytes
		n, err := r.Read(p)
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
		pos, crlf := searchForSplit(buf.Bytes()[searched:])
		if pos >= 0 {
			// we found the split, return the data
			hdr := make([]byte, pos)
			_, _ = buf.Read(hdr)
			return hdr, crlf, &remainder{buf.Bytes(), r}, nil
		}

		// No split found and EOF? Let's break out and then we'll process as if
		// the entire message is just header.
		if isEof {
			break
		}

		// The last 3 bytes might be the prefix to the split point
		searched = buf.Len() - 3
	}

	// If we're here, we were unable to find a header/body split. We will just
	// assume the message is all header, no body. Let's see if we can find what
	// to use as a break.
	for _, s := range splits {
		crlf := s[0 : len(s)/2]
		if bytes.Contains(buf.Bytes(), crlf) {
			return buf.Bytes(), crlf, &bytes.Buffer{}, nil
		}
	}

	// Or the ultimate fallback is...
	return buf.Bytes(), []byte("\x0d"), &bytes.Buffer{}, nil
}

// // ParseOpaque will turn the given input into an Opaque by detecting the line
// // break used to split the header from the body, using that break to split the
// // header part from the body part, and parsing the header. The body, whatever it
// // is, is kept as an opaque value provided in the io.Reader part of the
// // constructed Opaque object.
// func ParseOpaque(r io.Reader, opts ...ParseOption) (*Opaque, error) {
// 	pr := defaultParser.clone()
// 	for _, opt := range opts {
// 		opt(pr)
// 	}
//
// 	return pr.parseOpaque(r)
// }

// parseOpaque turns a reader into an Opaque.
func (pr *parser) parseToOpaque(r io.Reader) (*Opaque, error) {
	hdr, crlf, body, err := pr.splitHeadFromBody(r)
	if err != nil {
		return nil, err
	}

	head, err := header.Parse(hdr, crlf)
	if err != nil {
		return nil, err
	}

	return &Opaque{*head, body}, nil
}

// Parse will transform a *Opaque into a *Multipart or return a *Opaque if
// the object does not represent a multipart message. The parse of the message
// will proceed as follows:
//
// 1. We check to see if the Content-type is a multipart/* or message/* type. If
// it is not, the original message will be returned as-is and no parsing of the
// body of the message will be attempted.
//
// 2. We check to see if the boundary parameter is set on the Content-type
// header. If not, the original message is returned without attempting to read
// the body, but an error is returned, ErrNoBoundary, because a multipart
// message cannot be parsed without a boundary.
//
// 3. If Content-type and boundary checks both pass, the message body will be
// read to search for boundary markers. If no initial boundary marker is found
// when reading starts, a new message will be returned containing the original
// message data and an ErrMissingBoundary error will be returned.
//
// 4. If we get thi far and the initial boundary is found, then the remaining
// boundaries continue to be read until we reach the end of the message. At this
// point a *Multipart will be returned with all the parts broken up into
// pieces. If the end boundary is missing, it will also return an error
// ErrMissingEndBoundary.
func Parse(r io.Reader, opts ...ParseOption) (Generic, error) {
	pr := defaultParser.clone()
	for _, opt := range opts {
		opt(pr)
	}

	msg, err := pr.parseToOpaque(r)
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
	// For the purpose of capturing content, the newline before the boundary is
	// left with the prefix or the part, but the newline after the boundary is
	// considered part of the boundary.
	sb := []byte(fmt.Sprintf("--%s%s", pv.Boundary(), msg.Break()))
	mb := []byte(fmt.Sprintf("%s--%s%s", msg.Break(), pv.Boundary(), msg.Break()))
	eb := []byte(fmt.Sprintf("%s--%s--%s", msg.Break(), pv.Boundary(), msg.Break()))
	brkLen := len(msg.Break().Bytes())

	const (
		modeStart = iota
		modeMiddle
		modeEnd
	)

	// This scanner split function splits on any email message boundary. It
	// returns the parts as tokens, but the prefix and suffix, it captures
	// itself in the prefix/suffix vars.
	sc := bufio.NewScanner(msg.Reader)
	var prefix, suffix []byte
	mode := modeStart
	awaitingPrefix := true
	sc.Split(
		func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if mode == modeStart {
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
				}
				// else, we don't have enough data to know if we've got a
				// zero-length prefix yet or not.

			} else if mode == modeMiddle {
				// we are now looking for parts or possibly the prefix if it is
				// not a zero byte prefix
				if ix := bytes.Index(data, mb); ix >= 0 {
					// we found a \n--boundary\n string:
					// |-> advance past the boundary for the next token
					// |-> if awaitingPrefix, capture prefix
					// |-> if not awaitingPrefix, return token
					// |-> prefix/token will keep their final newline
					advance = ix + len(mb)
					if awaitingPrefix {
						// this is the first boundary, so the input so far is
						// the prefix
						prefix = data[:ix-brkLen]
						awaitingPrefix = false
					} else {
						// this is a subsequent boundary, so the input is a part
						token = data[:ix-brkLen]
					}
				} else if atEOF {
					// we didn't find a regular boundary, but we're at EOF, so
					// it's time to search for the final boundary
					mode = modeEnd
				}
				// else, we aren't at EOF, so there's more input and we may
				// yet find more interior boundaries to split on
			} else if mode == modeEnd {
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
					// we found the final \n--boundary--\n string:
					// |-> capture the suffix, which is everything after the
					// |   boundary
					// |-> capture the token to return as the final part
					token = data[:ix-brkLen]
					suffix = data[ix+len(eb):]
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
			} else {
				// never happens, right?
				panic("unexpected parser state")
			}
			return
		},
	)

	// This function will recover the original message if we get an error
	// parsing a sub-part.
	parts := make([][]byte, 0, 10)
	originalMessage := func() *Opaque {
		// finish accumulating the parts and find the suffix (if any)
		for sc.Scan() {
			part := sc.Bytes()
			parts = append(parts, part)
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
		}
	}

	// All returned tokens are parts
	msgParts := make([]Generic, 0, 10)
	for sc.Scan() {
		part := sc.Bytes()
		parts = append(parts, part)

		// parse each part as a simple message first
		opMsg, err := pr.parseToOpaque(bytes.NewReader(part))
		if err != nil {
			return originalMessage(), err
		}

		msg, err := pr.parse(opMsg, depth-1)
		if err != nil {
			return originalMessage(), err
		}

		msgParts = append(msgParts, msg)
	}

	return &Multipart{
		Header: msg.Header,
		prefix: prefix,
		suffix: suffix,
		parts:  msgParts,
	}, nil
}

// remainder takes the bytes already read from an io.Reader and make a new
// reader that returns those bytes first and then passes the reads from the
// unread part of the io.Reader on to the caller.
type remainder struct {
	prefix []byte
	r      io.Reader
}

// Read perform a read from the prefix buffer first, if any bytes remain. Once
// those bytes have been consumed, it starts consuming bytes from the io.Reader.
func (r *remainder) Read(p []byte) (n int, err error) {
	// read from prefix first
	if len(r.prefix) > 0 {
		n = copy(p, r.prefix)
		r.prefix = r.prefix[n:]
	}

	// if reading from prefix did not fill p, read from reader too
	if n < len(p) {
		var rn int
		rn, err = r.Read(p[n:])
		n += rn
	}

	return n, err
}

// Close implements io.Closer, just in case the nested io.Reader needs it. It
// passes the Close() call through. If the io.Reader is not an io.Closer, this
// is a no-op.
func (r *remainder) Close() error {
	if c, isCloser := r.r.(io.Closer); isCloser {
		return c.Close()
	}
	return nil
}
