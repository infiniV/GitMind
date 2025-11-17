package ui

import (
	"fmt"
	"strings"

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

	return MergeViewModel{
		analysis:          analysis,
		selectedIndex:     0,
		strategies:        strategies,
		confirmed:         false,
		returnToDashboard: false,
		hasDecision:       false,
	}
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.strategies)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Signal that a decision has been made
			m.hasDecision = true
			m.confirmed = true
			return m, nil
		}
	}

	return m, nil
}

// View renders the UI.
func (m MergeViewModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Render(fmt.Sprintf("ERROR: %v\n", m.err))
	}

	var sections []string

	// Header
	header := headerStyle.Render("GitMind - Merge Assistant")
	sections = append(sections, header)

	// Merge info
	mergeInfo := m.renderMergeInfo()
	sections = append(sections, mergeInfo)

	// Conflict warning
	if !m.analysis.CanMerge {
		warning := warningStyle.Render("[WARNING]") + " " +
			lipgloss.NewStyle().Foreground(colorError).Render(
				"Merge conflicts detected! Manual resolution required.")
		sections = append(sections, warning)

		// Show conflicts
		conflictList := m.renderConflicts()
		sections = append(sections, conflictList)
	}

	// Commits to merge
	commitsSection := m.renderCommits()
	sections = append(sections, commitsSection)

	// Merge message box
	messageBox := m.renderMergeMessage()
	sections = append(sections, messageBox)

	// AI reasoning
	reasoning := m.renderReasoning()
	sections = append(sections, reasoning)

	// Strategy selection
	strategySection := m.renderStrategies()
	sections = append(sections, strategySection)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m MergeViewModel) renderMergeInfo() string {
	var lines []string

	// Source and target branches
	branchLine := repoLabelStyle.Render("Merge:") + " " +
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(m.analysis.SourceBranchInfo.Name()) +
		lipgloss.NewStyle().Foreground(colorMuted).Render(" → ") +
		lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render(m.analysis.TargetBranch)
	lines = append(lines, branchLine)

	// Commit count
	commitCount := repoLabelStyle.Render("Commits:") + " " +
		repoValueStyle.Render(fmt.Sprintf("%d", m.analysis.CommitCount))
	lines = append(lines, commitCount)

	// Status
	status := repoLabelStyle.Render("Status:") + " "
	if m.analysis.CanMerge {
		status += lipgloss.NewStyle().Foreground(colorSuccess).Render("✓ Ready to merge")
	} else {
		status += lipgloss.NewStyle().Foreground(colorError).Render("✗ Conflicts detected")
	}
	lines = append(lines, status)

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderConflicts() string {
	if len(m.analysis.Conflicts) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, sectionTitleStyle.Render("Conflicts:"))

	for i, conflict := range m.analysis.Conflicts {
		if i >= 5 { // Limit to 5 conflicts
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(
				fmt.Sprintf("  ... and %d more", len(m.analysis.Conflicts)-5)))
			break
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(colorError).Render("  • "+conflict))
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderCommits() string {
	var lines []string
	lines = append(lines, sectionTitleStyle.Render("Commits to merge:"))

	maxCommits := len(m.analysis.Commits)
	if maxCommits > 5 {
		maxCommits = 5 // Show only first 5 commits
	}

	for i := 0; i < maxCommits; i++ {
		commit := m.analysis.Commits[i]
		commitLine := lipgloss.NewStyle().Foreground(colorMuted).Render("  • ") +
			lipgloss.NewStyle().Foreground(colorPrimary).Render(commit.Hash[:7]) + " " +
			repoValueStyle.Render(commit.Message)
		lines = append(lines, commitLine)
	}

	if len(m.analysis.Commits) > maxCommits {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(
			fmt.Sprintf("  ... and %d more commits", len(m.analysis.Commits)-maxCommits)))
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderMergeMessage() string {
	if m.analysis.MergeMessage == nil {
		return ""
	}

	var lines []string
	lines = append(lines, sectionTitleStyle.Render("Merge message:"))

	messageContent := m.analysis.MergeMessage.FullMessage()
	messageBox := commitBoxStyle.Render(messageContent)
	lines = append(lines, messageBox)

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderReasoning() string {
	if m.analysis.Reasoning == "" {
		return ""
	}

	reasoning := descriptionStyle.Render("💡 " + m.analysis.Reasoning)
	return reasoning
}

func (m MergeViewModel) renderStrategies() string {
	var lines []string
	lines = append(lines, sectionTitleStyle.Render("Select merge strategy:"))

	for i, strategy := range m.strategies {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = optionCursorStyle.Render("▶ ")
		}

		label := strategy.Label
		if strategy.Recommended {
			label += lipgloss.NewStyle().Foreground(colorSuccess).Render(" (recommended)")
		}

		if i == m.selectedIndex {
			label = optionSelectedStyle.Render(label)
		} else {
			label = optionNormalStyle.Render(label)
		}

		line := cursor + label
		lines = append(lines, line)

		// Add description
		desc := descriptionStyle.Render(strategy.Description)
		lines = append(lines, desc)
	}

	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderFooter() string {
	shortcuts := []string{
		shortcutKeyStyle.Render("↑/k") + " " + shortcutDescStyle.Render("up"),
		shortcutKeyStyle.Render("↓/j") + " " + shortcutDescStyle.Render("down"),
		shortcutKeyStyle.Render("enter") + " " + shortcutDescStyle.Render("confirm"),
		shortcutKeyStyle.Render("esc") + " " + shortcutDescStyle.Render("cancel"),
	}

	footer := footerStyle.Render(strings.Join(shortcuts, "  •  "))

	// Add metadata
	metadata := metadataStyle.Render(fmt.Sprintf("Model: %s  •  Tokens: %d",
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
