package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/ai"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/usecase"
)

// GitHubOperations defines the interface for GitHub operations
type GitHubOperations interface {
	ViewRepoWeb(ctx context.Context, repoPath string) error
	GetRepoInfo(ctx context.Context, repoPath string) (*github.RepoInfo, error)
}

// GitHubOps is a simple implementation of GitHubOperations
type GitHubOps struct{}

// ViewRepoWeb opens the repository in the browser
func (g GitHubOps) ViewRepoWeb(ctx context.Context, repoPath string) error {
	return github.ViewRepoWeb(ctx, repoPath)
}

// GetRepoInfo retrieves repository information
func (g GitHubOps) GetRepoInfo(ctx context.Context, repoPath string) (*github.RepoInfo, error) {
	return github.GetRepoInfo(ctx, repoPath)
}

// AppState represents the current state of the application
type AppState int

const (
	StateDashboard AppState = iota
	StateCommitAnalyzing
	StateCommitView
	StateCommitExecuting
	StateMergeAnalyzing
	StateMergeView
	StateMergeExecuting
	StateOnboarding
)

// Tab constants
type Tab int

const (
	TabDashboard Tab = iota
	TabSettings
)

// AppModel is the root model that manages the entire application lifecycle
type AppModel struct {
	// State management
	state AppState

	// Tab management
	currentTab Tab

	// Child models
	dashboard      *DashboardModel
	commitView     *CommitViewModel
	mergeView      *MergeViewModel
	settingsView   *SettingsView
	onboardingView *OnboardingModel

	// Dependencies
	gitOps     git.Operations
	aiProvider ai.Provider
	githubOps  GitHubOperations
	cfg        *domain.Config
	cfgManager *config.Manager
	repoPath   string

	// App info
	version string

	// Window dimensions
	windowWidth  int
	windowHeight int

	// Loading state
	loadingMessage string
	loadingDots    int

	// Results from async operations
	commitAnalysisResult *usecase.AnalyzeCommitResponse
	commitAnalysisError  error
	mergeAnalysisResult  *usecase.AnalyzeMergeResponse
	mergeAnalysisError   error

	// Action parameters from dashboard
	actionParams map[string]interface{}

	// Confirmation dialog state
	showingConfirmation     bool
	confirmationMessage     string
	confirmationCallback    func() tea.Cmd
	confirmationSelectedBtn int // 0 = No (default), 1 = Yes

	// Error modal state
	showingError bool
	errorMessage string
}

// NewAppModel creates a new root application model
func NewAppModel(gitOps git.Operations, aiProvider ai.Provider, cfg *domain.Config, cfgManager *config.Manager, repoPath, version string) AppModel {
	dashboard := NewDashboardModel(gitOps, repoPath, cfg)
	dashboard.SetVersion(version)
	githubOps := GitHubOps{}

	return AppModel{
		state:        StateDashboard,
		currentTab:   TabDashboard,
		dashboard:    &dashboard,
		gitOps:       gitOps,
		aiProvider:   aiProvider,
		githubOps:    githubOps,
		cfg:          cfg,
		cfgManager:   cfgManager,
		repoPath:     repoPath,
		version:      version,
		windowWidth:  150,
		windowHeight: 40,
		actionParams: make(map[string]interface{}),
	}
}

// NewAppModelWithOnboarding creates an AppModel that starts in onboarding mode
func NewAppModelWithOnboarding(gitOps git.Operations, cfg *domain.Config, cfgManager *config.Manager, repoPath, version string) AppModel {
	githubOps := GitHubOps{}
	onboarding := NewOnboardingModel(cfg, cfgManager, gitOps, repoPath)

	return AppModel{
		state:          StateOnboarding,
		currentTab:     TabDashboard,
		onboardingView: &onboarding,
		gitOps:         gitOps,
		githubOps:      githubOps,
		cfg:            cfg,
		cfgManager:     cfgManager,
		repoPath:       repoPath,
		version:        version,
		windowWidth:    150,
		windowHeight:   40,
		actionParams:   make(map[string]interface{}),
	}
}

// Messages for async operations

type commitAnalysisMsg struct {
	result *usecase.AnalyzeCommitResponse
	err    error
}

type mergeAnalysisMsg struct {
	result *usecase.AnalyzeMergeResponse
	err    error
}

