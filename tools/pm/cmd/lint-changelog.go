package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/tools/pm/changes"
	"github.com/zostay/go-email/v2/tools/pm/release"
)

var (
	lintChangelogCmd = &cobra.Command{
		Use:   "lint",
		Short: "Check the changelog file for problems",
		Args:  cobra.NoArgs,
		Run:   LintChangelog,
	}

	isRelease    bool
	isPreRelease bool
)

func init() {
	lintChangelogCmd.Flags().BoolVarP(&isRelease, "release", "r", false, "verify that there's no WIP section")
	lintChangelogCmd.Flags().BoolVarP(&isPreRelease, "pre-release", "p", false, "verify that the WIP section looks good")
}

func LintChangelog(_ *cobra.Command, _ []string) {
	changelog, err := os.Open(release.GoEmailConfig.Changelog)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to open Changes file: %v", err)
		os.Exit(1)
	}

	var mode changes.CheckMode
	switch {
	case isRelease:
		mode = changes.CheckRelease
	case isPreRelease:
		mode = changes.CheckPreRelease
	default:
		mode = changes.CheckStandard
	}

	linter := changes.NewLinter(changelog, mode)
	err = linter.Check()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
