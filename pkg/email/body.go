package email

// Body is a basic wrapper around a slice of bytes of data.
type Body struct {
	content []byte
}

// NewBody builds a new body with the given content bytes.
func NewBody(content []byte) *Body {
	return &Body{content}
}

// ContentString returns the message body as a string.
func (m *Body) ContentString() string { return string(m.content) }

// Content returns the message body.
func (m *Body) Content() []byte { return []byte(m.ContentString()) }

// SetContent sets the message body.
func (m *Body) SetContent(c []byte) { m.content = c }

// SetContentString sets the message body from a string.
func (m *Body) SetContentString(s string) { m.content = []byte(s) }

// String returns the email message as a string.
func (m *Body) String() string {
	return m.ContentString()
}

// Bytes returns the email message as a slice of bytes.
func (m *Body) Bytes() []byte {
	return m.Content()
}