type commitExecutionMsg struct {
	err error
}

type mergeExecutionMsg struct {
	err error
}

type loadingTickMsg time.Time

// Init initializes the application
func (m AppModel) Init() tea.Cmd {
	// If in onboarding state, init onboarding
	if m.state == StateOnboarding && m.onboardingView != nil {
		return m.onboardingView.Init()
	}

	// Otherwise init dashboard
	if m.dashboard != nil {
		return m.dashboard.Init()
	}

	return nil
}

// Update handles messages and updates the application state
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update window dimensions
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Forward to child views
		var cmd tea.Cmd
		if m.dashboard != nil {
			_, cmd = m.dashboard.Update(msg)
		}
		if m.commitView != nil {
			_, _ = m.commitView.Update(msg)
		}
		if m.mergeView != nil {
			_, _ = m.mergeView.Update(msg)
		}
		if m.settingsView != nil {
			_, _ = m.settingsView.Update(msg)
		}
		if m.onboardingView != nil {
			_, _ = m.onboardingView.Update(msg)
		}
		return m, cmd

	case tea.KeyMsg:
		// Handle error modal
		if m.showingError {
			// Any key dismisses error modal
			m.showingError = false
			m.errorMessage = ""
			return m, nil
		}

		// Handle confirmation dialog
		if m.showingConfirmation {
			switch msg.String() {
			case "left", "h":
				m.confirmationSelectedBtn = 0 // No
				return m, nil
			case "right", "l":
				m.confirmationSelectedBtn = 1 // Yes
				return m, nil
			case "tab":
				m.confirmationSelectedBtn = (m.confirmationSelectedBtn + 1) % 2
				return m, nil
			case "enter":
				m.showingConfirmation = false
				selectedYes := m.confirmationSelectedBtn == 1
				m.confirmationSelectedBtn = 0 // Reset for next time

				if selectedYes && m.confirmationCallback != nil {
					// Execute callback and return to dashboard
					m.state = StateDashboard
					cmd := m.confirmationCallback()
					return m, cmd
				}
				return m, nil
			case "esc":
				// ESC always means No
				m.showingConfirmation = false
				m.confirmationSelectedBtn = 0
				return m, nil
			}
			return m, nil
		}

		// Handle tab switching (only in dashboard state)
		if m.state == StateDashboard {
			switch msg.String() {
			case "1":
				m.currentTab = TabDashboard
				return m, nil
			case "2":
				m.currentTab = TabSettings
				// Lazy-init settings view
				if m.settingsView == nil {
					settings := NewSettingsView(m.cfg, m.cfgManager)
					m.settingsView = settings
				}
				return m, nil
			case "ctrl+tab":
				m.currentTab = (m.currentTab + 1) % 2
				// Lazy-init settings if needed
				if m.currentTab == TabSettings && m.settingsView == nil {
					settings := NewSettingsView(m.cfg, m.cfgManager)
					m.settingsView = settings
				}
				return m, nil
			case "ctrl+shift+tab":
				m.currentTab = (m.currentTab - 1 + 2) % 2
				// Lazy-init settings if needed
				if m.currentTab == TabSettings && m.settingsView == nil {
					settings := NewSettingsView(m.cfg, m.cfgManager)
					m.settingsView = settings
				}
				return m, nil
			}
		}

		// Handle quit in dashboard (q or esc when no submenu and on Dashboard tab)
		if m.state == StateDashboard && m.currentTab == TabDashboard && m.dashboard.activeSubmenu == NoSubmenu {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
		}

		// Handle Esc in different states
		if msg.String() == "esc" {
			switch m.state {
			case StateCommitAnalyzing:
				// Show confirmation to cancel analysis
				m.showingConfirmation = true
				m.confirmationSelectedBtn = 0 // Default to No
				m.confirmationMessage = "Cancel commit analysis?"
				m.confirmationCallback = func() tea.Cmd {
					return m.dashboard.Init()
				}
				return m, nil

			case StateCommitView:
				// Show confirmation to return to dashboard
				m.showingConfirmation = true
				m.confirmationSelectedBtn = 0 // Default to No
				m.confirmationMessage = "Return to dashboard without committing?"
				m.confirmationCallback = func() tea.Cmd {
					return m.dashboard.Init()
				}
				return m, nil

			case StateMergeAnalyzing:
				m.showingConfirmation = true
				m.confirmationSelectedBtn = 0 // Default to No
				m.confirmationMessage = "Cancel merge analysis?"
				m.confirmationCallback = func() tea.Cmd {
					return m.dashboard.Init()
				}
				return m, nil

			case StateMergeView:
				m.showingConfirmation = true
				m.confirmationSelectedBtn = 0 // Default to No
				m.confirmationMessage = "Return to dashboard without merging?"
				m.confirmationCallback = func() tea.Cmd {
					return m.dashboard.Init()
				}
				return m, nil
			}
		}

		// Handle quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case commitAnalysisMsg:
		m.commitAnalysisResult = msg.result
		m.commitAnalysisError = msg.err

		if msg.err != nil {
			// Show error modal instead of returning immediately
			m.showingError = true
			m.errorMessage = fmt.Sprintf("Commit Analysis Failed\n\n%v\n\nPress any key to continue", msg.err)
			m.state = StateDashboard
			return m, m.dashboard.Init()
		}

		// Transition to commit view
		m.state = StateCommitView
		m.commitView = NewCommitViewModel(
			msg.result.Repository,
			msg.result.BranchInfo,
			msg.result.Decision,
			msg.result.TokensUsed,
			msg.result.Model,
			m.windowWidth,
			m.windowHeight,
		)
		return m, m.commitView.Init()

	case mergeAnalysisMsg:
		m.mergeAnalysisResult = msg.result
		m.mergeAnalysisError = msg.err

		if msg.err != nil {
			// Show error modal instead of returning immediately
			m.showingError = true
			m.errorMessage = fmt.Sprintf("Merge Analysis Failed\n\n%v\n\nPress any key to continue", msg.err)
			m.state = StateDashboard
			return m, m.dashboard.Init()
		}

		// Transition to merge view
		m.state = StateMergeView
		mergeView := NewMergeViewModel(msg.result)
		m.mergeView = &mergeView
		return m, m.mergeView.Init()

	case commitExecutionMsg:
		if msg.err != nil {
			PrintError(fmt.Sprintf("Commit failed: %v", msg.err))
		} else {
			PrintSuccess("Commit successful!")
		}
		// Return to dashboard
		m.state = StateDashboard
		return m, m.dashboard.Init()

	case mergeExecutionMsg:
		if msg.err != nil {
			PrintError(fmt.Sprintf("Merge failed: %v", msg.err))
		} else {
			PrintSuccess("Merge successful!")
		}
		// Return to dashboard
		m.state = StateDashboard
		return m, m.dashboard.Init()

	case loadingTickMsg:
		// Animate loading dots
		if m.state == StateCommitAnalyzing || m.state == StateMergeAnalyzing || m.state == StateCommitExecuting || m.state == StateMergeExecuting {
			m.loadingDots = (m.loadingDots + 1) % 4
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				return loadingTickMsg(t)
			})
		}
		return m, nil
	}

	// Route messages to appropriate child model based on state
	switch m.state {
	case StateDashboard:
		// Route to active tab
		if m.currentTab == TabSettings && m.settingsView != nil {
			// Settings view handles its own messages
			updated, cmd := m.settingsView.Update(msg)
			m.settingsView = &updated
			return m, cmd
		}

		// Dashboard tab
		updated, cmd := m.dashboard.Update(msg)
		dashModel := updated.(DashboardModel)
		m.dashboard = &dashModel

		// Check if dashboard has an action to perform
		action := m.dashboard.GetAction()
		params := m.dashboard.GetActionParams()

		// Reset the action immediately after consuming it to prevent loops
		if action != ActionNone {
			m.dashboard.action = ActionNone
			m.dashboard.actionParams = make(map[string]interface{})
		}

		switch action {
		case ActionCommit:
			// Start commit analysis
			m.actionParams = params
			m.state = StateCommitAnalyzing
			m.loadingMessage = "Analyzing changes with AI"
			return m, tea.Batch(
				m.startCommitAnalysis(params),
				tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
					return loadingTickMsg(t)
				}),
			)

		case ActionMerge:
			// Start merge analysis
			m.actionParams = params
			m.state = StateMergeAnalyzing
			m.loadingMessage = "Analyzing merge with AI"
			return m, tea.Batch(
				m.startMergeAnalysis(params),
				tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
					return loadingTickMsg(t)
				}),
			)

		case ActionSwitchBranch:
			// Handle branch switching
			branch, _ := params["branch"].(string)
			if branch != "" {
				ctx := context.Background()
				if err := m.gitOps.CheckoutBranch(ctx, m.repoPath, branch); err != nil {
					PrintError(fmt.Sprintf("Failed to switch branch: %v", err))
				} else {
					PrintSuccess(fmt.Sprintf("Switched to branch: %s", branch))
				}
				// Refresh dashboard
				return m, m.dashboard.Init()
			}

		case ActionFetch:
			// Fetch updates from remote
			ctx := context.Background()
			PrintInfo("Fetching from remote...")
			if err := m.gitOps.Fetch(ctx, m.repoPath); err != nil {
				PrintError(fmt.Sprintf("Failed to fetch: %v", err))
			} else {
				PrintSuccess("Fetched updates from remote")
			}
			// Refresh dashboard to show new sync status
			return m, m.dashboard.Init()

		case ActionPull:
			// Pull changes from remote
			ctx := context.Background()
			PrintInfo("Pulling from remote...")
			if err := m.gitOps.Pull(ctx, m.repoPath); err != nil {
				PrintError(fmt.Sprintf("Failed to pull: %v", err))
			} else {
				PrintSuccess("Pulled changes from remote")
			}
			// Refresh dashboard
			return m, m.dashboard.Init()

		case ActionPush:
			// Push commits to remote
			ctx := context.Background()
			branch, _ := m.gitOps.GetCurrentBranch(ctx, m.repoPath)
			PrintInfo(fmt.Sprintf("Pushing to remote (%s)...", branch))
			if err := m.gitOps.Push(ctx, m.repoPath, branch, false); err != nil {
				PrintError(fmt.Sprintf("Failed to push: %v", err))
			} else {
				PrintSuccess("Pushed commits to remote")
			}
			// Refresh dashboard
			return m, m.dashboard.Init()

		case ActionViewGitHub:
			// Open repository in browser using gh CLI
			ctx := context.Background()
			PrintInfo("Opening repository in browser...")
			if err := m.githubOps.ViewRepoWeb(ctx, m.repoPath); err != nil {
				PrintError(fmt.Sprintf("Failed to open repository: %v", err))
			} else {
				PrintSuccess("Opened repository in browser")
			}
			// Stay on dashboard
			return m, cmd

		case ActionShowGitHubInfo:
			// Show GitHub repository information
			ctx := context.Background()
			PrintInfo("Fetching GitHub repository info...")
			info, err := m.githubOps.GetRepoInfo(ctx, m.repoPath)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to get repository info: %v", err))
			} else {
				// Display basic info
				PrintInfo(fmt.Sprintf("\nGitHub Repository: %s", info.FullName))
				if info.Description != "" {
					PrintInfo(fmt.Sprintf("Description: %s", info.Description))
				}
				if info.IsPrivate {
					PrintInfo("Visibility: Private")
				} else {
					PrintInfo("Visibility: Public")
				}
				if info.HTMLURL != "" {
					PrintInfo(fmt.Sprintf("URL: %s", info.HTMLURL))
				}
			}
			// Stay on dashboard
			return m, cmd

		case ActionSetupRemote:
			// Transition to onboarding GitHub step
			PrintInfo("Launching remote setup...")
			onboarding := NewOnboardingModel(m.cfg, m.cfgManager, m.gitOps, m.repoPath)
			// Jump directly to GitHub step
			onboarding.state = OnboardingGitHub
			onboarding.currentStep = 3 // GitHub is step 3
			screen := NewOnboardingGitHubScreen(3, 8, m.cfg, m.repoPath)
			onboarding.githubScreen = &screen
			m.onboardingView = &onboarding
			m.state = StateOnboarding
			return m, screen.Init()

		case ActionRefresh:
			// Refresh dashboard
			PrintInfo("Refreshing dashboard...")
			return m, m.dashboard.Init()
		}

		return m, cmd

	case StateOnboarding:
		if m.onboardingView == nil {
			return m, nil
		}

		updated, cmd := m.onboardingView.Update(msg)
		m.onboardingView = &updated

		// Check if onboarding is complete
		if m.onboardingView.IsCompleted() {
			// Reload config after onboarding
			cfg, err := m.cfgManager.Load()
			if err == nil {
				m.cfg = cfg
			}

			// Initialize dashboard
			dashboard := NewDashboardModel(m.gitOps, m.repoPath, m.cfg)
			dashboard.SetVersion(m.version)
			m.dashboard = &dashboard

			// Transition to dashboard
			m.state = StateDashboard
			PrintSuccess("Setup complete! Welcome to GitMind.")
			return m, m.dashboard.Init()
		}

		// Check if onboarding was cancelled
		if m.onboardingView.IsCancelled() {
			// Exit application
			PrintInfo("Setup cancelled. You can run 'gm onboard' to set up GitMind later.")
			return m, tea.Quit
		}

		return m, cmd

	case StateCommitView:
		if m.commitView == nil {
			return m, nil
		}

		updated, cmd := m.commitView.Update(msg)
		commitModel := updated.(CommitViewModel)
		m.commitView = &commitModel

		// Check if commit view wants to return to dashboard
		if m.commitView.ShouldReturnToDashboard() {
			m.state = StateDashboard
			return m, m.dashboard.Init()
		}

		// Check if commit view has a decision
		if m.commitView.HasDecision() {
			selectedOption := m.commitView.GetSelectedOption()
			m.state = StateCommitExecuting
			m.loadingMessage = "Executing commit"
			return m, tea.Batch(
				m.executeCommit(selectedOption),
				tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
					return loadingTickMsg(t)
				}),
			)
		}

		return m, cmd

	case StateMergeView:
		if m.mergeView == nil {
			return m, nil
		}

		updated, cmd := m.mergeView.Update(msg)
		mergeModel := updated.(MergeViewModel)
		m.mergeView = &mergeModel

		// Check if merge view wants to return to dashboard
		if m.mergeView.ShouldReturnToDashboard() {
			m.state = StateDashboard
			return m, m.dashboard.Init()
		}

		// Check if merge view has a decision
		if m.mergeView.HasDecision() {
			strategy := m.mergeView.GetSelectedStrategy()
			message := m.mergeView.GetMergeMessage()
			m.state = StateMergeExecuting
			m.loadingMessage = "Executing merge"
			return m, tea.Batch(
				m.executeMerge(strategy, message),
				tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
					return loadingTickMsg(t)
				}),
			)
		}

		return m, cmd
	}

	return m, nil
}

