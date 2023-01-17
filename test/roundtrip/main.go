package main

import (
	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/test/roundtrip/cmd"
)

func main() {
	err := cmd.Execute()
	cobra.CheckErr(err)
}
