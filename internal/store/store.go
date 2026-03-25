package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lazyscrum/internal/model"
)

// hashString returns the hex-encoded SHA256 hash of s, taking only the first
// 8 bytes (16 hex characters) to keep directory names concise.
func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

// gitBranch runs `git rev-parse --abbrev-ref HEAD` in dir and returns the
// branch name. Returns "default" if the command fails (e.g. not a git repo).
func GitBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "default"
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "default"
	}
	return branch
}

// ResolvePath builds the path for the state file:
//
//	~/.config/lazyscrum/<sha256(projectDir)>/<sha256(branch)>/state.json
func ResolvePath(projectDir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	branch := GitBranch(projectDir)

	base := filepath.Join(
		home,
		".config",
		"lazyscrum",
		hashString(projectDir),
		hashString(branch),
		"state.json",
	)
	return base, nil
}

// Save marshals state as indented JSON and writes it to path, creating all
// intermediate directories as needed.
func Save(path string, state *model.State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Load reads state from path and unmarshals it. If the file does not exist,
// it returns an empty NewState() with no error.
func Load(path string) (*model.State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.NewState(), nil
		}
		return nil, err
	}

	var state model.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
