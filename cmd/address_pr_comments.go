package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bsaliba1/jira-claude/internal/claude"
	"github.com/bsaliba1/jira-claude/internal/git"
	"github.com/bsaliba1/jira-claude/internal/github"
	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	flagPRNumber    int
	flagNoPush      bool
	flagWithReplies bool
)

var addressPRCommentsCmd = &cobra.Command{
	Use:   "address-pr-comments",
	Short: "Address PR review comments using Claude",
	Long: `Fetches PR review comments from GitHub, uses Claude to address them by making
code changes, and commits/pushes the changes.

If no PR number is provided, it will attempt to detect the PR from the current branch.`,
	RunE: runAddressPRComments,
}

func init() {
	addressPRCommentsCmd.Flags().IntVarP(&flagPRNumber, "pr", "n", 0, "PR number (auto-detect from current branch if omitted)")
	addressPRCommentsCmd.Flags().StringVarP(&flagRepo, "repo", "r", "", "Path to git repository (defaults to current directory)")
	addressPRCommentsCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Preview without making changes")
	addressPRCommentsCmd.Flags().StringVarP(&flagPromptPrefix, "prompt-prefix", "p", "", "Additional context for Claude")
	addressPRCommentsCmd.Flags().BoolVar(&flagNoPush, "no-push", false, "Skip automatic push after commit")
	addressPRCommentsCmd.Flags().BoolVar(&flagWithReplies, "with-replies", false, "Post reply summaries to comments after making changes")
}

func runAddressPRComments(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	l := log.Ctx(ctx)

	// Resolve repository path
	repoPath := flagRepo
	if repoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return pkgerrors.Wrap(err, "failed to get current directory")
		}
		repoPath = cwd
	}
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to resolve repository path")
	}

	ghClient := github.New(repoPath)
	gitClient := git.New(repoPath)

	// Determine PR number
	prNumber := flagPRNumber
	if prNumber == 0 {
		l.Info().Msg("detecting PR from current branch")
		detected, err := ghClient.GetPRForBranch()
		if err != nil {
			return pkgerrors.Wrap(err, "failed to detect PR (use --pr to specify)")
		}
		prNumber = detected
	}

	l.Info().Int("pr", prNumber).Msg("fetching PR comments")

	// Fetch PR comments
	comments, err := ghClient.GetPRComments(prNumber)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to fetch PR comments")
	}

	if len(comments.Comments) == 0 {
		l.Info().Msg("no review comments to address")
		fmt.Println("No review comments found on this PR.")
		return nil
	}

	l.Info().Int("count", len(comments.Comments)).Msg("found review comments")

	// Check git state
	hasChanges, err := gitClient.HasChanges()
	if err != nil {
		return pkgerrors.Wrap(err, "failed to check git status")
	}
	if hasChanges {
		l.Warn().Msg("working directory has uncommitted changes")
	}

	// Format comments as prompt
	prompt := github.FormatCommentsAsPrompt(comments, flagPromptPrefix)

	if flagDryRun {
		l.Info().Msg("[dry-run] would invoke Claude with the following prompt:")
		fmt.Println("\n--- PROMPT ---")
		fmt.Println(prompt)
		fmt.Println("--- END PROMPT ---")
		return nil
	}

	// Invoke Claude
	l.Info().Msg("invoking Claude Code to address comments")
	claudeClient := claude.New(repoPath)
	if err := claudeClient.Run(prompt); err != nil {
		return pkgerrors.Wrap(err, "Claude Code failed")
	}

	// Check for changes
	hasChanges, err = gitClient.HasChanges()
	if err != nil {
		return pkgerrors.Wrap(err, "failed to check for changes")
	}

	if !hasChanges {
		l.Info().Msg("no code changes were made")
		fmt.Println("No code changes were made by Claude.")
		return nil
	}

	// Commit changes
	commitMsg := fmt.Sprintf("Address PR #%d review comments\n\nAddressed by Claude Code", prNumber)
	if err := gitClient.AddAll(); err != nil {
		return pkgerrors.Wrap(err, "failed to stage changes")
	}
	if err := gitClient.Commit(commitMsg); err != nil {
		return pkgerrors.Wrap(err, "failed to commit changes")
	}
	l.Info().Msg("committed changes")

	// Push unless --no-push
	if !flagNoPush {
		if err := gitClient.Push(); err != nil {
			return pkgerrors.Wrap(err, "failed to push changes")
		}
		l.Info().Msg("pushed changes to origin")
	} else {
		l.Info().Msg("skipping push (--no-push specified)")
	}

	// Post replies if requested
	if flagWithReplies {
		l.Info().Msg("posting replies to comments")
		replyBody := "Addressed in latest commit."
		for _, comment := range comments.Comments {
			if err := ghClient.ReplyToComment(prNumber, comment.ID, replyBody); err != nil {
				l.Warn().Err(err).Int64("commentID", comment.ID).Msg("failed to post reply")
			}
		}
		l.Info().Msg("posted replies to all comments")
	}

	l.Info().Msg("finished addressing PR comments")
	fmt.Printf("\nSuccessfully addressed %d review comments on PR #%d\n", len(comments.Comments), prNumber)
	fmt.Printf("PR: %s\n", comments.PRURL)

	return nil
}
