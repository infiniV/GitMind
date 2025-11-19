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
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	sections = append(sections, "")
	sections = append(sections, "")

	// ASCII Art Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Align(lipgloss.Center)

	logo := logoStyle.Render(
		`
   ██████╗ ██╗████████╗    ███╗   ███╗██╗███╗   ██╗██████╗
  ██╔════╝ ██║╚══██╔══╝    ████╗ ████║██║████╗  ██║██╔══██╗
  ██║  ███╗██║   ██║       ██╔████╔██║██║██╔██╗ ██║██║  ██║
  ██║   ██║██║   ██║       ██║╚██╔╝██║██║██║╚██╗██║██║  ██║
  ╚██████╔╝██║   ██║       ██║ ╚═╝ ██║██║██║ ╚████║██████╔╝
   ╚═════╝ ╚═╝   ╚═╝       ╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═════╝`)

	sections = append(sections, logo)

	// Tagline
	taglineStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(2)

	tagline := taglineStyle.Render("AI-Powered Git Workflow Intelligence")
	sections = append(sections, tagline)

	// Progress indicator
	progressStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Align(lipgloss.Center).
		MarginBottom(3)

	progressBar := m.renderProgressBar()
	sections = append(sections, progressStyle.Render(progressBar))

	// Welcome message
	welcomeStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Align(lipgloss.Center).
		Width(70).
		MarginBottom(2)

	welcome := welcomeStyle.Render(
		"Welcome to GitMind! This wizard will help you configure your workspace.\n\n" +
			"We'll set up Git integration, AI providers, and workflow preferences.\n" +
			"The setup takes approximately 2-3 minutes to complete.",
	)
	sections = append(sections, welcome)

	// Separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Align(lipgloss.Center)

	sections = append(sections, separatorStyle.Render(strings.Repeat("─", 70)))
	sections = append(sections, "")

	// Footer with enhanced styling
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Align(lipgloss.Center).
		MarginTop(1)

	footer := footerStyle.Render(
		styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Continue") + "    " +
			styles.ShortcutKey.Render("Esc") + " " + styles.ShortcutDesc.Render("Skip setup"),
	)
	sections = append(sections, footer)

	// Center everything
	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(100).
		Render(content)
}

// renderProgressBar creates a visual progress indicator
func (m OnboardingWelcomeScreen) renderProgressBar() string {
	totalDots := 8
	currentDot := m.step

	styles := GetGlobalThemeManager().GetStyles()
	var dots []string
	for i := 1; i <= totalDots; i++ {
		if i == currentDot {
			dots = append(dots, lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Render("☑"))
		} else if i < currentDot {
			dots = append(dots, lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("✓"))
		} else {
			dots = append(dots, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("☐"))
		}
	}

	progressText := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		fmt.Sprintf("Step %d of %d", m.step, m.totalSteps),
	)

	return progressText + "  " + strings.Join(dots, " ")
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingWelcomeScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldSkip returns true if user wants to skip
func (m OnboardingWelcomeScreen) ShouldSkip() bool {
	return m.shouldSkip
}
