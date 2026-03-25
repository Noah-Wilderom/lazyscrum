package tui

import (
	"fmt"
	"strings"

	"lazyscrum/internal/model"
	"lazyscrum/internal/store"
	"lazyscrum/internal/tracker"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// pane identifies which pane is active.
type pane int

const (
	listPane pane = iota
	detailPane
)

// mode identifies the current interaction mode.
type mode int

const (
	modeNormal mode = iota
	modeCreate
	modeEdit
	modeDelete
	modeTicket
	modeImport
)

// formField identifies a field in the create/edit form.
type formField int

const (
	fieldTitle formField = iota
	fieldDescription
	fieldPriority
	fieldCount
)

// Model is the main Bubble Tea model for the TUI.
type Model struct {
	state      *model.State
	storePath  string
	projectDir string
	cursor    int
	pane      pane
	mode      mode
	width     int
	height    int

	// Form fields
	formFields    [fieldCount]textinput.Model
	activeField   formField
	editingID     string
	formPriority  model.Priority
	confirmDelete bool

	// Tracker integration
	tracker   tracker.Tracker
	ticketSel ticketSelector

	// Import acceptance criteria
	pendingAC      []string // AC detected from ticket description
	importSelected []bool   // which AC are selected for import
	importCursor   int
}

// New creates a new TUI model. The tracker parameter is optional (can be nil).
func New(state *model.State, storePath, projectDir string, t tracker.Tracker) Model {
	m := Model{
		state:        state,
		storePath:    storePath,
		projectDir:   projectDir,
		formPriority: model.PriorityMedium,
		tracker:      t,
	}
	m.initFormFields()
	return m
}

func (m *Model) initFormFields() {
	for i := 0; i < int(fieldCount); i++ {
		ti := textinput.New()
		switch formField(i) {
		case fieldTitle:
			ti.Placeholder = "Title"
			ti.Prompt = "  "
		case fieldDescription:
			ti.Placeholder = "Description"
			ti.Prompt = "  "
		case fieldPriority:
			ti.Placeholder = "high/medium/low"
			ti.Prompt = "  "
		}
		m.formFields[i] = ti
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case ticketsFetchedMsg:
		if msg.err != nil {
			m.ticketSel.setError(msg.err)
		} else {
			m.ticketSel.setIssues(msg.issues)
			// Pre-select autodetected ticket from branch
			branch := store.GitBranch(m.projectDir)
			issueKey := tracker.ExtractIssueKey(branch)
			if issueKey != "" {
				for i, issue := range m.ticketSel.filtered {
					if issue.Key == issueKey {
						m.ticketSel.cursor = i
						break
					}
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeNormal:
			return m.updateNormal(msg)
		case modeCreate, modeEdit:
			return m.updateForm(msg)
		case modeDelete:
			return m.updateDelete(msg)
		case modeTicket:
			return m.updateTicket(msg)
		case modeImport:
			return m.updateImport(msg)
		}
	}
	return m, nil
}

type ticketsFetchedMsg struct {
	issues []tracker.Issue
	err    error
}

func (m Model) fetchTickets() tea.Cmd {
	return func() tea.Msg {
		branch := store.GitBranch(m.projectDir)
		issueKey := tracker.ExtractIssueKey(branch)

		var issues []tracker.Issue
		var err error

		if issueKey != "" {
			parts := strings.SplitN(issueKey, "-", 2)
			if len(parts) == 2 {
				issues, err = m.tracker.ListProjectIssues(parts[0])
			}
		}

		if issues == nil && err == nil {
			issues, err = m.tracker.SearchIssues("updated >= -30d ORDER BY updated DESC")
		}

		return ticketsFetchedMsg{issues: issues, err: err}
	}
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.state.Items)-1 {
			m.cursor++
		}

	case key.Matches(msg, keys.New):
		m.mode = modeCreate
		m.pane = detailPane
		m.resetForm()
		m.formFields[fieldTitle].Focus()
		return m, nil

	case key.Matches(msg, keys.Edit):
		if len(m.state.Items) == 0 {
			return m, nil
		}
		item := m.state.Items[m.cursor]
		m.mode = modeEdit
		m.pane = detailPane
		m.editingID = item.ID
		m.formFields[fieldTitle].SetValue(item.Title)
		m.formFields[fieldDescription].SetValue(item.Description)
		m.formFields[fieldPriority].SetValue(string(item.Priority))
		m.formPriority = item.Priority
		m.activeField = fieldTitle
		m.formFields[fieldTitle].Focus()
		return m, nil

	case key.Matches(msg, keys.Delete):
		if len(m.state.Items) == 0 {
			return m, nil
		}
		m.mode = modeDelete
		m.confirmDelete = false

	case key.Matches(msg, keys.Tab):
		if m.pane == listPane {
			m.pane = detailPane
		} else {
			m.pane = listPane
		}

	case key.Matches(msg, keys.Cycle):
		if len(m.state.Items) > 0 {
			m.state.Items[m.cursor].CycleStatus()
			m.save()
		}

	case key.Matches(msg, keys.Import):
		if len(m.pendingAC) == 0 {
			return m, nil
		}
		m.mode = modeImport
		m.importCursor = 0
		m.importSelected = make([]bool, len(m.pendingAC))
		for i := range m.importSelected {
			m.importSelected[i] = true // select all by default
		}
		return m, nil

	case key.Matches(msg, keys.Jira):
		m.ticketSel = newTicketSelector()
		m.mode = modeTicket
		if m.tracker == nil {
			m.ticketSel.noConnection = true
			return m, nil
		}
		m.ticketSel.loading = true
		return m, m.fetchTickets()
	}
	return m, nil
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		m.pane = listPane
		return m, nil

	case msg.String() == "tab":
		m.formFields[m.activeField].Blur()
		m.activeField = (m.activeField + 1) % fieldCount
		m.formFields[m.activeField].Focus()
		return m, nil

	case msg.String() == "shift+tab":
		m.formFields[m.activeField].Blur()
		m.activeField = (m.activeField - 1 + fieldCount) % fieldCount
		m.formFields[m.activeField].Focus()
		return m, nil

	case key.Matches(msg, keys.Enter):
		if m.activeField < fieldCount-1 {
			// Advance to next field
			m.formFields[m.activeField].Blur()
			m.activeField++
			m.formFields[m.activeField].Focus()
			return m, nil
		}
		// Submit on last field
		return m.submitForm()
	}

	// Pass the key to the active text input
	var cmd tea.Cmd
	m.formFields[m.activeField], cmd = m.formFields[m.activeField].Update(msg)
	return m, cmd
}

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.formFields[fieldTitle].Value())
	if title == "" {
		return m, nil
	}

	description := strings.TrimSpace(m.formFields[fieldDescription].Value())
	priority := parsePriority(strings.TrimSpace(m.formFields[fieldPriority].Value()))

	if m.mode == modeCreate {
		ac := model.NewAcceptanceCriterion(title, description, priority)
		m.state.Add(ac)
		m.cursor = len(m.state.Items) - 1
	} else if m.mode == modeEdit {
		for i, item := range m.state.Items {
			if item.ID == m.editingID {
				m.state.Items[i].Title = title
				m.state.Items[i].Description = description
				m.state.Items[i].Priority = priority
				m.state.Update(m.state.Items[i])
				break
			}
		}
	}

	m.save()
	m.mode = modeNormal
	m.pane = listPane
	return m, nil
}

