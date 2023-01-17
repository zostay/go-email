package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/message"
)

var oneCmd = &cobra.Command{
	Use:   "one message",
	Short: "Shows the diff of a single message round-trip",
	Run:   RunOne,
}

func init() {
	rootCmd.AddCommand(oneCmd)
}

func RunOne(cmd *cobra.Command, args []string) {
	path := args[0]
	msgFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() { _ = msgFile.Close() }()

	m, err := message.Parse(msgFile, message.WithUnlimitedRecursion())
	if err != nil {
		panic(err)
	}

	rtFile, err := os.CreateTemp(os.TempDir(), "rtmsg-")
	if err != nil {
		panic(err)
	}

	_, err = m.WriteTo(rtFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("path = %s\n", path)
	fmt.Printf("tmp  = %s\n", rtFile.Name())

	diff := exec.Command("/usr/bin/diff", "-u", path, rtFile.Name())
	diff.Stdout = os.Stdout
	_ = diff.Run()
}
