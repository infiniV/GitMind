package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/usecase"
)

// MergeViewModel represents the state of the merge view.
type MergeViewModel struct {
	analysis          *usecase.AnalyzeMergeResponse
	selectedIndex     int
	strategies        []MergeStrategy
	confirmed         bool
	returnToDashboard bool
	hasDecision       bool
	err               error
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
}

// MergeStrategy represents a selectable merge strategy.
type MergeStrategy struct {
	Strategy    string
	Label       string
	Description string
	Recommended bool
}

// NewMergeViewModel creates a new merge view model.
func NewMergeViewModel(analysis *usecase.AnalyzeMergeResponse) MergeViewModel {
	strategies := buildMergeStrategies(analysis)

	// Initialize viewport with default size (will be updated on first WindowSizeMsg)
	vp := viewport.New(50, 20)

	m := MergeViewModel{
		analysis:          analysis,
		selectedIndex:     0,
		strategies:        strategies,
		confirmed:         false,
		returnToDashboard: false,
		hasDecision:       false,
		viewport:          vp,
		ready:             true, // Set ready immediately
		windowWidth:       120,  // Default width
		windowHeight:      30,   // Default height
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderStrategiesContent())

	return m
}

func buildMergeStrategies(analysis *usecase.AnalyzeMergeResponse) []MergeStrategy {
	strategies := []MergeStrategy{}

	// Determine which strategy is recommended
	recommended := analysis.SuggestedStrategy
	if recommended == "" {
		recommended = "regular"
	}

	// Always offer squash and regular
	strategies = append(strategies, MergeStrategy{
		Strategy:    "squash",
		Label:       "Squash merge",
		Description: "Combine all commits into a single commit",
		Recommended: recommended == "squash",
	})

	strategies = append(strategies, MergeStrategy{
		Strategy:    "regular",
		Label:       "Regular merge",
		Description: "Preserve all individual commits",
		Recommended: recommended == "regular",
	})

	// Only offer fast-forward if there are no conflicts and suggested
	if analysis.CanMerge && recommended == "fast-forward" {
		strategies = append(strategies, MergeStrategy{
			Strategy:    "fast-forward",
			Label:       "Fast-forward",
			Description: "Fast-forward without creating merge commit",
			Recommended: true,
		})
	}

	return strategies
}

