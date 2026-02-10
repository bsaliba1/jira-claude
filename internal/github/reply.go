package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ReplyToComment posts a reply to a review comment.
func (g *GitHub) ReplyToComment(prNumber int, commentID int64, body string) error {
	repoInfo, err := g.getRepoInfo()
	if err != nil {
		return err
	}

	// Create reply via gh api
	apiPath := fmt.Sprintf("repos/%s/pulls/%d/comments/%d/replies", repoInfo, prNumber, commentID)

	// Build JSON payload
	payload := map[string]string{"body": body}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to marshal reply payload")
	}

	cmd := exec.Command("gh", "api", apiPath, "-X", "POST", "--input", "-")
	cmd.Dir = g.repoPath
	cmd.Stdin = bytes.NewReader(payloadBytes)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().Int("pr", prNumber).Int64("commentID", commentID).Msg("posting reply to comment")

	if err := cmd.Run(); err != nil {
		return pkgerrors.Wrapf(err, "failed to post reply: %s", stderr.String())
	}

	return nil
}
