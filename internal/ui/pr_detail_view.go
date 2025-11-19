package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/ui/layout"
)

// PRDetailState represents the state of the PR detail view.
type PRDetailState int

const (
	PRDetailStateBrowsing PRDetailState = iota
	PRDetailStateConfirmAction
)

// PRAction represents an action that can be performed on a PR.
type PRAction struct {
	Action      domain.PRAction
	Label       string
	Description string
	RequiresConfirm bool
}

// PRDetailViewModel represents the PR detail view.
type PRDetailViewModel struct {
	pr                *domain.PRInfo
	actions           []PRAction
	selectedIndex     int
	state             PRDetailState
	returnToList      bool
	hasAction         bool
	selectedAction    *PRAction
	err               error
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
	confirmationBtn   int // 0: No, 1: Yes
}

// NewPRDetailViewModel creates a new PR detail view model.
func NewPRDetailViewModel(pr *domain.PRInfo) PRDetailViewModel {
	actions := buildPRActions(pr)

	// Initialize viewport with default size
	vp := viewport.New(50, 20)

	m := PRDetailViewModel{
		pr:              pr,
		actions:         actions,
		selectedIndex:   0,
		state:           PRDetailStateBrowsing,
		returnToList:    false,
		hasAction:       false,
		viewport:        vp,
		ready:           true,
		windowWidth:     120,
		windowHeight:    30,
		confirmationBtn: 0,
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderPRDetailsContent())

	return m
}

// buildPRActions builds available actions based on PR state.
func buildPRActions(pr *domain.PRInfo) []PRAction {
	var actions []PRAction

	switch pr.State() {
	case domain.PRStatusOpen:
		if pr.IsDraft() {
			// Draft PR actions
			actions = append(actions, PRAction{
				Action:          domain.PRActionMarkReady,
				Label:           "▸ Mark as ready for review",
				Description:     "Convert from draft to ready state",
				RequiresConfirm: false,
			})
		} else {
			// Ready PR actions
			actions = append(actions, PRAction{
				Action:          domain.PRActionConvertToDraft,
				Label:           "▸ Convert to draft",
				Description:     "Mark as work in progress",
				RequiresConfirm: false,
			})
		}

		// Merge actions (only for non-draft)
		if !pr.IsDraft() {
			actions = append(actions, PRAction{
				Action:          domain.PRActionMerge,
				Label:           "◆ Merge pull request",
				Description:     "Merge this PR into base branch",
				RequiresConfirm: true,
			})
		}

		// Close action
		actions = append(actions, PRAction{
			Action:          domain.PRActionClose,
			Label:           "✗ Close pull request",
			Description:     "Close without merging",
			RequiresConfirm: true,
		})

	case domain.PRStatusMerged:
		// No actions for merged PRs
		actions = append(actions, PRAction{
			Action:          domain.PRActionNone,
			Label:           "✓ Already merged",
			Description:     "This PR has been merged",
			RequiresConfirm: false,
		})

	case domain.PRStatusClosed:
		// No actions for closed PRs
		actions = append(actions, PRAction{
			Action:          domain.PRActionNone,
			Label:           "✗ Already closed",
			Description:     "This PR has been closed",
			RequiresConfirm: false,
		})
	}

	return actions
}

// Init initializes the PR detail view.
func (m PRDetailViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PRDetailViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Update viewport size
		cardWidth := msg.Width - 4
		if cardWidth < 80 {
			cardWidth = 80
		}
		innerWidth := cardWidth - 4
		viewportWidth := innerWidth - 2

		viewportHeight := msg.Height - 17
		if viewportHeight < 5 {
			viewportHeight = 5
		}

		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight

		// Refresh content
		m.viewport.SetContent(m.renderPRDetailsContent())

		return m, nil

	case tea.KeyMsg:
		// Handle confirmation state
		if m.state == PRDetailStateConfirmAction {
			switch msg.String() {
			case "left", "h":
				m.confirmationBtn = 0
				return m, nil
			case "right", "l":
				m.confirmationBtn = 1
				return m, nil
			case "tab":
				m.confirmationBtn = (m.confirmationBtn + 1) % 2
				return m, nil
			case "enter":
				if m.confirmationBtn == 1 {
					// Yes - execute action
					m.hasAction = true
					m.state = PRDetailStateBrowsing
				} else {
					// No - cancel
					m.state = PRDetailStateBrowsing
				}
				m.confirmationBtn = 0
				return m, nil
			case "esc":
				m.state = PRDetailStateBrowsing
				m.confirmationBtn = 0
				return m, nil
			}
			return m, nil
		}

		// Handle browsing state
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.viewport.SetContent(m.renderPRDetailsContent())
			}

		case "down", "j":
			if m.selectedIndex < len(m.actions)-1 {
				m.selectedIndex++
				m.viewport.SetContent(m.renderPRDetailsContent())
			}

		case "enter":
			// Execute selected action
			if len(m.actions) > 0 && m.selectedIndex < len(m.actions) {
				action := m.actions[m.selectedIndex]

				// Skip view-only actions
				if action.Action == domain.PRActionNone {
					return m, nil
				}

				m.selectedAction = &action

				if action.RequiresConfirm {
					// Show confirmation
					m.state = PRDetailStateConfirmAction
				} else {
					// Execute immediately
					m.hasAction = true
				}
			}
			return m, nil

		case "esc", "q":
			// Return to list
			m.returnToList = true
			return m, nil
		}
	}

	// Update viewport (handles scrolling)
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// View renders the PR detail view.
func (m PRDetailViewModel) View() string {
	styles := GetGlobalThemeManager().GetStyles()

	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Bold(true).
			Render(fmt.Sprintf("ERROR: %v\n", m.err))
	}

	if !m.ready {
		return lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Render("Loading pull request details...")
	}

	// Show confirmation modal if needed
	if m.state == PRDetailStateConfirmAction && m.selectedAction != nil {
		return m.renderConfirmationModal()
	}

	// Build layout
	var content strings.Builder

	// Logo
	content.WriteString(m.renderLogo())
	content.WriteString("\n\n")

	// PR header
	content.WriteString(m.renderPRHeader())
	content.WriteString("\n\n")

	// PR details and actions
	content.WriteString(m.renderPRContent())
	content.WriteString("\n\n")

	// Footer
	content.WriteString(m.renderFooter())

	return content.String()
}

