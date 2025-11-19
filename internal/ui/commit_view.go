package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// ViewState represents the current state of the view
type ViewState int

const (
	ViewStateBrowsing ViewState = iota
	ViewStateConfirm
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

	// Input handling
	state             ViewState
	msgInput          textinput.Model
	branchInput       textinput.Model
	confirmationFocus int // 0: Msg, 1: Branch, 2: Confirm, 3: Cancel
	customMessage     string
	customBranch      string
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
	windowWidth int,
	windowHeight int,
) *CommitViewModel {
	// Initialize text inputs
	msgInput := textinput.New()
	msgInput.CharLimit = 72 // Conventional commit header limit
	msgInput.Width = 50
	msgInput.Placeholder = "Enter commit message"

	branchInput := textinput.New()
	branchInput.CharLimit = 100
	branchInput.Width = 50
	branchInput.Placeholder = "Enter branch name"

	m := &CommitViewModel{
		repo:              repo,
		branchInfo:        branchInfo,
		decision:          decision,
		tokensUsed:        tokensUsed,
		model:             model,
		selectedIndex:     0,
		confirmed:         false,
		returnToDashboard: false,
		hasDecision:       false,
		ready:             true,
		windowWidth:       windowWidth,
		windowHeight:      windowHeight,
		state:             ViewStateBrowsing,
		msgInput:          msgInput,
		branchInput:       branchInput,
	}

	// Initialize options
	m.options = m.buildOptions()

	// Calculate viewport size based on window dimensions
	totalMargins := 4
	dividerWidth := 1
	usableWidth := windowWidth - totalMargins - dividerWidth
	rightWidth := usableWidth - int(float64(usableWidth)*0.48)
	viewportWidth := rightWidth - 2
	viewportHeight := windowHeight - 15

	// Initialize viewport with calculated size
	vp := viewport.New(viewportWidth, viewportHeight)
	m.viewport = vp

	// Set initial viewport content
	m.viewport.SetContent(m.renderOptionsContent())

	return m
}

func (m *CommitViewModel) buildOptions() []CommitOption {
	options := []CommitOption{}

	// Determine effective message and branch
	var msg *domain.CommitMessage
	if m.customMessage != "" {
		// Create a new message from custom input
		// We ignore error here as the input is already constrained by the text input model if needed
		// or we just accept it. NewCommitMessage handles truncation.
		var err error
		msg, err = domain.NewCommitMessage(m.customMessage)
		if err != nil {
			// If validation fails (e.g. empty), fallback to suggested
			msg = m.decision.SuggestedMessage()
		}
	} else {
		msg = m.decision.SuggestedMessage()
	}
	
	branchName := m.decision.BranchName()
	if m.customBranch != "" {
		branchName = m.customBranch
	}

	// Primary option based on AI decision
	primaryOption := CommitOption{
		Action:      m.decision.Action(),
		Label:       getPrimaryLabel(m.decision, branchName),
		Description: m.decision.Reasoning(),
		Message:     msg,
		BranchName:  branchName,
		Confidence:  m.decision.Confidence(),
	}
	options = append(options, primaryOption)

	// Add alternatives
	for _, alt := range m.decision.Alternatives() {
		option := CommitOption{
			Action:      alt.Action,
			Label:       getAlternativeLabel(alt.Action),
			Description: alt.Description,
			Message:     msg, // Use the effective message for alternatives too
			Confidence:  alt.Confidence,
		}
		options = append(options, option)
	}

	return options
}

