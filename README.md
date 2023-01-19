# go-email

This is a package primarily aimed at parsing, manipulating, and outputting email message documents. The v2 release of this module is a full-on Goland release. The original version (available [here](TODO)) was mroe closely based on the work of Simon Cozens, Ric Signes, and others on Perl-based modules [Email::Simple](https://metacpan.org/pod/Email::Simple) and [Email::MIME](https://metacpan.org/pod/Email::MIME). That foundation has not gone away, but the v2 release is a rewrite of the API to make this library more memory efficient and have a more Go-ish interface. This also keeps the round-trip guarantees of the previous version. That is, if you parse a message and immediately write it back out, the result should be identical. Also, if you modify a part of the message, only the modified parts will change, the rest will remain byte-for-byte identical to the parsed in data.

As such, this library attempts to keep it's pedigree by adopting the capabilities of mature libraries while implementing the latest features of RFC 5322, and to do so in a way that should be comfortable to Golang developers.

# Example Applications

## Processing Header Keywords

Here's a quick example showing some code that is will manipulate the headers of a program:

```go
package main

import (
	"io"
	"os"

	"github.com/zostay/go-email/v2/message"
)

func main() {
	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	// WithoutMultipart() means we want the top level headers only.
	m, err := message.Parse(msg, message.WithoutMultipart())
	if err != nil {
		panic(err)
	}

	// update the keywords of the new message
	if kws, err := m.GetHeader().GetKeywords(); err == nil && len(kws) > 0 {

		for _, kw := range kws {
			if kw == "Snuffle" {
				out := &message.Buffer{}
				out.Header = *m.GetHeader() // copy the original header
				content := m.GetReader()
				_, err = io.Copy(out, content) // copy the original message body
				if err != nil {
					panic(err)
				}

				// add Upagus to Keywords
				outKws := make([]string, len(kws)+1)
				outKws[len(kws)] = "Upagus"
				out.SetKeywords(outKws...)

				outMsg, err := os.Create("output.msg")
				if err != nil {
					panic(err)
				}

				_, err = out.WriteTo(outMsg)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
```

## Saving Off Message Attachments

Or if you are interested in processing message contents, consider this example:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/walker"
)

func main() {
	var fileCount = 0
	isUnsafeExt := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsDigit(c)
	}

	outputSafeFilename := func(fn string) string {
		safeExt := filepath.Ext(fn)
		if strings.IndexFunc(safeExt, isUnsafeExt) > -1 {
			safeExt = ".wasnotsafe" // check your input
		}
		fileCount++
		return fmt.Sprintf("%d.%s", fileCount, safeExt)
	}

	var saveAttachments walker.Parts = func(depth, i int, part message.Part) error {
		h := part.GetHeader()

		presentation, err := h.GetPresentation()
		if err != nil {
			panic(err)
		}

		fn, err := h.GetFilename()
		if err != nil {
			panic(err)
		}

		if presentation == "attachment" && fn != "" {
			of := outputSafeFilename(fn)
			outMsg, err := os.Create(of)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(outMsg, part.GetReader())
			if err != nil {
				panic(err)
			}
		}
	}

	msg, err := os.Open("input.msg")
	if err != nil {
		panic(err)
	}

	// we want to decode the transfer encoding to make sure we get the original
	// binary values of the message contents when saving off attachments
	m, err := message.Parse(msg, message.DecodeTransferEncoding())
	if err != nil {
		panic(err)
	}

	_ = saveAttachments.WalkOpaque(m)
}
```

## Generating a New Message

You can also create new messages for sending using this library:

```go
package main

import (
	"fmt"
	"net/smtp"

	"github.com/zostay/go-email/v2/message"
	"github.com/zostay/go-email/v2/message/transfer"
)

