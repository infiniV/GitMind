package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OnboardingWelcomeScreen is the welcome/intro screen
type OnboardingWelcomeScreen struct {
	step         int
	totalSteps   int
	shouldContinue bool
	shouldSkip    bool
}

// NewOnboardingWelcomeScreen creates a new welcome screen
func NewOnboardingWelcomeScreen(step, totalSteps int) OnboardingWelcomeScreen {
	return OnboardingWelcomeScreen{
		step:           step,
		totalSteps:     totalSteps,
		shouldContinue: false,
		shouldSkip:     false,
	}
}

// Init initializes the screen
func (m OnboardingWelcomeScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingWelcomeScreen) Update(msg tea.Msg) (OnboardingWelcomeScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.shouldContinue = true
			return m, nil
		case "esc", "q":
			m.shouldSkip = true
			return m, nil
		}
	}

	return m, nil
}

// View renders the welcome screen
func (m OnboardingWelcomeScreen) View() string {
	var sections []string

	// Header
	header := headerStyle.Render("Welcome to GitMind")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, metadataStyle.Render(progress))

	sections = append(sections, "")

	// Welcome message
	welcome := lipgloss.NewStyle().
		Foreground(colorText).
		Width(70).
		Render(
			"GitMind is an AI-powered Git workflow manager that helps you:\n\n" +
				"  • Generate intelligent commit messages\n" +
				"  • Make smart branching decisions\n" +
				"  • Automate merge workflows\n" +
				"  • Follow best practices\n\n" +
				"This setup wizard will help you configure GitMind for your workspace.\n" +
				"It will take about 2-3 minutes to complete.",
		)
	sections = append(sections, welcome)

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))
	sections = append(sections, "")

	// What we'll configure
	configList := lipgloss.NewStyle().
		Foreground(colorText).
		Render(
			"We'll configure:\n\n" +
				"  1. Git repository setup\n" +
				"  2. GitHub integration (optional)\n" +
				"  3. Branch preferences\n" +
				"  4. Commit conventions\n" +
				"  5. Branch naming patterns\n" +
				"  6. AI provider settings",
		)
	sections = append(sections, configList)

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footer := footerStyle.Render(
		shortcutKeyStyle.Render("Enter") + " " + shortcutDescStyle.Render("Continue") + "  " +
			shortcutKeyStyle.Render("Esc") + " " + shortcutDescStyle.Render("Skip setup"),
	)
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingWelcomeScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldSkip returns true if user wants to skip
func (m OnboardingWelcomeScreen) ShouldSkip() bool {
	return m.shouldSkip
}
