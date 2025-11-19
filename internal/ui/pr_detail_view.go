package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// PRDetailViewModel represents the state of the PR detail view.
type PRDetailViewModel struct {
	pr                *domain.PRInfo
	action            string // "", "update", "close", "merge", "convert-to-draft", "mark-ready"
	returnToList      bool
	returnToDashboard bool
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
	repoPath          string

	// Update form state
	updateMode        bool
	titleInput        textinput.Model
	bodyInput         textinput.Model
	currentInputField int // 0: title, 1: body
}

// NewPRDetailViewModel creates a new PR detail view model.
func NewPRDetailViewModel(pr *domain.PRInfo, repoPath string) PRDetailViewModel {
	// Initialize viewport
	vp := viewport.New(50, 20)

	// Initialize text inputs for update mode
	titleInput := textinput.New()
	titleInput.CharLimit = 72
	titleInput.Width = 70
	titleInput.Placeholder = "PR Title"
	if pr != nil {
		titleInput.SetValue(pr.Title())
	}

	bodyInput := textinput.New()
	bodyInput.CharLimit = 500
	bodyInput.Width = 70
	bodyInput.Placeholder = "PR Description"
	if pr != nil {
		bodyInput.SetValue(pr.Body())
	}

	m := PRDetailViewModel{
		pr:                pr,
		action:            "",
		returnToList:      false,
		returnToDashboard: false,
		viewport:          vp,
		ready:             true,
		windowWidth:       120,
		windowHeight:      30,
		repoPath:          repoPath,
		updateMode:        false,
		titleInput:        titleInput,
		bodyInput:         bodyInput,
		currentInputField: 0,
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderPRDetailContent())

	return m
}

// Init initializes the PR detail view.
func (m PRDetailViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the PR detail view.
func (m PRDetailViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Update viewport size
		headerHeight := 8 // Logo + PR header
		footerHeight := 3
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - headerHeight - footerHeight

		// Update content
		if !m.updateMode {
			m.viewport.SetContent(m.renderPRDetailContent())
		}

		if !m.ready {
			m.ready = true
		}

		return m, nil

	case tea.KeyMsg:
		if m.updateMode {
			return m.handleUpdateModeKeys(msg)
		}
		return m.handleViewModeKeys(msg)
	}

	// Update viewport
	if !m.updateMode {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	// Update text inputs in update mode
	var cmd tea.Cmd
	if m.currentInputField == 0 {
		m.titleInput, cmd = m.titleInput.Update(msg)
	} else {
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
}

// handleViewModeKeys handles keyboard input in viewing mode.
func (m PRDetailViewModel) handleViewModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.returnToList = true
		return m, nil

	case "u":
		// Enter update mode
		m.updateMode = true
		m.currentInputField = 0
		m.titleInput.Focus()
		return m, nil

	case "c":
		// Close PR
		m.action = "close"
		return m, nil

	case "m":
		// Merge PR
		if m.pr.State() == domain.PRStatusOpen && !m.pr.IsDraft() {
			m.action = "merge"
		}
		return m, nil

	case "d":
		// Toggle draft status
		if m.pr.IsDraft() {
			m.action = "mark-ready"
		} else {
			m.action = "convert-to-draft"
		}
		return m, nil

	case "up", "k":
		m.viewport.ScrollUp(1)
		return m, nil

	case "down", "j":
		m.viewport.ScrollDown(1)
		return m, nil

	case "pgup":
		m.viewport.PageUp()
		return m, nil

	case "pgdown":
		m.viewport.PageDown()
		return m, nil
	}

	return m, nil
}

// handleUpdateModeKeys handles keyboard input in update mode.
func (m PRDetailViewModel) handleUpdateModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit update mode
		m.updateMode = false
		m.titleInput.Blur()
		m.bodyInput.Blur()
		return m, nil

	case "tab":
		// Switch input field
		if m.currentInputField == 0 {
			m.titleInput.Blur()
			m.bodyInput.Focus()
			m.currentInputField = 1
		} else {
			m.bodyInput.Blur()
			m.titleInput.Focus()
			m.currentInputField = 0
		}
		return m, nil

	case "enter":
		// Save update
		m.action = "update"
		m.updateMode = false
		m.titleInput.Blur()
		m.bodyInput.Blur()
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.currentInputField == 0 {
		m.titleInput, cmd = m.titleInput.Update(msg)
	} else {
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
}

// View renders the PR detail view.
func (m PRDetailViewModel) View() string {
	if !m.ready || m.pr == nil {
		return "Loading..."
	}

	styles := GetGlobalThemeManager().GetStyles()

	// Render logo
	logo := m.renderLogo()

	// Render PR header
	prHeader := m.renderPRHeader()

	var content string
	if m.updateMode {
		content = m.renderUpdateForm()
	} else {
		content = styles.ViewportStyle.Render(m.viewport.View())
	}

	// Render footer
	footer := m.renderFooter()

	// Combine sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		"",
		prHeader,
		"",
		content,
		"",
		footer,
	)
}

// ShouldReturnToList returns whether the view wants to return to PR list.
func (m PRDetailViewModel) ShouldReturnToList() bool {
	return m.returnToList
}

