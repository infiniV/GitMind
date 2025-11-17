package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// CommitViewModel represents the state of the commit view.
type CommitViewModel struct {
	repo              *domain.Repository
	branchInfo        *domain.BranchInfo
	decision          *domain.Decision
	tokensUsed        int
	model             string
	selectedIndex     int
	options           []CommitOption
	confirmed         bool
	returnToDashboard bool
	hasDecision       bool
	err               error
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
func NewCommitViewModel(
	repo *domain.Repository,
	branchInfo *domain.BranchInfo,
	decision *domain.Decision,
	tokensUsed int,
	model string,
) *CommitViewModel {
	options := buildOptions(decision)

	return &CommitViewModel{
		repo:              repo,
		branchInfo:        branchInfo,
		decision:          decision,
		tokensUsed:        tokensUsed,
		model:             model,
		selectedIndex:     0,
		options:           options,
		confirmed:         false,
		returnToDashboard: false,
		hasDecision:       false,
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
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.options)-1 {
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
	repoInfo := m.renderRepoInfo()
	sections = append(sections, repoInfo)

	// Warning if no remote
	if !m.repo.HasRemote() {
		warning := warningStyle.Render("[WARNING]") + " " +
			lipgloss.NewStyle().Foreground(colorMuted).Render(
				"No remote configured. Use 'git remote add origin <url>' to add one.")
		sections = append(sections, warning)
	}

	// Separator
	sections = append(sections, renderSeparator(80))

	// Commit message box
	commitBox := m.renderCommitMessage()
	sections = append(sections, commitBox)

	// Separator
	sections = append(sections, renderSeparator(80))

	// Options
	optionsSection := m.renderOptions()
	sections = append(sections, optionsSection)

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m CommitViewModel) renderRepoInfo() string {
	var lines []string

	lines = append(lines, sectionTitleStyle.Render("Repository"))

	pathLine := repoLabelStyle.Render("Path:") + " " + repoValueStyle.Render(m.repo.Path())
	lines = append(lines, pathLine)

	branchLine := repoLabelStyle.Render("Branch:") + " " + repoValueStyle.Render(m.repo.CurrentBranch())
	lines = append(lines, branchLine)

	changesLine := repoLabelStyle.Render("Changes:") + " " + repoValueStyle.Render(m.repo.ChangeSummary())
	lines = append(lines, changesLine)

	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderCommitMessage() string {
	var content string

	if m.decision.SuggestedMessage() != nil {
		content = m.decision.SuggestedMessage().Title()
	} else {
		content = "No message generated"
	}

	title := sectionTitleStyle.Render("Suggested Commit Message")

	// Show AI analysis reasoning
	var reasoning string
	if m.decision.Reasoning() != "" {
		reasoning = descriptionStyle.Render("AI Analysis: " + m.decision.Reasoning())
	}

	box := commitBoxStyle.Render(content)

	if reasoning != "" {
		return title + "\n" + box + "\n" + reasoning
	}
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
	isSelected := index == m.selectedIndex

	// Build option content
	var content []string

	// Label with number
	number := fmt.Sprintf("%d.", index+1)
	label := fmt.Sprintf("%s %s", number, option.Label)
	if isSelected {
		content = append(content, optionLabelStyle.Render(label))
	} else {
		content = append(content, optionNormalStyle.Render(label))
	}

	// Confidence badge on same line
	badge := getConfidenceBadge(option.Confidence)
	content[0] = content[0] + " " + badge

	// Description (indented)
	if option.Description != "" {
		wrapped := wrapText(option.Description, 70)
		desc := optionDescStyle.Render(wrapped)
		content = append(content, desc)
	}

	// Join content
	optionContent := strings.Join(content, "\n")

	// Wrap in box style
	var boxStyle lipgloss.Style
	if isSelected {
		boxStyle = selectedOptionBoxStyle
	} else {
		boxStyle = normalOptionBoxStyle
	}

	return boxStyle.Render(optionContent)
}

func (m CommitViewModel) renderFooter() string {
	var lines []string

	// Keyboard shortcuts
	shortcuts := []string{
		shortcutKeyStyle.Render("↑/↓") + " " + shortcutDescStyle.Render("Navigate"),
		shortcutKeyStyle.Render("Enter") + " " + shortcutDescStyle.Render("Confirm"),
		shortcutKeyStyle.Render("Esc") + " " + shortcutDescStyle.Render("Cancel"),
	}
	shortcutLine := strings.Join(shortcuts, "  ")
	lines = append(lines, shortcutLine)

	// Metadata
	metadata := metadataStyle.Render(fmt.Sprintf("Model: %s  |  Tokens: %d",
		m.model, m.tokensUsed))
	lines = append(lines, metadata)

	return footerStyle.Render(strings.Join(lines, "\n"))
}

// GetSelectedOption returns the currently selected option as a domain.Alternative.
func (m CommitViewModel) GetSelectedOption() *domain.Alternative {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.options) {
		opt := m.options[m.selectedIndex]
		return &domain.Alternative{
			Action:      opt.Action,
			Description: opt.Description,
			Confidence:  opt.Confidence,
			BranchName:  opt.BranchName,
		}
	}
	return nil
}

// ShouldReturnToDashboard returns true if the view should return to dashboard.
func (m CommitViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// HasDecision returns true if the user has made a decision.
func (m CommitViewModel) HasDecision() bool {
	return m.hasDecision
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
