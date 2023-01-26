package release

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
)

type Process struct {
	Config

	gh     *github.Client
	repo   *git.Repository
	remote *git.Remote
	wc     *git.Worktree

	cleanupActions []func()

	addFiles []string
}

func (p *Process) ToAdd(fn string) {
	if p.addFiles == nil {
		p.addFiles = []string{}
	}
	p.addFiles = append(p.addFiles, fn)
}

func (p *Process) ForCleanup(action func()) {
	if p.cleanupActions == nil {
		p.cleanupActions = make([]func(), 0, 10)
	}
	p.cleanupActions = append(p.cleanupActions, action)
}

func (p *Process) Cleanup() {
	for i := len(p.cleanupActions) - 1; i >= 0; i-- {
		action := p.cleanupActions[i]
		action()
	}
}

func (p *Process) Choke(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "Failed: %s\n", msg)
	_, _ = fmt.Fprintln(os.Stderr, "Cancelling release.")
	p.Cleanup()
	os.Exit(1)
}

func (p *Process) Chokef(f string, args ...interface{}) {
	p.Choke(fmt.Sprintf(f, args...))
}

func initializeProcess(
	ctx context.Context,
	cfg *Config,
) (*Process, error) {
	p := &Process{
		Config: *cfg,
	}

	err := p.setupGithubClient(ctx)
	if err != nil {
		return nil, err
	}
	p.SetupGitRepo()

	return p, nil
}

func (p *Process) completeInitialization(v string) error {
	var err error
	p.Version, err = semver.NewVersion(v)
	if err != nil {
		return err
	}
	p.Branch = "release-v" + p.Version.String()
	p.Tag = "v" + p.Version.String()
	p.Today = time.Now().Format("2006-01-02")

	return nil
}

func NewProcess(ctx context.Context, v string, cfg *Config) (*Process, error) {
	p, err := initializeProcess(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = p.completeInitialization(v)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func NewProcessContinuation(ctx context.Context, cfg *Config) (*Process, error) {
	p, err := initializeProcess(ctx, cfg)
	if err != nil {
		return nil, err
	}

	headRef, err := p.repo.Head()
	if err != nil {
		p.Chokef("unable to find HEAD: %v", err)
	}

	const releasePrefix = "refs/heads/release-v"
	if !strings.HasPrefix(string(headRef.Name()), releasePrefix) {
		p.Choke("you must be on the release branch to finish the process")
	}

	v := string(headRef.Name()[len(releasePrefix):])
	err = p.completeInitialization(v)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Process) setupGithubClient(ctx context.Context) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is missing")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	p.gh = github.NewClient(tc)

	return nil
}

func (p *Process) SetupGitRepo() {
	l, err := git.PlainOpen(".")
	if err != nil {
		p.Chokef("unable to open git repository at .: %v", err)
	}

	p.repo = l

	r, err := p.repo.Remote("origin")
	if err != nil {
		p.Chokef("unable to connect to remote origin: %v", err)
	}

	p.remote = r

	w, err := p.repo.Worktree()
	if err != nil {
		p.Chokef("unable to examine the working copy: %v", err)
	}

	p.wc = w
}
