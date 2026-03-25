package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Noah-Wilderom/lazyscrum/internal/model"
)

// TestResolvePathFallback verifies that ResolvePath returns a non-empty path
// ending in "state.json" even when given a directory that is not a git repo.
func TestResolvePathFallback(t *testing.T) {
	dir := "/tmp/lazyscrum-test-not-a-repo"
	// Do not create the directory — git will fail regardless.

	path, err := ResolvePath(dir)
	if err != nil {
		t.Fatalf("ResolvePath returned unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("ResolvePath returned empty path")
	}
	if !strings.HasSuffix(path, "state.json") {
		t.Fatalf("expected path to end with state.json, got: %s", path)
	}
}

// TestSaveAndLoad verifies that a state written with Save can be read back
// correctly with Load and that all fields match.
func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	original := model.NewState()
	item := model.NewAcceptanceCriterion("Test title", "Test description", model.PriorityHigh)
	original.Add(item)

	if err := Save(statePath, original); err != nil {
		t.Fatalf("Save returned unexpected error: %v", err)
	}

	loaded, err := Load(statePath)
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	if len(loaded.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(loaded.Items))
	}

	got := loaded.Items[0]
	if got.ID != item.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, item.ID)
	}
	if got.Title != item.Title {
		t.Errorf("Title mismatch: got %q, want %q", got.Title, item.Title)
	}
	if got.Description != item.Description {
		t.Errorf("Description mismatch: got %q, want %q", got.Description, item.Description)
	}
	if got.Priority != item.Priority {
		t.Errorf("Priority mismatch: got %q, want %q", got.Priority, item.Priority)
	}
	if got.Status != item.Status {
		t.Errorf("Status mismatch: got %q, want %q", got.Status, item.Status)
	}
}

// TestLoadNonExistent verifies that loading from a path that does not exist
// returns an empty State (not nil) with no error.
func TestLoadNonExistent(t *testing.T) {
	path := "/tmp/lazyscrum-test-nonexistent-state/state.json"
	// Ensure the file does not exist.
	os.Remove(path)

	state, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned unexpected error for non-existent file: %v", err)
	}
	if state == nil {
		t.Fatal("Load returned nil state for non-existent file")
	}
	if len(state.Items) != 0 {
		t.Fatalf("expected empty state, got %d items", len(state.Items))
	}
}
