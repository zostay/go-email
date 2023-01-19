// Package message is the heart of this library. It provides objects for
// flexibly parsing and reading email messages (that survive even when the input
// is not strictly correct) and for generating new messages that are strictly
// correct. You can pair the parsing/reading tools with the generating tools to
// perform advanced email transformations.
//
// You can deal with any message as an Opaque message object. You can create
// these from existing messages by calling the ParseOpaque() function. You can
// generate these objects by using a Buffer and then calling the Opaque()
// method.
//
// If you want to be able to work with individual parts of a multipart method,
// you can use the Multipart message object instead. This is achieved for
// existing messages by parsing an Opaque message gotten via ParseOpaque() via
// the Parse() function:
//
//	opMsg, err := message.ParseOpaque(in)
//	if err != nil {
//	  panic(err)
//	}
//
//	msg, err := message.Parse(opMsg)
//	if err != nil {
//	  panic(err)
//	}
//
// Or you can generate these messages by using a Buffer to add parts to a
// message and then call the Multipart() method.
package message