// Init initializes the model.
func (m MergeViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m MergeViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Update viewport size on window resize
		// Match the pane width calculation (48/52 split)
		totalMargins := 4
		dividerWidth := 1
		usableWidth := msg.Width - totalMargins - dividerWidth
		leftWidth := int(float64(usableWidth) * 0.48)
		rightWidth := usableWidth - leftWidth

		// Viewport should be nearly as wide as the right pane to allow content to expand
		// Just subtract small margin for title/padding
		viewportWidth := rightWidth - 2
		viewportHeight := msg.Height - 15 // Account for header, footer, etc.
		if viewportHeight < 5 {
			viewportHeight = 5
		}
		if viewportWidth < 30 {
			viewportWidth = 30
		}
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				// Update viewport content to reflect selection
				m.viewport.SetContent(m.renderStrategiesContent())
			}

		case "down", "j":
			if m.selectedIndex < len(m.strategies)-1 {
				m.selectedIndex++
				// Update viewport content to reflect selection
				m.viewport.SetContent(m.renderStrategiesContent())
			}

		case "enter":
			// Signal that a decision has been made
			m.hasDecision = true
			m.confirmed = true
			return m, nil
		}
	}

	// Update viewport (handles scrolling)
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// View renders the UI with two-pane layout.
func (m MergeViewModel) View() string {
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
			Render("Initializing merge view...")
	}

	// Calculate pane widths with divider space
	// Use almost all available width, leaving small margins
	// Left pane: 45%, Right pane: 55% (right needs more for descriptions)
	totalMargins := 4 // Small margins on edges
	dividerWidth := 1
	usableWidth := m.windowWidth - totalMargins - dividerWidth

	leftWidth := int(float64(usableWidth) * 0.48)
	rightWidth := usableWidth - leftWidth

	// Ensure minimum widths
	if leftWidth < 50 {
		leftWidth = 50
	}
	if rightWidth < 50 {
		rightWidth = 50
	}

	// LEFT PANE: ASCII art, merge info, commits, merge message
	var leftSections []string

	// ASCII art header for MERGE
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	logo := logoStyle.Render(
		`  ███╗   ███╗███████╗██████╗  ██████╗ ███████╗
  ████╗ ████║██╔════╝██╔══██╗██╔════╝ ██╔════╝
  ██╔████╔██║█████╗  ██████╔╝██║  ███╗█████╗
  ██║╚██╔╝██║██╔══╝  ██╔══██╗██║   ██║██╔══╝
  ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝███████╗
  ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝`)

	leftSections = append(leftSections, logo)
	leftSections = append(leftSections, "")

	mergeInfo := m.renderMergeInfo()
	leftSections = append(leftSections, mergeInfo)

	// Conflict warning
	if !m.analysis.CanMerge {
		warning := styles.Warning.Render("[WARNING]") + " " +
			lipgloss.NewStyle().Foreground(styles.ColorError).Render(
				"Conflicts detected")
		leftSections = append(leftSections, warning)

		// Show conflicts
		conflictList := m.renderConflicts()
		leftSections = append(leftSections, conflictList)
	}

	leftSections = append(leftSections, "")

	// Commits to merge
	commitsSection := m.renderCommits()
	leftSections = append(leftSections, commitsSection)

	leftSections = append(leftSections, "")

	// Merge message box
	messageBox := m.renderMergeMessage()
	leftSections = append(leftSections, messageBox)

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		MaxWidth(leftWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, leftSections...))

	// DIVIDER: Vertical line separator
	dividerHeight := m.windowHeight - 5 // Account for footer
	if dividerHeight < 5 {
		dividerHeight = 5
	}
	dividerLines := make([]string, dividerHeight)
	dividerChar := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Render("│")
	for i := range dividerLines {
		dividerLines[i] = dividerChar
	}
	divider := strings.Join(dividerLines, "\n")

	// RIGHT PANE: Strategy selection with viewport
	var rightSections []string

	strategiesTitle := styles.SectionTitle.Render("Select merge strategy:")
	rightSections = append(rightSections, strategiesTitle)
	rightSections = append(rightSections, "")

	// Viewport with scrollable strategies
	viewportContent := m.viewport.View()
	rightSections = append(rightSections, viewportContent)

	// AI reasoning (if available)
	if m.analysis.Reasoning != "" {
		rightSections = append(rightSections, "")
		reasoning := m.renderReasoning()
		rightSections = append(rightSections, reasoning)
	}

	// Scroll indicator
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollIndicator := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Render(fmt.Sprintf("(%d%% scrolled)", int(m.viewport.ScrollPercent()*100)))
		rightSections = append(rightSections, "")
		rightSections = append(rightSections, scrollIndicator)
	}

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		MaxWidth(rightWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, rightSections...))

	// Combine panes horizontally with divider
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		divider,
		rightPane,
	)

	// Footer (full width)
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, "", footer)
}

