package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/usecase"
)

// CommitViewModel represents the state of the commit view.
type CommitViewModel struct {
	analysis      *usecase.AnalyzeCommitResponse
	selectedIndex int
	options       []CommitOption
	confirmed     bool
	cancelled     bool
	err           error
}

// CommitOption represents a user-selectable option.
type CommitOption struct {
	Action      domain.ActionType
	Label       string
	Description string
	Message     *domain.CommitMessage
	BranchName  string
}

// NewCommitViewModel creates a new commit view model.
func NewCommitViewModel(analysis *usecase.AnalyzeCommitResponse) CommitViewModel {
	options := buildOptions(analysis.Decision)

	return CommitViewModel{
		analysis:      analysis,
		selectedIndex: 0,
		options:       options,
		confirmed:     false,
		cancelled:     false,
	}
}

func buildOptions(decision *domain.Decision) []CommitOption {
	options := []CommitOption{}

	// Primary option based on AI decision
	primaryOption := CommitOption{
		Action:      decision.Action(),
		Label:       getPrimaryLabel(decision),
		Description: decision.Reasoning(),
		Message:     decision.SuggestedMessage(),
		BranchName:  decision.BranchName(),
	}
	options = append(options, primaryOption)

	// Add alternatives
	for _, alt := range decision.Alternatives() {
		option := CommitOption{
			Action:      alt.Action,
			Label:       getAlternativeLabel(alt.Action),
			Description: alt.Description,
			Message:     decision.SuggestedMessage(), // Reuse the same message
		}
		options = append(options, option)
	}

	return options
}

func getPrimaryLabel(decision *domain.Decision) string {
	switch decision.Action() {
	case domain.ActionCommitDirect:
		return fmt.Sprintf("✓ Commit directly (confidence: %.0f%%)", decision.Confidence()*100)
	case domain.ActionCreateBranch:
		return fmt.Sprintf("✓ Create branch '%s' (confidence: %.0f%%)", decision.BranchName(), decision.Confidence()*100)
	case domain.ActionReview:
		return "⚠ Manual review recommended"
	default:
		return "Unknown action"
	}
}

func getAlternativeLabel(action domain.ActionType) string {
	switch action {
	case domain.ActionCommitDirect:
		return "Commit directly instead"
	case domain.ActionCreateBranch:
		return "Create a new branch instead"
	case domain.ActionReview:
		return "Review manually"
	default:
		return "Other option"
	}
}

// Init initializes the model.
func (m CommitViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m CommitViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.options)-1 {
				m.selectedIndex++
			}

		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI.
func (m CommitViewModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var sb strings.Builder

	// Header
	sb.WriteString("╔══════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                     GitMind - AI Commit Assistant                    ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════════╝\n\n")

	// Repository info
	repo := m.analysis.Repository
	sb.WriteString(fmt.Sprintf("Repository: %s\n", repo.Path()))
	sb.WriteString(fmt.Sprintf("Branch: %s\n", repo.CurrentBranch()))

	// Warn if no remote configured
	if !repo.HasRemote() {
		sb.WriteString("⚠️  No remote repository configured (use 'git remote add origin <url>')\n")
	}

	sb.WriteString(fmt.Sprintf("Changes: %s\n\n", repo.ChangeSummary()))

	// AI suggested commit message
	decision := m.analysis.Decision
	sb.WriteString("Suggested Commit Message:\n")
	sb.WriteString("┌────────────────────────────────────────────────────────────────────┐\n")
	if decision.SuggestedMessage() != nil {
		sb.WriteString(fmt.Sprintf("│ %s\n", wrapText(decision.SuggestedMessage().Title(), 68)))
	}
	sb.WriteString("└────────────────────────────────────────────────────────────────────┘\n\n")

	// Options
	sb.WriteString("Select an action:\n\n")
	for i, option := range m.options {
		cursor := " "
		if i == m.selectedIndex {
			cursor = "▶"
		}

		sb.WriteString(fmt.Sprintf("%s %d. %s\n", cursor, i+1, option.Label))
		sb.WriteString(fmt.Sprintf("     %s\n\n", wrapText(option.Description, 65)))
	}

	// Footer
	sb.WriteString("────────────────────────────────────────────────────────────────────\n")
	sb.WriteString("↑/↓: Navigate  Enter: Confirm  Esc/Q: Cancel\n")
	sb.WriteString(fmt.Sprintf("AI Model: %s | Tokens: %d\n", m.analysis.Model, m.analysis.TokensUsed))

	return sb.String()
}

// GetSelectedOption returns the currently selected option.
func (m CommitViewModel) GetSelectedOption() *CommitOption {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.options) {
		return &m.options[m.selectedIndex]
	}
	return nil
}

// IsConfirmed returns true if the user confirmed the selection.
func (m CommitViewModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if the user cancelled.
func (m CommitViewModel) IsCancelled() bool {
	return m.cancelled
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	return text[:width-3] + "..."
}
