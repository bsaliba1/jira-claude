package jira

import (
	"github.com/andygrunwald/go-jira"
	pkgerrors "github.com/pkg/errors"
)

type Client interface {
	GetTicket(ticketKey string) (*Ticket, error)
}

var _ Client = (*JiraClient)(nil)

type JiraClient struct {
	client *jira.Client
}

func NewClient(host, username, apiToken string) (*JiraClient, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: apiToken,
	}

	client, err := jira.NewClient(tp.Client(), host)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "failed to create Jira client")
	}

	return &JiraClient{client: client}, nil
}

func (c *JiraClient) GetTicket(ticketKey string) (*Ticket, error) {
	issue, _, err := c.client.Issue.Get(ticketKey, nil)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to get ticket %s", ticketKey)
	}

	ticket := &Ticket{
		Key:        issue.Key,
		Summary:    issue.Fields.Summary,
		ProjectKey: issue.Fields.Project.Key,
	}

	if issue.Fields.Description != "" {
		ticket.Description = issue.Fields.Description
	}

	if issue.Fields.Type.Name != "" {
		ticket.IssueType = issue.Fields.Type.Name
	}

	if issue.Fields.Priority != nil {
		ticket.Priority = issue.Fields.Priority.Name
	}

	if len(issue.Fields.Labels) > 0 {
		ticket.Labels = issue.Fields.Labels
	}

	// Try to extract acceptance criteria from custom field if present
	if issue.Fields.Unknowns != nil {
		// Common custom field IDs for acceptance criteria
		for _, fieldID := range []string{"customfield_10016", "customfield_10017", "customfield_10001"} {
			if ac, ok := issue.Fields.Unknowns[fieldID]; ok {
				if acStr, ok := ac.(string); ok && acStr != "" {
					ticket.AcceptanceCrit = acStr
					break
				}
			}
		}
	}

	return ticket, nil
}