func main() {
	// Build a part that will be the attached document
	resume, _ := message.AttachmentFile(
        "resume.pdf",
        "application/pdf",
        transfer.Base64,
	)

	// Build a part that will contain the message content as text
	text := &message.Buffer{}
	text.SetMediaType("text/plain")
	_, _ = fmt.Fprintln(text, "You will find my awesome resume attached.")

	// Build a part that will contain the message content as HTML
	html := &message.Buffer{}
	html.SetMediaType("text/html")
	_, _ = fmt.Fprintln(html, "You will find my <strong>awesome</strong> resume attached.")

    // Build the top-level message from the parts.
    main := &message.Buffer{}
    main.SetSubject("My resume")
    main.SetTo("recruiter@example.com")
    main.SetFrom("me@example.com")
    main.SetMediaType("multipart/mixed")
    main.Add(
        message.MultipartAlternative(html.Opaque(), text.Opaque()),
        resume,
    )
    mainMsg := main.Opaque()

    // send the message via SMTP
    c, err := smtp.Dial("smtp.example.com:25")
    if err != nil {
        panic(err)
    }

    _ = c.Hello("me")
    _ = c.Mail("me@example.com")
    _ = c.Rcpt("recruiter@example.com")
    w, err := c.Data()
    if err != nil {
        panic(err)
    }
    _, _ = mainMsg.WriteTo(w)
    _ = w.Close()
    _ = c.Quit()
}
```

# Message Components

The top-level message objects are divided roughly into three major components:

1. An `message.Opaque` represents a parsed or constructed email message whose body is treated as a slice of bytes (available via an io.Reader). The content of the message might be a multipart MIME message or might just be text.
2. A `message.Multipart` represents a parsed or constructed email message that represents a multipart MIME message with at least one layer of parts below it. These parts might be other Multipart message or they might be Opaque messages.
3. A `message.Buffer` is a tool for building messages, which can be returned as either `message.Opaque` objects or (if they are multipart MIME messages) as `message.Multipart` objects.

It is important to note that any time you encounter a `message.Opaque` object, it might contain a simple message without sub-parts. Or it might contain a complex MIME multipart message, but we just haven't broken it up into parts. However, a `message.Multipart` object always contains a MIME multipart message that has been broken up for at least that layer, but the layers within might or might not be broken up. It is up to you how you want to slice and dice your messages.

# Parsing Messages

To do anything with an existing message, you'll need to parse it. The message parsing features of this library provide several levers to let you control how message parsing proceeds. Using just the defaults, parsing is quite simple. 

For example:

```go
r := strings.NewReader("Subject: test\n\nThis is a test.")
msg, err := message.Parse(r)
```

Parsing is performed in three phases:

* Phase #1: Finds the break between header a message body and determines what kind of line break to use when processing this message (usually `"\n"` or `"\r\n"`).
* Phase #2: Recursively parsing the body content for multipart content.
* Phase #3: Decoding the transfer encoding of each message body.

You may control aspects of these phases using the options described in the sections below.

## Parsing Multipart MIME Messages

The defaults will usually parse a message completely. In the example above, the message is a single layer deep and the object returned in `msg` will be a `*message.Opaque` object. If the message is a multipart MIME message, the returned message will be a `*message.Multipart` object instead, with all the parts and sub-parts parsed out, both implement the `message.Generic` interface, which allows you to work with either implementation. 

Multipart processing is performed breadth-first. The top-most parts are split and then any parts within those are split and on downward. The default settings will parse almost any message completely, down to the deepest sub-part. However, if a message is especially complicated (probably unreasonably complicated), some multipart sub-parts will be returned unparsed as `*message.Opaque` objects. The default depth where stopping happens is 10. 

If you want to change how this multipart parsing is performed, there are a number of settings:

```go
// This will only parse to the 5th layer deep.
m0, _ := message.Parse(r0, message.WithMaxDepth(5))

// This will not parse even the first layer.
// This always returns an *message.Opaque object.
m1, _ := message.Parse(r1, message.WithoutMultipart()) // same as WithMaxDepth(0)

// This will parse the first layer, but no further. If the message is a
// multipart message it will be *message.Multipart but all sub-parts are
// guaranteed to be *message.Opaque. Otherwise, it may return *message.Opaque.
m2, _ := message.Parse(r2, message.WithoutRecursion()) // same as WithMaxDepth(1)

