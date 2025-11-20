package ui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingGitHubScreen handles GitHub repository creation
type OnboardingGitHubScreen struct {
	step       int
	totalSteps int
	config     *domain.Config
	repoPath   string

	ghAvailable    bool
	ghAuthenticated bool
	checkComplete   bool
	hasRemote      bool

	// Form fields
	focusedField   int
	repoName       TextInput
	description    TextInput
	visibility     RadioGroup
	license        Dropdown
	gitignore      Dropdown
	addReadme      Checkbox
	enableIssues   Checkbox
	enableWiki     Checkbox
	enableProjects Checkbox

	// State
	creating       bool
	createComplete bool
	error          string
	shouldContinue bool
	shouldGoBack   bool
	shouldSkip     bool
	
	width  int
	height int
}

// NewOnboardingGitHubScreen creates a new GitHub screen
func NewOnboardingGitHubScreen(step, totalSteps int, config *domain.Config, repoPath string) OnboardingGitHubScreen {
	// Get repo name from current directory
	defaultRepoName := filepath.Base(repoPath)

	// Check if remote exists
	hasRemote := false
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	if err := cmd.Run(); err == nil {
		hasRemote = true
	}

	screen := OnboardingGitHubScreen{
		step:       step,
		totalSteps: totalSteps,
		config:     config,
		repoPath:   repoPath,
		hasRemote:  hasRemote,

		repoName:    NewTextInput("Repository Name", defaultRepoName),
		description: NewTextInput("Description", "Created with GitMind"),
		visibility: NewRadioGroup("Visibility", []string{"Public", "Private"}, 0),
		license:    NewDropdown("License", github.GetLicenseTemplates(), 0),
		gitignore:  NewDropdown(".gitignore Template", github.GetGitIgnoreTemplates(), 0),
		addReadme:      NewCheckbox("Add README.md", true),
		enableIssues:   NewCheckbox("Enable Issues", true),
		enableWiki:     NewCheckbox("Enable Wiki", false),
		enableProjects: NewCheckbox("Enable Projects", false),

		focusedField: 0,
		width:        100,
		height:       40,
	}

	// Set default values from config
	screen.repoName.Value = defaultRepoName
	screen.description.Value = ""
	if config.GitHub.DefaultVisibility == "private" {
		screen.visibility.Selected = 1
	}

	return screen
}

// Init initializes the screen
func (m OnboardingGitHubScreen) Init() tea.Cmd {
	return m.checkGitHub()
}

// checkGitHub checks if gh CLI is available and authenticated
func (m OnboardingGitHubScreen) checkGitHub() tea.Cmd {
	return func() tea.Msg {
		ghAvailable := github.CheckGHAvailable()
		if !ghAvailable {
			return githubCheckMsg{available: false, authenticated: false}
		}

		ctx := context.Background()
		authenticated, _ := github.CheckGHAuthenticated(ctx)
		return githubCheckMsg{available: true, authenticated: authenticated}
	}
}

type githubCheckMsg struct {
	available     bool
	authenticated bool
}

type githubCreateMsg struct {
	success bool
	error   string
}

