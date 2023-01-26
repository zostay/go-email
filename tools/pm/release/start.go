package release

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v49/github"

	"github.com/zostay/go-email/v2/tools/pm/changes"
)

var ignoreStatus = map[string]struct{}{
	".session.vim": {},
}

// IsDirty returns true if we consider the tree dirty. We do not consider
// Untracked to dirty the directory and we also ignore some filenames that are
// in the global .gitignore and not in the local .gitignore.
func IsDirty(status git.Status) bool {
	for fn, fstat := range status {
		if _, ignorable := ignoreStatus[fn]; ignorable {
			continue
		}

		if fstat.Worktree != git.Unmodified && fstat.Worktree != git.Untracked {
			return true
		}

		if fstat.Staging != git.Unmodified && fstat.Staging != git.Untracked {
			return true
		}
	}
	return false
}

// CheckGitCleanliness ensures that the current git repository is clean and that
// we are on the correct branch from which to trigger a release.
func (p *Process) CheckGitCleanliness() {
	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to find HEAD: %v", err)
	}

	if headRef.Name() != p.TargetBranchRefName() {
		p.Chokef("you must checkout %s to release", p.TargetBranch)
	}

	remoteRefs, err := p.remote.List(&git.ListOptions{})
	if err != nil {
		p.Chokef("unable to list remote git references: %v", err)
	}

	var masterRef *plumbing.Reference
	for _, ref := range remoteRefs {
		if ref.Name() == p.TargetBranchRefName() {
			masterRef = ref
			break
		}
	}

	if headRef.Hash() != masterRef.Hash() {
		p.Choke("local copy differs from remote, you need to push or pull")
	}

	stat, err := p.wc.Status()
	if err != nil {
		p.Chokef("unable to check working copy status: %v", err)
	}

	if IsDirty(stat) {
		p.Choke("your working copy is dirty")
	}
}

// LintChangelog performs a check to ensure the changelog is ready for release.
func (p *Process) LintChangelog(mode changes.CheckMode) {
	changelog, err := os.Open(p.Changelog)
	if err != nil {
		p.Chokef("unable to open Changes file: %v", err)
	}

	linter := changes.NewLinter(changelog, mode)
	err = linter.Check()
	if err != nil {
		p.Chokef("%v", err)
	}
}

// MakeReleaseBranch creates the branch that will be used to manage the release.
func (p *Process) MakeReleaseBranch() {
	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to retrieve the HEAD ref: %v", err)
	}

	err = p.wc.Checkout(&git.CheckoutOptions{
		Hash:   headRef.Hash(),
		Branch: p.BranchRefName(),
		Create: true,
	})
	if err != nil {
		p.Chokef("unable to checkout branch %s: %v", p.Branch, err)
	}

	p.ForCleanup(func() { _ = p.repo.Storer.RemoveReference(p.BranchRefName()) })
	p.ForCleanup(func() {
		_ = p.wc.Checkout(&git.CheckoutOptions{
			Branch: p.TargetBranchRefName(),
		})
	})
}

// FixupChangelog alters the changelog to prepare it for release.
func (p *Process) FixupChangelog() {
	r, err := os.Open(p.Changelog)
	if err != nil {
		p.Chokef("unable to open %s: %v", p.Changelog, err)
	}

	newChangelog := p.Changelog + ".new"

	w, err := os.Create(newChangelog)
	if err != nil {
		p.Chokef("unable to create %s: %v", newChangelog, err)
	}

	p.ForCleanup(func() { _ = os.Remove(newChangelog) })

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if line == "WIP" || line == "WIP  TBD" {
			_, _ = fmt.Fprintf(w, "v%s  %s\n", p.Version, p.Today)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	_ = r.Close()
	err = w.Close()
	if err != nil {
		p.Chokef("unable to close %s: %v", newChangelog, err)
	}

	err = os.Rename(newChangelog, p.Changelog)
	if err != nil {
		p.Chokef("unable to overwrite %s with %s: %v", p.Changelog, newChangelog, err)
	}

	p.ToAdd(p.Changelog)
}

// AddAndCommit adds changes made as part of the release process to the release
// branch.
func (p *Process) AddAndCommit() {
	for _, fn := range p.addFiles {
		_, err := p.wc.Add(fn)
		if err != nil {
			p.Chokef("error adding file %s to git: %v", fn, err)
		}
	}

	_, err := p.wc.Commit("releng: v"+p.Version.String(), &git.CommitOptions{})
	if err != nil {
		p.Chokef("error committing changes to git: %v", err)
	}
}

// PushReleaseBranch pushes the release branch to github for release testing.
func (p *Process) PushReleaseBranch() {
	err := p.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{p.BranchRefSpec()},
	})
	if err != nil {
		p.Chokef("error pushing changes to github: %v", err)
	}

	p.ForCleanup(func() {
		_ = p.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{p.BranchRefSpec()},
			Prune:      true,
		})
	})
}

// CreateGithubPullRequest creates the PR on github for monitoring the test
// results for release testing. This will also be used to merge the release
// branch when testing passes.
func (p *Process) CreateGithubPullRequest(ctx context.Context) {
	_, _, err := p.gh.PullRequests.Create(ctx, p.Owner, p.Project, &github.NewPullRequest{
		Title: github.String("Release v" + p.Version.String()),
		Head:  github.String(p.Branch),
		Base:  github.String(p.TargetBranch),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", p.Version)),
	})
	if err != nil {
		p.Chokef("unable to create pull request: %v", err)
	}
}
