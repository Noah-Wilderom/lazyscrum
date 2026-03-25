package jira

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != "test@example.com" || pass != "test-token" {
			t.Errorf("expected basic auth test@example.com:test-token, got %s:%s", user, pass)
		}

		resp := searchResponse{
			Issues: []issueResponse{
				{
					Key: "PROJ-123",
					Fields: issueFields{
						Summary:  "Fix login bug",
						Status:   statusField{Name: "In Progress"},
						Priority: priorityField{Name: "High"},
						Assignee: &assigneeField{DisplayName: "Noah"},
					},
				},
				{
					Key: "PROJ-456",
					Fields: issueFields{
						Summary:  "Add logout button",
						Status:   statusField{Name: "Todo"},
						Priority: priorityField{Name: "Medium"},
						Assignee: nil,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "test@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	issues, err := client.SearchIssues("project = PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}
	if issues[0].Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", issues[0].Key)
	}
	if issues[0].Assignee != "Noah" {
		t.Errorf("expected assignee 'Noah', got '%s'", issues[0].Assignee)
	}
	if issues[1].Assignee != "" {
		t.Errorf("expected empty assignee, got '%s'", issues[1].Assignee)
	}
}

func TestGetIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/PROJ-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := issueResponse{
			Key: "PROJ-123",
			Fields: issueFields{
				Summary:  "Fix login bug",
				Status:   statusField{Name: "Done"},
				Priority: priorityField{Name: "High"},
				Assignee: &assigneeField{DisplayName: "Noah"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "test@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	issue, err := client.GetIssue("PROJ-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", issue.Key)
	}
	if issue.Status != "Done" {
		t.Errorf("expected status 'Done', got '%s'", issue.Status)
	}
}

func TestListProjectIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jql := r.URL.Query().Get("jql")
		if jql != "project = PROJ ORDER BY updated DESC" {
			t.Errorf("unexpected JQL: %s", jql)
		}

		resp := searchResponse{
			Issues: []issueResponse{
				{
					Key: "PROJ-1",
					Fields: issueFields{
						Summary:  "Task 1",
						Status:   statusField{Name: "Todo"},
						Priority: priorityField{Name: "Low"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		email:      "test@example.com",
		apiToken:   "test-token",
		httpClient: http.DefaultClient,
	}

	issues, err := client.ListProjectIssues("PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Key != "PROJ-1" {
		t.Errorf("expected key PROJ-1, got %s", issues[0].Key)
	}
}

func TestExtractAcceptanceCriteria(t *testing.T) {
	// Simulates the ADF structure Jira returns for:
	// ## Acceptatiecriteria
	// - First criterion
	// - Second criterion
	doc := &adfDocument{
		Type: "doc",
		Content: []adfNode{
			{
				Type: "heading",
				Content: []adfNode{
					{Type: "text", Text: "Acceptatiecriteria"},
				},
			},
			{
				Type: "bulletList",
				Content: []adfNode{
					{
						Type: "listItem",
						Content: []adfNode{
							{
								Type: "paragraph",
								Content: []adfNode{
									{Type: "text", Text: "First criterion"},
								},
							},
						},
					},
					{
						Type: "listItem",
						Content: []adfNode{
							{
								Type: "paragraph",
								Content: []adfNode{
									{Type: "text", Text: "Second criterion"},
								},
							},
						},
					},
				},
			},
		},
	}

	criteria := extractAcceptanceCriteria(doc)
	if len(criteria) != 2 {
		t.Fatalf("expected 2 criteria, got %d", len(criteria))
	}
	if criteria[0] != "First criterion" {
		t.Errorf("expected 'First criterion', got '%s'", criteria[0])
	}
	if criteria[1] != "Second criterion" {
		t.Errorf("expected 'Second criterion', got '%s'", criteria[1])
	}
}

func TestExtractAcceptanceCriteriaEnglish(t *testing.T) {
	doc := &adfDocument{
		Type: "doc",
		Content: []adfNode{
			{
				Type: "heading",
				Content: []adfNode{
					{Type: "text", Text: "Acceptance Criteria"},
				},
			},
			{
				Type: "bulletList",
				Content: []adfNode{
					{
						Type: "listItem",
						Content: []adfNode{
							{
								Type: "paragraph",
								Content: []adfNode{
									{Type: "text", Text: "Users can log in"},
								},
							},
						},
					},
				},
			},
		},
	}

	criteria := extractAcceptanceCriteria(doc)
	if len(criteria) != 1 {
		t.Fatalf("expected 1 criterion, got %d", len(criteria))
	}
	if criteria[0] != "Users can log in" {
		t.Errorf("expected 'Users can log in', got '%s'", criteria[0])
	}
}

func TestExtractAcceptanceCriteriaNone(t *testing.T) {
	doc := &adfDocument{
		Type: "doc",
		Content: []adfNode{
			{
				Type: "paragraph",
				Content: []adfNode{
					{Type: "text", Text: "Just a regular description."},
				},
			},
		},
	}

	criteria := extractAcceptanceCriteria(doc)
	if len(criteria) != 0 {
		t.Errorf("expected 0 criteria, got %d", len(criteria))
	}
}

func TestExtractAcceptanceCriteriaWithRichText(t *testing.T) {
	// Simulates bold/italic text within list items
	doc := &adfDocument{
		Type: "doc",
		Content: []adfNode{
			{
				Type: "heading",
				Content: []adfNode{
					{Type: "text", Text: "Acceptatiecriteria"},
				},
			},
			{
				Type: "bulletList",
				Content: []adfNode{
					{
						Type: "listItem",
						Content: []adfNode{
							{
								Type: "paragraph",
								Content: []adfNode{
									{Type: "text", Text: "Email sent to "},
									{Type: "text", Text: "support@fanstr.nl"},
								},
							},
						},
					},
				},
			},
		},
	}

	criteria := extractAcceptanceCriteria(doc)
	if len(criteria) != 1 {
		t.Fatalf("expected 1 criterion, got %d", len(criteria))
	}
	if criteria[0] != "Email sent to support@fanstr.nl" {
		t.Errorf("expected 'Email sent to support@fanstr.nl', got '%s'", criteria[0])
	}
}