func (m MergeViewModel) renderMergeInfo() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Source and target branches
	branchLine := styles.RepoLabel.Render("Merge:") + " " +
		lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Render(m.analysis.SourceBranchInfo.Name()) +
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(" → ") +
		lipgloss.NewStyle().Foreground(styles.ColorSuccess).Bold(true).Render(m.analysis.TargetBranch)
	lines = append(lines, branchLine)

	// Commit count
	commitCount := styles.RepoLabel.Render("Commits:") + " " +
		styles.RepoValue.Render(fmt.Sprintf("%d", m.analysis.CommitCount))
	lines = append(lines, commitCount)

	// Status
	status := styles.RepoLabel.Render("Status:") + " "
	if m.analysis.CanMerge {
		status += lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("✓ Ready to merge")
	} else {
		status += lipgloss.NewStyle().Foreground(styles.ColorError).Render("✗ Conflicts detected")
	}
	lines = append(lines, status)

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderConflicts() string {
	styles := GetGlobalThemeManager().GetStyles()
	if len(m.analysis.Conflicts) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, styles.SectionTitle.Render("Conflicts:"))

	for i, conflict := range m.analysis.Conflicts {
		if i >= 5 { // Limit to 5 conflicts
			lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				fmt.Sprintf("  ... and %d more", len(m.analysis.Conflicts)-5)))
			break
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorError).Render("  • "+conflict))
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderCommits() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.SectionTitle.Render("Commits to merge:"))

	maxCommits := len(m.analysis.Commits)
	if maxCommits > 5 {
		maxCommits = 5 // Show only first 5 commits
	}

	for i := 0; i < maxCommits; i++ {
		commit := m.analysis.Commits[i]
		commitLine := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("  • ") +
			lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render(commit.Hash[:7]) + " " +
			styles.RepoValue.Render(commit.Message)
		lines = append(lines, commitLine)
	}

	if len(m.analysis.Commits) > maxCommits {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
			fmt.Sprintf("  ... and %d more commits", len(m.analysis.Commits)-maxCommits)))
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderMergeMessage() string {
	styles := GetGlobalThemeManager().GetStyles()
	if m.analysis.MergeMessage == nil {
		return ""
	}

	var lines []string
	lines = append(lines, styles.SectionTitle.Render("Merge message:"))

	// Calculate available width for merge message box
	usableWidth := m.windowWidth - 4 - 1 // margins and divider
	leftWidth := int(float64(usableWidth) * 0.48)
	boxWidth := leftWidth - 4 // Account for box padding/borders

	messageContent := m.analysis.MergeMessage.FullMessage()
	wrappedContent := wrapTextMerge(messageContent, boxWidth)
	messageBox := styles.CommitBox.Render(wrappedContent)
	lines = append(lines, messageBox)

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderReasoning() string {
	styles := GetGlobalThemeManager().GetStyles()
	if m.analysis.Reasoning == "" {
		return ""
	}

	reasoning := styles.Description.Render("💡 " + m.analysis.Reasoning)
	return reasoning
}

// renderStrategiesContent returns just the strategies text for viewport (no title)
func (m MergeViewModel) renderStrategiesContent() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	for i, strategy := range m.strategies {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = styles.OptionCursor.Render("▶ ")
		}

		label := strategy.Label
		if strategy.Recommended {
			label += lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render(" (recommended)")
		}

		if i == m.selectedIndex {
			label = styles.OptionSelected.Render(label)
		} else {
			label = styles.OptionNormal.Render(label)
		}

		line := cursor + label
		lines = append(lines, line)

		// Add description with proper wrapping
		if strategy.Description != "" {
			// Calculate available width based on right pane width
			totalMargins := 4
			dividerWidth := 1
			usableWidth := m.windowWidth - totalMargins - dividerWidth
			rightPaneWidth := usableWidth - int(float64(usableWidth)*0.48)
			if rightPaneWidth < 50 {
				rightPaneWidth = 50
			}

			// Account for cursor, padding
			maxWidth := rightPaneWidth - 10
			if maxWidth < 30 {
				maxWidth = 30
			}

			wrapped := wrapTextMerge(strategy.Description, maxWidth)
			desc := styles.Description.Render("  " + wrapped)
			lines = append(lines, desc)
		}

		if i < len(m.strategies)-1 {
			lines = append(lines, "") // Space between strategies
		}
	}

	return strings.Join(lines, "\n")
}

// wrapTextMerge wraps text to the specified width
func wrapTextMerge(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if currentLine != "" {
			testLine += " " + word
		} else {
			testLine = word
		}

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n  ") // Indent continuation lines
}

func (m MergeViewModel) renderStrategies() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.SectionTitle.Render("Select merge strategy:"))

	for i, strategy := range m.strategies {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = styles.OptionCursor.Render("▶ ")
		}

		label := strategy.Label
		if strategy.Recommended {
			label += lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render(" (recommended)")
		}

		if i == m.selectedIndex {
			label = styles.OptionSelected.Render(label)
		} else {
			label = styles.OptionNormal.Render(label)
		}

		line := cursor + label
		lines = append(lines, line)

		// Add description
		desc := styles.Description.Render(strategy.Description)
		lines = append(lines, desc)
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()
	shortcuts := []string{
		styles.ShortcutKey.Render("↑/k") + " " + styles.ShortcutDesc.Render("up"),
		styles.ShortcutKey.Render("↓/j") + " " + styles.ShortcutDesc.Render("down"),
		styles.ShortcutKey.Render("enter") + " " + styles.ShortcutDesc.Render("confirm"),
		styles.ShortcutKey.Render("esc") + " " + styles.ShortcutDesc.Render("cancel"),
	}

	footer := styles.Footer.Render(strings.Join(shortcuts, "  •  "))

	// Add metadata
	metadata := styles.Metadata.Render(fmt.Sprintf("Model: %s  •  Tokens: %d",
		m.analysis.Model, m.analysis.TokensUsed))

	return footer + "\n" + metadata
}

// GetSelectedStrategy returns the selected merge strategy.
func (m MergeViewModel) GetSelectedStrategy() string {
	if m.confirmed && m.selectedIndex < len(m.strategies) {
		return m.strategies[m.selectedIndex].Strategy
	}
	return ""
}

// IsConfirmed returns whether the user confirmed the merge.
func (m MergeViewModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns whether the user cancelled.
func (m MergeViewModel) IsCancelled() bool {
	return m.returnToDashboard
}

// ShouldReturnToDashboard returns true if the view should return to dashboard.
func (m MergeViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// HasDecision returns true if the user has made a decision.
func (m MergeViewModel) HasDecision() bool {
	return m.hasDecision
}
