package email

// Message represents an email message. It is composed of a Header and a Body.
type Message interface {
	Body
	Header
	Outputter
}
