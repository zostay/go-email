v2.3.1  2023-01-30

 * Bugfix: Handle another strange date I have come across in my sample data.

v2.3.0  2023-01-30

 * On `header.ErrManyFields` failure, `(*header.Header).Get()` now returns the value of the first field found with the error.
 * Add a `header.ParseTime()` function which provides the same time parsing facilities that `(*header.Header).GetTime()` and `(*header.Header).GetDate()` use.
 * Add a `header.ParseAddressList()`function which provides the same `addr.AddressList` parsing facilities that are built in to the various address parsing methods of `header.Header`.

v2.2.1  2023-01-30

 * Bugfix: Fixed a typo in `header.ErrNoSuchField` error message.

v2.2.0  2023-01-30

 * Deprecation: The contents of the `walker` package are deprecated and will be removed in a future release. Use the features of the `walk` package instead.
 * Add the `walk` package.
 * Add `walk.AndProcess` and `walk.Processor` to handle walking a message part tree with the goal of generic message processing.
 * Add `walk.AndTransform` and `walk.Transformer` to handle walking a message part tree, which is a process tool that aids with message transformation.
 * Add `walk.AndProcessSingle` to process single parts of a message.
 * Add `walk.AndProcessMultipart` to process multipart parts of a message.
 * Add `message.NewBuffer` to copy message.Part into a message.Buffer.
 * Add `message.NewBlankBuffer` to copy `message.Part` into a `message.Buffer`, but without body content or sub-parts.
 * Add `(*message.Buffer).AddBuffers` method for convenience.

v2.1.0  2023-01-27

 * Deprecation: The `OpaqueAlreadyEncoded()` method should no longer be used. Instead, call `SetEncoded(true)` on a mesage.Buffer and then call `Opaque()` to archive the same result.
 * Calls to `message.Buffer`'s `Opaque()` and `Multipart()` methods are now guaranteed to work if called repeatedly.
 * Add the `Buffer.SetMultipart` method for setting the `message.BufferMode` of a `message.Buffer` to `message.ModeMultipart` and setting the capacity of the internal parts slice.
 * It is now possible to use message.Buffer as a `message.Part` directly.
 * Add the `Buffer.SetOpaque` method for setting the `message.BufferMode` of a `message.Buffer` to `message.ModeOpaque`.
 * Add a linter to check changelog correctness.
 * Add a linter to check for bad patterns in golang.
 * Add tools to manage the release process.

v2.0.3  2023-01-22

 * Bugfix: Base64 transfer encoding inserts line breaks at column 76 now.
 * Bugfix: `AttachmentFile` now sets `Content-disposition` to `attachment` and sets the `filename` parameter correctly.
 * Bugfix: Junk at the start of the message is more recoverable than it was. The message parsed without junk will be returned with the `field.BadStartError`.
 * Add `Clone()` method to `field.Field`, `header.Base`, and `header.Header`
 * More test coverage for transfer encoding support.
 * More test coverage for `message.Opaque` features.
 * More test coverage for `message.Parse` features.

v2.0.2  2023-01-19

 * Docfix: Made some additional correction to examples in the docs.

v2.0.1  2023-01-18

 * Docfix: Fixed a missing return in an example in the docs.

v2.0.0  2023-01-18

 * Major rewrite of the original. This is the first official, tagged release.
 * Provides support for parsing via `message.Parse()`
 * Containers for messages are `message.Opaque` for any kind of message or message part and `message.Multipart` for MIME multipart messages and parts.
 * Provide `message.Buffer` for creating new messages.
 * The header package provides high- and mid-level access to email message headers.
 * The field package provides low-level access to email message header fields.
 * The transfer package provides transfer encoding support.
 * The encoding package enables full charset support (at the cost of a much larger compiled binary).
 * The param package provides support for parameterized header field bodies as are used for `Content-type` and `Content-disposition` headers.
 * The walker package provides experimental support for iterating through message parts.

v1.0.0  2022-12-05

 * Unofficial tagged release.
