package github

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type GitHub struct {
	repoPath string
}

func New(repoPath string) *GitHub {
	return &GitHub{repoPath: repoPath}
}

// CreatePR creates a pull request using the gh CLI.
// Returns the PR URL.
func (g *GitHub) CreatePR(title, body, baseBranch string) (string, error) {
	args := []string{
		"pr", "create",
		"--title", title,
		"--body", body,
		"--base", baseBranch,
		"--draft",
	}

	cmd := exec.Command("gh", args...)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info().Str("title", title).Str("base", baseBranch).Msg("creating PR via gh CLI")

	if err := cmd.Run(); err != nil {
		return "", pkgerrors.Wrapf(err, "gh pr create failed: %s", stderr.String())
	}

	prURL := strings.TrimSpace(stdout.String())
	return prURL, nil
}

// FormatPRBody creates a PR body with ticket reference and summary.
func FormatPRBody(ticketKey, ticketSummary, jiraHost string) string {
	var sb strings.Builder

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("Implements [%s](%s/browse/%s): %s\n\n", ticketKey, jiraHost, ticketKey, ticketSummary))
	sb.WriteString("## Changes\n\n")
	sb.WriteString("_Changes implemented by Claude Code based on Jira ticket._\n\n")
	sb.WriteString("## Test Plan\n\n")
	sb.WriteString("- [ ] Review changes\n")
	sb.WriteString("- [ ] Run tests\n")
	sb.WriteString("- [ ] Manual verification\n")

	return sb.String()
}
