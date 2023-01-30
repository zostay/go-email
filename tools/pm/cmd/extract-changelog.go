package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/tools/pm/changes"
	"github.com/zostay/go-email/v2/tools/pm/release"
)

var (
	extractChangelogCmd = &cobra.Command{
		Use:   "extract <version>",
		Short: "extract the bullets for the changelog section for the given version",
		Args:  cobra.ExactArgs(1),
		Run:   ExtractChangelog,
	}
)

func ExtractChangelog(_ *cobra.Command, args []string) {
	r, err := changes.ExtractSection(release.GoEmailConfig.Changelog, args[0])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read changelog section: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read changelog data: %v\n", err)
		os.Exit(1)
	}
}
