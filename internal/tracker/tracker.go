package tracker

// Issue represents an issue from any tracker (Jira, Linear, etc.).
type Issue struct {
	Key                string
	Summary            string
	Status             string
	Assignee           string
	Priority           string
	AcceptanceCriteria []string
}

// Tracker is a read-only interface for issue trackers.
type Tracker interface {
	Name() string
	SearchIssues(query string) ([]Issue, error)
	GetIssue(key string) (*Issue, error)
	ListProjectIssues(projectKey string) ([]Issue, error)
}
