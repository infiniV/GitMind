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

// PRListState represents the state of the PR list view.
type PRListState int

const (
	PRListStateBrowsing PRListState = iota
	PRListStateFiltering
)

// PRListViewModel represents the PR list view.
type PRListViewModel struct {
	prs               []*domain.PRInfo
	filteredPRs       []*domain.PRInfo
	selectedIndex     int
	currentFilter     string // "all", "open", "closed", "merged", "draft"
	state             PRListState
	returnToDashboard bool
	viewPRDetail      bool
	selectedPR        *domain.PRInfo
	err               error
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
}

// NewPRListViewModel creates a new PR list view model.
func NewPRListViewModel(prs []*domain.PRInfo) PRListViewModel {
	// Initialize viewport with default size
	vp := viewport.New(50, 20)

	m := PRListViewModel{
		prs:               prs,
		filteredPRs:       prs, // Initially show all
		selectedIndex:     0,
		currentFilter:     "all",
		state:             PRListStateBrowsing,
		returnToDashboard: false,
		viewPRDetail:      false,
		viewport:          vp,
		ready:             true,
		windowWidth:       120,
		windowHeight:      30,
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderPRListContent())

	return m
}

// Init initializes the PR list view.
func (m PRListViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m PRListViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.viewport.SetContent(m.renderPRListContent())

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.viewport.SetContent(m.renderPRListContent())
			}

		case "down", "j":
			if m.selectedIndex < len(m.filteredPRs)-1 {
				m.selectedIndex++
				m.viewport.SetContent(m.renderPRListContent())
			}

		case "enter":
			// View PR detail
			if len(m.filteredPRs) > 0 && m.selectedIndex < len(m.filteredPRs) {
				m.selectedPR = m.filteredPRs[m.selectedIndex]
				m.viewPRDetail = true
			}
			return m, nil

		case "f":
			// Cycle through filters
			m.cycleFilter()
			m.applyFilter()
			m.selectedIndex = 0
			m.viewport.SetContent(m.renderPRListContent())
			return m, nil

		case "r":
			// Refresh (signal to parent to reload)
			// Parent will handle reloading
			return m, nil

		case "esc", "q":
			// Return to dashboard
			m.returnToDashboard = true
			return m, nil
		}
	}

	// Update viewport (handles scrolling)
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// View renders the PR list view.
func (m PRListViewModel) View() string {
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
			Render("Loading pull requests...")
	}

	// Build layout
	var content strings.Builder

	// Logo
	content.WriteString(m.renderLogo())
	content.WriteString("\n\n")

	// Header with filter info
	content.WriteString(m.renderHeader())
	content.WriteString("\n\n")

	// PR list
	content.WriteString(m.renderPRList())
	content.WriteString("\n\n")

	// Footer
	content.WriteString(m.renderFooter())

	return content.String()
}

// renderLogo renders the PR list logo.
func (m PRListViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	ascii := `  ██████╗ ██████╗     ██╗     ██╗███████╗████████╗
  ██╔══██╗██╔══██╗    ██║     ██║██╔════╝╚══██╔══╝
  ██████╔╝██████╔╝    ██║     ██║███████╗   ██║
  ██╔═══╝ ██╔══██╗    ██║     ██║╚════██║   ██║
  ██║     ██║  ██║    ███████╗██║███████║   ██║
  ╚═╝     ╚═╝  ╚═╝    ╚══════╝╚═╝╚══════╝   ╚═╝`

	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Modal)).
		Bold(true).
		Padding(0, layout.SpacingMD)

	return logoStyle.Render(ascii)
}

// renderHeader renders the header with filter and count info.
func (m PRListViewModel) renderHeader() string {
	styles := GetGlobalThemeManager().GetStyles()

	filterText := fmt.Sprintf("Filter: %s", m.currentFilter)
	countText := fmt.Sprintf("Showing %d of %d PRs", len(m.filteredPRs), len(m.prs))

	headerStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted)

	return headerStyle.Render(filterText) + "  " + mutedStyle.Render(countText)
}

