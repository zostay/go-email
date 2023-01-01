// Package email represents a newly rewritten version of
// github.com/zostay/go-email. The original version was a largely a port from
// Perl libraries named Email::Simple and Email::MIME. This version takes all
// the information I've learned from Golang best practices, knowledge of email,
// code ported from those Perl modules, and tried to create a very go-ish
// implementation of the library.
//
// Rather than splitting up the code according to a Simple/MIME email
// boundary, the code is split according to part of message.
// The Simple/MIME handling of the previous library was slightly unnatural,
// even in the original Perl code and largely came from the fact that each
// was a separate library with a different purpose. Instead,
// we primarily treat a message as either a message.Opaque or a
// message.Multipart. The message.Opaque means this library simply treats the
// message body as an io.Reader and assigns no meaning to the contents. The
// message.Multipart breaks multipart messages up into sub-parts, which can
// then be dealt with separately. These will either be another message.Multipart
// that can be further broken up, or a message.Opaque that is treated as an
// io.Reader. The end user is then free to pack multipart messages in to
// message.Opaque objects or deal with them in parts via message.Multipart.
// Other message parts content is up to the end user to deal with as they
// choose with some additional general helps as appropriate and implemented.
//
// If you want to build new messages, a message.Buffer is provided that has the
// capability for creating either message.Opaque or message.Multipart or
// combinations of these.
//
// If you want to transform messages, you can start by using the message.Parse()
// method to parse the message as far as you are interested. From there, you may
// make simple changes by modifying the header. You can transform message.Opaque
// message body by wrapping the io.Reader of the start reader and then writing
// the modified data back out using the WriteTo() method. You can transform
// message.Multipart message parts by iterating over the parts and inserting new
// parts, omitting parts, or manipulating parts as you like in a destination
// message.Buffer. Round-tripping in this last case are not as good as they were
// before, but I figure if you're doing something as extreme as culling attached
// executables or something like that, you can afford to cause the message to be
// reformatted more strictly in the process.
//
// For dealing with messages headers, the high-level interface is provided via
// header.Header. Low-level access can be granted by using header.Header to
// work directly with field.Field objects, but typically, just working with
// fields as part of the header object is probably all most end users will need.
//
// In addition, I've attempted to incorporate some improvements from RFC 5322 to
// help ensure messages generated from this library are correct and using this
// library is convenient.
//
// Finally, as much as possible, I've tried to preserve the round-tripping
// capabilities from the previous library. If you modify some aspect of a
// message, the rest will remain byte-for-byte identical on output, insofaras I
// can manage to pull that off. As I use this as the basis for some mail
// filtering software, it is important that messages get written out the same as
// they started.
package email
