package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	Confidence  float64
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
		Confidence:  decision.Confidence(),
	}
	options = append(options, primaryOption)

	// Add alternatives
	for _, alt := range decision.Alternatives() {
		option := CommitOption{
			Action:      alt.Action,
			Label:       getAlternativeLabel(alt.Action),
			Description: alt.Description,
			Message:     decision.SuggestedMessage(),
			Confidence:  alt.Confidence,
		}
		options = append(options, option)
	}

	return options
}

func getPrimaryLabel(decision *domain.Decision) string {
	switch decision.Action() {
	case domain.ActionCommitDirect:
		return "Commit to current branch"
	case domain.ActionCreateBranch:
		return fmt.Sprintf("Create branch '%s'", decision.BranchName())
	case domain.ActionReview:
		return "Manual review required"
	case domain.ActionMerge:
		return "Merge to parent branch"
	default:
		return "Unknown action"
	}
}

func getAlternativeLabel(action domain.ActionType) string {
	switch action {
	case domain.ActionCommitDirect:
		return "Commit directly"
	case domain.ActionCreateBranch:
		return "Create new branch"
	case domain.ActionReview:
		return "Review manually"
	case domain.ActionMerge:
		return "Merge to parent"
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
		return lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Render(fmt.Sprintf("ERROR: %v\n", m.err))
	}

	var sections []string

	// Header
	header := headerStyle.Render("GitMind - Commit Assistant")
	sections = append(sections, header)

	// Repository info
	repo := m.analysis.Repository
	repoInfo := m.renderRepoInfo(repo)
	sections = append(sections, repoInfo)

	// Warning if no remote
	if !repo.HasRemote() {
		warning := warningStyle.Render("[WARNING]") + " " +
			lipgloss.NewStyle().Foreground(colorMuted).Render(
				"No remote configured. Use 'git remote add origin <url>' to add one.")
		sections = append(sections, warning)
	}

	// Commit message box
	decision := m.analysis.Decision
	commitBox := m.renderCommitMessage(decision)
	sections = append(sections, commitBox)

	// Options
	optionsSection := m.renderOptions()
	sections = append(sections, optionsSection)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m CommitViewModel) renderRepoInfo(repo *domain.Repository) string {
	var lines []string

	lines = append(lines, sectionTitleStyle.Render("Repository"))

	pathLine := repoLabelStyle.Render("Path:") + " " + repoValueStyle.Render(repo.Path())
	lines = append(lines, pathLine)

	branchLine := repoLabelStyle.Render("Branch:") + " " + repoValueStyle.Render(repo.CurrentBranch())
	lines = append(lines, branchLine)

	changesLine := repoLabelStyle.Render("Changes:") + " " + repoValueStyle.Render(repo.ChangeSummary())
	lines = append(lines, changesLine)

	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderCommitMessage(decision *domain.Decision) string {
	var content string

	if decision.SuggestedMessage() != nil {
		content = decision.SuggestedMessage().Title()
	} else {
		content = "No message generated"
	}

	title := sectionTitleStyle.Render("Suggested Commit Message")
	box := commitBoxStyle.Render(content)

	return title + "\n" + box
}

func (m CommitViewModel) renderOptions() string {
	var lines []string

	title := sectionTitleStyle.Render("Actions")
	lines = append(lines, title)
	lines = append(lines, "") // Empty line

	for i, option := range m.options {
		optionStr := m.renderOption(i, option)
		lines = append(lines, optionStr)
		lines = append(lines, "") // Space between options
	}

	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderOption(index int, option CommitOption) string {
	var parts []string

	// Cursor and number
	cursor := "  "
	numberStyle := optionNormalStyle
	labelStyle := optionNormalStyle
	confStyle := getConfidenceStyle(option.Confidence)

	if index == m.selectedIndex {
		cursor = optionCursorStyle.Render("> ")
		numberStyle = optionSelectedStyle
		labelStyle = optionSelectedStyle
	}

	number := numberStyle.Render(fmt.Sprintf("%d.", index+1))
	label := labelStyle.Render(option.Label)

	// Confidence badge
	confLabel := getConfidenceLabel(option.Confidence)
	confBadge := confStyle.Render(fmt.Sprintf("[%s: %.0f%%]", confLabel, option.Confidence*100))

	// First line: cursor + number + label + confidence
	firstLine := cursor + number + " " + label + " " + confBadge
	parts = append(parts, firstLine)

	// Description (wrapped and indented)
	if option.Description != "" {
		wrapped := wrapText(option.Description, 65)
		desc := descriptionStyle.Render(wrapped)
		parts = append(parts, desc)
	}

	return strings.Join(parts, "\n")
}

func (m CommitViewModel) renderFooter() string {
	var lines []string

	// Keyboard shortcuts
	shortcuts := []string{
		shortcutKeyStyle.Render("↑/↓") + " " + shortcutDescStyle.Render("Navigate"),
		shortcutKeyStyle.Render("Enter") + " " + shortcutDescStyle.Render("Confirm"),
		shortcutKeyStyle.Render("Esc/Q") + " " + shortcutDescStyle.Render("Cancel"),
	}
	shortcutLine := strings.Join(shortcuts, "  ")
	lines = append(lines, shortcutLine)

	// Metadata
	metadata := metadataStyle.Render(fmt.Sprintf("Model: %s  |  Tokens: %d",
		m.analysis.Model, m.analysis.TokensUsed))
	lines = append(lines, metadata)

	return footerStyle.Render(strings.Join(lines, "\n"))
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

	// Try to wrap at word boundary
	if lastSpace := strings.LastIndex(text[:width], " "); lastSpace > width-20 {
		return text[:lastSpace] + "..."
	}

	return text[:width-3] + "..."
}
