package mime_test

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

func Example() {
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
