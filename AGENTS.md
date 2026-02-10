# AGENTS.md

This file provides context for Claude Code agents working on the jira-claude repository.

## Project Overview

jira-claude is a CLI tool that automates the workflow from Jira ticket to GitHub PR. It invokes Claude Code to implement tickets and address PR review comments.

## Commands

| Command | Description |
|---------|-------------|
| `work` | Takes a Jira ticket, implements it with Claude, and creates a PR |
| `address-pr-comments` | Fetches PR review comments and uses Claude to address them |

## Directory Structure

```
jira-claude/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command and command wiring
│   ├── work.go            # work command implementation
│   └── address_pr_comments.go  # address-pr-comments command
├── internal/              # Internal packages
│   ├── claude/            # Claude Code CLI invocation
│   ├── config/            # Configuration loading (env vars)
│   ├── git/               # Git operations (branch, commit, push)
│   ├── github/            # GitHub operations via gh CLI (PR, comments, replies)
│   └── jira/              # Jira API client
├── main.go                # Entry point
├── Makefile               # Build targets
└── README.md              # User documentation
```

## Key Patterns

### CLI Framework
- Uses [Cobra](https://github.com/spf13/cobra) for command-line parsing
- Commands defined in `cmd/` with `init()` for flag registration
- Commands wired in `root.go` via `wireCommands()`

### External Tool Invocation
- Claude Code: Invoked via `claude` CLI with `--print` flag for non-interactive mode
- GitHub: Uses `gh` CLI for all GitHub operations (PRs, comments, API calls)
- Git: Executes git commands via `os/exec`

### Logging
- Uses [zerolog](https://github.com/rs/zerolog) with console writer for CLI output
- Context-based logging via `log.Ctx(ctx)`

### Error Handling
- Uses [pkg/errors](https://github.com/pkg/errors) for error wrapping
- Commands return errors; root command handles exit codes

## Building and Testing

```bash
# Build
make build

# Install to GOPATH/bin
make install

# Run directly
go run main.go work --ticket=PROJ-123
```

## Dependencies

- Go 1.21+
- Claude Code CLI (`claude`)
- GitHub CLI (`gh`)
- Git

## Environment Variables

See README.md for required configuration. Key variables:
- `JIRA_CLAUDE_JIRA_HOST` - Jira instance URL
- `JIRA_CLAUDE_JIRA_USERNAME` - Jira email
- `JIRA_CLAUDE_JIRA_API_TOKEN` - Jira API token
