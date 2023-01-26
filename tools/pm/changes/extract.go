package changes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// ExtractSection will return a reader that will output the bullets that are
// written below the changelog heading for the given version. The first argument
// is the file name of the changelog to open and the second line is a semver
// string to search for. Returns an error if there is a problem reading the file
// or if the version heading is not found.
func ExtractSection(fn string, vstring string) (io.Reader, error) {
	r, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	var (
		vprefix = vstring + "  "
		sc      = bufio.NewScanner(r)
		started = false
		blanks  = 0
		buf     = &bytes.Buffer{}
	)

	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, vprefix) {
			started = true
			continue
		}

		if !started {
			continue
		}

		if line == "" {
			blanks++
			continue
		}

		if blanks >= 2 {
			break
		}

		buf.WriteString(line)
		buf.WriteRune('\n')
	}

	if !started {
		return nil, fmt.Errorf("a change log section for version %s was not found", vstring)
	}

	return buf, nil
}
