package jira

import (
	"fmt"
	"strings"
)

type Ticket struct {
	Key             string
	Summary         string
	Description     string
	AcceptanceCrit  string
	IssueType       string
	Priority        string
	Labels          []string
	ProjectKey      string
}

func (t *Ticket) FormatAsPrompt(promptPrefix string) string {
	var sb strings.Builder

	if promptPrefix != "" {
		sb.WriteString(promptPrefix)
		sb.WriteString("\n\n")
	}

	sb.WriteString(fmt.Sprintf("# Jira Ticket: %s\n\n", t.Key))
	sb.WriteString(fmt.Sprintf("## Summary\n%s\n\n", t.Summary))

	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", t.Description))
	}

	if t.AcceptanceCrit != "" {
		sb.WriteString(fmt.Sprintf("## Acceptance Criteria\n%s\n\n", t.AcceptanceCrit))
	}

	if t.IssueType != "" {
		sb.WriteString(fmt.Sprintf("**Type:** %s\n", t.IssueType))
	}

	if t.Priority != "" {
		sb.WriteString(fmt.Sprintf("**Priority:** %s\n", t.Priority))
	}

	if len(t.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("**Labels:** %s\n", strings.Join(t.Labels, ", ")))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("Please implement this ticket. Follow best practices and existing code patterns in the repository.")

	return sb.String()
}
