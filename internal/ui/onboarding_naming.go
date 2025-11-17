package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingNamingScreen handles branch naming convention configuration
type OnboardingNamingScreen struct {
	step       int
	totalSteps int
	config     *domain.Config

	// Form fields
	focusedField   int
	enforce        Checkbox
	pattern        TextInput
	allowedPrefixes CheckboxGroup
	customPrefix   TextInput

	// Preview
	previewExample string

	shouldContinue bool
	shouldGoBack   bool
}

// NewOnboardingNamingScreen creates a new naming screen
func NewOnboardingNamingScreen(step, totalSteps int, config *domain.Config) OnboardingNamingScreen {
	// Default prefixes
	defaultPrefixes := []string{"feature", "hotfix", "bugfix", "release", "refactor"}
	checked := make([]bool, len(defaultPrefixes))

	// Check which prefixes are currently allowed
	if config.Naming.Enforce && len(config.Naming.AllowedPrefixes) > 0 {
		for i, prefix := range defaultPrefixes {
			for _, allowed := range config.Naming.AllowedPrefixes {
				if prefix == allowed {
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

	screen := OnboardingNamingScreen{
		step:       step,
		totalSteps: totalSteps,
		config:     config,

		enforce:         NewCheckbox("Enforce branch naming patterns", config.Naming.Enforce),
		pattern:         NewTextInput("Branch Pattern", "feature/{description}"),
		allowedPrefixes: NewCheckboxGroup("Allowed Prefixes", defaultPrefixes, checked),
		customPrefix:    NewTextInput("Add Custom Prefix", ""),

		focusedField: 0,
	}

	// Set current pattern
	if config.Naming.Pattern != "" {
		screen.pattern.Value = config.Naming.Pattern
	} else {
		screen.pattern.Value = "feature/{description}"
	}

	screen.updatePreview()

	return screen
}

// Init initializes the screen
func (m OnboardingNamingScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingNamingScreen) Update(msg tea.Msg) (OnboardingNamingScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focusedField == 4 {
				// Save and continue
				m.saveToConfig()
				m.shouldContinue = true
				return m, nil
			} else if m.focusedField == 3 {
				// Add custom prefix
				if m.customPrefix.Value != "" {
					m.allowedPrefixes.Items = append(m.allowedPrefixes.Items,
						NewCheckbox(m.customPrefix.Value, true))
					m.customPrefix.Value = ""
					m.updatePreview()
				}
				return m, nil
			}
			return m, nil

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 5
			return m, nil

		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 5) % 5
			return m, nil

		case "left":
			if m.focusedField == 0 {
				m.shouldGoBack = true
				return m, nil
			} else if m.focusedField == 2 {
				m.allowedPrefixes.Previous()
				return m, nil
			}
			return m, nil

		case "right":
			if m.focusedField == 2 {
				m.allowedPrefixes.Next()
				return m, nil
			}
			return m, nil

		case "space":
			if m.focusedField == 0 {
				m.enforce.Toggle()
				m.updatePreview()
			} else if m.focusedField == 2 {
				m.allowedPrefixes.Toggle()
				m.updatePreview()
			}
			return m, nil

		default:
			// Handle text input
			if m.focusedField == 1 {
				m.pattern.Update(msg)
				m.updatePreview()
			} else if m.focusedField == 3 {
				m.customPrefix.Update(msg)
			}
			return m, nil
		}
	}

	return m, nil
}

// updatePreview updates the preview example
func (m *OnboardingNamingScreen) updatePreview() {
	if !m.enforce.Checked {
		m.previewExample = "Any branch name is allowed\n\nExamples:\n  my-feature\n  fix-bug-123\n  experiment"
		return
	}

	pattern := m.pattern.Value
	if pattern == "" {
		pattern = "feature/{description}"
	}

	prefixes := m.allowedPrefixes.GetChecked()
	if len(prefixes) == 0 {
		m.previewExample = "No prefixes selected"
		return
	}

	// Generate examples for each prefix
	var examples []string
	for i, prefix := range prefixes {
		if i >= 3 {
			break // Limit to 3 examples
		}
		example := strings.ReplaceAll(pattern, "{prefix}", prefix)
		example = strings.ReplaceAll(example, "{description}", "user-authentication")
		example = strings.ReplaceAll(example, "{issue}", "123")

		// If pattern doesn't have {prefix}, prepend it
		if !strings.Contains(pattern, "{prefix}") {
			example = prefix + "/" + example
		}

		examples = append(examples, "  "+example)
	}

	m.previewExample = "Valid branch names:\n\n" + strings.Join(examples, "\n")
}

// saveToConfig saves the configuration
func (m *OnboardingNamingScreen) saveToConfig() {
	m.config.Naming.Enforce = m.enforce.Checked
	m.config.Naming.Pattern = m.pattern.Value
	m.config.Naming.AllowedPrefixes = m.allowedPrefixes.GetChecked()
}

// View renders the naming screen
func (m OnboardingNamingScreen) View() string {
	var sections []string

	// Header
	header := headerStyle.Render("Branch Naming Patterns")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, metadataStyle.Render(progress))

	sections = append(sections, "")

	// Description
	desc := lipgloss.NewStyle().Foreground(colorMuted).Render(
		"Configure branch naming conventions for consistency.")
	sections = append(sections, desc)

	sections = append(sections, "")

	// Enforce checkbox
	m.enforce.Focused = (m.focusedField == 0)
	sections = append(sections, m.enforce.View())
	sections = append(sections, HelpText{Text: "Require branches to follow naming patterns"}.View())

	sections = append(sections, "")

	// Show pattern options if enforcing
	if m.enforce.Checked {
		// Pattern
		m.pattern.Focused = (m.focusedField == 1)
		sections = append(sections, m.pattern.View())
		sections = append(sections, HelpText{
			Text: "Use placeholders: {prefix}, {description}, {issue}",
		}.View())

		sections = append(sections, "")

		// Allowed prefixes
		sections = append(sections, m.allowedPrefixes.View())
		sections = append(sections, HelpText{Text: "Select which branch prefixes are allowed"}.View())

		sections = append(sections, "")

		// Custom prefix
		m.customPrefix.Focused = (m.focusedField == 3)
		sections = append(sections, m.customPrefix.View())
		sections = append(sections, HelpText{Text: "Press Enter to add custom prefix to list"}.View())
	} else {
		info := lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(1, 2).
			Render("Branch naming is not enforced. You can name branches freely.")
		sections = append(sections, info)
	}

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// Preview
	sections = append(sections, formLabelStyle.Render("Preview:"))
	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Padding(0, 1).
		Foreground(colorText).
		Width(66).
		Render(m.previewExample)
	sections = append(sections, previewBox)

	sections = append(sections, "")

	// Continue button
	continueBtn := NewButton("Continue")
	continueBtn.Focused = (m.focusedField == 4)
	sections = append(sections, continueBtn.View())

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footer := footerStyle.Render(
		shortcutKeyStyle.Render("Tab/↑↓")+" "+shortcutDescStyle.Render("Navigate")+"  "+
			shortcutKeyStyle.Render("Space")+" "+shortcutDescStyle.Render("Toggle")+"  "+
			shortcutKeyStyle.Render("←")+" "+shortcutDescStyle.Render("Back"))
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingNamingScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingNamingScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
