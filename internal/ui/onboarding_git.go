package ui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/git"
)

// OnboardingGitInitScreen handles git initialization
type OnboardingGitInitScreen struct {
	step       int
	totalSteps int
	gitOps     git.Operations
	repoPath   string

	isGitRepo      bool
	hasRemote      bool
	initComplete   bool
	shouldContinue bool
	shouldGoBack   bool
	error          string
}

// NewOnboardingGitInitScreen creates a new git init screen
func NewOnboardingGitInitScreen(step, totalSteps int, gitOps git.Operations, repoPath string) OnboardingGitInitScreen {
	ctx := context.Background()
	isRepo, _ := gitOps.IsGitRepo(ctx, repoPath)

	// Check if remote exists
	hasRemote := false
	if isRepo {
		cmd := exec.Command("git", "remote", "get-url", "origin")
		cmd.Dir = repoPath
		if err := cmd.Run(); err == nil {
			hasRemote = true
		}
	}

	return OnboardingGitInitScreen{
		step:       step,
		totalSteps: totalSteps,
		gitOps:     gitOps,
		repoPath:   repoPath,
		isGitRepo:  isRepo,
		hasRemote:  hasRemote,
	}
}

// Init initializes the screen
func (m OnboardingGitInitScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingGitInitScreen) Update(msg tea.Msg) (OnboardingGitInitScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.isGitRepo || m.initComplete {
				m.shouldContinue = true
			} else if !m.initComplete {
				// Initialize git repo
				cmd := exec.Command("git", "init")
				cmd.Dir = m.repoPath
				if err := cmd.Run(); err != nil {
					m.error = err.Error()
				} else {
					m.initComplete = true
					m.isGitRepo = true
				}
			}
			return m, nil
		case "left":
			m.shouldGoBack = true
			return m, nil
		case "s", "S":
			// Skip git init
			m.shouldContinue = true
			return m, nil
		}
	}

	return m, nil
}

// View renders the git init screen
func (m OnboardingGitInitScreen) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("Git Repository Setup")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, styles.Metadata.Render(progress))

	sections = append(sections, "")

	// Status
	if m.isGitRepo {
		status := styles.StatusOk.Render("✓") + " " +
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("Git repository detected")
		sections = append(sections, status)

		// Check for remote
		if !m.hasRemote {
			remoteStatus := styles.StatusWarning.Render("!") + " " +
				lipgloss.NewStyle().Foreground(styles.ColorText).Render("No remote configured")
			sections = append(sections, remoteStatus)
			sections = append(sections, "")
			sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"Your repository doesn't have a remote origin.\n"+
				"You can configure GitHub integration in the next step."))
		} else {
			sections = append(sections, "")
			sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				"Your workspace is already a git repository with remote. You're all set!"))
		}
	} else if m.initComplete {
		status := styles.StatusOk.Render("✓") + " " +
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("Git repository initialized")
		sections = append(sections, status)
	} else {
		status := styles.StatusWarning.Render("!") + " " +
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("No git repository found")
		sections = append(sections, status)
		sections = append(sections, "")
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
			"GitMind works best with git repositories. Would you like to initialize one now?"))
	}

	if m.error != "" {
		sections = append(sections, "")
		sections = append(sections, styles.StatusError.Render("Error: "+m.error))
	}

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footerText := ""
	if m.isGitRepo || m.initComplete {
		footerText = styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Continue")
	} else {
		footerText = styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Initialize") + "  " +
			styles.ShortcutKey.Render("S") + " " + styles.ShortcutDesc.Render("Skip")
	}
	footerText += "  " + styles.ShortcutKey.Render("←") + " " + styles.ShortcutDesc.Render("Back")

	footer := styles.Footer.Render(footerText)
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingGitInitScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingGitInitScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
