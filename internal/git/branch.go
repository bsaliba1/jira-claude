package git

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	multiDash       = regexp.MustCompile(`-+`)
)

// GenerateBranchName creates a branch name from ticket key and summary.
// Format: prefix/ticket-key-summary-slug
// Example: feature/proj-123-add-user-authentication
func GenerateBranchName(prefix, ticketKey, summary string) string {
	// Lowercase and normalize the summary
	slug := strings.ToLower(summary)

	// Replace non-alphanumeric characters with dashes
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")

	// Collapse multiple dashes
	slug = multiDash.ReplaceAllString(slug, "-")

	// Trim leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// Limit slug length to keep branch names reasonable
	const maxSlugLen = 50
	if len(slug) > maxSlugLen {
		slug = slug[:maxSlugLen]
		// Don't end with a dash
		slug = strings.TrimRight(slug, "-")
	}

	// Lowercase the ticket key
	ticketLower := strings.ToLower(ticketKey)

	// Build branch name
	branchName := prefix + ticketLower
	if slug != "" {
		branchName += "-" + slug
	}

	return branchName
}