// View renders the application
func (m AppModel) View() string {
	var content string

	// Handle onboarding state (full screen, no tabs)
	if m.state == StateOnboarding {
		if m.onboardingView != nil {
			return m.onboardingView.View()
		}
		return "Loading onboarding..."
	}

	// For non-dashboard states, render overlays directly without tabs
	if m.state != StateDashboard {
		var overlayView string

		// Render overlay based on state
		switch m.state {
		case StateCommitAnalyzing, StateCommitExecuting:
			overlayView = m.renderLoadingOverlay()

		case StateCommitView:
			if m.commitView != nil {
				overlayView = m.commitView.View()
			}

		case StateMergeAnalyzing, StateMergeExecuting:
			overlayView = m.renderLoadingOverlay()

		case StateMergeView:
			if m.mergeView != nil {
				overlayView = m.mergeView.View()
			}
		}

		// Show confirmation dialog if active (completely blocks screen)
		if m.showingConfirmation {
			return m.renderConfirmationDialog()
		}

		// Show error modal if active (completely blocks screen)
		if m.showingError {
			return m.renderErrorModal()
		}

		return overlayView
	}

	// Show confirmation dialog if active (highest priority - blocks dashboard)
	if m.showingConfirmation {
		return m.renderConfirmationDialog()
	}

	// Show error modal if active (blocks dashboard)
	if m.showingError {
		return m.renderErrorModal()
	}

	// Render tab bar
	tabBar := m.renderTabBar()

	// Render active tab content
	switch m.currentTab {
	case TabDashboard:
		content = m.dashboard.View()
	case TabSettings:
		if m.settingsView != nil {
			content = m.settingsView.View()
		} else {
			content = "Loading settings..."
		}
	}

	// Combine tab bar and content
	view := tabBar + "\n" + content

	return view
}

