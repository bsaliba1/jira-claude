package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ReviewComment represents a single review comment on a PR.
type ReviewComment struct {
	ID       int64  `json:"id"`
	Author   string `json:"user.login"`
	Body     string `json:"body"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	DiffHunk string `json:"diff_hunk"`
	URL      string `json:"html_url"`
}

// reviewCommentJSON matches the GitHub API response structure.
type reviewCommentJSON struct {
	ID       int64  `json:"id"`
	Body     string `json:"body"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	DiffHunk string `json:"diff_hunk"`
	HTMLURL  string `json:"html_url"`
	User     struct {
		Login string `json:"login"`
	} `json:"user"`
}

// PRComments contains all review comments for a PR.
type PRComments struct {
	PRNumber int
	PRTitle  string
	PRURL    string
	Comments []ReviewComment
}

// prViewJSON matches the gh pr view --json output.
type prViewJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// GetPRForBranch detects the PR number for the current branch.
func (g *GitHub) GetPRForBranch() (int, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "number")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().Msg("detecting PR for current branch")

	if err := cmd.Run(); err != nil {
		return 0, pkgerrors.Wrapf(err, "failed to detect PR for branch: %s", stderr.String())
	}

	var result prViewJSON
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return 0, pkgerrors.Wrap(err, "failed to parse gh pr view output")
	}

	return result.Number, nil
}

// GetPRDetails fetches PR title and URL.
func (g *GitHub) GetPRDetails(prNumber int) (title, url string, err error) {
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", prNumber), "--json", "number,title,url")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().Int("pr", prNumber).Msg("fetching PR details")

	if err := cmd.Run(); err != nil {
		return "", "", pkgerrors.Wrapf(err, "failed to get PR details: %s", stderr.String())
	}

	var result prViewJSON
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", "", pkgerrors.Wrap(err, "failed to parse PR details")
	}

	return result.Title, result.URL, nil
}

// GetPRComments fetches all review comments for a PR.
func (g *GitHub) GetPRComments(prNumber int) (*PRComments, error) {
	// Get PR details first
	title, url, err := g.GetPRDetails(prNumber)
	if err != nil {
		return nil, err
	}

	// Get repository owner/name from remote
	repoInfo, err := g.getRepoInfo()
	if err != nil {
		return nil, err
	}

	// Fetch review comments via gh api
	apiPath := fmt.Sprintf("repos/%s/pulls/%d/comments", repoInfo, prNumber)
	cmd := exec.Command("gh", "api", apiPath)
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().Int("pr", prNumber).Str("api", apiPath).Msg("fetching PR comments")

	if err := cmd.Run(); err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to fetch PR comments: %s", stderr.String())
	}

	var rawComments []reviewCommentJSON
	if err := json.Unmarshal(stdout.Bytes(), &rawComments); err != nil {
		return nil, pkgerrors.Wrap(err, "failed to parse PR comments")
	}

	// Convert to our ReviewComment type
	comments := make([]ReviewComment, 0, len(rawComments))
	for _, rc := range rawComments {
		comments = append(comments, ReviewComment{
			ID:       rc.ID,
			Author:   rc.User.Login,
			Body:     rc.Body,
			Path:     rc.Path,
			Line:     rc.Line,
			DiffHunk: rc.DiffHunk,
			URL:      rc.HTMLURL,
		})
	}

	return &PRComments{
		PRNumber: prNumber,
		PRTitle:  title,
		PRURL:    url,
		Comments: comments,
	}, nil
}

// getRepoInfo returns the owner/repo string from the git remote.
func (g *GitHub) getRepoInfo() (string, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	cmd.Dir = g.repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", pkgerrors.Wrapf(err, "failed to get repo info: %s", stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// FormatCommentsAsPrompt formats PR comments as a prompt for Claude.
func FormatCommentsAsPrompt(comments *PRComments, prefix string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# PR Review Comments for PR #%d: %s\n\n", comments.PRNumber, comments.PRTitle))

	if prefix != "" {
		sb.WriteString("## Additional Context\n\n")
		sb.WriteString(prefix)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Review Comments to Address\n\n")

	for i, comment := range comments.Comments {
		sb.WriteString(fmt.Sprintf("### Comment %d by @%s\n", i+1, comment.Author))
		sb.WriteString(fmt.Sprintf("**File:** `%s`", comment.Path))
		if comment.Line > 0 {
			sb.WriteString(fmt.Sprintf(" (line %d)", comment.Line))
		}
		sb.WriteString("\n")

		if comment.DiffHunk != "" {
			sb.WriteString("**Code context:**\n")
			sb.WriteString("```\n")
			sb.WriteString(comment.DiffHunk)
			sb.WriteString("\n```\n")
		}

		sb.WriteString("**Comment:**\n")
		sb.WriteString(comment.Body)
		sb.WriteString("\n\n---\n\n")
	}

	sb.WriteString(`## Instructions
Please address these review comments by making the necessary code changes.
For each comment, either:
1. Make the requested code changes directly
2. If the comment is unclear or needs discussion, note what clarification is needed

Focus on implementing the requested changes accurately and completely.
`)

	return sb.String()
}