func getPrimaryLabel(decision *domain.Decision, branchName string) string {
	switch decision.Action() {
	case domain.ActionCommitDirect:
		return "Commit to current branch"
	case domain.ActionCreateBranch:
		return fmt.Sprintf("Create branch '%s'", branchName)
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
		// Use full width for vertical layout
		cardWidth := msg.Width - 4
		if cardWidth < 80 {
			cardWidth = 80
		}
		innerWidth := cardWidth - 4
		
		viewportWidth := innerWidth - 2 // Account for padding

		// Calculate available height for viewport using layout helper
		viewportHeight := msg.Height - 15 // Logo + header + footer + margins
		if viewportHeight < 5 {
			viewportHeight = 5
		}
		
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight

		return m, nil

	case tea.KeyMsg:
		// Handle confirmation state
		if m.state == ViewStateConfirm {
			switch msg.String() {
			case "tab":
				// Cycle focus
				// 0: Msg, 1: Branch (if visible), 2: Confirm, 3: Cancel
				m.confirmationFocus++
				
				// Skip branch input if not creating branch
				selectedOption := m.options[m.selectedIndex]
				if m.confirmationFocus == 1 && selectedOption.Action != domain.ActionCreateBranch {
					m.confirmationFocus++
				}
				
				if m.confirmationFocus > 3 {
					m.confirmationFocus = 0
				}
				
				// Update focus state of inputs
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
					m.branchInput.Blur()
				} else if m.confirmationFocus == 1 {
					m.msgInput.Blur()
					m.branchInput.Focus()
				} else {
					m.msgInput.Blur()
					m.branchInput.Blur()
				}
				return m, textinput.Blink

			case "shift+tab":
				m.confirmationFocus--
				
				// Skip branch input if not creating branch
				selectedOption := m.options[m.selectedIndex]
				if m.confirmationFocus == 1 && selectedOption.Action != domain.ActionCreateBranch {
					m.confirmationFocus--
				}
				
				if m.confirmationFocus < 0 {
					m.confirmationFocus = 3
				}
				
				// Update focus state of inputs
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
					m.branchInput.Blur()
				} else if m.confirmationFocus == 1 {
					m.msgInput.Blur()
					m.branchInput.Focus()
				} else {
					m.msgInput.Blur()
					m.branchInput.Blur()
				}
				return m, textinput.Blink

			case "enter":
				if m.confirmationFocus == 2 { // Confirm button
					// Save values
					m.customMessage = m.msgInput.Value()
					m.customBranch = m.branchInput.Value()
					
					// Rebuild options to reflect changes
					m.options = m.buildOptions()
					
					// Signal decision
					m.hasDecision = true
					m.confirmed = true
					return m, nil
				} else if m.confirmationFocus == 3 { // Cancel button
					m.state = ViewStateBrowsing
					m.msgInput.Blur()
					m.branchInput.Blur()
					return m, nil
				}
				// If on input, maybe move to next field?
				// For now, let's just treat enter as confirm if not on cancel
				// Or better, let enter on input just be enter (newline?) or move focus
				// Since these are single line inputs, enter usually submits
				// Let's make Enter on inputs move to next field
				m.confirmationFocus++
				selectedOption := m.options[m.selectedIndex]
				if m.confirmationFocus == 1 && selectedOption.Action != domain.ActionCreateBranch {
					m.confirmationFocus++
				}
				if m.confirmationFocus > 3 {
					m.confirmationFocus = 0 // Loop back or stop at confirm?
					// Let's stop at confirm (2)
					m.confirmationFocus = 2
				}
				
				// Update focus
				if m.confirmationFocus == 0 {
					m.msgInput.Focus()
					m.branchInput.Blur()
				} else if m.confirmationFocus == 1 {
					m.msgInput.Blur()
					m.branchInput.Focus()
				} else {
					m.msgInput.Blur()
					m.branchInput.Blur()
				}
				return m, nil

			case "esc":
				m.state = ViewStateBrowsing
				m.msgInput.Blur()
				m.branchInput.Blur()
				return m, nil
			}

			// Pass messages to inputs
			var cmd tea.Cmd
			if m.confirmationFocus == 0 {
				m.msgInput, cmd = m.msgInput.Update(msg)
				return m, cmd
			} else if m.confirmationFocus == 1 {
				m.branchInput, cmd = m.branchInput.Update(msg)
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
				m.viewport.SetContent(m.renderOptionsContent())
			}

		case "down", "j":
			if m.selectedIndex < len(m.options)-1 {
				m.selectedIndex++
				// Update viewport content to reflect selection
				m.viewport.SetContent(m.renderOptionsContent())
			}

		case "enter":
			// Transition to confirmation state
			m.state = ViewStateConfirm
			m.confirmationFocus = 0 // Start at message
			
			// Initialize inputs with current values
			selectedOption := m.options[m.selectedIndex]
			
			// Message
			if selectedOption.Message != nil {
				m.msgInput.SetValue(selectedOption.Message.Title())
			} else {
				m.msgInput.SetValue("")
			}
			
			// Branch
			if selectedOption.BranchName != "" {
				m.branchInput.SetValue(selectedOption.BranchName)
			} else {
				m.branchInput.SetValue("")
			}
			
			m.msgInput.Focus()
			return m, textinput.Blink
		}
	}

	// Update viewport (handles scrolling)
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// View renders the UI with a master-detail layout.
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

	if m.state == ViewStateConfirm {
		return m.renderConfirmationModal()
	}

	// Layout Dimensions
	headerHeight := 8 // Logo (6) + Info (1) + Padding (1)
	footerHeight := 2
	contentHeight := m.windowHeight - headerHeight - footerHeight
	if contentHeight < 10 {
		contentHeight = 10
	}

	// 1. Header Section (Logo + Repo Info)
	logo := m.renderLogo()
	repoInfo := m.renderRepoInfoCompact()
	header := lipgloss.JoinVertical(lipgloss.Left, logo, repoInfo)

	// 2. Main Content (Split View)
	// Left: Options Menu (30%)
	// Right: Details & Context (70%)
	
	totalWidth := m.windowWidth - 4
	leftWidth := int(float64(totalWidth) * 0.35)
	rightWidth := totalWidth - leftWidth - 3 // -3 for divider/padding

	if leftWidth < 25 { leftWidth = 25 }
	if rightWidth < 40 { rightWidth = 40 }

	// Left Pane: Options List
	m.viewport.Width = leftWidth
	m.viewport.Height = contentHeight
	m.viewport.SetContent(m.renderOptionList(leftWidth))
	
	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight).
		Render(m.viewport.View())

	// Right Pane: Details
	rightPane := m.renderDetailsPane(rightWidth, contentHeight)

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Height(contentHeight).
		Render(" │ ")

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		leftPane,
		divider,
		rightPane,
	)

	// Wrap main content in a card/box if desired, or just keep it clean
	// The user wants "compact", so minimal borders is better.
	
	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"", // Spacer
		mainContent,
		footer,
	)
}

