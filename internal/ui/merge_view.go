package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

	// Input handling
	state             ViewState
	msgInput          textinput.Model
	confirmationFocus int // 0: Msg, 1: Confirm, 2: Cancel
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

	// Initialize text input
	msgInput := textinput.New()
	msgInput.CharLimit = 72
	msgInput.Width = 50
	msgInput.Placeholder = "Enter merge message"

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
		state:             ViewStateBrowsing,
		msgInput:          msgInput,
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
		
		// Refresh content with new width
		m.viewport.SetContent(m.renderStrategiesContent())

		return m, nil

	case tea.KeyMsg:
		// Handle confirmation state
		if m.state == ViewStateConfirm {
			switch msg.String() {
			case "tab":
				m.confirmationFocus++
				if m.confirmationFocus > 2 {
					m.confirmationFocus = 0
				}
				
				// Update focus state
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
				} else {
					m.msgInput.Blur()
				}
				return m, textinput.Blink

			case "shift+tab":
				m.confirmationFocus--
				if m.confirmationFocus < 0 {
					m.confirmationFocus = 2
				}
				
				// Update focus state
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
				} else {
					m.msgInput.Blur()
				}
				return m, textinput.Blink

			case "enter":
				if m.confirmationFocus == 1 { // Confirm button
					// Signal decision
					m.hasDecision = true
					m.confirmed = true
					return m, nil
				} else if m.confirmationFocus == 2 { // Cancel button
					m.state = ViewStateBrowsing
					m.msgInput.Blur()
					return m, nil
				}
				
				// If on input, move to next field
				m.confirmationFocus++
				if m.confirmationFocus > 2 {
					m.confirmationFocus = 1 // Go to confirm button
				}
				
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
				} else {
					m.msgInput.Blur()
				}
				return m, nil

			case "esc":
				m.state = ViewStateBrowsing
				m.msgInput.Blur()
				return m, nil
			}

			// Pass messages to input
			if m.confirmationFocus == 0 {
				m.msgInput, cmd = m.msgInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// Handle browsing state
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
			// Transition to confirmation state
			m.state = ViewStateConfirm
			m.confirmationFocus = 0 // Start at message
			
			// Initialize input with default message
			if m.analysis.MergeMessage != nil {
				m.msgInput.SetValue(m.analysis.MergeMessage.Title())
			} else {
				m.msgInput.SetValue("Merge branch '" + m.analysis.SourceBranchInfo.Name() + "'")
			}
			
			m.msgInput.Focus()
			return m, textinput.Blink
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

	// Render confirmation modal if in confirm state
	if m.state == ViewStateConfirm {
		return m.renderConfirmationModal()
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
		warning := styles.Warning.Render("✗") + " " +
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

func (m MergeViewModel) renderConfirmationModal() string {
	styles := GetGlobalThemeManager().GetStyles()
	selectedStrategy := m.strategies[m.selectedIndex]

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Render("Confirm Merge")

	// Action description
	actionDesc := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Render(selectedStrategy.Label)

	// Message Input
	msgLabel := styles.FormLabel.Render("Merge Message:")
	msgInput := m.msgInput.View()
	if m.confirmationFocus == 0 {
		msgInput = styles.FormInputFocused.Render(m.msgInput.View())
	} else {
		msgInput = styles.FormInput.Render(m.msgInput.View())
	}

	// Buttons
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorMuted)

	buttonActiveStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(2).
		Bold(true).
		Background(styles.ColorPrimary).
		Foreground(lipgloss.Color("#000000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary)

	// Render buttons
	confirmBtn := "Confirm"
	cancelBtn := "Cancel"

	if m.confirmationFocus == 1 {
		confirmBtn = buttonActiveStyle.Render(confirmBtn)
		cancelBtn = buttonStyle.Render(cancelBtn)
	} else if m.confirmationFocus == 2 {
		confirmBtn = buttonStyle.Render(confirmBtn)
		cancelBtn = buttonActiveStyle.Render(cancelBtn)
	} else {
		confirmBtn = buttonStyle.Render(confirmBtn)
		cancelBtn = buttonStyle.Render(cancelBtn)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, confirmBtn, cancelBtn)

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render("Tab to navigate  •  Enter to confirm/next  •  Esc to cancel")

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		actionDesc,
		"",
		msgLabel,
		msgInput,
		"",
		buttons,
		"",
		helpText,
	)

	// Create a modal box
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Background(lipgloss.Color("#1a1a1a")). // Dark background
		Width(70)

	return lipgloss.Place(
		m.windowWidth, m.windowHeight,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content),
	)
}

func (m MergeViewModel) renderMergeInfo() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Source and target branches
	branchLine := styles.RepoLabel.Render("Merge:") + " " +
		lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Render(m.analysis.SourceBranchInfo.Name()) +
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(" -> ") +
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
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorError).Render("  ✗ "+conflict))
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
		commitLine := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("  - ") +
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

	reasoning := styles.Description.Render("INFO: " + m.analysis.Reasoning)
	return reasoning
}

