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
		// Use full width for vertical layout
		cardWidth := msg.Width - 4
		if cardWidth < 80 {
			cardWidth = 80
		}
		innerWidth := cardWidth - 4
		
		viewportWidth := innerWidth - 2 // Account for padding

		// Calculate available height for viewport using consistent calculation
		viewportHeight := msg.Height - 17 // Logo + header + content + footer + margins
		if viewportHeight < 5 {
			viewportHeight = 5
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
				switch m.confirmationFocus {
				case 1: // Confirm button
					// Signal decision
					m.hasDecision = true
					m.confirmed = true
					return m, nil
				case 2: // Cancel button
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

// View renders the UI with a master-detail layout.
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

	// 1. Header Section (Logo + Merge Info)
	logo := m.renderLogo()
	mergeInfo := m.renderMergeInfoCompact()
	header := lipgloss.JoinVertical(lipgloss.Left, logo, mergeInfo)

	// 2. Main Content (Split View)
	// Left: Strategies Menu (35%)
	// Right: Details & Context (65%)
	
	totalWidth := m.windowWidth - 4
	leftWidth := int(float64(totalWidth) * 0.35)
	rightWidth := totalWidth - leftWidth - 3

	if leftWidth < 25 { leftWidth = 25 }
	if rightWidth < 40 { rightWidth = 40 }

	// Left Pane: Strategies List
	m.viewport.Width = leftWidth
	m.viewport.Height = contentHeight
	m.viewport.SetContent(m.renderStrategyList(leftWidth))
	
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

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"", // Spacer
		mainContent,
		footer,
	)
}

func (m MergeViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	return lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Render(
		`  ███╗   ███╗███████╗██████╗  ██████╗ ███████╗
  ████╗ ████║██╔════╝██╔══██╗██╔════╝ ██╔════╝
  ██╔████╔██║█████╗  ██████╔╝██║  ███╗█████╗
  ██║╚██╔╝██║██╔══╝  ██╔══██╗██║   ██║██╔══╝
  ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝███████╗
  ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝`)
}

func (m MergeViewModel) renderStrategyList(width int) string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.SectionTitle.Render("STRATEGIES"))
	lines = append(lines, "")

	for i, strategy := range m.strategies {
		isSelected := i == m.selectedIndex
		
		label := fmt.Sprintf("%d. %s", i+1, strategy.Label)
		
		var style lipgloss.Style
		if isSelected {
			style = styles.TabActive.Width(width).Padding(0, 1)
			label = "> " + label
		} else {
			style = styles.TabInactive.Width(width).Padding(0, 1)
			label = "  " + label
		}
		
		lines = append(lines, style.Render(label))
		lines = append(lines, "") // Spacing
	}
	
	return strings.Join(lines, "\n")
}

func (m MergeViewModel) renderDetailsPane(width, height int) string {
	styles := GetGlobalThemeManager().GetStyles()
	selectedStrategy := m.strategies[m.selectedIndex]
	
	var sections []string
	
	// 1. Description
	title := styles.SectionTitle.Render("DETAILS")
	sections = append(sections, title)
	
	desc := wrapTextMerge(selectedStrategy.Description, width)
	sections = append(sections, styles.Description.Render(desc))
	
	if selectedStrategy.Recommended {
		rec := lipgloss.NewStyle().Foreground(styles.ColorSuccess).Bold(true).Render("✓ Recommended by AI")
		sections = append(sections, rec)
	}
	
	sections = append(sections, "")
	sections = append(sections, styles.SectionTitle.Render("CONTEXT"))
	
	// 2. Conflicts (if any)
	if !m.analysis.CanMerge {
		warn := styles.Warning.Render("Conflicts Detected:")
		sections = append(sections, warn)
		for i, c := range m.analysis.Conflicts {
			if i >= 3 { break }
			sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorError).Render("- "+c))
		}
	} else {
		ok := lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("✓ No conflicts")
		sections = append(sections, ok)
	}
	
	sections = append(sections, "")
	
	// 3. Merge Message Preview
	if m.analysis.MergeMessage != nil {
		msgBox := styles.CommitBox.Width(width).Render(
			wrapTextMerge(m.analysis.MergeMessage.FullMessage(), width-4))
		sections = append(sections, msgBox)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m MergeViewModel) renderStrategiesContent() string {
	return m.renderStrategyList(m.viewport.Width)
}

func (m MergeViewModel) renderConfirmationModal() string {
	styles := GetGlobalThemeManager().GetStyles()
	
	// Calculate dimensions
	width := 60
	height := 12
	
	// Title
	title := styles.SectionTitle.Render("CONFIRM MERGE")
	
	// Message Input
	inputStyle := styles.FormInput.Width(width - 4)
	if m.confirmationFocus == 0 {
		inputStyle = styles.FormInputFocused.Width(width - 4)
	}
	inputView := inputStyle.Render(m.msgInput.View())

	// Buttons
	btnStyle := styles.TabInactive.Padding(0, 2)
	activeBtnStyle := styles.TabActive.Padding(0, 2)
	
	confirmBtn := btnStyle.Render("Confirm")
	if m.confirmationFocus == 1 {
		confirmBtn = activeBtnStyle.Render("Confirm")
	}
	
	cancelBtn := btnStyle.Render("Cancel")
	if m.confirmationFocus == 2 {
		cancelBtn = activeBtnStyle.Render("Cancel")
	}
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, confirmBtn, "  ", cancelBtn)
	
	// Content
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		"Enter merge message:",
		inputView,
		"",
		buttons,
	)
	
	// Box
	box := styles.CommitBox.
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		Render(content)
		
	// Center in window
	return lipgloss.Place(m.windowWidth, m.windowHeight, 
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func (m MergeViewModel) renderMergeInfoCompact() string {
	styles := GetGlobalThemeManager().GetStyles()
	
	source := m.analysis.SourceBranchInfo.Name()
	target := m.analysis.TargetBranch
	
	branchStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	textStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	return lipgloss.NewStyle().
		Padding(0, 2).
		Render(fmt.Sprintf("%s %s %s %s", 
			branchStyle.Render(source),
			textStyle.Render("→"),
			branchStyle.Render(target),
			mutedStyle.Render(fmt.Sprintf("(%d commits)", len(m.analysis.Commits))),
		))
}

func (m MergeViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()
	
	help := "↑/↓: Select • Enter: Merge • Esc: Cancel"
	if m.state == ViewStateConfirm {
		help = "Tab: Next • Enter: Select • Esc: Back"
	}
	
	return styles.Footer.Render(help)
}

// ShouldReturnToDashboard returns true if the view should return to dashboard.
func (m MergeViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// HasDecision returns true if the user has made a decision.
func (m MergeViewModel) HasDecision() bool {
	return m.hasDecision
}

// GetSelectedStrategy returns the selected merge strategy.
func (m MergeViewModel) GetSelectedStrategy() string {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.strategies) {
		return m.strategies[m.selectedIndex].Strategy
	}
	return "regular" // Default
}

// GetMergeMessage returns the merge message.
func (m MergeViewModel) GetMergeMessage() string {
	return m.msgInput.Value()
}

func wrapTextMerge(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).Render(text)
}