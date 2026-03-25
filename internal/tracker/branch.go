package tracker

import "regexp"

var issueKeyPattern = regexp.MustCompile(`[A-Z]+-\d+`)

// ExtractIssueKey scans a branch name for a Jira-style issue key (e.g. PROJ-123).
// Returns the first match, or empty string if none found.
func ExtractIssueKey(branch string) string {
	return issueKeyPattern.FindString(branch)
}
