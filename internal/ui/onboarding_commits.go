package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingCommitsScreen handles commit convention configuration
type OnboardingCommitsScreen struct {
	step       int
	totalSteps int
	config     *domain.Config

	// Form fields
	focusedField    int
	convention      RadioGroup
	commitTypes     CheckboxGroup
	requireScope    Checkbox
	requireBreaking Checkbox
	customTemplate  TextInput

	// Preview
	previewExample string

	shouldContinue bool
	shouldGoBack   bool
}

// NewOnboardingCommitsScreen creates a new commits screen
func NewOnboardingCommitsScreen(step, totalSteps int, config *domain.Config) OnboardingCommitsScreen {
	// Default commit types
	defaultTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	checked := make([]bool, len(defaultTypes))

	// Check which types are currently enabled
	if config.Commits.Convention == "conventional" && len(config.Commits.Types) > 0 {
		for i, commitType := range defaultTypes {
			for _, enabled := range config.Commits.Types {
				if commitType == enabled {
					checked[i] = true
					break
				}
			}
		}
	} else {
		// Default all enabled
		for i := range checked {
			checked[i] = true
		}
	}

	// Determine convention index
	conventionIdx := 0 // Conventional Commits
	switch config.Commits.Convention {
	case "custom":
		conventionIdx = 1
	case "none":
		conventionIdx = 2
	}

	screen := OnboardingCommitsScreen{
		step:       step,
		totalSteps: totalSteps,
		config:     config,

		convention: NewRadioGroup("Commit Convention", []string{
			"Conventional Commits (recommended)",
			"Custom Template",
			"None (freeform)",
		}, conventionIdx),
		commitTypes:     NewCheckboxGroup("Allowed Commit Types", defaultTypes, checked),
		requireScope:    NewCheckbox("Require scope in commits", config.Commits.RequireScope),
		requireBreaking: NewCheckbox("Require breaking change marker", config.Commits.RequireBreaking),
		customTemplate:  NewTextInput("Custom Template", "{type}({scope}): {description}"),

		focusedField: 0,
	}

	if config.Commits.CustomTemplate != "" {
		screen.customTemplate.Value = config.Commits.CustomTemplate
	}

	screen.updatePreview()

	return screen
}

