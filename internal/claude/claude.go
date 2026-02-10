package claude

import (
	"bytes"
	"os"
	"os/exec"

	pkgerrors "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Claude struct {
	workDir string
}

func New(workDir string) *Claude {
	return &Claude{workDir: workDir}
}

// Run executes Claude Code with the given prompt.
// It runs claude -p "<prompt>" in the working directory.
func (c *Claude) Run(prompt string) error {
	cmd := exec.Command("claude", "-p", prompt, "--allowedTools", "Write,Edit,Read,Bash,Grep,Glob", "--permission-mode", "bypassPermissions")
	cmd.Dir = c.workDir

	// Pass through environment for AWS credentials (Bedrock)
	cmd.Env = os.Environ()

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	log.Info().Str("workDir", c.workDir).Msg("invoking Claude Code")
	log.Debug().Str("prompt", prompt).Msg("claude prompt")

	if err := cmd.Run(); err != nil {
		return pkgerrors.Wrapf(err, "claude command failed: %s", stderr.String())
	}

	return nil
}

// RunWithOutput executes Claude Code and returns the output.
func (c *Claude) RunWithOutput(prompt string) (string, error) {
	cmd := exec.Command("claude", "-p", prompt, "--allowedTools", "Write,Edit,Read,Bash,Grep,Glob", "--permission-mode", "bypassPermissions")
	cmd.Dir = c.workDir

	// Pass through environment for AWS credentials (Bedrock)
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info().Str("workDir", c.workDir).Msg("invoking Claude Code")
	log.Debug().Str("prompt", prompt).Msg("claude prompt")

	if err := cmd.Run(); err != nil {
		return "", pkgerrors.Wrapf(err, "claude command failed: %s", stderr.String())
	}

	return stdout.String(), nil
}
