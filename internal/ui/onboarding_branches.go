package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingBranchesScreen handles branch configuration
type OnboardingBranchesScreen struct {
	step       int
	totalSteps int
	config     *domain.Config

	// Form fields
	focusedField      int
	mainBranch        TextInput
	protectedBranches CheckboxGroup
	customProtected   TextInput
	autoPush          Checkbox
	autoPull          Checkbox

	shouldContinue bool
	shouldGoBack   bool
	
	width  int
	height int
}

// NewOnboardingBranchesScreen creates a new branches screen
func NewOnboardingBranchesScreen(step, totalSteps int, config *domain.Config) OnboardingBranchesScreen {
	// Default protected branches
	defaultBranches := []string{"main", "master", "develop", "production"}
	checked := make([]bool, len(defaultBranches))

	// Check which ones are currently protected
	for i, branch := range defaultBranches {
		for _, protected := range config.Git.ProtectedBranches {
			if branch == protected {
				checked[i] = true
				break
			}
		}
	}

	screen := OnboardingBranchesScreen{
		step:       step,
		totalSteps: totalSteps,
		config:     config,

		mainBranch:        NewTextInput("Main Branch", "main"),
		protectedBranches: NewCheckboxGroup("Protected Branches", defaultBranches, checked),
		customProtected:   NewTextInput("Add Custom Branch", ""),
		autoPush:          NewCheckbox("Auto-push after commit", config.Git.AutoPush),
		autoPull:          NewCheckbox("Auto-pull before operations", config.Git.AutoPull),

		focusedField: 0,
		width:        100,
		height:       40,
	}

	// Set current main branch
	if config.Git.MainBranch != "" {
		screen.mainBranch.Value = config.Git.MainBranch
	} else {
		screen.mainBranch.Value = "main"
	}

	return screen
}

