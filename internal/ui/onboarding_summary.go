package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingSummaryScreen shows configuration summary
type OnboardingSummaryScreen struct {
	step       int
	totalSteps int
	config     *domain.Config

	selectedOption int // 0 = Save & Continue, 1 = Go Back

	shouldSave   bool
	shouldGoBack bool
}

// NewOnboardingSummaryScreen creates a new summary screen
func NewOnboardingSummaryScreen(step, totalSteps int, config *domain.Config) OnboardingSummaryScreen {
	return OnboardingSummaryScreen{
		step:           step,
		totalSteps:     totalSteps,
		config:         config,
		selectedOption: 0,
	}
}

// Init initializes the screen
func (m OnboardingSummaryScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingSummaryScreen) Update(msg tea.Msg) (OnboardingSummaryScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.selectedOption == 0 {
				m.shouldSave = true
			} else {
				m.shouldGoBack = true
			}
			return m, nil

		case "left", "up":
			if m.selectedOption == 1 {
				m.selectedOption = 0
			} else {
				m.shouldGoBack = true
			}
			return m, nil

		case "right", "down", "tab":
			m.selectedOption = (m.selectedOption + 1) % 2
			return m, nil
		}
	}

	return m, nil
}

// View renders the summary screen
func (m OnboardingSummaryScreen) View() string {
	var sections []string

	// Header
	header := headerStyle.Render("Configuration Summary")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, metadataStyle.Render(progress))

	sections = append(sections, "")

	// Intro
	intro := lipgloss.NewStyle().Foreground(colorText).Render(
		"Review your configuration before saving:")
	sections = append(sections, intro)

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// Git Configuration
	sections = append(sections, sectionHeaderStyle.Render("Git Configuration"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Main Branch", m.config.Git.MainBranch))
	sections = append(sections, m.renderKeyValue("Protected Branches", strings.Join(m.config.Git.ProtectedBranches, ", ")))
	sections = append(sections, m.renderKeyValue("Auto-push", m.boolToString(m.config.Git.AutoPush)))
	sections = append(sections, m.renderKeyValue("Auto-pull", m.boolToString(m.config.Git.AutoPull)))

	sections = append(sections, "")

	// GitHub Configuration
	sections = append(sections, sectionHeaderStyle.Render("GitHub Integration"))
	sections = append(sections, "")
	if m.config.GitHub.Enabled {
		sections = append(sections, m.renderKeyValue("Enabled", statusOkStyle.Render("Yes")))
		sections = append(sections, m.renderKeyValue("Default Visibility", m.config.GitHub.DefaultVisibility))
		sections = append(sections, m.renderKeyValue("Default License", m.config.GitHub.DefaultLicense))
		sections = append(sections, m.renderKeyValue("Default .gitignore", m.config.GitHub.DefaultGitIgnore))
	} else {
		sections = append(sections, m.renderKeyValue("Enabled", statusWarningStyle.Render("No")))
	}

	sections = append(sections, "")

	// Commit Configuration
	sections = append(sections, sectionHeaderStyle.Render("Commit Conventions"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Convention", m.capitalizeFirst(m.config.Commits.Convention)))
	if m.config.Commits.Convention == "conventional" {
		sections = append(sections, m.renderKeyValue("Allowed Types", strings.Join(m.config.Commits.Types, ", ")))
		sections = append(sections, m.renderKeyValue("Require Scope", m.boolToString(m.config.Commits.RequireScope)))
		sections = append(sections, m.renderKeyValue("Require Breaking", m.boolToString(m.config.Commits.RequireBreaking)))
	} else if m.config.Commits.Convention == "custom" {
		sections = append(sections, m.renderKeyValue("Template", m.config.Commits.CustomTemplate))
	}

	sections = append(sections, "")

	// Naming Configuration
	sections = append(sections, sectionHeaderStyle.Render("Branch Naming"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Enforce Patterns", m.boolToString(m.config.Naming.Enforce)))
	if m.config.Naming.Enforce {
		sections = append(sections, m.renderKeyValue("Pattern", m.config.Naming.Pattern))
		sections = append(sections, m.renderKeyValue("Allowed Prefixes", strings.Join(m.config.Naming.AllowedPrefixes, ", ")))
	}

	sections = append(sections, "")

	// AI Configuration
	sections = append(sections, sectionHeaderStyle.Render("AI Provider"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Provider", m.capitalizeFirst(m.config.AI.Provider)))
	sections = append(sections, m.renderKeyValue("API Key", m.maskAPIKey(m.config.AI.APIKey)))
	sections = append(sections, m.renderKeyValue("Tier", m.capitalizeFirst(m.config.AI.APITier)))
	sections = append(sections, m.renderKeyValue("Default Model", m.config.AI.DefaultModel))
	sections = append(sections, m.renderKeyValue("Fallback Model", m.config.AI.FallbackModel))
	sections = append(sections, m.renderKeyValue("Include Context", m.boolToString(m.config.AI.IncludeContext)))

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// Buttons
	saveBtn := NewButton("Save & Continue")
	saveBtn.Focused = (m.selectedOption == 0)
	backBtn := NewButton("Go Back")
	backBtn.Focused = (m.selectedOption == 1)

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, saveBtn.View(), "  ", backBtn.View())
	sections = append(sections, buttons)

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footer := footerStyle.Render(
		shortcutKeyStyle.Render("Tab/←→")+" "+shortcutDescStyle.Render("Navigate")+"  "+
			shortcutKeyStyle.Render("Enter")+" "+shortcutDescStyle.Render("Confirm"))
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

// Helper methods
func (m OnboardingSummaryScreen) renderKeyValue(key, value string) string {
	keyStyle := lipgloss.NewStyle().Foreground(colorText).Width(20).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(colorMuted)

	return "  " + keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

func (m OnboardingSummaryScreen) boolToString(b bool) string {
	if b {
		return statusOkStyle.Render("Yes")
	}
	return statusWarningStyle.Render("No")
}

func (m OnboardingSummaryScreen) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (m OnboardingSummaryScreen) maskAPIKey(key string) string {
	if len(key) == 0 {
		return statusErrorStyle.Render("Not set")
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

var sectionHeaderStyle = lipgloss.NewStyle().
	Foreground(colorPrimary).
	Bold(true)

// ShouldSave returns true if user wants to save
func (m OnboardingSummaryScreen) ShouldSave() bool {
	return m.shouldSave
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingSummaryScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