func (m Model) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if len(m.state.Items) > 0 {
			m.state.Remove(m.state.Items[m.cursor].ID)
			if m.cursor >= len(m.state.Items) && m.cursor > 0 {
				m.cursor--
			}
			m.save()
		}
		m.mode = modeNormal

	case "n", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) updateTicket(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.ticketSel.cursor > 0 {
			m.ticketSel.cursor--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.ticketSel.cursor < len(m.ticketSel.filtered)-1 {
			m.ticketSel.cursor++
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		issue := m.ticketSel.selectedIssue()
		if issue != nil {
			m.state.JiraTicket = m.ticketSel.toJiraTicket()
			m.pendingAC = issue.AcceptanceCriteria
			m.save()
		}
		m.mode = modeNormal
		return m, nil
	}

	// Pass to search input for filtering
	var cmd tea.Cmd
	m.ticketSel.searchInput, cmd = m.ticketSel.searchInput.Update(msg)
	m.ticketSel.filter()
	return m, cmd
}

func (m Model) updateImport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.mode = modeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.importCursor > 0 {
			m.importCursor--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.importCursor < len(m.pendingAC)-1 {
			m.importCursor++
		}
		return m, nil

	case msg.String() == " ":
		// Toggle selection
		m.importSelected[m.importCursor] = !m.importSelected[m.importCursor]
		return m, nil

	case key.Matches(msg, keys.Enter):
		// Import selected criteria
		for i, ac := range m.pendingAC {
			if m.importSelected[i] {
				m.state.Add(model.NewAcceptanceCriterion(ac, "", model.PriorityMedium))
			}
		}
		m.pendingAC = nil
		m.importSelected = nil
		m.cursor = 0
		m.save()
		m.mode = modeNormal
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate available dimensions
	availWidth := m.width - 4 // account for appStyle padding
	availHeight := m.height - 4 // account for appStyle padding, title, and help

	// Title bar
	title := titleStyle.Render("LazyScrum")

	// Calculate pane widths
	listWidth := availWidth * 40 / 100
	detailWidth := availWidth - listWidth

	// Calculate pane height (subtract title and help bar lines)
	paneHeight := availHeight - 3
	if paneHeight < 3 {
		paneHeight = 3
	}

	// Render panes
	listContent := m.renderList(listWidth-4, paneHeight-2) // account for border + padding
	detailContent := m.renderDetail(detailWidth-6, paneHeight-2) // account for border + padding

	var listPaneView, detailPaneView string
	if m.pane == listPane {
		listPaneView = activePaneBorderStyle.Width(listWidth - 4).Height(paneHeight - 2).Render(listContent)
		detailPaneView = inactivePaneBorderStyle.Width(detailWidth - 4).Height(paneHeight - 2).Render(detailContent)
	} else {
		listPaneView = inactivePaneBorderStyle.Width(listWidth - 4).Height(paneHeight - 2).Render(listContent)
		detailPaneView = activePaneBorderStyle.Width(detailWidth - 4).Height(paneHeight - 2).Render(detailContent)
	}

	panes := lipgloss.JoinHorizontal(lipgloss.Top, listPaneView, detailPaneView)

	// Help bar
	help := m.renderHelp()

	content := lipgloss.JoinVertical(lipgloss.Left, title, panes, help)
	rendered := appStyle.Render(content)

	// Render import overlay
	if m.mode == modeImport {
		olWidth := m.width * 70 / 100
		olHeight := m.height * 60 / 100
		olContent := m.renderImport(olWidth-6, olHeight-4)
		overlay := overlayStyle.Width(olWidth - 4).Height(olHeight - 4).Render(olContent)

		x := (m.width - olWidth) / 2
		y := (m.height - olHeight) / 2
		rendered = placeOverlay(x, y, overlay, rendered)
	}

	// Render ticket selector overlay
	if m.mode == modeTicket {
		olWidth := m.width * 70 / 100
		olHeight := m.height * 60 / 100
		olContent := m.ticketSel.view(olWidth-6, olHeight-4)
		overlay := overlayStyle.Width(olWidth - 4).Height(olHeight - 4).Render(olContent)

		x := (m.width - olWidth) / 2
		y := (m.height - olHeight) / 2
		rendered = placeOverlay(x, y, overlay, rendered)
	}

	return rendered
}

func placeOverlay(x, y int, overlay, background string) string {
	bgLines := strings.Split(background, "\n")
	olLines := strings.Split(overlay, "\n")

	for i, olLine := range olLines {
		bgIdx := y + i
		if bgIdx < 0 || bgIdx >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgIdx]
		bgRunes := []rune(bgLine)

		for len(bgRunes) < x+lipgloss.Width(olLine) {
			bgRunes = append(bgRunes, ' ')
		}

		before := string(bgRunes[:x])
		after := ""
		afterStart := x + lipgloss.Width(olLine)
		if afterStart < len(bgRunes) {
			after = string(bgRunes[afterStart:])
		}

		bgLines[bgIdx] = before + olLine + after
	}

	return strings.Join(bgLines, "\n")
}