func (m CommitViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	return lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Render(
		`  ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗
 ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝
 ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║
 ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║
 ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║
  ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝`)
}

func (m CommitViewModel) renderOptionList(width int) string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.SectionTitle.Render("ACTIONS"))
	lines = append(lines, "")

	for i, option := range m.options {
		isSelected := i == m.selectedIndex
		
		label := fmt.Sprintf("%d. %s", i+1, option.Label)
		
		var style lipgloss.Style
		if isSelected {
			style = styles.TabActive.Copy().Width(width).Padding(0, 1)
			label = "> " + label
		} else {
			style = styles.TabInactive.Copy().Width(width).Padding(0, 1)
			label = "  " + label
		}
		
		lines = append(lines, style.Render(label))
		lines = append(lines, "") // Spacing
	}
	
	return strings.Join(lines, "\n")
}

func (m CommitViewModel) renderDetailsPane(width, height int) string {
	styles := GetGlobalThemeManager().GetStyles()
	selectedOption := m.options[m.selectedIndex]
	
	var sections []string
	
	// 1. Description of Action
	title := styles.SectionTitle.Render("DETAILS")
	sections = append(sections, title)
	
	desc := wrapText(selectedOption.Description, width)
	sections = append(sections, styles.Description.Render(desc))
	
	sections = append(sections, "")
	sections = append(sections, styles.SectionTitle.Render("CONTEXT"))
	
	// 2. Commit Message Preview (if applicable)
	if selectedOption.Message != nil {
		msgBox := styles.CommitBox.Copy().Width(width).Render(
			wrapText(selectedOption.Message.Title(), width-4))
		sections = append(sections, msgBox)
	}
	
	// 3. Branch Info (if applicable)
	if selectedOption.BranchName != "" {
		branchInfo := fmt.Sprintf("Target Branch: %s", selectedOption.BranchName)
		sections = append(sections, styles.RepoValue.Render(branchInfo))
	}
	
	// 4. Confidence
	conf := fmt.Sprintf("AI Confidence: %.0f%%", selectedOption.Confidence*100)
	sections = append(sections, styles.Metadata.Render(conf))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m CommitViewModel) renderConfirmationModal() string {
	styles := GetGlobalThemeManager().GetStyles()
	selectedOption := m.options[m.selectedIndex]

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Render("Confirm Action")

	// Action description
	actionDesc := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Render(selectedOption.Label)

	// Message Input
	msgLabel := styles.FormLabel.Render("Commit Message:")
	msgInput := m.msgInput.View()
	if m.confirmationFocus == 0 {
		// Highlight the input if focused
		// We can't easily style the internal text of textinput.View() without rebuilding it
		// But textinput handles its own styling.
		// Let's just rely on the cursor blinking which textinput provides.
		// Maybe add a border or indicator?
		msgInput = styles.FormInputFocused.Render(m.msgInput.Value())
		// Wait, if we render just the value, we lose the cursor.
		// Let's stick to m.msgInput.View() but maybe wrap it in a border?
		// Actually, let's just use the View() output.
		// The issue is that textinput.View() returns a string.
		// If we want to show focus, we can wrap it.
		msgInput = styles.FormInputFocused.Render(m.msgInput.View())
	} else {
		msgInput = styles.FormInput.Render(m.msgInput.View())
	}

	// Branch Input (only if creating branch)
	var branchSection string
	if selectedOption.Action == domain.ActionCreateBranch {
		branchLabel := styles.FormLabel.Render("Branch Name:")
		branchView := m.branchInput.View()
		if m.confirmationFocus == 1 {
			branchView = styles.FormInputFocused.Render(branchView)
		} else {
			branchView = styles.FormInput.Render(branchView)
		}
		branchSection = lipgloss.JoinVertical(lipgloss.Left, "", branchLabel, branchView)
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

	if m.confirmationFocus == 2 {
		confirmBtn = buttonActiveStyle.Render(confirmBtn)
		cancelBtn = buttonStyle.Render(cancelBtn)
	} else if m.confirmationFocus == 3 {
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
		branchSection,
		"",
		buttons,
		"",
		helpText,
	)

	// Create a modal box
	theme := GetGlobalThemeManager().GetCurrentTheme()
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Confirmation)).
		Width(70)

	return lipgloss.Place(
		m.windowWidth, m.windowHeight,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content),
	)
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

