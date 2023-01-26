package transfer

import "io"

// writer is an internal helper to make wrapping easier.
type writer struct {
	io.Writer
	io.Closer
}

// Close will close the nested writer if performClose is true.
func (w *writer) Close() error {
	if w.Closer != nil {
		return w.Closer.Close()
	}
	return nil
}
