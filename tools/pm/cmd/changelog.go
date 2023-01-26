package cmd

import "github.com/spf13/cobra"

var (
	changelogCmd = &cobra.Command{
		Use:   "changelog",
		Short: "Commands related to change logs",
	}
)

func init() {
	changelogCmd.AddCommand(lintChangelogCmd)
	changelogCmd.AddCommand(extractChangelogCmd)
}