// renderStrategiesContent returns just the strategies text for viewport (no title)
func (m MergeViewModel) renderStrategiesContent() string {
	var lines []string

	for i, strategy := range m.strategies {
		strategyStr := m.renderStrategy(i, strategy)
		lines = append(lines, strategyStr)
		if i < len(m.strategies)-1 {
			lines = append(lines, "") // Space between strategies
		}
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderStrategy(index int, strategy MergeStrategy) string {
	styles := GetGlobalThemeManager().GetStyles()
	isSelected := index == m.selectedIndex

	// Build strategy content
	var content []string

	// Label with number
	number := fmt.Sprintf("%d.", index+1)
	label := fmt.Sprintf("%s %s", number, strategy.Label)
	if isSelected {
		content = append(content, styles.OptionLabel.Render(label))
	} else {
		content = append(content, styles.OptionNormal.Render(label))
	}

	// Recommended badge
	if strategy.Recommended {
		badge := lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Bold(true).
			Render("✓ RECOMMENDED")
		content[0] = content[0] + " " + badge
	}

	// Calculate right pane width once
	totalMargins := 4
	dividerWidth := 1
	usableWidth := m.windowWidth - totalMargins - dividerWidth
	rightPaneWidth := usableWidth - int(float64(usableWidth)*0.48)
	if rightPaneWidth < 50 {
		rightPaneWidth = 50
	}

	// Calculate interior width for content (inside the box borders and padding)
	// Box has: rounded border (2 chars each side) + padding (1 each side) = 6 total
	interiorWidth := rightPaneWidth - 6
	if interiorWidth < 40 {
		interiorWidth = 40
	}

	// Calculate max width for text wrapping
	// Account for OptionDesc paddingLeft (2 chars)
	maxWidth := interiorWidth - 2
	if maxWidth < 30 {
		maxWidth = 30
	}

	// Description (wrapped to fit within the box)
	if strategy.Description != "" {
		wrapped := wrapTextMerge(strategy.Description, maxWidth)
		desc := styles.OptionDesc.Render(wrapped)
		content = append(content, desc)
	}

	// Join content
	strategyContent := strings.Join(content, "\n")

	// Use Place to force content to fill the full interior width
	placedContent := lipgloss.Place(
		interiorWidth,
		lipgloss.Height(strategyContent),
		lipgloss.Left,
		lipgloss.Top,
		strategyContent,
	)

	// Wrap in box style
	var boxStyle lipgloss.Style
	if isSelected {
		boxStyle = styles.SelectedOptionBox
	} else {
		boxStyle = styles.NormalOptionBox
	}

	return boxStyle.Render(placedContent)
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

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Keyboard shortcuts
	shortcuts := []string{
		styles.ShortcutKey.Render("↑/↓") + " " + styles.ShortcutDesc.Render("Navigate"),
		styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Confirm"),
		styles.ShortcutKey.Render("Esc") + " " + styles.ShortcutDesc.Render("Cancel"),
	}
	shortcutLine := strings.Join(shortcuts, "  ")
	lines = append(lines, shortcutLine)

	// Metadata
	metadata := styles.Metadata.Render(fmt.Sprintf("Model: %s  |  Tokens: %d",
		m.analysis.Model, m.analysis.TokensUsed))
	lines = append(lines, metadata)

	return styles.Footer.Render(strings.Join(lines, "\n"))
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

// GetMergeMessage returns the edited merge message.
func (m MergeViewModel) GetMergeMessage() string {
	return m.msgInput.Value()
}
