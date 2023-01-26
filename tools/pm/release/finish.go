package release

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/google/go-github/v49/github"

	"github.com/zostay/go-email/v2/tools/pm/changes"
)

// CaptureChangesInfo loads the bullets for the changelog section relevant to
// this release into the process configuration for use when creating the release
// later.
func (p *Process) CaptureChangesInfo() {
	vstring := "v" + p.Version.String()
	cr, err := changes.ExtractSection(p.Changelog, vstring)
	if err != nil {
		p.Chokef("unable to get log of changes: %v", err)
	}

	chgs, err := io.ReadAll(cr)
	if err != nil {
		p.Chokef("unable to read log of changes: %v", err)
	}

	p.ChangesInfo = string(chgs)
}

// CheckReadyForMerge ensures that all the required tests are passing.
func (p *Process) CheckReadyForMerge(ctx context.Context) {
	bp, _, err := p.gh.Repositories.GetBranchProtection(ctx, p.Owner, p.Project, p.TargetBranch)
	if err != nil {
		p.Chokef("unable to get branches %s: %v", p.Branch, err)
	}

	checks := bp.GetRequiredStatusChecks().Checks
	passage := make(map[string]bool, len(checks))
	for _, check := range checks {
		passage[check.Context] = false
	}

	crs, _, err := p.gh.Checks.ListCheckRunsForRef(ctx, p.Owner, p.Project, p.Branch, &github.ListCheckRunsOptions{})
	if err != nil {
		p.Chokef("unable to list check runs for branch %s: %v", p.Branch, err)
	}

	for _, run := range crs.CheckRuns {
		passage[run.GetName()] =
			run.GetStatus() == "completed" &&
				run.GetConclusion() == "success"
	}

	for k, v := range passage {
		if !v {
			p.Chokef("cannot merge release branch because it has not passed check %q", k)
		}
	}
}

// MergePullRequest merges the PR into master.
func (p *Process) MergePullRequest(ctx context.Context) {
	prs, _, err := p.gh.PullRequests.List(ctx, p.Owner, p.Project, &github.PullRequestListOptions{})
	if err != nil {
		p.Chokef("unable to list pull requests: %v", err)
	}

	prId := 0
	for _, pr := range prs {
		if pr.Head.GetRef() == p.Branch {
			prId = pr.GetNumber()
			break
		}
	}

	if prId == 0 {
		p.Chokef("cannot find pull request for branch %s", p.Branch)
	}

	m, _, err := p.gh.PullRequests.Merge(ctx, p.Owner, p.Project, prId, "Merging release branch.", &github.PullRequestOptions{})
	if err != nil {
		p.Chokef("unable to merge pull request %d: %v", prId, err)
	}

	if !m.GetMerged() {
		p.Chokef("failed to merge pull request %d", prId)
	}
}

// TagRelease creates and pushes a tag for the newly merged release on master.
func (p *Process) TagRelease() {
	err := p.wc.Checkout(&git.CheckoutOptions{
		Branch: p.TargetBranchRefName(),
	})
	if err != nil {
		p.Chokef("unable to switch to %s branch: %v", p.TargetBranch, err)
	}

	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to get HEAD ref of %s branch: %v", p.TargetBranch, err)
	}

	head := headRef.Hash()
	_, err = p.repo.CreateTag(p.Tag, head, &git.CreateTagOptions{
		Message: fmt.Sprintf("Release tag for v%s", p.Version.String()),
	})
	if err != nil {
		p.Chokef("unable to tag release %s: %v", p.Tag, err)
	}

	p.ForCleanup(func() { _ = p.repo.DeleteTag(p.Tag) })

	err = p.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{p.TagRefSpec()},
	})
	if err != nil {
		p.Chokef("unable to push tags to origin: %v", err)
	}

	p.ForCleanup(func() {
		_ = p.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{p.TagRefSpec()},
			Prune:      true,
		})
	})
}

// CreateRelease creates a release on github for the release.
func (p *Process) CreateRelease(ctx context.Context) {
	releaseName := fmt.Sprintf("Release v%s", p.Version)
	_, _, err := p.gh.Repositories.CreateRelease(ctx, p.Owner, p.Project, &github.RepositoryRelease{
		TagName:              github.String(p.Tag),
		Name:                 github.String(releaseName),
		Body:                 github.String(p.ChangesInfo),
		Draft:                github.Bool(false),
		Prerelease:           github.Bool(false),
		GenerateReleaseNotes: github.Bool(false),
		MakeLatest:           github.String("true"),
	})

	if err != nil {
		p.Chokef("failed to create release %q: %v", releaseName, err)
	}
}
