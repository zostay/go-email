# go-email

This is a package primarily aimed at parsing, manipulating, and outputting email
message documents. It is aimed at being very flexible and is loosely based on
the work of Simon Cozens, Ric Signes, and others on their Perl based
[Email::Simple](https://metacpan.org/pod/Email::Simple) and
[Email::MIME](https://metacpan.org/pod/Email::MIME) libraries. I wouldn't go so
far as to call it a port of these libraries.  However, the style and shape of
the parser and other features do bear some resemblance.

These libraries have been workhorses for email processing for decades and are
good repositories of practical knowledge when it comes to how things actually
work in the wild, not just what RFC 822, RFC 2822, RFC 5322 and others have to
say about how they ought to work.

# Synopsis

Here's a quick example showing some code that is just interested in manipulating
headers:

```Go
package main

import (
	"io/ioutil"
	"strings"

	"github.com/zostay/go-email/pkg/email/simple"
)

func main() {
	msg, err := ioutil.ReadFile("input.msg")
	if err != nil {
		panic(err)
	}

	m, err := simple.Parse(msg)
	if err != nil {
		panic(err)
	}

	if kw := m.HeaderGet("Keywords"); kw != "" {
		kws := strings.Split(kw, ",")
		for _, k := range kws {
			if strings.TrimSpace(k) == "Snuffle" {
				kw += ", Upagus"
				_ = m.HeaderSet("Keywords", kw)
				_ = ioutil.WriteFile("output.msg", m.Bytes(), 0644)
			}
		}
	}
}
```

Or if you are interested in message bodies:

```Go
package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zostay/go-email/pkg/email/mime"
)

var fileCount = 0

func isUnsafeExt(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsDigit(c)
}

func outputSafeFilename(fn string) string {
	safeExt := filepath.Ext(fn)
	if strings.IndexFunc(safeExt, isUnsafeExt) > -1 {
		safeExt = ".wasnotsafe" // CHECK INPUT YOU CRAZY PERSON
	}
	fileCount++
	return fmt.Sprintf("%d%s", fileCount, safeExt)
}

func saveAttachments(m *mime.Message) {
	if fn := m.Filename(); fn != "" {
		of := outputSafeFilename(fn)
		b, _ := m.ContentUnicode()
		_ = ioutil.WriteFile(of, []byte(b), 0644)
	} else {
		for _, p := range m.Parts {
			saveAttachments(p)
		}
	}
}

func main() {
	msg, err := ioutil.ReadFile("input.msg")
	if err != nil {
		panic(err)
	}

	m, err := mime.Parse(msg)
	if err != nil {
		panic(err)
	}

	saveAttachments(m)
}
```

# Goals

There are a few fundamental goals that guided this work. The include the
following:

1. **Round-tripping.** When working with messages, the parsing should be
   performed in a byte-for-byte reversible fashion. If you parse a message, do
   nothing to it and then call `String()`, you should get back exactly what was
   parsed. Furthermore, if you modify any aspect of the message, all the other
   aspects should continue to be round-trip-able insofar as possible.

2. **Logically Useful.** The parsing should provide enough information to allow
   the developer to work with the important content easily. This means that
   headers like To and From must be provided as fully parsed address lists.
   Dates must be provided as `time.Time` objects, and `Content-type` and
   `Content-disposition` headers must be provided in structured form. Not all
   headers are logically available this way, but the most important ones are.
   Also, message bodies can be decoded and have their text reformatted as
   Go-native Unicode or plain binary handling as required by the message content
   and as requested by the developer.

3. **Liberal Acceptance.** This library attempts to read messages, even those
   that are illegal, if there's a reasonable heuristic for guessing what is
   meant. For example, email headers may be folded across lines, but such lines
   are required to have white space indents. However, in the real world, some
   mailers fail to correctly do this. If a line is encountered that is folded
   incorrectly, this library will continue parsing the data anyway.

4. **Prefer Strict Output on Change.** If something is badly formatted on input,
   it will still be round-tripped as-is on output. However, if a part is
   updated, this library will attempt to correct the error and output strictly
   correct messages according to RFC 5322 or whatever else is the latest
   standard on the subject. (Please report a bug if the library does not do
   this.)

5. **Allow Bad Output on Change.** There may be cases where you want to preserve
   the bad output. Methods are provided to allow you to bypass the checks and
   output header fields and message bodies that are badly formatted. This
   library will not help you with this, but it will let you do it.

6. **Resume-able Error Handling.** When possible, errors will be returned with
   partially correct objects. If a parse error occurs, this library will attempt
   to give you what it has. For example, if a message has junk text at the top
   that does not appear to a mail header, it will read through that text until
   it gets something that looks like an email. Then it will return the parsed
   message found and return an error that includes the junk content at the
   front. Consider this idiomatic error handling code:

   ```go
   m, err := simple.Parse([]byte("junk text\nSubject: bob\n\nMessage\n"))
   if err != nil {
       return err
   }
   ```

   In this case, the message is parsed and thrown away. However, if you want to
   keep the message anyway, you can:

   ```go
   m, err := simple.Parse([]byte("junk text\nSubject: bob\n\nMessage\n"))
   if err != nil {
       fmt.Println(err)
   }
   // continue working with m
   fmt.Println("Got an email with subject: " + m.HeaderGet("Subject"))
   ```

   This is not always possible or reasonable, but where it is, this style of
   error handling is preferred.

# Anti-Goals

There are also some things I deliberately avoided trying to solve when writing
this library. This includes:

1. **Unattended Production Use.** I recommend using this library in an
   unattended production situation. It makes several decisions that will make
   production use in a server or other automated situation risk. I'm intending
   to use this library in a command-line tool I run by hand and monitor the
   output of directly.

2. **Memory Performance.** In memory performance of this library is absolutely
   abysmal. I haven't measured it, but I wouldn't be surprised if the parsed the
   structures are 3 or 4 times larger than the original file. If you parse a 100
   MB email, you might need 1 GB of memory available to complete the operation.
   I might look into ways of avoiding this to some extent, but it is not a
   priority.

3. **Rigid Strictness.** I'm not going to enforce rigid strictness on output. If
   you are a terrible developer or have need of doing terrible things with your
   email, so be it. I ain't your momma. I am not going to give you any support
   either, so don't file an issue if you're doing something broken or weird.
   You're on your own.

4. **Support.** Speaking of support. I'm just a guy doing this in my spare time.
   I don't do much email messaging for work, so I'm not going to support this
   very well.  Patches welcome. I will apply them if they make sense, have good
   docs and tests, and don't break anything I need, but I'm not likely to tinker
   with these packages.

5. **Email Creation.** It should be possible to create email messages
   programmatically, but this aspect of things is pretty skimpy right now. This
   library is primarily aimed at parse-modify-format operations.

# Copyright & License

Please note, some of the test data used to test this project are under a
different license and copyright. See the `LICENSE` file in the `test` directory.

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
