package release

import "github.com/coreos/go-semver/semver"

type Config struct {
	// Version is the semantic version of the release being processed.
	Version *semver.Version

	// Branch is the name of the release branch.
	Branch string

	// Tag is the name of the final release tag.
	Tag string

	// Today is the date YYYY-MM-DD date of the release.
	Today string

	// Changelog is the name of the file holding the change log.
	Changelog string

	// Owner is the name of the owner of the project on github.
	Owner string

	// Project is the name of the repository on github.
	Project string

	// TargetBranch is the branch we are merging into (usually master).
	TargetBranch string
}

var GoEmailConfig = Config{
	Changelog: "Changes",
	Owner:     "zostay",
	Project:   "go-email",

	TargetBranch: "master",
}
