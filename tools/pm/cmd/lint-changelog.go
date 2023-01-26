package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/tools/pm/changes"
)

var (
	lintChangelogCmd = &cobra.Command{
		Use:   "lint-changelog",
		Short: "Check the changelog file for problems",
		Args:  cobra.NoArgs,
		Run:   LintChangelog,
	}

	isRelease bool
)

func init() {
	lintChangelogCmd.Flags().BoolVarP(&isRelease, "release", "r", false, "verify the changelog is ready for release")
}

func LintChangelog(_ *cobra.Command, _ []string) {
	changelog, err := os.Open("Changes")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to open Changes file: %v", err)
		os.Exit(1)
	}

	linter := changes.NewLinter(changelog, isRelease)
	err = linter.Check()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
