package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Git struct {
	repoPath string
}

func New(repoPath string) *Git {
	return &Git{repoPath: repoPath}
}

func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().Strs("args", args).Str("repo", g.repoPath).Msg("running git command")

	if err := cmd.Run(); err != nil {
		return "", pkgerrors.Wrapf(err, "git %s failed: %s", strings.Join(args, " "), stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CurrentBranch returns the name of the current branch.
func (g *Git) CurrentBranch() (string, error) {
	return g.run("rev-parse", "--abbrev-ref", "HEAD")
}

// CreateBranch creates and checks out a new branch.
func (g *Git) CreateBranch(branchName string) error {
	_, err := g.run("checkout", "-b", branchName)
	return err
}

// Checkout switches to an existing branch.
func (g *Git) Checkout(branchName string) error {
	_, err := g.run("checkout", branchName)
	return err
}

// BranchExists checks if a branch exists.
func (g *Git) BranchExists(branchName string) bool {
	_, err := g.run("rev-parse", "--verify", branchName)
	return err == nil
}

// DeleteBranch deletes a local branch.
func (g *Git) DeleteBranch(branchName string) error {
	_, err := g.run("branch", "-D", branchName)
	return err
}

// Fetch fetches from remote.
func (g *Git) Fetch() error {
	_, err := g.run("fetch", "origin")
	return err
}

// Pull pulls the current branch from remote.
func (g *Git) Pull() error {
	_, err := g.run("pull")
	return err
}

// HasChanges returns true if there are uncommitted changes.
func (g *Git) HasChanges() (bool, error) {
	status, err := g.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return status != "", nil
}

// AddAll stages all changes.
func (g *Git) AddAll() error {
	_, err := g.run("add", "-A")
	return err
}

// Commit creates a commit with the given message.
func (g *Git) Commit(message string) error {
	_, err := g.run("commit", "-m", message)
	return err
}

// Push pushes the current branch to origin.
func (g *Git) Push() error {
	branch, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	_, err = g.run("push", "-u", "origin", branch)
	return err
}

// GetRemoteURL returns the remote URL for origin.
func (g *Git) GetRemoteURL() (string, error) {
	return g.run("remote", "get-url", "origin")
}

// EnsureClean returns an error if the working directory is not clean.
func (g *Git) EnsureClean() error {
	hasChanges, err := g.HasChanges()
	if err != nil {
		return err
	}
	if hasChanges {
		return fmt.Errorf("working directory has uncommitted changes")
	}
	return nil
}

// RebaseOnto rebases current branch onto the given base.
func (g *Git) RebaseOnto(base string) error {
	_, err := g.run("rebase", base)
	return err
}
