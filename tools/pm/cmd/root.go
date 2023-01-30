package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "pm",
		Short: "Golang project management tools by zostay",
	}
)

func init() {
	rootCmd.AddCommand(changelogCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(templateFileCmd)
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
