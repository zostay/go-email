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

func (p *Process) CheckGitCleanliness() {
	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to find HEAD: %v", err)
	}

	if headRef.Name() != "refs/heads/master" {
		p.Choke("you must checkout master to release")
	}

	remoteRefs, err := p.remote.List(&git.ListOptions{})
	if err != nil {
		p.Chokef("unable to list remote git references: %v", err)
	}

	var masterRef *plumbing.Reference
	for _, ref := range remoteRefs {
		if ref.Name() == "refs/heads/master" {
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

	if !stat.IsClean() {
		p.Choke("your working copy is dirty")
	}
}

func (p *Process) LintChangelog() {
	changelog, err := os.Open(p.Changelog)
	if err != nil {
		p.Chokef("unable to open Changes file: %v", err)
	}

	linter := changes.NewLinter(changelog, true)
	err = linter.Check()
	if err != nil {
		p.Chokef("%v", err)
	}
}

func (p *Process) MakeReleaseBranch() {
	err := p.repo.CreateBranch(&config.Branch{
		Name:   p.Branch,
		Remote: "origin",
		Merge:  plumbing.ReferenceName("refs/head/" + p.Branch),
	})
	if err != nil {
		p.Chokef("unable to create release branch %s: %v", p.Branch, err)
	}

	p.ForCleanup(func() { _ = p.repo.DeleteBranch(p.Branch) })
}

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
			_, _ = fmt.Fprintf(w, "%s  %s\n", p.Version, p.Today)
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

func (p *Process) PushReleaseBranch() {
	err := p.repo.Push(&git.PushOptions{})
	if err != nil {
		p.Chokef("error pushing changes to github: %v", err)
	}
}

func (p *Process) CreateGithubPullRequest(ctx context.Context) {
	_, _, err := p.gh.PullRequests.Create(ctx, p.Owner, p.Project, &github.NewPullRequest{
		Title: github.String("Release v" + p.Version.String()),
		Head:  github.String(p.Branch),
		Base:  github.String("master"),
		Body:  github.String(fmt.Sprintf("Pull request to release v%s of go-email.", p.Version)),
	})
	if err != nil {
		p.Chokef("unable to create pull request: %v", err)
	}
}
