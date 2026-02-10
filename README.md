# jira-claude

A CLI tool that takes a Jira ticket, invokes Claude Code to implement it, and creates a GitHub PR.

## Installation

```bash
go install github.com/bsaliba1/jira-claude@latest
```

Or clone and build locally:

```bash
git clone https://github.com/bsaliba1/jira-claude.git
cd jira-claude
make install
```

## Configuration

Set the following environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JIRA_CLAUDE_JIRA_HOST` | Yes | - | Your Jira instance URL (e.g., `https://yourcompany.atlassian.net`) |
| `JIRA_CLAUDE_JIRA_USERNAME` | Yes | - | Your Jira username (email) |
| `JIRA_CLAUDE_JIRA_API_TOKEN` | Yes | - | Your Jira API token |
| `JIRA_CLAUDE_BRANCH_PREFIX` | No | `feature/` | Prefix for created branches |
| `JIRA_CLAUDE_DEFAULT_BASE_BRANCH` | No | `main` | Default base branch for PRs |

### Getting a Jira API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a label and copy the token

## Prerequisites

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) CLI installed and authenticated
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated

## Usage

```bash
jira-claude work --ticket=PROJ-123
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--ticket` | `-t` | Jira ticket key (required) |
| `--repo` | `-r` | Path to git repository (defaults to current directory) |
| `--base-branch` | `-b` | Base branch for PR (defaults to config or 'main') |
| `--prompt-prefix` | `-p` | Additional context to prepend to the prompt |
| `--dry-run` | - | Print what would be done without making changes |

### Examples

```bash
# Work on a ticket in the current directory
jira-claude work --ticket=SUI-640

# Work on a ticket in a different repository
jira-claude work --ticket=SUI-640 --repo=/path/to/repo

# Use a different base branch
jira-claude work --ticket=SUI-640 --base-branch=develop

# Add additional context for Claude
jira-claude work --ticket=SUI-640 --prompt-prefix="Use React hooks for this implementation"

# Preview what would happen without making changes
jira-claude work --ticket=SUI-640 --dry-run
```

### Address PR Comments

Address review comments on a PR using Claude:

```bash
# Address comments on PR detected from current branch
jira-claude address-pr-comments

# Address comments on specific PR
jira-claude address-pr-comments --pr 123

# Preview what would be done
jira-claude address-pr-comments --dry-run

# Also post replies to comments after making changes
jira-claude address-pr-comments --with-replies

# Add context for Claude
jira-claude address-pr-comments --prompt-prefix "This is a React codebase"
```

#### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--pr` | `-n` | PR number (auto-detect from current branch if omitted) |
| `--repo` | `-r` | Path to git repository (defaults to current directory) |
| `--dry-run` | - | Preview without making changes |
| `--prompt-prefix` | `-p` | Additional context for Claude |
| `--no-push` | - | Skip automatic push after commit |
| `--with-replies` | - | Post reply summaries to comments after making changes |

## Workflow

### Work Command

When you run `jira-claude work`, it:

1. Fetches the Jira ticket details
2. Ensures the git repository is clean
3. Checks out the base branch and pulls latest
4. Creates a feature branch (e.g., `feature/SUI-640-ticket-summary`)
5. Invokes Claude Code with the ticket details as a prompt
6. Commits any changes made by Claude
7. Pushes the branch to origin
8. Creates a GitHub PR linking back to the Jira ticket

### Address PR Comments Command

When you run `jira-claude address-pr-comments`, it:

1. Detects the PR from the current branch (or uses `--pr`)
2. Fetches all review comments from the PR
3. Formats the comments into a prompt for Claude
4. Invokes Claude Code to address the comments
5. Commits any changes made by Claude
6. Pushes the branch to origin (unless `--no-push`)
7. Optionally posts replies to each comment (with `--with-replies`)

## License

MIT
