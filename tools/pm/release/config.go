package release

import (
	"path"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

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

	// ChangesInfo is the bullets in the change log to put into the release
	// body.
	ChangesInfo string
}

var GoEmailConfig = Config{
	Changelog: "Changes.md",
	Owner:     "zostay",
	Project:   "go-email",

	TargetBranch: "master",
}

func ref(t, n string) string {
	return path.Join("refs", t, n)
}

func refSpec(r string) config.RefSpec {
	return config.RefSpec(strings.Join([]string{r, r}, ":"))
}

func (c *Config) BranchRef() string {
	return ref("heads", c.Branch)
}

func (c *Config) BranchRefName() plumbing.ReferenceName {
	return plumbing.ReferenceName(c.BranchRef())
}

func (c *Config) BranchRefSpec() config.RefSpec {
	return refSpec(c.BranchRef())
}

func (c *Config) TargetBranchRef() string {
	return ref("heads", c.TargetBranch)
}

func (c *Config) TargetBranchRefName() plumbing.ReferenceName {
	return plumbing.ReferenceName(c.TargetBranchRef())
}

func (c *Config) TagRef() string {
	return ref("tags", c.Tag)
}

func (c *Config) TagRefSpec() config.RefSpec {
	return refSpec(c.TagRef())
}