// ShouldReturnToDashboard returns whether the view wants to return to dashboard.
func (m PRDetailViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// GetAction returns the current action to perform.
func (m PRDetailViewModel) GetAction() string {
	return m.action
}

// GetUpdatedTitle returns the updated PR title.
func (m PRDetailViewModel) GetUpdatedTitle() string {
	return m.titleInput.Value()
}

// GetUpdatedBody returns the updated PR body.
func (m PRDetailViewModel) GetUpdatedBody() string {
	return m.bodyInput.Value()
}

// renderLogo renders the PR detail logo.
func (m PRDetailViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	logo := styles.Header.Render("PULL REQUEST DETAILS")
	if m.repoPath != "" {
		repoInfo := styles.RepoLabel.Render("Repository: ") + styles.RepoValue.Render(m.repoPath)
		return lipgloss.JoinVertical(lipgloss.Left, logo, repoInfo)
	}
	return logo
}

// renderPRHeader renders the PR header with number, title, and state.
func (m PRDetailViewModel) renderPRHeader() string {
	styles := GetGlobalThemeManager().GetStyles()

	// PR number and state
	stateStr := m.getStateString()
	stateStyle := m.getStateStyle()

	header := fmt.Sprintf("%s #%d: %s %s",
		styles.Header.Render("Pull Request"),
		m.pr.Number(),
		m.pr.Title(),
		stateStyle.Render(stateStr),
	)

	return header
}

// renderPRDetailContent renders the PR detail content for the viewport.
func (m PRDetailViewModel) renderPRDetailContent() string {
	styles := GetGlobalThemeManager().GetStyles()

	var lines []string

	// Basic info
	lines = append(lines, styles.SectionTitle.Render("Details"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Author:     %s", m.pr.Author()))
	lines = append(lines, fmt.Sprintf("  Base:       %s", m.pr.BaseRef()))
	lines = append(lines, fmt.Sprintf("  Head:       %s", m.pr.HeadRef()))
	lines = append(lines, fmt.Sprintf("  State:      %s", m.getStateString()))
	if m.pr.IsDraft() {
		lines = append(lines, fmt.Sprintf("  Draft:      %s", styles.StatusWarning.Render("Yes")))
	}
	lines = append(lines, "")

	// Labels
	if len(m.pr.Labels()) > 0 {
		lines = append(lines, styles.SectionTitle.Render("Labels"))
		lines = append(lines, "")
		for _, label := range m.pr.Labels() {
			lines = append(lines, fmt.Sprintf("  • %s", label))
		}
		lines = append(lines, "")
	}

	// Description
	lines = append(lines, styles.SectionTitle.Render("Description"))
	lines = append(lines, "")
	if m.pr.Body() != "" {
		bodyLines := strings.Split(m.pr.Body(), "\n")
		for _, line := range bodyLines {
			lines = append(lines, "   "+line)
		}
	} else {
		mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
		lines = append(lines, mutedStyle.Render("   No description provided"))
	}
	lines = append(lines, "")

	// URL
	lines = append(lines, styles.SectionTitle.Render("URL"))
	lines = append(lines, "")
	lines = append(lines, "   "+m.pr.HTMLURL())

	return strings.Join(lines, "\n")
}

// renderUpdateForm renders the update form.
func (m PRDetailViewModel) renderUpdateForm() string {
	styles := GetGlobalThemeManager().GetStyles()

	var lines []string

	lines = append(lines, styles.SectionTitle.Render("Update Pull Request"))
	lines = append(lines, "")
	lines = append(lines, styles.FormLabel.Render("Title:"))
	lines = append(lines, m.titleInput.View())
	lines = append(lines, "")
	lines = append(lines, styles.FormLabel.Render("Description:"))
	lines = append(lines, m.bodyInput.View())
	lines = append(lines, "")
	lines = append(lines, styles.FormHelp.Render("Press Tab to switch fields, Enter to save, Esc to cancel"))

	return strings.Join(lines, "\n")
}

// getStateString returns the string representation of the PR state.
func (m PRDetailViewModel) getStateString() string {
	if m.pr.IsDraft() {
		return "Draft"
	}
	switch m.pr.State() {
	case domain.PRStatusMerged:
		return "Merged"
	case domain.PRStatusClosed:
		return "Closed"
	case domain.PRStatusOpen:
		return "Open"
	default:
		return "Unknown"
	}
}

// getStateStyle returns the style for the PR state.
func (m PRDetailViewModel) getStateStyle() lipgloss.Style {
	styles := GetGlobalThemeManager().GetStyles()

	if m.pr.IsDraft() {
		return styles.StatusWarning
	}
	switch m.pr.State() {
	case domain.PRStatusMerged:
		return styles.StatusInfo
	case domain.PRStatusClosed:
		return styles.StatusError
	case domain.PRStatusOpen:
		return styles.StatusOk
	default:
		return lipgloss.NewStyle().Foreground(styles.ColorMuted)
	}
}

// renderFooter renders the footer with keyboard shortcuts.
func (m PRDetailViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()

	var help string
	if m.updateMode {
		help = "tab: next field • enter: save • esc: cancel"
	} else {
		help = "↑↓: scroll • u: update • c: close • m: merge • d: toggle draft • esc: back"
	}

	metadata := fmt.Sprintf("PR #%d", m.pr.Number())

	footer := styles.Footer.Render(help)
	if metadata != "" {
		footer = footer + " " + styles.Metadata.Render(metadata)
	}

	return footer
}
