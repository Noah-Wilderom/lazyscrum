package model

import (
	"testing"
	"time"
)

func TestNewAcceptanceCriterion(t *testing.T) {
	title := "User can log in"
	description := "Given valid credentials, user can authenticate"
	priority := PriorityHigh

	before := time.Now()
	ac := NewAcceptanceCriterion(title, description, priority)
	after := time.Now()

	if ac.ID == "" {
		t.Error("expected non-empty ID")
	}
	if ac.Title != title {
		t.Errorf("expected title %q, got %q", title, ac.Title)
	}
	if ac.Description != description {
		t.Errorf("expected description %q, got %q", description, ac.Description)
	}
	if ac.Status != StatusTodo {
		t.Errorf("expected status %q, got %q", StatusTodo, ac.Status)
	}
	if ac.Priority != priority {
		t.Errorf("expected priority %q, got %q", priority, ac.Priority)
	}
	if ac.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if ac.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
	if ac.CreatedAt.Before(before) || ac.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", ac.CreatedAt, before, after)
	}
	if ac.UpdatedAt.Before(before) || ac.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v not within expected range [%v, %v]", ac.UpdatedAt, before, after)
	}
}

func TestCycleStatus(t *testing.T) {
	ac := NewAcceptanceCriterion("title", "desc", PriorityLow)

	if ac.Status != StatusTodo {
		t.Fatalf("initial status should be todo, got %q", ac.Status)
	}

	prevUpdated := ac.UpdatedAt

	// todo -> in-progress
	ac.CycleStatus()
	if ac.Status != StatusInProgress {
		t.Errorf("after first cycle expected in-progress, got %q", ac.Status)
	}
	if !ac.UpdatedAt.After(prevUpdated) {
		t.Error("UpdatedAt should have advanced after CycleStatus")
	}
	prevUpdated = ac.UpdatedAt

	// in-progress -> done
	ac.CycleStatus()
	if ac.Status != StatusDone {
		t.Errorf("after second cycle expected done, got %q", ac.Status)
	}
	if !ac.UpdatedAt.After(prevUpdated) {
		t.Error("UpdatedAt should have advanced after CycleStatus")
	}
	prevUpdated = ac.UpdatedAt

	// done -> todo
	ac.CycleStatus()
	if ac.Status != StatusTodo {
		t.Errorf("after third cycle expected todo, got %q", ac.Status)
	}
	if !ac.UpdatedAt.After(prevUpdated) {
		t.Error("UpdatedAt should have advanced after CycleStatus")
	}
}

func TestStateAddAndRemove(t *testing.T) {
	s := NewState()

	if len(s.Items) != 0 {
		t.Fatalf("new state should have 0 items, got %d", len(s.Items))
	}

	ac := NewAcceptanceCriterion("title", "desc", PriorityMedium)
	s.Add(ac)

	if len(s.Items) != 1 {
		t.Fatalf("after Add, expected 1 item, got %d", len(s.Items))
	}
	if s.Items[0].ID != ac.ID {
		t.Errorf("expected item ID %q, got %q", ac.ID, s.Items[0].ID)
	}

	s.Remove(ac.ID)

	if len(s.Items) != 0 {
		t.Fatalf("after Remove, expected 0 items, got %d", len(s.Items))
	}
}

func TestStateJiraTicket(t *testing.T) {
	s := NewState()

	if s.JiraTicket != nil {
		t.Fatal("expected nil JiraTicket on new state")
	}

	s.JiraTicket = &JiraTicket{
		Key:      "PROJ-123",
		Summary:  "Fix login",
		Status:   "In Progress",
		Assignee: "noah",
		Priority: "High",
	}

	if s.JiraTicket.Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", s.JiraTicket.Key)
	}
}

func TestStateUpdate(t *testing.T) {
	s := NewState()
	ac := NewAcceptanceCriterion("original title", "desc", PriorityLow)
	s.Add(ac)

	originalUpdatedAt := s.Items[0].UpdatedAt

	// Ensure time advances before updating
	time.Sleep(time.Millisecond)

	updated := ac
	updated.Title = "updated title"
	s.Update(updated)

	if len(s.Items) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(s.Items))
	}
	if s.Items[0].Title != "updated title" {
		t.Errorf("expected title %q, got %q", "updated title", s.Items[0].Title)
	}
	if !s.Items[0].UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("expected UpdatedAt to advance after Update; was %v, now %v", originalUpdatedAt, s.Items[0].UpdatedAt)
	}
}
