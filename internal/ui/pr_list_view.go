package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// PRListViewModel represents the state of the PR list view.
type PRListViewModel struct {
	prs               []*domain.PRInfo
	selectedIndex     int
	filterState       string // "all", "open", "closed", "merged", "draft"
	returnToDashboard bool
	viewDetail        bool // Navigate to detail view
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
	repoPath          string
}

// NewPRListViewModel creates a new PR list view model.
func NewPRListViewModel(prs []*domain.PRInfo, repoPath string) PRListViewModel {
	// Initialize viewport with default size
	vp := viewport.New(50, 20)

	m := PRListViewModel{
		prs:               prs,
		selectedIndex:     0,
		filterState:       "all",
		returnToDashboard: false,
		viewDetail:        false,
		viewport:          vp,
		ready:             true,
		windowWidth:       120,
		windowHeight:      30,
		repoPath:          repoPath,
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderPRListContent())

	return m
}

// Init initializes the PR list view.
func (m PRListViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the PR list view.
func (m PRListViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Update viewport size
		headerHeight := 12 // Logo + filter bar
		footerHeight := 3
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - headerHeight - footerHeight

		// Update content
		m.viewport.SetContent(m.renderPRListContent())

		if !m.ready {
			m.ready = true
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.returnToDashboard = true
			return m, nil

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.viewport.SetContent(m.renderPRListContent())
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.prs)-1 {
				m.selectedIndex++
				m.viewport.SetContent(m.renderPRListContent())
			}
			return m, nil

		case "enter":
			// View PR detail
			if len(m.prs) > 0 {
				m.viewDetail = true
			}
			return m, nil

		case "o":
			// Filter: open PRs
			m.filterState = "open"
			return m, nil

		case "c":
			// Filter: closed PRs
			m.filterState = "closed"
			return m, nil

		case "m":
			// Filter: merged PRs
			m.filterState = "merged"
			return m, nil

		case "d":
			// Filter: draft PRs
			m.filterState = "draft"
			return m, nil

		case "a":
			// Filter: all PRs
			m.filterState = "all"
			return m, nil

		case "r":
			// Refresh - return to trigger reload
			m.returnToDashboard = true
			return m, nil
		}
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the PR list view.
func (m PRListViewModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	styles := GetGlobalThemeManager().GetStyles()

	// Render logo
	logo := m.renderLogo()

	// Render filter bar
	filterBar := m.renderFilterBar()

	// Render viewport with PR list
	viewportContent := m.viewport.View()

	// Render footer
	footer := m.renderFooter()

	// Combine sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		"",
		filterBar,
		"",
		styles.ViewportStyle.Render(viewportContent),
		"",
		footer,
	)
}

// ShouldReturnToDashboard returns whether the view wants to return to dashboard.
func (m PRListViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// ShouldViewDetail returns whether the view wants to show PR detail.
func (m PRListViewModel) ShouldViewDetail() bool {
	return m.viewDetail
}

// GetSelectedPR returns the currently selected PR.
func (m PRListViewModel) GetSelectedPR() *domain.PRInfo {
	if len(m.prs) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.prs) {
		return nil
	}
	return m.prs[m.selectedIndex]
}

// GetFilterState returns the current filter state.
func (m PRListViewModel) GetFilterState() string {
	return m.filterState
}

// renderLogo renders the PR list logo.
func (m PRListViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	logo := styles.Header.Render("PULL REQUESTS")
	if m.repoPath != "" {
		repoInfo := styles.RepoLabel.Render("Repository: ") + styles.RepoValue.Render(m.repoPath)
		return lipgloss.JoinVertical(lipgloss.Left, logo, repoInfo)
	}
	return logo
}

// renderFilterBar renders the filter buttons.
func (m PRListViewModel) renderFilterBar() string {
	styles := GetGlobalThemeManager().GetStyles()

	filters := []struct {
		key   string
		label string
		state string
	}{
		{"a", "All", "all"},
		{"o", "Open", "open"},
		{"c", "Closed", "closed"},
		{"m", "Merged", "merged"},
		{"d", "Draft", "draft"},
	}

	var filterButtons []string
	for _, f := range filters {
		style := styles.FilterInactive
		if m.filterState == f.state {
			style = styles.FilterActive
		}
		filterButtons = append(filterButtons, style.Render(fmt.Sprintf("[%s] %s", f.key, f.label)))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(filterButtons, "  "),
	)
}

// renderPRListContent renders the PR list content for the viewport.
func (m PRListViewModel) renderPRListContent() string {
	if len(m.prs) == 0 {
		return "      No pull requests found"
	}

	styles := GetGlobalThemeManager().GetStyles()

	var lines []string

	// Table header
	headerStyle := styles.StatusInfo.Bold(true)
	header := fmt.Sprintf("%-6s %-10s %-40s %-15s %-20s",
		"#", "State", "Title", "Author", "Branch")
	lines = append(lines, headerStyle.Render(header))
	lines = append(lines, strings.Repeat("─", m.windowWidth-4))

	// PR rows
	for i, pr := range m.prs {
		// Determine style based on selection and state
		var rowStyle lipgloss.Style
		if i == m.selectedIndex {
			rowStyle = styles.ListItemSelected
		} else {
			rowStyle = styles.ListItemNormal
		}

		// Status indicator
		statusIcon := m.getStatusIcon(pr)
		stateStr := fmt.Sprintf("%s %s", statusIcon, m.getStateString(pr))

		// Truncate title if needed
		title := pr.Title()
		if len(title) > 37 {
			title = title[:34] + "..."
		}

		// Truncate author if needed
		author := pr.Author()
		if len(author) > 12 {
			author = author[:9] + "..."
		}

		// Branch info
		branch := fmt.Sprintf("%s → %s", pr.HeadRef(), pr.BaseRef())
		if len(branch) > 18 {
			branch = branch[:15] + "..."
		}

		row := fmt.Sprintf("%-6d %-10s %-40s %-15s %-20s",
			pr.Number(), stateStr, title, author, branch)

		lines = append(lines, rowStyle.Render(row))
	}

	return strings.Join(lines, "\n")
}

// getStatusIcon returns the Unicode icon for the PR status.
func (m PRListViewModel) getStatusIcon(pr *domain.PRInfo) string {
	if pr.IsDraft() {
		return "◇" // Draft
	}
	switch pr.State() {
	case domain.PRStatusMerged:
		return "◆" // Merged
	case domain.PRStatusClosed:
		return "✕" // Closed
	case domain.PRStatusOpen:
		return "◉" // Open
	default:
		return "○" // Unknown
	}
}

// getStateString returns the string representation of the PR state.
func (m PRListViewModel) getStateString(pr *domain.PRInfo) string {
	if pr.IsDraft() {
		return "Draft"
	}
	switch pr.State() {
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

// renderFooter renders the footer with keyboard shortcuts.
func (m PRListViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()

	help := "↑↓: navigate • enter: view detail • a/o/c/m/d: filter • r: refresh • esc: back"
	metadata := fmt.Sprintf("Showing %d pull request(s)", len(m.prs))

	footer := styles.Footer.Render(help)
	if metadata != "" {
		footer = footer + " " + styles.Metadata.Render(metadata)
	}

	return footer
}