// Or you can turn off all limits and get everything...
m3, _ := message.Parse(r3, message.WithUnlimitedRecursion()) // same as WithMaxDepth(-1)
```

The `message.WithoutMultipart()` option is especially useful if you only plan to work with the headers of the top-level message.

## Decoding Transfer Encoding

The `Content-transfer-encoding` header is often important when processing the content of a message body. However, if perform this decoding automatically, it becomes costly to round-trip the message and keep unchanged messages byte-for-byte identical. As such, this library opts to not process transfer encoding unless asked to do so.

Therefore, if you want to process message contents, you will want to specify the `messsage.DecodeTransferEncoding()` option to enable that decoding. If you are just processing headers, you don't need to worry about it.

## Maximum Lengths

During parsing there are limits set on how long a message header and a message part may be. The `message.DefaultMaxHeaderLength` and the `message.DefaultMaxPartLength` provide defaults for these. These default values are, as of this writing, set to `bufio.MaxScanTokenSize`, which is 65KiB. This is a pretty reasonable setting for headers. This is likely to be too small, though, for messages with large attachments.

If you receive a `message.ErrLargeHeader` error ("the header exceeds the maximum parse length") while parsing a message, you should consider raising the size limit on header length using the `message.WithMaxHeaderLength()` option.

If you receie a `message.ErrLargePart` error ("a message part exceeds the maximum parse length") while parsing a message, you should consider raising the size limit on part length using the `message.WithMaxPartLength()` option.

## Chunking Performance

During the first two phases of parsing, the content of the message will be read in chunks. The default chunk size is called `message.DefaultChunkSize` and you can modify this with the `message.WithChunkSize()` option.

# Building Messages

To build a message using this library, you will want to use `message.Buffer`. This is an object whose purpose is to let you build up new messages from parts.

Any use of `message.Buffer` should involve setting one or more headers on the buffer. The `message.Buffer` is a `header.Header`, so you can call all the header manipulations methods on it directly.

From there, you may treat the `message.Buffer` as either an `io.Writer` or call the `Add()` method on it to add one or more `message.Part` objects to it. Be aware, though, that once you treat it as an `io.Writer`, you must only treat it as an `io.Writer` or the object will panic. Or if you call `Add()`, you must not treat it as an `io.Writer` or the object will panic.

When you have completely built your message, you may call either `Opaque()` or `Multipart()` to get that kind of object back out. You can do this regardless of how you worked with the buffer, with a couple caveats:

* If you used the buffer as an `io.Writer`, the message returned by `Opaque()` will contain an `io.Reader` with those byte in it.
* If you called the `Add()` method on the buffer, the message returned by `Opaque()` will have an `io.Reader` containing a serialized version of all the sub-parts (and their sub-parts, and so on).
* If you called the buffer as an `io.Writer`, the message returned by `Multipart()` must parse as a multipart MIME message. If not, an error will be returned instead of a `message.Multipart` object.
* If you called the `Add()` method on the buffer, the message returned by `Multipart()` will have all the parts added available via the `GetParts()` method. They will be returned exactly as added.

In either case, you should dispose of the `message.Buffer` object after calling either `Opaque()` or `Multipart()` methods. Continuing to modify the buffer after that might have unforeseen consequences.

## Alternative to Using Buffer

Another way to build a message is to create a `message.Opaque` object. Then, modify the header and set the `Reader` field to an `io.Reader` of your choice (e.g., an `os.File` containing the bytes you want to use). If this the file contains binary data, be sure to set the `Content-transfer-encoding` header to `base64` or something to ensure the data will be transferred properly.

# Memory Considerations

This library is somewhat careful with memory. If you use `message.Buffer` to create a new message, the amount stored will typically be very close to the number of bytes you write to the header and body of your message.

Parsing, however, tends to be about twice as expensive for headers as each header field will be stored in an original raw form as well as a decoded form. As headers are typically small anyway, this is not considered a serious problem.

If the `message.WithoutMultipart()` option is used during parsing, then the reader will be read as far as necessary to find and parse the message. But the message body will not be read in full into memory. For large files, this is likely a significant memory savings.

However, any other setting for handling multipart handling, may result in the entire content of the message being read into memory. If the returned object is a `message.Multipart`, then the entire message has already been read into memory.

# Copyright & License

Please note, some of the test data used to test this project are under a
different license and copyright. See the `LICENSE` file in the `test/data` directory.

Copyright 2020 Andrew Sterling Hanenkamp

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
