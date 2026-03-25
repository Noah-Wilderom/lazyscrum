package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Cycle  key.Binding
	Tab    key.Binding
	Enter  key.Binding
	Escape key.Binding
	Quit   key.Binding
	Jira   key.Binding
	Import key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Cycle: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter/space", "cycle status"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch pane"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Jira: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "jira tickets"),
	),
	Import: key.NewBinding(
		key.WithKeys("I"),
		key.WithHelp("I", "import AC"),
	),
}