// renderLoadingOverlay renders a loading message overlay
func (m AppModel) renderLoadingOverlay() string {
	styles := GetGlobalThemeManager().GetStyles()

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary).
		Render("ℹ AI ANALYSIS")

	// Operation type
	operation := "Analyzing Changes"
	switch m.state {
	case StateMergeAnalyzing:
		operation = "Analyzing Merge"
	case StateCommitExecuting:
		operation = "Executing Commit"
	case StateMergeExecuting:
		operation = "Executing Merge"
	}

	opText := lipgloss.NewStyle().
		Foreground(styles.ColorSecondary).
		Bold(true).
		Render(operation)

	// Loading animation
	dots := ""
	for i := 0; i < m.loadingDots; i++ {
		dots += "."
	}
	// Pad dots to avoid layout jumping
	dots = fmt.Sprintf("%-3s", dots)

	loadingText := styles.Loading.Render(m.loadingMessage + dots)

	// Content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		opText,
		"",
		loadingText,
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Please wait while we process your request..."),
	)

	// Create a centered box
	box := styles.CommitBox.
		Padding(2, 4).
		Width(60).
		Align(lipgloss.Center).
		Render(content)

	return "\n\n" + lipgloss.Place(
		m.windowWidth, m.windowHeight-4, // Adjust for margins
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

// renderConfirmationDialog renders a full-screen confirmation dialog with buttons
func (m AppModel) renderConfirmationDialog() string {
	styles := GetGlobalThemeManager().GetStyles()

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Render("ℹ Confirmation")

	// Message
	message := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(m.confirmationMessage)

	// Button styles
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorMuted)

	buttonActiveStyle := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(2).
		Bold(true).
		Background(styles.ColorPrimary).
		Foreground(lipgloss.Color("#000000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary)

	// Render buttons
	noBtn := "No"
	yesBtn := "Yes"

	if m.confirmationSelectedBtn == 0 {
		noBtn = buttonActiveStyle.Render(noBtn)
		yesBtn = buttonStyle.Render(yesBtn)
	} else {
		noBtn = buttonStyle.Render(noBtn)
		yesBtn = buttonActiveStyle.Render(yesBtn)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, noBtn, yesBtn)

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render("←/→ or Tab to switch  •  Enter to confirm  •  Esc to cancel")

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		message,
		"",
		"",
		buttons,
		"",
		helpText,
	)

	// Create a modal box with primary color background
	theme := GetGlobalThemeManager().GetCurrentTheme()
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Confirmation)).
		Width(60)

	return "\n\n" + lipgloss.Place(
		80, 20,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderErrorModal renders an error modal
func (m AppModel) renderErrorModal() string {
	styles := GetGlobalThemeManager().GetStyles()

	title := lipgloss.NewStyle().
		Foreground(styles.ColorError).
		Bold(true).
		Render("✗ ERROR")

	message := lipgloss.NewStyle().
		Foreground(styles.ColorError).
		Render(m.errorMessage)

	content := title + "\n\n" + message

	return styles.CommitBox.
		BorderForeground(styles.ColorError).
		Render(content)
}

// renderTabBar renders the tab bar at the top
func (m AppModel) renderTabBar() string {
	styles := GetGlobalThemeManager().GetStyles()
	var tabs []string

	// Dashboard tab
	if m.currentTab == TabDashboard {
		tabs = append(tabs, styles.TabActive.Render("[1] Dashboard"))
	} else {
		tabs = append(tabs, styles.TabInactive.Render("[1] Dashboard"))
	}

	// Spacer
	tabs = append(tabs, "  ")

	// Settings tab
	if m.currentTab == TabSettings {
		tabs = append(tabs, styles.TabActive.Render("[2] Settings"))
	} else {
		tabs = append(tabs, styles.TabInactive.Render("[2] Settings"))
	}

	tabLine := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return styles.TabBar.Render(tabLine)
}

// startCommitAnalysis initiates the commit analysis workflow
func (m AppModel) startCommitAnalysis(params map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get parameters
		customMessage, _ := params["message"].(string)
		useConventional, _ := params["conventional"].(bool)

		// Create use case
		analyzeUC := usecase.NewAnalyzeCommitUseCase(m.gitOps, m.aiProvider)

		// Create API key
		apiKey, err := domain.NewAPIKey(m.cfg.AI.APIKey, m.cfg.AI.Provider)
		if err != nil {
			return commitAnalysisMsg{result: nil, err: err}
		}
		tier, err := domain.ParseAPITier(m.cfg.AI.APITier)
		if err != nil {
			tier = domain.TierUnknown
		}
		apiKey.SetTier(tier)

		// Build request
		req := usecase.AnalyzeCommitRequest{
			RepoPath:               m.repoPath,
			ProtectedBranches:      m.cfg.Git.ProtectedBranches,
			UseConventionalCommits: useConventional,
			UserPrompt:             customMessage,
			APIKey:                 apiKey,
		}

		// Execute analysis
		result, err := analyzeUC.Execute(ctx, req)

		return commitAnalysisMsg{result: result, err: err}
	}
}