func (m Model) renderList(width, height int) string {
	if len(m.state.Items) == 0 {
		return normalItemStyle.Render("No acceptance criteria yet.\nPress 'n' to create one.")
	}

	var lines []string
	for i, item := range m.state.Items {
		statusIcon := statusIcon(item.Status)
		priorityIcon := priorityIcon(item.Priority)

		var line string
		if i == m.cursor {
			line = selectedItemStyle.Render(fmt.Sprintf("▸ %s %s %s", statusIcon, priorityIcon, item.Title))
		} else {
			line = normalItemStyle.Render(fmt.Sprintf("  %s %s %s", statusIcon, priorityIcon, item.Title))
		}

		if width > 0 && lipgloss.Width(line) > width {
			line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		}

		lines = append(lines, line)
		if len(lines) >= height {
			break
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderDetail(width, height int) string {
	switch m.mode {
	case modeCreate:
		return m.renderForm("New Acceptance Criterion")
	case modeEdit:
		return m.renderForm("Edit Acceptance Criterion")
	case modeDelete:
		if len(m.state.Items) == 0 {
			return ""
		}
		item := m.state.Items[m.cursor]
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			labelStyle.Render("Delete this item?"),
			normalItemStyle.Render(item.Title),
			helpStyle.Render("y/n"),
		)
	default:
		return m.renderDetailNormal()
	}
}

func (m Model) renderDetailNormal() string {
	if len(m.state.Items) == 0 {
		return helpStyle.Render("Select an item to view details.")
	}

	item := m.state.Items[m.cursor]

	var b strings.Builder

	if m.state.JiraTicket != nil {
		b.WriteString(labelStyle.Render("Linked Ticket: "))
		b.WriteString(ticketKeyStyle.Render(m.state.JiraTicket.Key))
		b.WriteString(normalItemStyle.Render(" " + m.state.JiraTicket.Summary))
		b.WriteString("\n\n")
	}

	b.WriteString(selectedItemStyle.Render(item.Title))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Status: "))
	b.WriteString(statusText(item.Status))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Priority: "))
	b.WriteString(priorityText(item.Priority))
	b.WriteString("\n\n")

	if item.Description != "" {
		b.WriteString(labelStyle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(normalItemStyle.Render(item.Description))
		b.WriteString("\n\n")
	}

	b.WriteString(labelStyle.Render("Created: "))
	b.WriteString(helpStyle.Render(item.CreatedAt.Format("2006-01-02 15:04")))

	return b.String()
}

func (m Model) renderForm(title string) string {
	var b strings.Builder

	b.WriteString(selectedItemStyle.Render(title))
	b.WriteString("\n\n")

	labels := [fieldCount]string{"Title", "Description", "Priority"}
	for i := 0; i < int(fieldCount); i++ {
		b.WriteString(labelStyle.Render(labels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(m.formFields[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("tab/shift+tab: cycle fields • enter: next/submit • esc: cancel"))

	return b.String()
}

func (m Model) renderImport(width, height int) string {
	var b strings.Builder

	b.WriteString(overlayTitleStyle.Render("Import Acceptance Criteria"))
	b.WriteString("\n\n")

	if len(m.pendingAC) == 0 {
		b.WriteString(helpStyle.Render("No acceptance criteria found in ticket."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: close"))
		return b.String()
	}

	maxVisible := height - 6
	if maxVisible < 3 {
		maxVisible = 3
	}

	start := 0
	if m.importCursor >= maxVisible {
		start = m.importCursor - maxVisible + 1
	}

	for i := start; i < len(m.pendingAC) && i < start+maxVisible; i++ {
		checkbox := "[ ]"
		if m.importSelected[i] {
			checkbox = "[x]"
		}

		text := m.pendingAC[i]
		maxTextWidth := width - 10
		if maxTextWidth > 0 && len(text) > maxTextWidth {
			text = text[:maxTextWidth-1] + "…"
		}

		line := fmt.Sprintf("%s %s", checkbox, text)

		if i == m.importCursor {
			line = selectedItemStyle.Render("▸ " + line)
		} else {
			line = normalItemStyle.Render("  " + line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: navigate • space: toggle • enter: import selected • esc: cancel"))

	return b.String()
}

func (m Model) renderHelp() string {
	if m.mode != modeNormal {
		return ""
	}

	parts := []string{
		fmt.Sprintf("%s %s", helpStyle.Render("↑/k"), helpStyle.Render("up")),
		fmt.Sprintf("%s %s", helpStyle.Render("↓/j"), helpStyle.Render("down")),
		fmt.Sprintf("%s %s", helpStyle.Render("n"), helpStyle.Render("new")),
		fmt.Sprintf("%s %s", helpStyle.Render("e"), helpStyle.Render("edit")),
		fmt.Sprintf("%s %s", helpStyle.Render("d"), helpStyle.Render("delete")),
		fmt.Sprintf("%s %s", helpStyle.Render("enter/space"), helpStyle.Render("cycle")),
		fmt.Sprintf("%s %s", helpStyle.Render("tab"), helpStyle.Render("switch pane")),
		fmt.Sprintf("%s %s", helpStyle.Render("J"), helpStyle.Render("jira")),
	}

	if len(m.pendingAC) > 0 {
		parts = append(parts, fmt.Sprintf("%s %s", helpStyle.Render("I"), helpStyle.Render("import AC")))
	}

	parts = append(parts, fmt.Sprintf("%s %s", helpStyle.Render("q"), helpStyle.Render("quit")))

	return helpStyle.Render(strings.Join(parts, "  •  "))
}

func (m *Model) resetForm() {
	for i := 0; i < int(fieldCount); i++ {
		m.formFields[i].SetValue("")
		m.formFields[i].Blur()
	}
	m.activeField = fieldTitle
	m.editingID = ""
	m.formPriority = model.PriorityMedium
}

func (m *Model) save() {
	_ = store.Save(m.storePath, m.state)
}

// parsePriority parses a string into a Priority value.
func parsePriority(s string) model.Priority {
	switch strings.ToLower(s) {
	case "high", "h":
		return model.PriorityHigh
	case "low", "l":
		return model.PriorityLow
	default:
		return model.PriorityMedium
	}
}

// statusIcon returns a status indicator icon.
func statusIcon(s model.Status) string {
	switch s {
	case model.StatusTodo:
		return statusTodoStyle.Render("○")
	case model.StatusInProgress:
		return statusInProgressStyle.Render("◐")
	case model.StatusDone:
		return statusDoneStyle.Render("●")
	default:
		return "○"
	}
}

// priorityIcon returns a priority indicator icon.
func priorityIcon(p model.Priority) string {
	switch p {
	case model.PriorityHigh:
		return priorityHighStyle.Render("▲")
	case model.PriorityMedium:
		return priorityMediumStyle.Render("■")
	case model.PriorityLow:
		return priorityLowStyle.Render("▽")
	default:
		return "■"
	}
}

// statusText returns a styled string for a status.
func statusText(s model.Status) string {
	switch s {
	case model.StatusTodo:
		return statusTodoStyle.Render("Todo")
	case model.StatusInProgress:
		return statusInProgressStyle.Render("In Progress")
	case model.StatusDone:
		return statusDoneStyle.Render("Done")
	default:
		return string(s)
	}
}

// priorityText returns a styled string for a priority.
func priorityText(p model.Priority) string {
	switch p {
	case model.PriorityHigh:
		return priorityHighStyle.Render("High")
	case model.PriorityMedium:
		return priorityMediumStyle.Render("Medium")
	case model.PriorityLow:
		return priorityLowStyle.Render("Low")
	default:
		return string(p)
	}
}
