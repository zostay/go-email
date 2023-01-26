package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "pm",
	Short: "Golang project management tools by zostay",
}

func init() {
	rootCmd.AddCommand(lintChangelogCmd)
	rootCmd.AddCommand(startReleaseCmd)
	rootCmd.AddCommand(finishReleaseCmd)
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