// startMergeAnalysis initiates the merge analysis workflow
func (m AppModel) startMergeAnalysis(params map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get parameters
		sourceBranch, _ := params["source"].(string)
		targetBranch, _ := params["target"].(string)

		// Create use case
		analyzeUC := usecase.NewAnalyzeMergeUseCase(m.gitOps, m.aiProvider)

		// Create API key
		apiKey, err := domain.NewAPIKey(m.cfg.AI.APIKey, m.cfg.AI.Provider)
		if err != nil {
			return mergeAnalysisMsg{result: nil, err: err}
		}
		tier, err := domain.ParseAPITier(m.cfg.AI.APITier)
		if err != nil {
			tier = domain.TierUnknown
		}
		apiKey.SetTier(tier)

		// Build request
		req := usecase.AnalyzeMergeRequest{
			RepoPath:          m.repoPath,
			SourceBranch:      sourceBranch,
			TargetBranch:      targetBranch,
			ProtectedBranches: m.cfg.Git.ProtectedBranches,
			APIKey:            apiKey,
		}

		// Execute analysis
		result, err := analyzeUC.Execute(ctx, req)

		return mergeAnalysisMsg{result: result, err: err}
	}
}

// executeCommit executes the selected commit action
func (m AppModel) executeCommit(option *CommitOption) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create execute use case
		executeUC := usecase.NewExecuteCommitUseCase(m.gitOps)

		// Use the message from the option if available, otherwise fallback to decision
		msg := option.Message
		if msg == nil {
			msg = m.commitAnalysisResult.Decision.SuggestedMessage()
		}

		// Build request
		req := usecase.ExecuteCommitRequest{
			RepoPath:      m.repoPath,
			Decision:      m.commitAnalysisResult.Decision,
			Action:        option.Action,
			CommitMessage: msg,
			BranchName:    option.BranchName,
			StageAll:      true,
		}

		// Execute commit
		resp, err := executeUC.Execute(ctx, req)
		if err != nil {
			return commitExecutionMsg{err: err}
		}

		// If manual review, don't push
		if req.Action == domain.ActionReview {
			return commitExecutionMsg{err: nil}
		}

		// Determine branch to push
		branchToPush := req.BranchName
		if branchToPush == "" {
			if resp.BranchCreated != "" {
				branchToPush = resp.BranchCreated
			} else {
				var err error
				branchToPush, err = m.gitOps.GetCurrentBranch(ctx, m.repoPath)
				if err != nil {
					return commitExecutionMsg{err: fmt.Errorf("commit successful but failed to get current branch for push: %w", err)}
				}
			}
		}

		// Push changes
		// The Push implementation automatically handles -u if upstream is missing
		if err := m.gitOps.Push(ctx, m.repoPath, branchToPush, false); err != nil {
			return commitExecutionMsg{err: fmt.Errorf("commit successful but push failed: %w", err)}
		}

		return commitExecutionMsg{err: nil}
	}
}

// executeMerge executes the selected merge strategy
func (m AppModel) executeMerge(strategy string, message string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create execute use case
		executeUC := usecase.NewExecuteMergeUseCase(m.gitOps)

		// Create commit message from string
		mergeMsg, _ := domain.NewCommitMessage(message)

		// Build request
		req := usecase.ExecuteMergeRequest{
			RepoPath:     m.repoPath,
			SourceBranch: m.mergeAnalysisResult.SourceBranchInfo.Name(),
			TargetBranch: m.mergeAnalysisResult.TargetBranch,
			Strategy:     strategy,
			MergeMessage: mergeMsg,
		}

		// Execute merge
		_, err := executeUC.Execute(ctx, req)

		return mergeExecutionMsg{err: err}
	}
}
