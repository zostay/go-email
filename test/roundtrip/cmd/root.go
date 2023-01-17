package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "roundtrip",
	Short: "Tools for testing message round-tripping",
}

func Execute() error {
	return rootCmd.Execute()
}