// Init initializes the screen
func (m OnboardingCommitsScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingCommitsScreen) Update(msg tea.Msg) (OnboardingCommitsScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focusedField == 5 {
				// Save and continue
				m.saveToConfig()
				m.shouldContinue = true
				return m, nil
			}
			return m, nil

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 6
			m.updatePreview()
			return m, nil

		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 6) % 6
			m.updatePreview()
			return m, nil

		case "left":
			switch m.focusedField {
			case 0:
				m.convention.Selected = (m.convention.Selected - 1 + len(m.convention.Options)) % len(m.convention.Options)
				m.updatePreview()
				return m, nil
			case 1:
				m.commitTypes.Previous()
				return m, nil
			}
			return m, nil

		case "right":
			switch m.focusedField {
			case 0:
				m.convention.Selected = (m.convention.Selected + 1) % len(m.convention.Options)
				m.updatePreview()
				return m, nil
			case 1:
				m.commitTypes.Next()
				return m, nil
			}
			return m, nil

		case " ": // Space character
			switch m.focusedField {
			case 0:
				m.convention.Selected = (m.convention.Selected + 1) % len(m.convention.Options)
				m.updatePreview()
			case 1:
				if m.commitTypes.FocusedIdx >= 0 && m.commitTypes.FocusedIdx < len(m.commitTypes.Items) {
					m.commitTypes.Items[m.commitTypes.FocusedIdx].Checked = !m.commitTypes.Items[m.commitTypes.FocusedIdx].Checked
				}
				m.updatePreview()
			case 2:
				m.requireScope.Checked = !m.requireScope.Checked
				m.updatePreview()
			case 3:
				m.requireBreaking.Checked = !m.requireBreaking.Checked
				m.updatePreview()
			}
			return m, nil

		case "backspace", "delete":
			// Handle text input deletion for custom template
			if m.focusedField == 4 && m.convention.Selected == 1 {
				if len(m.customTemplate.Value) > 0 {
					m.customTemplate.Value = m.customTemplate.Value[:len(m.customTemplate.Value)-1]
				}
				m.updatePreview()
			}
			return m, nil

		default:
			// Handle text input for custom template
			if m.focusedField == 4 && m.convention.Selected == 1 {
				key := msg.String()
				if key == "space" {
					m.customTemplate.Value += " "
					m.updatePreview()
				} else if len(key) == 1 {
					m.customTemplate.Value += key
					m.updatePreview()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// updatePreview updates the preview example
func (m *OnboardingCommitsScreen) updatePreview() {
	switch m.convention.Selected {
	case 0: // Conventional Commits
		types := m.commitTypes.GetChecked()
		if len(types) > 0 {
			exampleType := types[0]
			if m.requireScope.Checked {
				m.previewExample = fmt.Sprintf("%s(api): add user authentication endpoint", exampleType)
			} else {
				m.previewExample = fmt.Sprintf("%s: add user authentication endpoint", exampleType)
			}
			if m.requireBreaking.Checked {
				m.previewExample += "\n\nBREAKING CHANGE: authentication is now required"
			}
		} else {
			m.previewExample = "feat: add new feature"
		}

	case 1: // Custom Template
		template := m.customTemplate.Value
		if template == "" {
			template = "{type}({scope}): {description}"
		}
		// Replace placeholders with examples
		example := strings.ReplaceAll(template, "{type}", "feat")
		example = strings.ReplaceAll(example, "{scope}", "api")
		example = strings.ReplaceAll(example, "{description}", "add user authentication")
		m.previewExample = example

	case 2: // None (freeform)
		m.previewExample = "Add user authentication to API endpoints"
	}
}

// saveToConfig saves the configuration
func (m *OnboardingCommitsScreen) saveToConfig() {
	switch m.convention.Selected {
	case 0:
		m.config.Commits.Convention = "conventional"
		m.config.Commits.Types = m.commitTypes.GetChecked()
		m.config.Commits.RequireScope = m.requireScope.Checked
		m.config.Commits.RequireBreaking = m.requireBreaking.Checked
		m.config.Commits.CustomTemplate = ""

	case 1:
		m.config.Commits.Convention = "custom"
		m.config.Commits.CustomTemplate = m.customTemplate.Value
		m.config.Commits.Types = []string{}
		m.config.Commits.RequireScope = false
		m.config.Commits.RequireBreaking = false

	default:
		m.config.Commits.Convention = "none"
		m.config.Commits.Types = []string{}
		m.config.Commits.RequireScope = false
		m.config.Commits.RequireBreaking = false
		m.config.Commits.CustomTemplate = ""
	}
}

// View renders the commits screen
func (m OnboardingCommitsScreen) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("Commit Conventions")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, styles.Metadata.Render(progress))

	sections = append(sections, "")

	// Description
	desc := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		"Choose how you want to format commit messages.")
	sections = append(sections, desc)

	sections = append(sections, "")

	// Convention selection
	m.convention.Focused = (m.focusedField == 0)
	sections = append(sections, m.convention.View())

	sections = append(sections, "")

	// Show different options based on convention
	switch m.convention.Selected {
	case 0: // Conventional Commits
		sections = append(sections, m.commitTypes.View())
		sections = append(sections, HelpText{Text: "Select which commit types are allowed in your project"}.View())

		sections = append(sections, "")

		// Options
		sections = append(sections, styles.FormLabel.Render("Options:"))
		m.requireScope.Focused = (m.focusedField == 2)
		sections = append(sections, "  "+m.requireScope.View())
		sections = append(sections, HelpText{Text: "Example: feat(api): ... instead of feat: ..."}.View())

		sections = append(sections, "")

		m.requireBreaking.Focused = (m.focusedField == 3)
		sections = append(sections, "  "+m.requireBreaking.View())
		sections = append(sections, HelpText{Text: "Require BREAKING CHANGE footer for major changes"}.View())

	case 1: // Custom Template
		m.customTemplate.Focused = (m.focusedField == 4)
		sections = append(sections, m.customTemplate.View())
		sections = append(sections, HelpText{
			Text: "Use placeholders: {type}, {scope}, {description}, {body}",
		}.View())

	case 2: // None
		info := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Padding(1, 2).
			Render("No commit format restrictions. You can write commit messages freely.")
		sections = append(sections, info)
	}

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// Preview
	sections = append(sections, styles.FormLabel.Render("Preview Example:"))
	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorMuted).
		Padding(0, 1).
		Foreground(styles.ColorPrimary).
		Width(66).
		Render(m.previewExample)
	sections = append(sections, previewBox)

	sections = append(sections, "")

	// Continue button
	continueBtn := NewButton("Continue")
	continueBtn.Focused = (m.focusedField == 5)
	sections = append(sections, continueBtn.View())

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footer := styles.Footer.Render(
		styles.ShortcutKey.Render("Tab/↑↓")+" "+styles.ShortcutDesc.Render("Navigate")+"  "+
			styles.ShortcutKey.Render("Space/←→")+" "+styles.ShortcutDesc.Render("Select")+"  "+
			styles.ShortcutKey.Render("←")+" "+styles.ShortcutDesc.Render("Back"))
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingCommitsScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingCommitsScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
