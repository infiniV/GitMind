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
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/usecase"
)

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
	state         AppState
	previousState AppState

	// Tab management
	currentTab Tab

	// Child models
	dashboard    *DashboardModel
	commitView   *CommitViewModel
	mergeView    *MergeViewModel
	settingsView *SettingsView

	// Dependencies
	gitOps     git.Operations
	aiProvider ai.Provider
	cfg        *domain.Config
	cfgManager *config.Manager
	repoPath   string

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
	showingConfirmation  bool
	confirmationMessage  string
	confirmationCallback func() tea.Cmd
}

// NewAppModel creates a new root application model
func NewAppModel(gitOps git.Operations, aiProvider ai.Provider, cfg *domain.Config, cfgManager *config.Manager, repoPath string) AppModel {
	dashboard := NewDashboardModel(gitOps, repoPath)

	return AppModel{
		state:        StateDashboard,
		currentTab:   TabDashboard,
		dashboard:    &dashboard,
		gitOps:       gitOps,
		aiProvider:   aiProvider,
		cfg:          cfg,
		cfgManager:   cfgManager,
		repoPath:     repoPath,
		actionParams: make(map[string]interface{}),
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
	return m.dashboard.Init()
}

// Update handles messages and updates the application state
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation dialog
		if m.showingConfirmation {
			switch msg.String() {
			case "y", "Y":
				m.showingConfirmation = false
				if m.confirmationCallback != nil {
					return m, m.confirmationCallback()
				}
				return m, nil
			case "n", "N", "enter", "esc":
				m.showingConfirmation = false
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
				m.confirmationMessage = "Cancel commit analysis?"
				m.confirmationCallback = func() tea.Cmd {
					m.state = StateDashboard
					return m.dashboard.Init()
				}
				return m, nil

			case StateCommitView:
				// Show confirmation to return to dashboard
				m.showingConfirmation = true
				m.confirmationMessage = "Return to dashboard without committing?"
				m.confirmationCallback = func() tea.Cmd {
					m.state = StateDashboard
					return m.dashboard.Init()
				}
				return m, nil

			case StateMergeAnalyzing:
				m.showingConfirmation = true
				m.confirmationMessage = "Cancel merge analysis?"
				m.confirmationCallback = func() tea.Cmd {
					m.state = StateDashboard
					return m.dashboard.Init()
				}
				return m, nil

			case StateMergeView:
				m.showingConfirmation = true
				m.confirmationMessage = "Return to dashboard without merging?"
				m.confirmationCallback = func() tea.Cmd {
					m.state = StateDashboard
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
			// Show error and return to dashboard
			PrintError(fmt.Sprintf("Analysis failed: %v", msg.err))
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
		)
		return m, m.commitView.Init()

	case mergeAnalysisMsg:
		m.mergeAnalysisResult = msg.result
		m.mergeAnalysisError = msg.err

		if msg.err != nil {
			PrintError(fmt.Sprintf("Analysis failed: %v", msg.err))
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
			m.state = StateMergeExecuting
			m.loadingMessage = "Executing merge"
			return m, tea.Batch(
				m.executeMerge(strategy),
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

	// For non-dashboard states, render overlays directly without tabs
	if m.state != StateDashboard {
		// Render overlay based on state
		switch m.state {
		case StateCommitAnalyzing, StateCommitExecuting:
			return m.renderLoadingOverlay()

		case StateCommitView:
			if m.commitView != nil {
				return m.commitView.View()
			}

		case StateMergeAnalyzing, StateMergeExecuting:
			return m.renderLoadingOverlay()

		case StateMergeView:
			if m.mergeView != nil {
				return m.mergeView.View()
			}
		}
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

	// Show confirmation dialog if active
	if m.showingConfirmation {
		return view + "\n\n" + m.renderConfirmationDialog()
	}

	return view
}

// renderLoadingOverlay renders a loading message overlay
func (m AppModel) renderLoadingOverlay() string {
	dots := ""
	for i := 0; i < m.loadingDots; i++ {
		dots += "."
	}

	loadingText := loadingStyle.Render(m.loadingMessage + dots)

	// Create a centered box
	box := commitBoxStyle.Render(loadingText)

	return "\n\n" + box
}

// renderConfirmationDialog renders a confirmation dialog
func (m AppModel) renderConfirmationDialog() string {
	message := warningStyle.Render(m.confirmationMessage)
	prompt := footerStyle.Render("[y/N]")

	content := message + "\n\n" + prompt

	return commitBoxStyle.Render(content)
}

// renderTabBar renders the tab bar at the top
func (m AppModel) renderTabBar() string {
	var tabs []string

	// Dashboard tab
	if m.currentTab == TabDashboard {
		tabs = append(tabs, tabActiveStyle.Render("[1] Dashboard"))
	} else {
		tabs = append(tabs, tabInactiveStyle.Render("[1] Dashboard"))
	}

	// Spacer
	tabs = append(tabs, "  ")

	// Settings tab
	if m.currentTab == TabSettings {
		tabs = append(tabs, tabActiveStyle.Render("[2] Settings"))
	} else {
		tabs = append(tabs, tabInactiveStyle.Render("[2] Settings"))
	}

	tabLine := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return tabBarStyle.Render(tabLine)
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
func (m AppModel) executeCommit(option *domain.Alternative) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create execute use case
		executeUC := usecase.NewExecuteCommitUseCase(m.gitOps)

		// Build request
		req := usecase.ExecuteCommitRequest{
			RepoPath:      m.repoPath,
			Decision:      m.commitAnalysisResult.Decision,
			Action:        option.Action,
			CommitMessage: m.commitAnalysisResult.Decision.SuggestedMessage(),
			BranchName:    option.BranchName,
			StageAll:      true,
		}

		// Execute commit
		_, err := executeUC.Execute(ctx, req)

		return commitExecutionMsg{err: err}
	}
}

// executeMerge executes the selected merge strategy
func (m AppModel) executeMerge(strategy string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create execute use case
		executeUC := usecase.NewExecuteMergeUseCase(m.gitOps)

		// Build request
		req := usecase.ExecuteMergeRequest{
			RepoPath:     m.repoPath,
			SourceBranch: m.mergeAnalysisResult.SourceBranchInfo.Name(),
			TargetBranch: m.mergeAnalysisResult.TargetBranch,
			Strategy:     strategy,
			MergeMessage: m.mergeAnalysisResult.MergeMessage,
		}

		// Execute merge
		_, err := executeUC.Execute(ctx, req)

		return mergeExecutionMsg{err: err}
	}
}
