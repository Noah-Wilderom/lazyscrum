package model

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type AcceptanceCriterion struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewAcceptanceCriterion creates a new AcceptanceCriterion with a UUID, sets
// status to StatusTodo, and timestamps both CreatedAt and UpdatedAt to now.
func NewAcceptanceCriterion(title, description string, priority Priority) AcceptanceCriterion {
	now := time.Now()
	return AcceptanceCriterion{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CycleStatus advances the status through the cycle: todo -> in-progress ->
// done -> todo, and updates UpdatedAt.
func (ac *AcceptanceCriterion) CycleStatus() {
	switch ac.Status {
	case StatusTodo:
		ac.Status = StatusInProgress
	case StatusInProgress:
		ac.Status = StatusDone
	case StatusDone:
		ac.Status = StatusTodo
	}
	ac.UpdatedAt = time.Now()
}

// JiraTicket holds the linked issue tracker ticket for this project/branch.
type JiraTicket struct {
	Key      string `json:"key"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
	Priority string `json:"priority"`
}

// State holds the application state.
type State struct {
	Items      []AcceptanceCriterion `json:"items"`
	JiraTicket *JiraTicket           `json:"jira_ticket,omitempty"`
}

// NewState creates an empty State.
func NewState() *State {
	return &State{
		Items: []AcceptanceCriterion{},
	}
}

// Add appends an AcceptanceCriterion to the state.
func (s *State) Add(ac AcceptanceCriterion) {
	s.Items = append(s.Items, ac)
}

// Remove removes the AcceptanceCriterion with the given ID from the state.
func (s *State) Remove(id string) {
	filtered := s.Items[:0]
	for _, item := range s.Items {
		if item.ID != id {
			filtered = append(filtered, item)
		}
	}
	s.Items = filtered
}

// Update replaces the AcceptanceCriterion matching ac.ID with the provided
// value and sets UpdatedAt to now.
func (s *State) Update(ac AcceptanceCriterion) {
	now := time.Now()
	for i, item := range s.Items {
		if item.ID == ac.ID {
			ac.UpdatedAt = now
			s.Items[i] = ac
			return
		}
	}
}
