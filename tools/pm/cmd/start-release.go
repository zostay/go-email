package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/tools/pm/release"
)

var (
	startReleaseCmd = &cobra.Command{
		Use:   "start-release <version>",
		Short: "Start a release",
		Args:  cobra.ExactArgs(1),
		RunE:  StartRelease,
	}
)

func StartRelease(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	process, err := release.NewProcess(ctx, args[0])
	if err != nil {
		return err
	}

	process.SetupGitRepo()
	process.CheckGitCleanliness()
	process.LintChangelog()
	process.MakeReleaseBranch()
	process.FixupChangelog()
	process.AddAndCommit()
	process.PushReleaseBranch()
	process.CreateGithubPullRequest(ctx)

	return nil
}
