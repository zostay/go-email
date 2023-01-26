package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/go-email/v2/tools/pm/release"
)

var (
	releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "commands related to software releases",
	}

	targetBranch string
)

func init() {
	releaseCmd.AddCommand(startReleaseCmd)
	releaseCmd.AddCommand(finishReleaseCmd)

	releaseCmd.PersistentFlags().StringVar(&targetBranch, "target-branch", "master", "the branch to merge into during release")
}

func MakeReleaseConfig() *release.Config {
	cfg := release.GoEmailConfig
	cfg.TargetBranch = targetBranch
	return &cfg
}
