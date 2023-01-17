package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/message"
)

// This is a special test I use to see how this library performs at
// round-tripping. Given the name of a folder, it will read all files in that
// fold and all sub-folders. Each file will be completely parsed as a message.
// Then, the original bytes will be compared to the bytes that are generated by
// calling WriteTo() on the message into a new buffer. It will record the number
// of differences between input and the proposed output and write that to the
// output file.

var allCmd = &cobra.Command{
	Use:   "all file.log maildir",
	Short: "Creates a report showing round-tripping for all files in a maildir",
	Args:  cobra.ExactArgs(2),
	Run:   CheckAll,
}

var (
	opaqueParse bool
)

func init() {
	allCmd.Flags().BoolVarP(&opaqueParse, "opaque", "o", false, "check round-trip of opaque parsing only")

	rootCmd.AddCommand(allCmd)
}

func CheckAll(cmd *cobra.Command, args []string) {
	outputFile, mailDir := args[0], args[1]
	of, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	logEntry := func(path, msg, detail string) {
		_, _ = fmt.Fprintf(of, "%s: %s %s\n", path, msg, detail)
	}

	differ := diffmatchpatch.New()

	counter := 0
	throb := func() {
		if counter > 0 {
			if counter%100 == 0 {
				fmt.Print(".")
			}

			if counter%7800 == 0 {
				fmt.Print("\n")
			}
		}

		switch counter % 4 {
		case 0:
			fmt.Print(`-`)
		case 1:
			fmt.Print(`\`)
		case 2:
			fmt.Print(`|`)
		case 3:
			fmt.Print(`/`)
		}

		fmt.Print("\b")
		counter++
	}

	mailDirWalker := func(
		path string,
		d fs.DirEntry,
		err error,
	) error {
		if d.IsDir() {
			return nil
		}

		if err != nil {
			logEntry(path, "WalkErr", err.Error())
			return nil
		}

		msgFile, err := os.Open(path)
		if err != nil {
			logEntry(path, "FailOpen", err.Error())
			return err
		}
		defer func() { _ = msgFile.Close() }()

		opt := message.WithUnlimitedRecursion()
		if opaqueParse {
			opt = message.WithoutMultipart()
		}

		m, err := message.Parse(msgFile, opt)
		if err != nil {
			logEntry(path, "FailParse", err.Error())
			return nil
		}

		buf := &bytes.Buffer{}
		_, err = m.WriteTo(buf)
		if err != nil {
			logEntry(path, "FailWrite", err.Error())
			return nil
		}

		_, err = msgFile.Seek(0, 0)
		if err != nil {
			logEntry(path, "FailSeek", err.Error())
			return nil
		}

		msgText, err := io.ReadAll(msgFile)
		if err != nil {
			logEntry(path, "FailReadAll", err.Error())
			return nil
		}

		diffs := differ.DiffMain(buf.String(), string(msgText), true)
		plus, minus := 0, 0
		for _, diff := range diffs {
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				plus += strings.Count(diff.Text, "\n")
			case diffmatchpatch.DiffDelete:
				minus += strings.Count(diff.Text, "\n")
			case diffmatchpatch.DiffEqual:
				// nop
			}
		}

		detail := fmt.Sprintf("+%d/-%d", plus, minus)
		logEntry(path, "Diff", detail)

		throb()

		return nil
	}

	err = filepath.WalkDir(mailDir, mailDirWalker)
	if err != nil {
		panic(err)
	}
}
