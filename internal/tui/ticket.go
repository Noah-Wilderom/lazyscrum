package tui

import (
	"fmt"
	"strings"

	"github.com/Noah-Wilderom/lazyscrum/internal/model"
	"github.com/Noah-Wilderom/lazyscrum/internal/tracker"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type ticketSelector struct {
	issues       []tracker.Issue
	filtered     []tracker.Issue
	cursor       int
	searchInput  textinput.Model
	loading      bool
	err          error
	noConnection bool
}

func newTicketSelector() ticketSelector {
	si := textinput.New()
	si.Placeholder = "Search tickets..."
	si.Prompt = "> "
	si.Focus()

	return ticketSelector{
		searchInput: si,
	}
}

func (ts *ticketSelector) setIssues(issues []tracker.Issue) {
	ts.issues = issues
	ts.filtered = issues
	ts.cursor = 0
	ts.loading = false
}

func (ts *ticketSelector) setError(err error) {
	ts.err = err
	ts.loading = false
}

func (ts *ticketSelector) filter() {
	query := strings.ToLower(ts.searchInput.Value())
	if query == "" {
		ts.filtered = ts.issues
		ts.cursor = 0
		return
	}

	var filtered []tracker.Issue
	for _, issue := range ts.issues {
		text := strings.ToLower(issue.Key + " " + issue.Summary + " " + issue.Assignee)
		if strings.Contains(text, query) {
			filtered = append(filtered, issue)
		}
	}
	ts.filtered = filtered
	if ts.cursor >= len(ts.filtered) {
		ts.cursor = 0
	}
}

func (ts *ticketSelector) selectedIssue() *tracker.Issue {
	if len(ts.filtered) == 0 {
		return nil
	}
	return &ts.filtered[ts.cursor]
}

func (ts *ticketSelector) toJiraTicket() *model.JiraTicket {
	issue := ts.selectedIssue()
	if issue == nil {
		return nil
	}
	return &model.JiraTicket{
		Key:      issue.Key,
		Summary:  issue.Summary,
		Status:   issue.Status,
		Assignee: issue.Assignee,
		Priority: issue.Priority,
	}
}

func (ts *ticketSelector) view(width, height int) string {
	var b strings.Builder

	b.WriteString(overlayTitleStyle.Render("Select Jira Ticket"))
	b.WriteString("\n\n")

	if ts.noConnection {
		b.WriteString(helpStyle.Render("Jira not connected.\nRun `lazyscrum connect jira` to set up."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: close"))
		return b.String()
	}

	b.WriteString(ts.searchInput.View())
	b.WriteString("\n\n")

	if ts.loading {
		b.WriteString(helpStyle.Render("Loading tickets..."))
		return b.String()
	}

	if ts.err != nil {
		b.WriteString(statusTodoStyle.Render(fmt.Sprintf("Error: %v", ts.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: close"))
		return b.String()
	}

	if len(ts.filtered) == 0 {
		b.WriteString(helpStyle.Render("No tickets found."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: close"))
		return b.String()
	}

	maxVisible := height - 8
	if maxVisible < 3 {
		maxVisible = 3
	}

	start := 0
	if ts.cursor >= maxVisible {
		start = ts.cursor - maxVisible + 1
	}

	for i := start; i < len(ts.filtered) && i < start+maxVisible; i++ {
		issue := ts.filtered[i]

		key := ticketKeyStyle.Render(issue.Key)
		status := ticketStatusStyle.Render(issue.Status)
		assignee := ticketAssigneeStyle.Render(issue.Assignee)
		summary := issue.Summary

		maxSummaryWidth := width - 44
		if maxSummaryWidth > 0 && lipgloss.Width(summary) > maxSummaryWidth {
			summary = lipgloss.NewStyle().MaxWidth(maxSummaryWidth).Render(summary)
		}

		line := fmt.Sprintf("%s %s %s %s", key, status, assignee, summary)

		if i == ts.cursor {
			line = selectedItemStyle.Render("▸ " + line)
		} else {
			line = normalItemStyle.Render("  " + line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel"))

	return b.String()
}
