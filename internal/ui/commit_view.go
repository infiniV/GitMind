package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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
	viewport          viewport.Model
	ready             bool
	windowWidth       int
	windowHeight      int
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

	// Initialize viewport with default size (will be updated on first WindowSizeMsg)
	vp := viewport.New(50, 20)

	m := &CommitViewModel{
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
		viewport:          vp,
		ready:             true, // Set ready immediately
		windowWidth:       120,  // Default width
		windowHeight:      30,   // Default height
	}

	// Set initial viewport content
	m.viewport.SetContent(m.renderOptionsContent())

	return m
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Update viewport size on window resize
		// Match the pane width calculation
		viewportWidth := (msg.Width / 2) - 8
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
				m.viewport.SetContent(m.renderOptionsContent())
			}

		case "down", "j":
			if m.selectedIndex < len(m.options)-1 {
				m.selectedIndex++
				// Update viewport content to reflect selection
				m.viewport.SetContent(m.renderOptionsContent())
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
func (m CommitViewModel) View() string {
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
			Render("Initializing commit view...")
	}

	// Calculate pane widths - more conservative to prevent cutoff
	leftWidth := (m.windowWidth / 2) - 4
	rightWidth := (m.windowWidth / 2) - 4

	// LEFT PANE: ASCII art, repo info, commit message
	var leftSections []string

	// ASCII art header for COMMIT
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	logo := logoStyle.Render(
		`  ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗
 ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝
 ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║
 ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║
 ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║
  ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝`)

	leftSections = append(leftSections, logo)
	leftSections = append(leftSections, "")

	repoInfo := m.renderRepoInfo()
	leftSections = append(leftSections, repoInfo)

	if !m.repo.HasRemote() {
		warning := styles.Warning.Render("[WARNING]") + " " +
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"No remote configured")
		leftSections = append(leftSections, warning)
	}

	leftSections = append(leftSections, "")
	leftSections = append(leftSections, renderSeparator(leftWidth))
	leftSections = append(leftSections, "")

	commitBox := m.renderCommitMessage()
	leftSections = append(leftSections, commitBox)

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, leftSections...))

	// RIGHT PANE: Options with viewport
	var rightSections []string

	optionsTitle := styles.SectionTitle.Render("Actions")
	rightSections = append(rightSections, optionsTitle)
	rightSections = append(rightSections, "")

	// Viewport with scrollable options
	viewportContent := m.viewport.View()
	rightSections = append(rightSections, viewportContent)

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
		Render(lipgloss.JoinVertical(lipgloss.Left, rightSections...))

	// Combine panes horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		"  ", // 2-space gap
		rightPane,
	)

	// Footer (full width)
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, "", footer)
}

func (m CommitViewModel) renderRepoInfo() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.SectionTitle.Render("Repository"))

	pathLine := styles.RepoLabel.Render("Path:") + " " + styles.RepoValue.Render(m.repo.Path())
	lines = append(lines, pathLine)

	branchLine := styles.RepoLabel.Render("Branch:") + " " + styles.RepoValue.Render(m.repo.CurrentBranch())
	lines = append(lines, branchLine)

	changesLine := styles.RepoLabel.Render("Changes:") + " " + styles.RepoValue.Render(m.repo.ChangeSummary())
	lines = append(lines, changesLine)

	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderCommitMessage() string {
	styles := GetGlobalThemeManager().GetStyles()
	var content string

	if m.decision.SuggestedMessage() != nil {
		content = m.decision.SuggestedMessage().Title()
	} else {
		content = "No message generated"
	}

	title := styles.SectionTitle.Render("Suggested Commit Message")

	// Show AI analysis reasoning
	var reasoning string
	if m.decision.Reasoning() != "" {
		reasoning = styles.Description.Render("AI Analysis: " + m.decision.Reasoning())
	}

	box := styles.CommitBox.Render(content)

	if reasoning != "" {
		return title + "\n" + box + "\n" + reasoning
	}
	return title + "\n" + box
}

// renderOptionsContent returns just the options text for viewport (no title)
func (m CommitViewModel) renderOptionsContent() string {
	var lines []string

	for i, option := range m.options {
		optionStr := m.renderOption(i, option)
		lines = append(lines, optionStr)
		if i < len(m.options)-1 {
			lines = append(lines, "") // Space between options
		}
	}

	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderOptions() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	title := styles.SectionTitle.Render("Actions")
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
	styles := GetGlobalThemeManager().GetStyles()
	isSelected := index == m.selectedIndex

	// Build option content
	var content []string

	// Label with number
	number := fmt.Sprintf("%d.", index+1)
	label := fmt.Sprintf("%s %s", number, option.Label)
	if isSelected {
		content = append(content, styles.OptionLabel.Render(label))
	} else {
		content = append(content, styles.OptionNormal.Render(label))
	}

	// Confidence badge on same line
	badge := getConfidenceBadge(option.Confidence)
	content[0] = content[0] + " " + badge

	// Description (wrapped to pane width minus padding)
	if option.Description != "" {
		// Calculate available width: half window - margins - padding - borders
		maxWidth := (m.windowWidth / 2) - 12
		if maxWidth < 30 {
			maxWidth = 30 // Minimum width
		}
		wrapped := wrapText(option.Description, maxWidth)
		desc := styles.OptionDesc.Render(wrapped)
		content = append(content, desc)
	}

	// Join content
	optionContent := strings.Join(content, "\n")

	// Wrap in box style
	var boxStyle lipgloss.Style
	if isSelected {
		boxStyle = styles.SelectedOptionBox
	} else {
		boxStyle = styles.NormalOptionBox
	}

	return boxStyle.Render(optionContent)
}

func (m CommitViewModel) renderFooter() string {
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
		m.model, m.tokensUsed))
	lines = append(lines, metadata)

	return styles.Footer.Render(strings.Join(lines, "\n"))
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
