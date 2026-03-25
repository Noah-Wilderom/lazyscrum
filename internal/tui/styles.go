package tui

import "github.com/charmbracelet/lipgloss"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6C63FF")).
			Padding(0, 1)

	activePaneBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6C63FF")).
				Padding(0, 1)

	inactivePaneBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#555555")).
				Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C63FF")).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD"))

	statusTodoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	statusInProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD93D"))

	statusDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6BCB77"))

	priorityHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true)

	priorityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD93D"))

	priorityLowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	detailPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 2)

	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6C63FF")).
			Padding(1, 2)

	overlayTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C63FF")).
				Bold(true)

	ticketKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C63FF")).
			Bold(true).
			Width(12)

	ticketStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD93D")).
				Width(14)

	ticketAssigneeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Width(12)
)
