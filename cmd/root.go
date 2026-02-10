package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "jira-claude",
	Short: "Jira to PR automation tool",
	Long:  `A CLI tool that takes a Jira ticket, invokes Claude Code to implement it, and creates a GitHub PR.`,
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	initLogger()
	ctx = log.Logger.WithContext(ctx)

	wireCommands()

	if err := root.ExecuteContext(ctx); err != nil {
		log.Error().Err(err).Msg("error running command")
		os.Exit(1)
	}
}

func wireCommands() {
	root.AddCommand(workCmd)
	root.AddCommand(addressPRCommentsCmd)
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Use console writer for better readability in CLI
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	zerolog.DefaultContextLogger = &log.Logger
}