// renderPRList renders the scrollable PR list.
func (m PRListViewModel) renderPRList() string {
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

// renderPRListContent renders the content inside the viewport.
func (m PRListViewModel) renderPRListContent() string {
	if len(m.filteredPRs) == 0 {
		styles := GetGlobalThemeManager().GetStyles()
		return lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Italic(true).
			Render("No pull requests found")
	}

	var lines []string
	styles := GetGlobalThemeManager().GetStyles()

	for i, pr := range m.filteredPRs {
		isSelected := i == m.selectedIndex

		// Build PR line
		prLine := m.formatPRLine(pr, isSelected)

		if isSelected {
			// Highlight selected item
			selectedStyle := lipgloss.NewStyle().
				Foreground(styles.ColorPrimary).
				Bold(true)
			prLine = selectedStyle.Render(prLine)
		}

		lines = append(lines, prLine)

		// Add description preview if selected
		if isSelected && pr.Body() != "" {
			preview := pr.Body()
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
			previewStyle := lipgloss.NewStyle().
				Foreground(styles.ColorMuted).
				Italic(true).
				PaddingLeft(layout.SpacingLG)
			lines = append(lines, previewStyle.Render(preview))
		}
	}

	return strings.Join(lines, "\n")
}

// formatPRLine formats a single PR line.
func (m PRListViewModel) formatPRLine(pr *domain.PRInfo, isSelected bool) string {
	styles := GetGlobalThemeManager().GetStyles()

	// Status indicator
	statusIndicator := m.getStatusIndicator(pr)
	statusStyle := m.getStatusStyle(pr)

	// PR number
	numberText := fmt.Sprintf("#%-4d", pr.Number())

	// Title (truncate if needed)
	title := pr.Title()
	maxTitleLen := 60
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen] + "..."
	}

	// Author
	author := pr.Author()
	if len(author) > 15 {
		author = author[:15] + "..."
	}

	// Build line
	var parts []string

	// Selection indicator
	if isSelected {
		parts = append(parts, "▸")
	} else {
		parts = append(parts, " ")
	}

	// Status
	parts = append(parts, statusStyle.Render(statusIndicator))

	// Number
	numberStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	parts = append(parts, numberStyle.Render(numberText))

	// Title
	parts = append(parts, title)

	// Author
	authorStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	parts = append(parts, authorStyle.Render(fmt.Sprintf("by %s", author)))

	return strings.Join(parts, " ")
}

// getStatusIndicator returns the Unicode indicator for PR status.
func (m PRListViewModel) getStatusIndicator(pr *domain.PRInfo) string {
	switch pr.State() {
	case domain.PRStatusOpen:
		if pr.IsDraft() {
			return "◇"
		}
		return "◆"
	case domain.PRStatusMerged:
		return "✓"
	case domain.PRStatusClosed:
		return "✗"
	default:
		return "◦"
	}
}

// getStatusStyle returns the style for PR status.
func (m PRListViewModel) getStatusStyle(pr *domain.PRInfo) lipgloss.Style {
	styles := GetGlobalThemeManager().GetStyles()

	switch pr.State() {
	case domain.PRStatusOpen:
		if pr.IsDraft() {
			return lipgloss.NewStyle().Foreground(styles.ColorWarning)
		}
		return lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	case domain.PRStatusMerged:
		return lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	case domain.PRStatusClosed:
		return lipgloss.NewStyle().Foreground(styles.ColorError)
	default:
		return lipgloss.NewStyle().Foreground(styles.ColorMuted)
	}
}

// renderFooter renders the footer with keyboard shortcuts.
func (m PRListViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()
	help := "↑/↓: Navigate • Enter: View Details • F: Filter • R: Refresh • Esc: Back"
	return styles.Footer.Render(help)
}

// cycleFilter cycles through available filters.
func (m *PRListViewModel) cycleFilter() {
	filters := []string{"all", "open", "draft", "merged", "closed"}
	currentIdx := 0

	for i, f := range filters {
		if f == m.currentFilter {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(filters)
	m.currentFilter = filters[nextIdx]
}

// applyFilter applies the current filter to the PR list.
func (m *PRListViewModel) applyFilter() {
	if m.currentFilter == "all" {
		m.filteredPRs = m.prs
		return
	}

	var filtered []*domain.PRInfo
	for _, pr := range m.prs {
		switch m.currentFilter {
		case "open":
			if pr.State() == domain.PRStatusOpen && !pr.IsDraft() {
				filtered = append(filtered, pr)
			}
		case "draft":
			if pr.State() == domain.PRStatusOpen && pr.IsDraft() {
				filtered = append(filtered, pr)
			}
		case "merged":
			if pr.State() == domain.PRStatusMerged {
				filtered = append(filtered, pr)
			}
		case "closed":
			if pr.State() == domain.PRStatusClosed {
				filtered = append(filtered, pr)
			}
		}
	}
	m.filteredPRs = filtered
}

// ShouldReturnToDashboard returns true if the view should return to dashboard.
func (m PRListViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// ShouldViewPRDetail returns true if a PR detail should be shown.
func (m PRListViewModel) ShouldViewPRDetail() bool {
	return m.viewPRDetail
}

// GetSelectedPR returns the selected PR for detail view.
func (m PRListViewModel) GetSelectedPR() *domain.PRInfo {
	return m.selectedPR
}

// GetCurrentFilter returns the current filter for reloading.
func (m PRListViewModel) GetCurrentFilter() string {
	return m.currentFilter
}