// renderLogo renders the PR detail logo.
func (m PRDetailViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	ascii := `  ██████╗ ██████╗     ██████╗ ███████╗████████╗ █████╗ ██╗██╗
  ██╔══██╗██╔══██╗    ██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██║██║
  ██████╔╝██████╔╝    ██║  ██║█████╗     ██║   ███████║██║██║
  ██╔═══╝ ██╔══██╗    ██║  ██║██╔══╝     ██║   ██╔══██║██║██║
  ██║     ██║  ██║    ██████╔╝███████╗   ██║   ██║  ██║██║███████╗
  ╚═╝     ╚═╝  ╚═╝    ╚═════╝ ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚══════╝`

	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Modal)).
		Bold(true).
		Padding(0, layout.SpacingMD)

	return logoStyle.Render(ascii)
}

// renderPRHeader renders the PR header with number, title, and status.
func (m PRDetailViewModel) renderPRHeader() string {
	styles := GetGlobalThemeManager().GetStyles()

	// Status badge
	statusStyle := m.getStatusStyle()
	statusText := m.getStatusText()
	status := statusStyle.Render(fmt.Sprintf(" %s ", statusText))

	// PR number and title
	numberStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Bold(true)
	number := numberStyle.Render(fmt.Sprintf("#%d", m.pr.Number()))

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Bold(true)
	title := titleStyle.Render(m.pr.Title())

	return fmt.Sprintf("%s  %s %s", status, number, title)
}

// renderPRContent renders the PR details and actions.
func (m PRDetailViewModel) renderPRContent() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(lipgloss.Color(theme.Backgrounds.Modal)).
		Padding(layout.SpacingSM, layout.SpacingMD).
		Width(m.windowWidth - 4)

	return cardStyle.Render(m.viewport.View())
}

// renderPRDetailsContent renders the content inside the viewport.
func (m PRDetailViewModel) renderPRDetailsContent() string {
	var lines []string
	styles := GetGlobalThemeManager().GetStyles()

	// PR metadata
	lines = append(lines, m.renderMetadata()...)
	lines = append(lines, "")

	// Description
	if m.pr.Body() != "" {
		descStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		lines = append(lines, descStyle.Render("Description:"))

		bodyLines := strings.Split(m.pr.Body(), "\n")
		for _, line := range bodyLines {
			lines = append(lines, "  "+line)
		}
		lines = append(lines, "")
	}

	// Labels
	if len(m.pr.Labels()) > 0 {
		labelStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		lines = append(lines, labelStyle.Render("Labels:"))

		for _, label := range m.pr.Labels() {
			labelBadge := lipgloss.NewStyle().
				Foreground(styles.ColorPrimary).
				Bold(true).
				Render(fmt.Sprintf("  ▸ %s", label))
			lines = append(lines, labelBadge)
		}
		lines = append(lines, "")
	}

	// Actions section
	actionHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Bold(true)
	lines = append(lines, actionHeaderStyle.Render("Available Actions:"))
	lines = append(lines, "")

	// Render actions
	for i, action := range m.actions {
		isSelected := i == m.selectedIndex

		actionLine := m.formatActionLine(action, isSelected)

		if isSelected {
			selectedStyle := lipgloss.NewStyle().
				Foreground(styles.ColorPrimary).
				Bold(true)
			actionLine = selectedStyle.Render(actionLine)
		}

		lines = append(lines, actionLine)

		// Show description if selected
		if isSelected {
			descStyle := lipgloss.NewStyle().
				Foreground(styles.ColorMuted).
				Italic(true).
				PaddingLeft(layout.SpacingXL)
			lines = append(lines, descStyle.Render(action.Description))
		}
	}

	return strings.Join(lines, "\n")
}

