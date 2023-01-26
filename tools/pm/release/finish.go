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

func (p *Process) CheckReadyForMerge(ctx context.Context) {
	bp, _, err := p.gh.Repositories.GetBranchProtection(ctx, p.Owner, p.Project, "master")
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
			p.Chokef("cannot merge release branch because it failed check %q", k)
		}
	}
}

func (p *Process) MergePullRequest(ctx context.Context) {
	prs, _, err := p.gh.PullRequests.List(ctx, p.Owner, p.Project, &github.PullRequestListOptions{})
	if err != nil {
		p.Chokef("unable to list pull requests: %v", err)
	}

	prId := 0
	for _, pr := range prs {
		if pr.Head.GetLabel() == p.Branch {
			prId = pr.GetNumber()
			break
		}
	}

	m, _, err := p.gh.PullRequests.Merge(ctx, p.Owner, p.Project, prId, "Merging release branch.", &github.PullRequestOptions{})
	if err != nil {
		p.Chokef("unable to merge pull request %d: %v", prId, err)
	}

	if !m.GetMerged() {
		p.Chokef("failed to merge pull request %d", prId)
	}
}

func (p *Process) TagRelease() {
	err := p.wc.Checkout(&git.CheckoutOptions{
		Branch: "master",
	})
	if err != nil {
		p.Chokef("unable to switch to master branch: %v", err)
	}

	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to get HEAD ref of master branch: %v", err)
	}

	head := headRef.Hash()
	_, err = p.repo.CreateTag(p.Tag, head, &git.CreateTagOptions{})
	if err != nil {
		p.Chokef("unable to tag release %s: %v", p.Tag, err)
	}

	p.ForCleanup(func() { _ = p.repo.DeleteTag(p.Tag) })

	tagRef := config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", p.Tag, p.Tag))
	err = p.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{tagRef},
	})
	if err != nil {
		p.Chokef("unable to push tags to origin: %v", err)
	}

	p.ForCleanup(func() {
		_ = p.remote.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs:   []config.RefSpec{tagRef},
			Prune:      true,
		})
	})
}

func (p *Process) CreateRelease(ctx context.Context) {
	cr, err := changes.ExtractSection(p.Changelog, p.Version.String())
	if err != nil {
		p.Chokef("unable to get log of changes: %v", err)
	}

	chgs, err := io.ReadAll(cr)
	if err != nil {
		p.Chokef("unable to read log of changes: %v", err)
	}

	chgStr := string(chgs)
	releaseName := fmt.Sprintf("Release v%s", p.Version)

	_, _, err = p.gh.Repositories.CreateRelease(ctx, p.Owner, p.Project, &github.RepositoryRelease{
		TagName:              github.String(p.Tag),
		Name:                 github.String(releaseName),
		Body:                 github.String(chgStr),
		Draft:                github.Bool(false),
		Prerelease:           github.Bool(false),
		GenerateReleaseNotes: github.Bool(false),
		MakeLatest:           github.String("true"),
	})

	if err != nil {
		p.Chokef("failed to create release %q: %v", releaseName, err)
	}
}
