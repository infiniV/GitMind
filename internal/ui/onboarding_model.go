package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingState represents the current onboarding step
type OnboardingState int

const (
	OnboardingWelcome OnboardingState = iota
	OnboardingGitInit
	OnboardingGitHub
	OnboardingBranches
	OnboardingCommits
	OnboardingNaming
	OnboardingAI
	OnboardingSummary
	OnboardingComplete
)

// OnboardingModel manages the onboarding workflow
type OnboardingModel struct {
	state      OnboardingState
	config     *domain.Config
	cfgManager *config.Manager
	gitOps     git.Operations
	repoPath   string

	// Current step tracking
	currentStep int
	totalSteps  int

	// Skip flag
	skipAll bool

	// Sub-models for each screen
	welcomeScreen   *OnboardingWelcomeScreen
	gitInitScreen   *OnboardingGitInitScreen
	githubScreen    *OnboardingGitHubScreen
	branchesScreen  *OnboardingBranchesScreen
	commitsScreen   *OnboardingCommitsScreen
	namingScreen    *OnboardingNamingScreen
	aiScreen        *OnboardingAIScreen
	summaryScreen   *OnboardingSummaryScreen

	// Window dimensions
	windowWidth  int
	windowHeight int

	// Completion flag
	completed bool
	cancelled bool
}

// NewOnboardingModel creates a new onboarding model
func NewOnboardingModel(cfg *domain.Config, cfgManager *config.Manager, gitOps git.Operations, repoPath string) OnboardingModel {
	// Initialize the welcome screen
	welcomeScreen := NewOnboardingWelcomeScreen(1, 8)

	return OnboardingModel{
		state:         OnboardingWelcome,
		config:        cfg,
		cfgManager:    cfgManager,
		gitOps:        gitOps,
		repoPath:      repoPath,
		currentStep:   1,
		totalSteps:    8,
		skipAll:       false,
		completed:     false,
		cancelled:     false,
		welcomeScreen: &welcomeScreen,
		windowWidth:   100, // Default fallback
		windowHeight:  40,  // Default fallback
	}
}

// Init initializes the onboarding
func (m OnboardingModel) Init() tea.Cmd {
	if m.welcomeScreen != nil {
		return m.welcomeScreen.Init()
	}
	return nil
}

// Update handles messages
func (m OnboardingModel) Update(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	// Handle window resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
	}

	// Handle global keys
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	// Route to current screen
	switch m.state {
	case OnboardingWelcome:
		return m.updateWelcomeScreen(msg)
	case OnboardingGitInit:
		return m.updateGitInitScreen(msg)
	case OnboardingGitHub:
		return m.updateGitHubScreen(msg)
	case OnboardingBranches:
		return m.updateBranchesScreen(msg)
	case OnboardingCommits:
		return m.updateCommitsScreen(msg)
	case OnboardingNaming:
		return m.updateNamingScreen(msg)
	case OnboardingAI:
		return m.updateAIScreen(msg)
	case OnboardingSummary:
		return m.updateSummaryScreen(msg)
	case OnboardingComplete:
		m.completed = true
		return m, nil
	}

	return m, nil
}

// View renders the current screen
func (m OnboardingModel) View() string {
	switch m.state {
	case OnboardingWelcome:
		if m.welcomeScreen != nil {
			return m.welcomeScreen.View()
		}
	case OnboardingGitInit:
		if m.gitInitScreen != nil {
			return m.gitInitScreen.View()
		}
	case OnboardingGitHub:
		if m.githubScreen != nil {
			return m.githubScreen.View()
		}
	case OnboardingBranches:
		if m.branchesScreen != nil {
			return m.branchesScreen.View()
		}
	case OnboardingCommits:
		if m.commitsScreen != nil {
			return m.commitsScreen.View()
		}
	case OnboardingNaming:
		if m.namingScreen != nil {
			return m.namingScreen.View()
		}
	case OnboardingAI:
		if m.aiScreen != nil {
			return m.aiScreen.View()
		}
	case OnboardingSummary:
		if m.summaryScreen != nil {
			return m.summaryScreen.View()
		}
	}

	return "Loading..."
}