// Update handles messages
func (m OnboardingGitHubScreen) Update(msg tea.Msg) (OnboardingGitHubScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case githubCheckMsg:
		m.ghAvailable = msg.available
		m.ghAuthenticated = msg.authenticated
		m.checkComplete = true
		return m, nil

	case githubCreateMsg:
		m.creating = false
		if msg.success {
			m.createComplete = true
			m.shouldContinue = true
		} else {
			m.error = msg.error
		}
		return m, nil

	case tea.KeyMsg:
		// If not available, not authenticated, or already has remote - simple navigation
		if !m.ghAvailable || !m.ghAuthenticated || m.hasRemote {
			switch msg.String() {
			case "enter":
				m.shouldContinue = true
				return m, nil
			case "esc":
				m.shouldGoBack = true
				return m, nil
			case "s", "S":
				m.shouldSkip = true
				m.shouldContinue = true
				return m, nil
			}
			return m, nil
		}

		// Full form navigation
		switch msg.String() {
		case "enter":
			// For button, submit form
			if m.focusedField == 9 {
				return m, m.createRepository()
			}
			// For dropdowns, toggle them
			switch m.focusedField {
			case 3: // License dropdown
				m.license.Toggle()
				return m, nil
			case 4: // Gitignore dropdown
				m.gitignore.Toggle()
				return m, nil
			}
			// For all other fields (text, radio, checkbox), move to next
			m.focusedField = (m.focusedField + 1) % 10
			return m, nil

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 10
			return m, nil

		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 10) % 10
			return m, nil

		case "esc":
			// Esc always goes back to previous screen
			m.shouldGoBack = true
			return m, nil

		case "left":
			// ONLY for navigating within radio/dropdown - NO back navigation
			if m.focusedField == 2 {
				m.visibility.Selected = (m.visibility.Selected - 1 + len(m.visibility.Options)) % len(m.visibility.Options)
			} else if m.focusedField == 3 && m.license.Open {
				m.license.Previous()
			} else if m.focusedField == 4 && m.gitignore.Open {
				m.gitignore.Previous()
			}
			return m, nil

		case "right":
			// ONLY for navigating within radio/dropdown
			if m.focusedField == 2 {
				m.visibility.Selected = (m.visibility.Selected + 1) % len(m.visibility.Options)
			} else if m.focusedField == 3 && m.license.Open {
				m.license.Next()
			} else if m.focusedField == 4 && m.gitignore.Open {
				m.gitignore.Next()
			}
			return m, nil

		case " ": // Space character
			// Toggle checkboxes or cycle radio
			if m.focusedField == 2 {
				m.visibility.Selected = (m.visibility.Selected + 1) % len(m.visibility.Options)
			} else if m.focusedField >= 5 && m.focusedField <= 8 {
				switch m.focusedField {
				case 5:
					m.addReadme.Checked = !m.addReadme.Checked
				case 6:
					m.enableIssues.Checked = !m.enableIssues.Checked
				case 7:
					m.enableWiki.Checked = !m.enableWiki.Checked
				case 8:
					m.enableProjects.Checked = !m.enableProjects.Checked
				}
			}
			return m, nil

		case "s", "S":
			// Skip this screen
			m.shouldSkip = true
			m.shouldContinue = true
			return m, nil

		case "backspace", "delete":
			// Handle text input deletion
			switch m.focusedField {
			case 0:
				if len(m.repoName.Value) > 0 {
					m.repoName.Value = m.repoName.Value[:len(m.repoName.Value)-1]
				}
			case 1:
				if len(m.description.Value) > 0 {
					m.description.Value = m.description.Value[:len(m.description.Value)-1]
				}
			}
			return m, nil

		default:
			// Handle all other text input (alphanumeric, symbols, etc.)
			if m.focusedField == 0 || m.focusedField == 1 {
				// Check if it's a printable character or space
				key := msg.String()
				if key == "space" {
					if m.focusedField == 0 {
						m.repoName.Value += " "
					} else {
						m.description.Value += " "
					}
				} else if len(key) == 1 {
					if m.focusedField == 0 {
						m.repoName.Value += key
					} else {
						m.description.Value += key
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// createRepository creates the GitHub repository
func (m OnboardingGitHubScreen) createRepository() tea.Cmd {
	return func() tea.Msg {
		m.creating = true

		ctx := context.Background()

		// Get current user for repo URL
		owner, err := github.GetCurrentUser(ctx)
		if err != nil {
			return githubCreateMsg{success: false, error: "Failed to get GitHub username: " + err.Error()}
		}

		// Build options
		opts := github.CreateRepoOptions{
			Name:        m.repoName.Value,
			Description: m.description.Value,
			Visibility:  strings.ToLower(m.visibility.GetSelected()),
			License:     m.license.GetSelected(),
			GitIgnore:   m.gitignore.GetSelected(),
			AddReadme:   m.addReadme.Checked,
			EnableIssues:   m.enableIssues.Checked,
			EnableWiki:     m.enableWiki.Checked,
			EnableProjects: m.enableProjects.Checked,
		}

		// Create repository
		err = github.CreateRepository(ctx, opts)
		if err != nil {
			return githubCreateMsg{success: false, error: err.Error()}
		}

		// Set remote
		repoURL := github.GetRepoURL(owner, m.repoName.Value)
		err = github.SetRemote(ctx, m.repoPath, repoURL)
		if err != nil {
			return githubCreateMsg{success: false, error: "Repository created but failed to set remote: " + err.Error()}
		}

		// Update config
		m.config.GitHub.Enabled = true
		m.config.GitHub.DefaultVisibility = strings.ToLower(m.visibility.GetSelected())
		m.config.GitHub.DefaultLicense = m.license.GetSelected()
		m.config.GitHub.DefaultGitIgnore = m.gitignore.GetSelected()
		m.config.GitHub.EnableIssues = m.enableIssues.Checked
		m.config.GitHub.EnableWiki = m.enableWiki.Checked
		m.config.GitHub.EnableProjects = m.enableProjects.Checked

		return githubCreateMsg{success: true, error: ""}
	}
}

// View renders the GitHub screen
func (m OnboardingGitHubScreen) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("GitHub Integration")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, styles.Metadata.Render(progress))

	sections = append(sections, "")

	// Check if checking status
	if !m.checkComplete {
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Checking GitHub CLI..."))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(sections, "\n"))
	}

	// If gh not available
	if !m.ghAvailable {
		sections = append(sections, styles.StatusWarning.Render("!")+" "+
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("GitHub CLI (gh) not found"))
		sections = append(sections, "")
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
			"The GitHub CLI is not installed or not in your PATH.\n\n"+
				"To use GitHub integration, install it from:\n"+
				"  https://cli.github.com/\n\n"+
				"You can configure this later via Settings."))
		sections = append(sections, "")
		sections = append(sections, renderSeparator(70))
		sections = append(sections, "")
		sections = append(sections, styles.Footer.Render(
			styles.ShortcutKey.Render("Enter")+" "+styles.ShortcutDesc.Render("Skip & Continue")+"  "+
				styles.ShortcutKey.Render("Esc")+" "+styles.ShortcutDesc.Render("Back")+"  "+
				styles.ShortcutKey.Render("S")+" "+styles.ShortcutDesc.Render("Skip")))
		
		content := lipgloss.JoinVertical(lipgloss.Left, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// If not authenticated
	if !m.ghAuthenticated {
		sections = append(sections, styles.StatusWarning.Render("!")+" "+
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("GitHub CLI not authenticated"))
		sections = append(sections, "")
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
			"You need to authenticate with GitHub first.\n\n"+
				"Run this command in a separate terminal:\n\n"+
				"  gh auth login\n\n"+
				"Then restart this wizard, or configure GitHub later via Settings."))
		sections = append(sections, "")
		sections = append(sections, renderSeparator(70))
		sections = append(sections, "")
		sections = append(sections, styles.Footer.Render(
			styles.ShortcutKey.Render("Enter")+" "+styles.ShortcutDesc.Render("Skip & Continue")+"  "+
				styles.ShortcutKey.Render("Esc")+" "+styles.ShortcutDesc.Render("Back")+"  "+
				styles.ShortcutKey.Render("S")+" "+styles.ShortcutDesc.Render("Skip")))
		
		content := lipgloss.JoinVertical(lipgloss.Left, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// If remote already exists
	if m.hasRemote {
		sections = append(sections, styles.StatusOk.Render("✓")+" "+
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("Git remote already configured"))
		sections = append(sections, "")
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
			"Your repository already has a remote origin configured.\n"+
				"You can skip this step or reconfigure it later via Settings."))
		sections = append(sections, "")
		sections = append(sections, renderSeparator(70))
		sections = append(sections, "")
		sections = append(sections, styles.Footer.Render(
			styles.ShortcutKey.Render("Enter")+" "+styles.ShortcutDesc.Render("Continue")+"  "+
				styles.ShortcutKey.Render("Esc")+" "+styles.ShortcutDesc.Render("Back")+"  "+
				styles.ShortcutKey.Render("S")+" "+styles.ShortcutDesc.Render("Skip")))
		
		content := lipgloss.JoinVertical(lipgloss.Left, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// If creating
	if m.creating {
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render("Creating GitHub repository..."))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(sections, "\n"))
	}

	// If create complete
	if m.createComplete {
		sections = append(sections, styles.StatusOk.Render("✓")+" "+
			lipgloss.NewStyle().Foreground(styles.ColorText).Render("Repository created successfully!"))
		sections = append(sections, "")
		sections = append(sections, renderSeparator(70))
		sections = append(sections, "")
		sections = append(sections, styles.Footer.Render(
			styles.ShortcutKey.Render("Enter")+" "+styles.ShortcutDesc.Render("Continue")))
		
		content := lipgloss.JoinVertical(lipgloss.Left, sections...)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Full form
	formTitle := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true).Render("Create a new GitHub repository")
	sections = append(sections, formTitle)
	sections = append(sections, "")

	// Repository name
	m.repoName.Focused = (m.focusedField == 0)
	sections = append(sections, m.repoName.View())
	sections = append(sections, "")

	// Description
	m.description.Focused = (m.focusedField == 1)
	sections = append(sections, m.description.View())
	sections = append(sections, "")

	// Visibility
	m.visibility.Focused = (m.focusedField == 2)
	sections = append(sections, m.visibility.View())
	sections = append(sections, "")

	// License
	m.license.Focused = (m.focusedField == 3)
	sections = append(sections, m.license.View())
	sections = append(sections, "")

	// .gitignore
	m.gitignore.Focused = (m.focusedField == 4)
	sections = append(sections, m.gitignore.View())
	sections = append(sections, "")

	// Checkboxes
	sections = append(sections, styles.FormLabel.Render("Options:"))
	m.addReadme.Focused = (m.focusedField == 5)
	sections = append(sections, m.addReadme.View())
	m.enableIssues.Focused = (m.focusedField == 6)
	sections = append(sections, m.enableIssues.View())
	m.enableWiki.Focused = (m.focusedField == 7)
	sections = append(sections, m.enableWiki.View())
	m.enableProjects.Focused = (m.focusedField == 8)
	sections = append(sections, m.enableProjects.View())

	sections = append(sections, "")

	// Create button
	createBtn := NewButton("Create Repository")
	createBtn.Focused = (m.focusedField == 9)
	sections = append(sections, createBtn.View())

	// Error
	if m.error != "" {
		sections = append(sections, "")
		sections = append(sections, styles.StatusError.Render("Error: "+m.error))
	}

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

	// Footer - simple, consistent instructions
	footerText := styles.ShortcutKey.Render("Tab/↑↓") + " " + styles.ShortcutDesc.Render("Navigate") + "  " +
		styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Next/Select") + "  " +
		styles.ShortcutKey.Render("Esc") + " " + styles.ShortcutDesc.Render("Back") + "  " +
		styles.ShortcutKey.Render("S") + " " + styles.ShortcutDesc.Render("Skip")

	// Add field-specific hints
	if m.focusedField == 0 || m.focusedField == 1 {
		// Text input
		footerText = styles.ShortcutKey.Render("Type") + " " + styles.ShortcutDesc.Render("to edit") + "  " + footerText
	} else if m.focusedField >= 5 && m.focusedField <= 8 {
		// Checkbox
		footerText = styles.ShortcutKey.Render("Space") + " " + styles.ShortcutDesc.Render("Toggle") + "  " + footerText
	} else if m.focusedField == 9 {
		// Button
		footerText = styles.ShortcutKey.Render("Enter") + " " + styles.ShortcutDesc.Render("Create Repository") + "  " +
			styles.ShortcutKey.Render("Esc") + " " + styles.ShortcutDesc.Render("Back")
	}

	footer := styles.Footer.Render(footerText)
	mainView = append(mainView, footer)

	// Center the whole view
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, mainView...),
	)
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingGitHubScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingGitHubScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