func (m CommitViewModel) renderRepoInfoCompact() string {
	styles := GetGlobalThemeManager().GetStyles()
	
	// Single line: Path | Branch | Changes
	path := styles.RepoValue.Render(m.repo.Path())
	branch := styles.RepoValue.Render(m.repo.CurrentBranch())
	changes := styles.RepoValue.Render(m.repo.ChangeSummary())
	
	labelStyle := styles.RepoLabel
	
	return fmt.Sprintf("%s %s  %s %s  %s %s", 
		labelStyle.Render("Path:"), path,
		labelStyle.Render("Branch:"), branch,
		labelStyle.Render("Changes:"), changes)
}

func (m CommitViewModel) renderCommitMessage() string {
	styles := GetGlobalThemeManager().GetStyles()
	var content string

	if m.customMessage != "" {
		content = m.customMessage
	} else if m.decision.SuggestedMessage() != nil {
		content = m.decision.SuggestedMessage().Title()
	} else {
		content = "No message generated"
	}

	title := styles.SectionTitle.Render("Suggested Commit Message")

	// Calculate available width for commit box
	usableWidth := m.windowWidth - 4 - 1 // margins and divider
	leftWidth := int(float64(usableWidth) * 0.48)
	boxWidth := leftWidth - 4 // Account for box padding/borders

	// Wrap content to fit in box
	wrappedContent := wrapText(content, boxWidth)

	box := styles.CommitBox.Render(wrappedContent)

	// Show AI analysis reasoning with wrapping
	var reasoning string
	if m.decision.Reasoning() != "" {
		reasoningText := "AI Analysis: " + m.decision.Reasoning()
		wrappedReasoning := wrapText(reasoningText, leftWidth-2)
		reasoning = styles.Description.Render(wrappedReasoning)
	}

	if reasoning != "" {
		return title + "\n" + box + "\n" + reasoning
	}
	return title + "\n" + box
}

func (m CommitViewModel) renderCommitMessageCompact() string {
	styles := GetGlobalThemeManager().GetStyles()
	var content string

	if m.customMessage != "" {
		content = m.customMessage
	} else if m.decision.SuggestedMessage() != nil {
		content = m.decision.SuggestedMessage().Title()
	} else {
		content = "No message generated"
	}

	// Just the box, no title, no reasoning (unless critical)
	// Calculate available width
	cardWidth := m.windowWidth - 4
	if cardWidth < 80 { cardWidth = 80 }
	boxWidth := cardWidth - 8 // padding

	wrappedContent := wrapText(content, boxWidth)
	return styles.CommitBox.Render(wrappedContent)
}

// renderOptionsContent returns just the options text for viewport
func (m CommitViewModel) renderOptionsContent() string {
	return m.renderOptionList(m.viewport.Width)
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

// GetSelectedOption returns the currently selected option.
func (m CommitViewModel) GetSelectedOption() *CommitOption {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.options) {
		opt := m.options[m.selectedIndex]
		return &opt
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