// Helper methods for screen updates (to be implemented with each screen)
func (m OnboardingModel) updateWelcomeScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.welcomeScreen == nil {
		return m, nil
	}

	updated, cmd := m.welcomeScreen.Update(msg)
	m.welcomeScreen = &updated

	if m.welcomeScreen.ShouldContinue() {
		m.state = OnboardingGitInit
		m.currentStep++
		// Initialize git init screen
		screen := NewOnboardingGitInitScreen(m.currentStep, m.totalSteps, m.gitOps, m.repoPath)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.gitInitScreen = &screen
		return m, screen.Init()
	}

	if m.welcomeScreen.ShouldSkip() {
		m.cancelled = true
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateGitInitScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.gitInitScreen == nil {
		return m, nil
	}

	updated, cmd := m.gitInitScreen.Update(msg)
	m.gitInitScreen = &updated

	if m.gitInitScreen.ShouldContinue() {
		m.state = OnboardingGitHub
		m.currentStep++
		screen := NewOnboardingGitHubScreen(m.currentStep, m.totalSteps, m.config, m.repoPath)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.githubScreen = &screen
		return m, screen.Init()
	}

	if m.gitInitScreen.ShouldGoBack() {
		m.state = OnboardingWelcome
		m.currentStep--
		// Welcome screen already exists, just return
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateGitHubScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.githubScreen == nil {
		return m, nil
	}

	updated, cmd := m.githubScreen.Update(msg)
	m.githubScreen = &updated

	if m.githubScreen.ShouldContinue() {
		m.state = OnboardingBranches
		m.currentStep++
		screen := NewOnboardingBranchesScreen(m.currentStep, m.totalSteps, m.config)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.branchesScreen = &screen
		return m, screen.Init()
	}

	if m.githubScreen.ShouldGoBack() {
		m.state = OnboardingGitInit
		m.currentStep--
		// Git init screen already exists
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateBranchesScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.branchesScreen == nil {
		return m, nil
	}

	updated, cmd := m.branchesScreen.Update(msg)
	m.branchesScreen = &updated

	if m.branchesScreen.ShouldContinue() {
		m.state = OnboardingCommits
		m.currentStep++
		screen := NewOnboardingCommitsScreen(m.currentStep, m.totalSteps, m.config)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.commitsScreen = &screen
		return m, screen.Init()
	}

	if m.branchesScreen.ShouldGoBack() {
		m.state = OnboardingGitHub
		m.currentStep--
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateCommitsScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.commitsScreen == nil {
		return m, nil
	}

	updated, cmd := m.commitsScreen.Update(msg)
	m.commitsScreen = &updated

	if m.commitsScreen.ShouldContinue() {
		m.state = OnboardingNaming
		m.currentStep++
		screen := NewOnboardingNamingScreen(m.currentStep, m.totalSteps, m.config)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.namingScreen = &screen
		return m, screen.Init()
	}

	if m.commitsScreen.ShouldGoBack() {
		m.state = OnboardingBranches
		m.currentStep--
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateNamingScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.namingScreen == nil {
		return m, nil
	}

	updated, cmd := m.namingScreen.Update(msg)
	m.namingScreen = &updated

	if m.namingScreen.ShouldContinue() {
		m.state = OnboardingAI
		m.currentStep++
		screen := NewOnboardingAIScreen(m.currentStep, m.totalSteps, m.config)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.aiScreen = &screen
		return m, screen.Init()
	}

	if m.namingScreen.ShouldGoBack() {
		m.state = OnboardingCommits
		m.currentStep--
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateAIScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.aiScreen == nil {
		return m, nil
	}

	updated, cmd := m.aiScreen.Update(msg)
	m.aiScreen = &updated

	if m.aiScreen.ShouldContinue() {
		m.state = OnboardingSummary
		m.currentStep++
		screen := NewOnboardingSummaryScreen(m.currentStep, m.totalSteps, m.config)
		screen.width = m.windowWidth
		screen.height = m.windowHeight
		m.summaryScreen = &screen
		return m, screen.Init()
	}

	if m.aiScreen.ShouldGoBack() {
		m.state = OnboardingNaming
		m.currentStep--
		return m, nil
	}

	return m, cmd
}

func (m OnboardingModel) updateSummaryScreen(msg tea.Msg) (OnboardingModel, tea.Cmd) {
	if m.summaryScreen == nil {
		return m, nil
	}

	updated, cmd := m.summaryScreen.Update(msg)
	m.summaryScreen = &updated

	if m.summaryScreen.ShouldSave() {
		// Save configuration
		if err := m.cfgManager.Save(m.config); err != nil {
			// Handle error (could show error screen)
			PrintError("Failed to save configuration: " + err.Error())
		}
		m.state = OnboardingComplete
		m.completed = true
		return m, nil
	}

	if m.summaryScreen.ShouldGoBack() {
		m.state = OnboardingAI
		m.currentStep--
		return m, nil
	}

	return m, cmd
}

// IsCompleted returns true if onboarding is complete
func (m OnboardingModel) IsCompleted() bool {
	return m.completed
}

// IsCancelled returns true if onboarding was cancelled
func (m OnboardingModel) IsCancelled() bool {
	return m.cancelled
}
