package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Tracker: &TrackerConfig{
			Provider: "jira",
			Domain:   "mycompany.atlassian.net",
			Email:    "user@example.com",
			APIToken: "test-api-token",
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Tracker == nil {
		t.Fatal("expected tracker config")
	}
	if loaded.Tracker.Provider != "jira" {
		t.Errorf("expected provider 'jira', got '%s'", loaded.Tracker.Provider)
	}
	if loaded.Tracker.Domain != "mycompany.atlassian.net" {
		t.Errorf("expected domain, got '%s'", loaded.Tracker.Domain)
	}
	if loaded.Tracker.Email != "user@example.com" {
		t.Errorf("expected email, got '%s'", loaded.Tracker.Email)
	}
	if loaded.Tracker.APIToken != "test-api-token" {
		t.Errorf("expected api token, got '%s'", loaded.Tracker.APIToken)
	}
}

func TestLoadNonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Tracker != nil {
		t.Error("expected nil tracker for fresh config")
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(path))
	}
}
