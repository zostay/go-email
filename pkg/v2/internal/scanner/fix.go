package scanner

import "bufio"

// This code has no business being in this module, but I don't have this utility
// anywhere else at the moment.
//
// It is my opinion, that the built-in bufio.Scanner has an annoying flaw and
// inconsistency in the design of SplitFunc. The signature of SplitFunc is:
//
//  type SplitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)
//
// This SplitFunc will cause the scanner to exit under any the following
// circumstances:
//
// * When err != nil
// * When atEOF && token == nil
//
// However, this makes the scanner more complex than it needs to be. Really,
// the scanner ought to quit under these circumstances instead:
//
// * When err != nil
// * When atEOF && len(data) - advance == 0
//
// This makes much more sense because is the signal that tells the caller how
// much data has been consumed. However, each chunk of consumed data may or may
// not be valuable, where "valuable" means we return it as a token. If the next
// chunk, however, is not valuable we can't just return nil. We have to have
// an inner loop that continues seeking for a valuable token to return because
// if we fail to return one after atEOF, we terminate the scan. However, the
// scanner already provides a loop. Why do we need an inner loop? We should
// just reuse the loop that is already provided by the scanner. Wasteful and
// it complicates the code inside the bufio.SplitFunc. This fixes it.
//
// I'm sure there's some justification for this choice, but I don't like it.

func MakeSplitFuncExitByAdvance(split bufio.SplitFunc) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		totalAdvance := 0
		for {
			// run the split function
			advance, token, err := split(data, atEOF)

			// if either a token is return or my preferred termination criteria
			// are met, return immediately
			//
			// Note #1: We do not check for atEOF here with len(data)-advance ==
			// 0 because if we advance to len(data)-advance == 0, there is no
			// more data to consume in any case. Either atEOF and this is the
			// termination criteria OR !atEOF and we don't want to violate the
			// part of the contract in the bufio documentation that says that
			// len(data) == 0 may happen if and only if atEOF.
			//
			// Note #2: We go with len(data)-advance <= 0 rather than
			// len(data)-advance == 0 because that's an error condition we want
			// to pass up the chain for the caller to handle.
			//
			// Note #3: Here's another problem with this inner loop thing the
			// bufio.SplitFunc causes: we have to make sure to accumulate all
			// our inner advances and return that or we'll advance by the wrong
			// amount in the outer loop. Silly.
			if token != nil || len(data)-advance <= 0 || err != nil {
				return totalAdvance + advance, token, err
			}

			// otherwise, advance and try for another token
			data = data[advance:]
			totalAdvance += advance
		}
	}
}
