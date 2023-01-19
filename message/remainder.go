package message

import "io"

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
		rn, err = r.r.Read(p[n:])
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
