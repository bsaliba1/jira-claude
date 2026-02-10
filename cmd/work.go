package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bsaliba1/jira-claude/internal/claude"
	"github.com/bsaliba1/jira-claude/internal/config"
	"github.com/bsaliba1/jira-claude/internal/git"
	"github.com/bsaliba1/jira-claude/internal/github"
	"github.com/bsaliba1/jira-claude/internal/jira"
	"github.com/kelseyhightower/envconfig"
	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	flagTicket       string
	flagRepo         string
	flagBaseBranch   string
	flagPromptPrefix string
	flagDryRun       bool
)

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Implement a Jira ticket and create a PR",
	Long: `Fetches a Jira ticket, creates a feature branch, invokes Claude Code to implement
the ticket, commits the changes, pushes the branch, and creates a GitHub PR.`,
	RunE: runWork,
}

func init() {
	workCmd.Flags().StringVarP(&flagTicket, "ticket", "t", "", "Jira ticket key (e.g., PROJ-123)")
	workCmd.Flags().StringVarP(&flagRepo, "repo", "r", "", "Path to git repository (defaults to current directory)")
	workCmd.Flags().StringVarP(&flagBaseBranch, "base-branch", "b", "", "Base branch for PR (defaults to config or 'main')")
	workCmd.Flags().StringVarP(&flagPromptPrefix, "prompt-prefix", "p", "", "Additional context to prepend to the prompt")
	workCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print what would be done without making changes")

	workCmd.MarkFlagRequired("ticket")
}

func runWork(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	l := log.Ctx(ctx).With().Str("ticket", flagTicket).Logger()

	// Load configuration
	var conf config.Config
	if err := envconfig.Process(config.EnvConfigPrefix, &conf); err != nil {
		return pkgerrors.Wrap(err, "failed to load configuration (check JIRA_CLAUDE_* env vars)")
	}

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

	// Resolve base branch
	baseBranch := flagBaseBranch
	if baseBranch == "" {
		baseBranch = conf.DefaultBaseBranch
	}

	l.Info().Str("repo", repoPath).Str("baseBranch", baseBranch).Msg("starting work on ticket")

	// Step 1: Create Jira client and fetch ticket
	l.Info().Msg("fetching Jira ticket")
	jiraClient, err := jira.NewClient(conf.JiraHost, conf.JiraUsername, conf.JiraAPIToken)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to create Jira client")
	}

	ticket, err := jiraClient.GetTicket(flagTicket)
	if err != nil {
		return pkgerrors.Wrap(err, "failed to fetch ticket")
	}

	l.Info().
		Str("summary", ticket.Summary).
		Str("type", ticket.IssueType).
		Msg("fetched ticket details")

	// Step 2: Initialize git and ensure clean state
	gitClient := git.New(repoPath)

	if err := gitClient.EnsureClean(); err != nil {
		return pkgerrors.Wrap(err, "repository must be clean before starting")
	}

	// Step 3: Checkout base branch and pull latest
	l.Info().Str("branch", baseBranch).Msg("checking out base branch")
	if !flagDryRun {
		if err := gitClient.Checkout(baseBranch); err != nil {
			return pkgerrors.Wrapf(err, "failed to checkout %s", baseBranch)
		}
		if err := gitClient.Pull(); err != nil {
			l.Warn().Err(err).Msg("failed to pull latest (continuing anyway)")
		}
	}

	// Step 4: Create feature branch
	branchName := git.GenerateBranchName(conf.BranchPrefix, ticket.Key, ticket.Summary)
	l.Info().Str("branch", branchName).Msg("creating feature branch")

	if flagDryRun {
		l.Info().Msg("[dry-run] would create branch")
	} else {
		if gitClient.BranchExists(branchName) {
			l.Info().Str("branch", branchName).Msg("branch exists, deleting and recreating")
			if err := gitClient.DeleteBranch(branchName); err != nil {
				return pkgerrors.Wrap(err, "failed to delete existing branch")
			}
		}
		if err := gitClient.CreateBranch(branchName); err != nil {
			return pkgerrors.Wrap(err, "failed to create feature branch")
		}
	}

	// Step 5: Generate prompt and invoke Claude
	prompt := ticket.FormatAsPrompt(flagPromptPrefix)
	l.Info().Msg("invoking Claude Code")

	if flagDryRun {
		l.Info().Str("prompt", prompt).Msg("[dry-run] would invoke Claude with prompt")
	} else {
		claudeClient := claude.New(repoPath)
		if err := claudeClient.Run(prompt); err != nil {
			return pkgerrors.Wrap(err, "Claude Code failed")
		}
	}

	// Step 6: Check for changes and commit
	l.Info().Msg("checking for changes")

	if flagDryRun {
		l.Info().Msg("[dry-run] would commit and push changes")
	} else {
		hasChanges, err := gitClient.HasChanges()
		if err != nil {
			return pkgerrors.Wrap(err, "failed to check for changes")
		}

		if !hasChanges {
			l.Warn().Msg("no changes were made by Claude")
			return nil
		}

		// Commit changes
		commitMsg := fmt.Sprintf("%s: %s\n\nImplemented by Claude Code", ticket.Key, ticket.Summary)
		if err := gitClient.AddAll(); err != nil {
			return pkgerrors.Wrap(err, "failed to stage changes")
		}
		if err := gitClient.Commit(commitMsg); err != nil {
			return pkgerrors.Wrap(err, "failed to commit changes")
		}
		l.Info().Msg("committed changes")

		// Push branch
		if err := gitClient.Push(); err != nil {
			return pkgerrors.Wrap(err, "failed to push branch")
		}
		l.Info().Msg("pushed branch to origin")
	}

	// Step 7: Create PR
	l.Info().Msg("creating pull request")

	if flagDryRun {
		l.Info().Msg("[dry-run] would create PR")
	} else {
		ghClient := github.New(repoPath)
		prTitle := fmt.Sprintf("%s: %s", ticket.Key, ticket.Summary)
		prBody := github.FormatPRBody(ticket.Key, ticket.Summary, conf.JiraHost)

		prURL, err := ghClient.CreatePR(prTitle, prBody, baseBranch)
		if err != nil {
			return pkgerrors.Wrap(err, "failed to create PR")
		}

		l.Info().Str("url", prURL).Msg("created pull request")
		fmt.Printf("\nPR created: %s\n", prURL)
	}

	l.Info().Msg("work complete")
	return nil
}
