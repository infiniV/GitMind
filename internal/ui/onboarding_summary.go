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
	
	width  int
	height int
}

// NewOnboardingSummaryScreen creates a new summary screen
func NewOnboardingSummaryScreen(step, totalSteps int, config *domain.Config) OnboardingSummaryScreen {
	return OnboardingSummaryScreen{
		step:           step,
		totalSteps:     totalSteps,
		config:         config,
		selectedOption: 0,
		width:          100,
		height:         40,
	}
}

// Init initializes the screen
func (m OnboardingSummaryScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingSummaryScreen) Update(msg tea.Msg) (OnboardingSummaryScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("Configuration Summary")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, styles.Metadata.Render(progress))

	sections = append(sections, "")

	// Intro
	intro := lipgloss.NewStyle().Foreground(styles.ColorText).Render(
		"Review your configuration before saving:")
	sections = append(sections, intro)

	sections = append(sections, "")
	
	// Git Configuration
	sections = append(sections, getSectionHeaderStyle().Render("Git Configuration"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Main Branch", m.config.Git.MainBranch))
	sections = append(sections, m.renderKeyValue("Protected Branches", strings.Join(m.config.Git.ProtectedBranches, ", ")))
	sections = append(sections, m.renderKeyValue("Auto-push", m.boolToString(m.config.Git.AutoPush)))
	sections = append(sections, m.renderKeyValue("Auto-pull", m.boolToString(m.config.Git.AutoPull)))

	sections = append(sections, "")

	// GitHub Configuration
	sections = append(sections, getSectionHeaderStyle().Render("GitHub Integration"))
	sections = append(sections, "")
	if m.config.GitHub.Enabled {
		sections = append(sections, m.renderKeyValue("Enabled", styles.StatusOk.Render("Yes")))
		sections = append(sections, m.renderKeyValue("Default Visibility", m.config.GitHub.DefaultVisibility))
		sections = append(sections, m.renderKeyValue("Default License", m.config.GitHub.DefaultLicense))
		sections = append(sections, m.renderKeyValue("Default .gitignore", m.config.GitHub.DefaultGitIgnore))
	} else {
		sections = append(sections, m.renderKeyValue("Enabled", styles.StatusWarning.Render("No")))
	}

	sections = append(sections, "")

	// Commit Configuration
	sections = append(sections, getSectionHeaderStyle().Render("Commit Conventions"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Convention", m.capitalizeFirst(m.config.Commits.Convention)))
	switch m.config.Commits.Convention {
	case "conventional":
		sections = append(sections, m.renderKeyValue("Allowed Types", strings.Join(m.config.Commits.Types, ", ")))
		sections = append(sections, m.renderKeyValue("Require Scope", m.boolToString(m.config.Commits.RequireScope)))
		sections = append(sections, m.renderKeyValue("Require Breaking", m.boolToString(m.config.Commits.RequireBreaking)))
	case "custom":
		sections = append(sections, m.renderKeyValue("Template", m.config.Commits.CustomTemplate))
	}

	sections = append(sections, "")

	// Naming Configuration
	sections = append(sections, getSectionHeaderStyle().Render("Branch Naming"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Enforce Patterns", m.boolToString(m.config.Naming.Enforce)))
	if m.config.Naming.Enforce {
		sections = append(sections, m.renderKeyValue("Pattern", m.config.Naming.Pattern))
		sections = append(sections, m.renderKeyValue("Allowed Prefixes", strings.Join(m.config.Naming.AllowedPrefixes, ", ")))
	}

	sections = append(sections, "")

	// AI Configuration
	sections = append(sections, getSectionHeaderStyle().Render("AI Provider"))
	sections = append(sections, "")
	sections = append(sections, m.renderKeyValue("Provider", m.capitalizeFirst(m.config.AI.Provider)))
	sections = append(sections, m.renderKeyValue("API Key", m.maskAPIKey(m.config.AI.APIKey)))
	sections = append(sections, m.renderKeyValue("Tier", m.capitalizeFirst(m.config.AI.APITier)))
	sections = append(sections, m.renderKeyValue("Default Model", m.config.AI.DefaultModel))
	sections = append(sections, m.renderKeyValue("Fallback Model", m.config.AI.FallbackModel))
	sections = append(sections, m.renderKeyValue("Include Context", m.boolToString(m.config.AI.IncludeContext)))

	sections = append(sections, "")

	// Buttons
	saveBtn := NewButton("Save & Continue")
	saveBtn.Focused = (m.selectedOption == 0)
	backBtn := NewButton("Go Back")
	backBtn.Focused = (m.selectedOption == 1)

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, saveBtn.View(), "  ", backBtn.View())
	sections = append(sections, buttons)

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
		styles.ShortcutKey.Render("Tab/←→")+" "+styles.ShortcutDesc.Render("Navigate")+"  "+
			styles.ShortcutKey.Render("Enter")+" "+styles.ShortcutDesc.Render("Confirm"))
	mainView = append(mainView, footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, mainView...),
	)
}

// Helper methods
func (m OnboardingSummaryScreen) renderKeyValue(key, value string) string {
	styles := GetGlobalThemeManager().GetStyles()
	keyStyle := lipgloss.NewStyle().Foreground(styles.ColorText).Width(20).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	return "  " + keyStyle.Render(key+":") + " " + valueStyle.Render(value)
}

func (m OnboardingSummaryScreen) boolToString(b bool) string {
	styles := GetGlobalThemeManager().GetStyles()
	if b {
		return styles.StatusOk.Render("Yes")
	}
	return styles.StatusWarning.Render("No")
}

func (m OnboardingSummaryScreen) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (m OnboardingSummaryScreen) maskAPIKey(key string) string {
	styles := GetGlobalThemeManager().GetStyles()
	if len(key) == 0 {
		return styles.StatusError.Render("Not set")
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func getSectionHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(GetGlobalThemeManager().GetStyles().ColorPrimary).
		Bold(true)
}

// ShouldSave returns true if user wants to save
func (m OnboardingSummaryScreen) ShouldSave() bool {
	return m.shouldSave
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingSummaryScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
