package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"lazyscrum/internal/tracker"
)

type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

func NewClient(domain, email, apiToken string) *Client {
	return &Client{
		baseURL:    fmt.Sprintf("https://%s", domain),
		email:      email,
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

func (c *Client) Name() string {
	return "jira"
}

// Jira REST API v3 response types

type searchResponse struct {
	Issues []issueResponse `json:"issues"`
}

type issueResponse struct {
	Key    string      `json:"key"`
	Fields issueFields `json:"fields"`
}

type issueFields struct {
	Summary     string          `json:"summary"`
	Status      statusField     `json:"status"`
	Priority    priorityField   `json:"priority"`
	Assignee    *assigneeField  `json:"assignee"`
	Description *adfDocument    `json:"description"`
}

// Atlassian Document Format (ADF) types

type adfDocument struct {
	Type    string    `json:"type"`
	Content []adfNode `json:"content"`
}

type adfNode struct {
	Type    string    `json:"type"`
	Content []adfNode `json:"content"`
	Text    string    `json:"text,omitempty"`
	Attrs   *adfAttrs `json:"attrs,omitempty"`
}

type adfAttrs struct {
	Level int `json:"level,omitempty"`
}

type statusField struct {
	Name string `json:"name"`
}

type priorityField struct {
	Name string `json:"name"`
}

type assigneeField struct {
	DisplayName string `json:"displayName"`
}

func (c *Client) SearchIssues(query string) ([]tracker.Issue, error) {
	u := fmt.Sprintf("%s/rest/api/3/search/jql?jql=%s&fields=summary,status,priority,assignee,description",
		c.baseURL, url.QueryEscape(query))
	return c.fetchIssues(u)
}

func (c *Client) GetIssue(key string) (*tracker.Issue, error) {
	u := fmt.Sprintf("%s/rest/api/3/issue/%s?fields=summary,status,priority,assignee,description",
		c.baseURL, url.PathEscape(key))

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	var ir issueResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	issue := toTrackerIssue(ir)
	return &issue, nil
}

func (c *Client) ListProjectIssues(projectKey string) ([]tracker.Issue, error) {
	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", projectKey)
	u := fmt.Sprintf("%s/rest/api/3/search/jql?jql=%s&fields=summary,status,priority,assignee,description",
		c.baseURL, url.QueryEscape(jql))
	return c.fetchIssues(u)
}

func (c *Client) fetchIssues(u string) ([]tracker.Issue, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	issues := make([]tracker.Issue, len(sr.Issues))
	for i, ir := range sr.Issues {
		issues[i] = toTrackerIssue(ir)
	}
	return issues, nil
}

func toTrackerIssue(ir issueResponse) tracker.Issue {
	assignee := ""
	if ir.Fields.Assignee != nil {
		assignee = ir.Fields.Assignee.DisplayName
	}
	var ac []string
	if ir.Fields.Description != nil {
		ac = extractAcceptanceCriteria(ir.Fields.Description)
	}
	return tracker.Issue{
		Key:                ir.Key,
		Summary:            ir.Fields.Summary,
		Status:             ir.Fields.Status.Name,
		Assignee:           assignee,
		Priority:           ir.Fields.Priority.Name,
		AcceptanceCriteria: ac,
	}
}

// extractAcceptanceCriteria finds a heading containing "acceptatiecriteria"
// or "acceptance criteria" in an ADF document, then extracts the bullet list
// items that follow it.
func extractAcceptanceCriteria(doc *adfDocument) []string {
	found := false
	var criteria []string

	for _, node := range doc.Content {
		if isACHeading(node) {
			found = true
			continue
		}

		if found {
			// Once we hit the AC heading, collect bullet list items
			if node.Type == "bulletList" || node.Type == "orderedList" {
				for _, item := range node.Content {
					if item.Type == "listItem" {
						text := extractText(item)
						text = strings.TrimSpace(text)
						if text != "" {
							criteria = append(criteria, text)
						}
					}
				}
				// Stop after the first list following the heading
				break
			}

			// If we hit another heading, stop looking
			if node.Type == "heading" {
				break
			}
		}
	}

	return criteria
}

// isACHeading checks if a node is a heading containing acceptance criteria keywords.
func isACHeading(node adfNode) bool {
	if node.Type != "heading" {
		return false
	}
	text := strings.ToLower(extractText(node))
	return strings.Contains(text, "acceptatiecriteria") ||
		strings.Contains(text, "acceptance criteria")
}

// extractText recursively extracts all text content from an ADF node.
func extractText(node adfNode) string {
	if node.Text != "" {
		return node.Text
	}
	var parts []string
	for _, child := range node.Content {
		if t := extractText(child); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "")
}
