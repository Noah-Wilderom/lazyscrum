package tracker

import "testing"

func TestExtractIssueKey(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"feature/PROJ-123-add-login", "PROJ-123"},
		{"PROJ-456-description", "PROJ-456"},
		{"bugfix/PROJ-789", "PROJ-789"},
		{"PROJ-1", "PROJ-1"},
		{"AB-99-something", "AB-99"},
		{"main", ""},
		{"feature/no-ticket-here", ""},
		{"", ""},
		{"feature/PROJ-123/sub/TASK-456", "PROJ-123"},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := ExtractIssueKey(tt.branch)
			if got != tt.want {
				t.Errorf("ExtractIssueKey(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}
