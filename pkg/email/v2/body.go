package email

// Body represents a generic message body.
type Body interface {
	Outputter

	// Content returns the content of the message body as a slice of bytes.
	Content() []byte

	// SetContent sets the message content to the given slice of bytes.
	SetContent([]byte)

	// ContentString returns the content of the message body as a string.
	ContentString() string

	// SetContentString sets the message content t the given string.
	SetContentString(string)
}