// Init initializes the screen
func (m OnboardingBranchesScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingBranchesScreen) Update(msg tea.Msg) (OnboardingBranchesScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.focusedField {
			case 4:
				// Save and continue
				m.saveToConfig()
				m.shouldContinue = true
				return m, nil
			case 1:
				// Toggle focused checkbox in group
				if m.protectedBranches.FocusedIdx >= 0 && m.protectedBranches.FocusedIdx < len(m.protectedBranches.Items) {
					m.protectedBranches.Items[m.protectedBranches.FocusedIdx].Checked = !m.protectedBranches.Items[m.protectedBranches.FocusedIdx].Checked
				}
				return m, nil
			case 2:
				// Add custom protected branch
				if m.customProtected.Value != "" {
					// Add to protected branches list
					m.protectedBranches.Items = append(m.protectedBranches.Items,
						NewCheckbox(m.customProtected.Value, true))
					m.customProtected.Value = ""
				}
				return m, nil
			}
			// Note: field 3 toggles are handled elsewhere
			return m, nil

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 5
			return m, nil

		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 5) % 5
			return m, nil

		case "left":
			switch m.focusedField {
			case 0:
				m.shouldGoBack = true
				return m, nil
			case 1:
				m.protectedBranches.Previous()
				return m, nil
			}
			return m, nil

		case "right":
			switch m.focusedField {
			case 1:
				m.protectedBranches.Next()
				return m, nil
			}
			return m, nil

		case " ": // Space character
			switch m.focusedField {
			case 1:
				if m.protectedBranches.FocusedIdx >= 0 && m.protectedBranches.FocusedIdx < len(m.protectedBranches.Items) {
					m.protectedBranches.Items[m.protectedBranches.FocusedIdx].Checked = !m.protectedBranches.Items[m.protectedBranches.FocusedIdx].Checked
				}
			case 3:
				// Toggle auto-push
				m.autoPush.Checked = !m.autoPush.Checked
			}
			return m, nil

		case "p", "P":
			// Quick toggle auto-pull
			m.autoPull.Checked = !m.autoPull.Checked
			return m, nil

		case "backspace", "delete":
			// Handle text input deletion
			switch m.focusedField {
			case 0:
				if len(m.mainBranch.Value) > 0 {
					m.mainBranch.Value = m.mainBranch.Value[:len(m.mainBranch.Value)-1]
				}
			case 2:
				if len(m.customProtected.Value) > 0 {
					m.customProtected.Value = m.customProtected.Value[:len(m.customProtected.Value)-1]
				}
			}
			return m, nil

		default:
			// Handle text input
			switch m.focusedField {
			case 0, 2:
				key := msg.String()
				if key == "space" {
					if m.focusedField == 0 {
						m.mainBranch.Value += " "
					} else {
						m.customProtected.Value += " "
					}
				} else if len(key) == 1 {
					if m.focusedField == 0 {
						m.mainBranch.Value += key
					} else {
						m.customProtected.Value += key
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// saveToConfig saves the configuration
func (m *OnboardingBranchesScreen) saveToConfig() {
	// Main branch
	m.config.Git.MainBranch = m.mainBranch.Value

	// Protected branches
	m.config.Git.ProtectedBranches = m.protectedBranches.GetChecked()

	// Auto push/pull
	m.config.Git.AutoPush = m.autoPush.Checked
	m.config.Git.AutoPull = m.autoPull.Checked
}

// View renders the branches screen
func (m OnboardingBranchesScreen) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("Branch Configuration")
	// sections = append(sections, header) // Moved to mainView

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	// sections = append(sections, styles.Metadata.Render(progress)) // Moved to mainView

	// sections = append(sections, "")

	// Description
	desc := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		"Configure your branch preferences and protected branches.")
	sections = append(sections, desc)

	sections = append(sections, "")

	// Main branch
	m.mainBranch.Focused = (m.focusedField == 0)
	sections = append(sections, m.mainBranch.View())
	sections = append(sections, HelpText{Text: "The primary branch for your repository (e.g., main, master)"}.View())

	sections = append(sections, "")

	// Protected branches
	sections = append(sections, m.protectedBranches.View())
	sections = append(sections, HelpText{Text: "Branches that require extra confirmation before operations"}.View())

	sections = append(sections, "")

	// Custom protected branch
	m.customProtected.Focused = (m.focusedField == 2)
	sections = append(sections, m.customProtected.View())
	sections = append(sections, HelpText{Text: "Press Enter to add a custom branch to protected list"}.View())

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// Options
	sections = append(sections, styles.FormLabel.Render("Git Workflow Options:"))
	m.autoPush.Focused = (m.focusedField == 3)
	sections = append(sections, "  "+m.autoPush.View())
	sections = append(sections, HelpText{Text: "Automatically push commits to remote after creating them"}.View())

	sections = append(sections, "")
	sections = append(sections, "  "+m.autoPull.View())
	sections = append(sections, HelpText{Text: "Automatically pull latest changes before operations (Press P to toggle)"}.View())

	sections = append(sections, "")

	// Continue button
	continueBtn := NewButton("Continue")
	continueBtn.Focused = (m.focusedField == 4)
	sections = append(sections, continueBtn.View())

	// Wrap in card
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	cardStyle := styles.DashboardCard.Padding(1, 2)

	// Main view assembly
	mainView := []string{
		header,
		styles.Metadata.Render(progress),
		"",
		cardStyle.Render(content),
		"",
		renderSeparator(70),
	}

	// Footer
	footer := styles.Footer.Render(
		styles.ShortcutKey.Render("Tab/↑↓")+" "+styles.ShortcutDesc.Render("Navigate")+"  "+
			styles.ShortcutKey.Render("Space")+" "+styles.ShortcutDesc.Render("Toggle")+"  "+
			styles.ShortcutKey.Render("←")+" "+styles.ShortcutDesc.Render("Back"))
	mainView = append(mainView, footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, mainView...),
	)
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingBranchesScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingBranchesScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
