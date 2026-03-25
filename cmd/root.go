package cmd

import (
	"fmt"
	"os"

	"lazyscrum/internal/config"
	"lazyscrum/internal/store"
	"lazyscrum/internal/tracker"
	"lazyscrum/internal/tracker/jira"
	"lazyscrum/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lazyscrum",
	Short: "A TUI for managing SCRUM acceptance criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		storePath, err := store.ResolvePath(cwd)
		if err != nil {
			return fmt.Errorf("resolving store path: %w", err)
		}

		state, err := store.Load(storePath)
		if err != nil {
			return fmt.Errorf("loading state: %w", err)
		}

		// Try to load tracker from config (optional)
		var t tracker.Tracker
		cfgPath, err := config.DefaultPath()
		if err == nil {
			cfg, err := config.Load(cfgPath)
			if err == nil && cfg.Tracker != nil && cfg.Tracker.Provider == "jira" {
				t = jira.NewClient(cfg.Tracker.Domain, cfg.Tracker.Email, cfg.Tracker.APIToken)
			}
		}

		m := tui.New(state, storePath, cwd, t)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