// renderMetadata renders PR metadata (author, branches, etc.).
func (m PRDetailViewModel) renderMetadata() []string {
	styles := GetGlobalThemeManager().GetStyles()
	labelStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	valueStyle := lipgloss.NewStyle().Foreground(styles.ColorText)

	var lines []string

	// Author
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("Author:"),
		valueStyle.Render(m.pr.Author())))

	// Branches
	lines = append(lines, fmt.Sprintf("%s %s → %s",
		labelStyle.Render("Branches:"),
		valueStyle.Render(m.pr.HeadRef()),
		valueStyle.Render(m.pr.BaseRef())))

	// URL
	lines = append(lines, fmt.Sprintf("%s %s",
		labelStyle.Render("URL:"),
		valueStyle.Render(m.pr.HTMLURL())))

	return lines
}

// formatActionLine formats a single action line.
func (m PRDetailViewModel) formatActionLine(action PRAction, isSelected bool) string {
	var prefix string
	if isSelected {
		prefix = "▸ "
	} else {
		prefix = "  "
	}

	return prefix + action.Label
}

// getStatusText returns the text representation of PR status.
func (m PRDetailViewModel) getStatusText() string {
	switch m.pr.State() {
	case domain.PRStatusOpen:
		if m.pr.IsDraft() {
			return "DRAFT"
		}
		return "OPEN"
	case domain.PRStatusMerged:
		return "MERGED"
	case domain.PRStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// getStatusStyle returns the style for PR status badge.
func (m PRDetailViewModel) getStatusStyle() lipgloss.Style {
	styles := GetGlobalThemeManager().GetStyles()

	switch m.pr.State() {
	case domain.PRStatusOpen:
		if m.pr.IsDraft() {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(styles.ColorWarning).
				Bold(true)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(styles.ColorSuccess).
			Bold(true)
	case domain.PRStatusMerged:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(styles.ColorPrimary).
			Bold(true)
	case domain.PRStatusClosed:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(styles.ColorError).
			Bold(true)
	default:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(styles.ColorMuted).
			Bold(true)
	}
}

// renderConfirmationModal renders the confirmation modal.
func (m PRDetailViewModel) renderConfirmationModal() string {
	if m.selectedAction == nil {
		return ""
	}

	styles := GetGlobalThemeManager().GetStyles()

	// Calculate dimensions
	width := 60
	height := 10

	// Title
	title := styles.SectionTitle.Render("CONFIRM ACTION")

	// Action text
	actionText := strings.TrimPrefix(m.selectedAction.Label, "▸ ")
	actionText = strings.TrimPrefix(actionText, "◆ ")
	actionText = strings.TrimPrefix(actionText, "✗ ")
	message := fmt.Sprintf("Are you sure you want to %s?", strings.ToLower(actionText))

	// Buttons
	btnStyle := styles.TabInactive.Padding(0, 2)
	activeBtnStyle := styles.TabActive.Padding(0, 2)

	yesBtn := btnStyle.Render("Yes")
	if m.confirmationBtn == 1 {
		yesBtn = activeBtnStyle.Render("Yes")
	}

	noBtn := btnStyle.Render("No")
	if m.confirmationBtn == 0 {
		noBtn = activeBtnStyle.Render("No")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, noBtn, "  ", yesBtn)

	// Content
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		message,
		"",
		buttons,
	)

	// Box
	box := styles.CommitBox.
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Render(content)

	// Center on screen
	topPadding := (m.windowHeight - height) / 2
	leftPadding := (m.windowWidth - width) / 2

	paddedBox := lipgloss.NewStyle().
		PaddingTop(topPadding).
		PaddingLeft(leftPadding).
		Render(box)

	return paddedBox
}

// renderFooter renders the footer with keyboard shortcuts.
func (m PRDetailViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()
	help := "↑/↓: Navigate • Enter: Execute Action • Esc: Back to List"
	return styles.Footer.Render(help)
}

// ShouldReturnToList returns true if the view should return to list.
func (m PRDetailViewModel) ShouldReturnToList() bool {
	return m.returnToList
}

// HasAction returns true if an action should be executed.
func (m PRDetailViewModel) HasAction() bool {
	return m.hasAction
}

// GetSelectedAction returns the selected action for execution.
func (m PRDetailViewModel) GetSelectedAction() domain.PRAction {
	if m.selectedAction != nil {
		return m.selectedAction.Action
	}
	return domain.PRActionNone
}

// GetPR returns the PR being viewed.
func (m PRDetailViewModel) GetPR() *domain.PRInfo {
	return m.pr
}
