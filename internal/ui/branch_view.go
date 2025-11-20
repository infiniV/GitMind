package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
	"github.com/yourusername/gitman/internal/ui/layout"
	"github.com/yourusername/gitman/internal/usecase"
)

// BranchViewState represents the current state of the branch view.
type BranchViewState int

const (
	BranchViewBrowsing BranchViewState = iota
	BranchViewExpanded
	BranchViewDeleting
	BranchViewForceDeletePrompt
	BranchViewDeleteRemotePrompt
	BranchViewRenaming
	BranchViewSettingUpstream
	BranchViewManaging
)

// BranchViewModel represents the state of the branch management view.
type BranchViewModel struct {
	// Data
	branches          []*domain.BranchInfo
	currentBranch     string
	repoPath          string
	config            *domain.Config

	// State
	state             BranchViewState
	selectedIndex     int
	expandedIndex     int // -1 when collapsed

	// UI Components
	viewport          viewport.Model
	detailViewport    viewport.Model
	renameInput       textinput.Model
	upstreamInput     textinput.Model

	// Actions
	deleteConfirmed     bool
	deleteRemote        bool
	forceDelete         bool
	selectedBranch      *domain.BranchInfo
	remoteName          string
	confirmSelectedBtn  int // 0 = No, 1 = Yes

	// Dimensions
	windowWidth       int
	windowHeight      int

	// Navigation
	returnToDashboard bool

	// Use cases
	manageBranchesUC  *usecase.ManageBranchesUseCase

	// Error handling
	errorMessage      string
	successMessage    string
}

// NewBranchViewModel creates a new branch view model.
func NewBranchViewModel(
	repoPath string,
	config *domain.Config,
	gitOps git.Operations,
) BranchViewModel {
	// Initialize viewports
	vp := viewport.New(50, 20)
	detailVp := viewport.New(50, 20)

	// Initialize text inputs
	renameInput := textinput.New()
	renameInput.Placeholder = "new-branch-name"
	renameInput.CharLimit = 50

	upstreamInput := textinput.New()
	upstreamInput.Placeholder = "origin/branch-name"
	upstreamInput.CharLimit = 50

	m := BranchViewModel{
		branches:          []*domain.BranchInfo{},
		currentBranch:     "",
		repoPath:          repoPath,
		config:            config,
		state:             BranchViewBrowsing,
		selectedIndex:     0,
		expandedIndex:     -1,
		viewport:          vp,
		detailViewport:    detailVp,
		renameInput:       renameInput,
		upstreamInput:     upstreamInput,
		deleteConfirmed:    false,
		deleteRemote:       false,
		confirmSelectedBtn: 0, // Default to No
		windowWidth:        120,
		windowHeight:       30,
		returnToDashboard:  false,
		manageBranchesUC:   usecase.NewManageBranchesUseCase(gitOps),
		errorMessage:       "",
		successMessage:     "",
	}

	// Set initial loading content
	m.viewport.SetContent("Loading branches...")

	return m
}

// Init initializes the branch view.
func (m BranchViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadBranches(),
		textinput.Blink,
	)
}

// loadBranches loads all branches with their information.
func (m BranchViewModel) loadBranches() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		branches, err := m.manageBranchesUC.GetAllBranches(
			ctx,
			m.repoPath,
			m.config.Git.ProtectedBranches,
		)
		if err != nil {
			return branchLoadErrorMsg{err}
		}

		return branchesLoadedMsg{branches}
	}
}

// branchesLoadedMsg is sent when branches are loaded successfully.
type branchesLoadedMsg struct {
	branches []*domain.BranchInfo
}

// branchLoadErrorMsg is sent when branch loading fails.
type branchLoadErrorMsg struct {
	err error
}

// branchDeletedMsg is sent when a branch is deleted successfully.
type branchDeletedMsg struct {
	response *usecase.DeleteBranchResponse
}

// branchRenamedMsg is sent when a branch is renamed successfully.
type branchRenamedMsg struct {
	response *usecase.RenameBranchResponse
}

// upstreamSetMsg is sent when upstream is set successfully.
type upstreamSetMsg struct {
	response *usecase.SetUpstreamResponse
}

// Update handles messages and updates the branch view.
func (m BranchViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		if m.state == BranchViewExpanded {
			// Calculate split widths for 35/65 layout
			leftWidth, rightWidth := layout.CalculateSplitWidths(msg.Width, layout.SplitRatio35_65)

			// Update viewport sizes
			headerHeight := 10
			footerHeight := 3
			contentHeight := msg.Height - headerHeight - footerHeight

			m.viewport.Width = leftWidth - 4
			m.viewport.Height = contentHeight

			m.detailViewport.Width = rightWidth - 4
			m.detailViewport.Height = contentHeight
		} else {
			// Full width table view
			headerHeight := 10
			footerHeight := 3
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}

		// Update content
		m.updateViewportContent()

		return m, nil

	case branchesLoadedMsg:
		m.branches = msg.branches
		// The first branch in the sorted list IS the current branch
		// (GetAllBranches sorts with current branch first)
		if len(m.branches) > 0 {
			m.currentBranch = m.branches[0].Name()
		}
		m.updateViewportContent()
		return m, nil

	case branchLoadErrorMsg:
		// Check if this is a "not fully merged" error during deletion
		errMsg := msg.err.Error()
		if strings.Contains(errMsg, "not fully merged") && m.selectedBranch != nil {
			// Offer force delete option
			m.state = BranchViewForceDeletePrompt
			m.confirmSelectedBtn = 0 // Default to No
			return m, nil
		}

		// Reset state back to browsing so error is visible
		m.state = BranchViewBrowsing
		m.errorMessage = fmt.Sprintf("Error: %v", msg.err)
		// Show error in viewport as well
		m.viewport.SetContent(fmt.Sprintf("Error loading branches:\n\n%v", msg.err))
		return m, nil

	case branchDeletedMsg:
		m.successMessage = msg.response.Message
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.confirmSelectedBtn = 0
		m.forceDelete = false

		// Check if we should prompt for remote deletion
		if msg.response.LocalDeleted && !msg.response.RemoteDeleted && msg.response.RemoteDeletionError == nil {
			// Local deleted but we didn't try remote yet - don't prompt, just refresh
			return m, m.loadBranches()
		}

		return m, m.loadBranches()

	case branchRenamedMsg:
		m.successMessage = msg.response.Message
		m.state = BranchViewBrowsing
		return m, m.loadBranches()

	case upstreamSetMsg:
		m.successMessage = msg.response.Message
		m.state = BranchViewBrowsing
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		// Handle state-specific keys
		switch m.state {
		case BranchViewBrowsing, BranchViewExpanded:
			return m.handleBrowsingKeys(msg)
		case BranchViewDeleting:
			return m.handleDeletingKeys(msg)
		case BranchViewForceDeletePrompt:
			return m.handleForceDeletePromptKeys(msg)
		case BranchViewDeleteRemotePrompt:
			return m.handleDeleteRemotePromptKeys(msg)
		case BranchViewRenaming:
			return m.handleRenamingKeys(msg)
		case BranchViewSettingUpstream:
			return m.handleUpstreamKeys(msg)
		case BranchViewManaging:
			// Allow Esc to cancel during processing
			if msg.String() == "esc" {
				m.state = BranchViewBrowsing
				m.errorMessage = "Operation cancelled"
				return m, nil
			}
			// Ignore other keys while managing
			return m, nil
		}
	}

	// Update viewport
	if m.state == BranchViewBrowsing || m.state == BranchViewExpanded {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

		if m.state == BranchViewExpanded {
			m.detailViewport, cmd = m.detailViewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleBrowsingKeys handles keyboard input in browsing state.
func (m BranchViewModel) handleBrowsingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.returnToDashboard = true
		return m, nil

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
			m.updateViewportContent()
			m.scrollToSelected()
		}
		return m, nil

	case "down", "j":
		if m.selectedIndex < len(m.branches)-1 {
			m.selectedIndex++
			m.updateViewportContent()
			m.scrollToSelected()
		}
		return m, nil

	case "enter":
		// Toggle expand/collapse
		if m.state == BranchViewBrowsing {
			m.state = BranchViewExpanded
			m.expandedIndex = m.selectedIndex
		} else {
			m.state = BranchViewBrowsing
			m.expandedIndex = -1
		}
		m.updateViewportContent()
		return m, nil

	case "d":
		// Delete branch
		if len(m.branches) == 0 {
			return m, nil
		}
		m.selectedBranch = m.branches[m.selectedIndex]
		m.state = BranchViewDeleting
		return m, nil

	case "r":
		// Rename branch
		if len(m.branches) == 0 {
			return m, nil
		}
		m.selectedBranch = m.branches[m.selectedIndex]
		m.renameInput.SetValue(m.selectedBranch.Name())
		m.renameInput.Focus()
		m.state = BranchViewRenaming
		return m, nil

	case "u":
		// Set upstream
		if len(m.branches) == 0 {
			return m, nil
		}
		m.selectedBranch = m.branches[m.selectedIndex]
		m.upstreamInput.SetValue("")
		m.upstreamInput.Focus()
		m.state = BranchViewSettingUpstream
		return m, nil

	case "R":
		// Refresh
		m.successMessage = ""
		m.errorMessage = ""
		return m, m.loadBranches()
	}

	return m, nil
}

// handleDeletingKeys handles keyboard input during deletion confirmation.
func (m BranchViewModel) handleDeletingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "right", "l":
		// Toggle button selection
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "tab":
		// Tab to switch buttons
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "enter":
		// Confirm selected button
		if m.confirmSelectedBtn == 1 {
			// Yes selected - delete
			m.state = BranchViewManaging
			m.confirmSelectedBtn = 0 // Reset for next time
			return m, m.deleteBranch(false)
		}
		// No selected - cancel
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.confirmSelectedBtn = 0
		return m, nil

	case "esc":
		// ESC always cancels
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.confirmSelectedBtn = 0
		return m, nil
	}

	return m, nil
}

// handleForceDeletePromptKeys handles keyboard input for force deletion prompt.
func (m BranchViewModel) handleForceDeletePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "right", "l":
		// Toggle button selection
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "tab":
		// Tab to switch buttons
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "enter":
		// Confirm selected button
		if m.confirmSelectedBtn == 1 {
			// Yes selected - force delete
			m.forceDelete = true
			m.state = BranchViewManaging
			m.confirmSelectedBtn = 0 // Reset for next time
			return m, m.deleteBranch(false)
		}
		// No selected - cancel
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.forceDelete = false
		m.confirmSelectedBtn = 0
		return m, nil

	case "esc":
		// ESC always cancels
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.forceDelete = false
		m.confirmSelectedBtn = 0
		return m, nil
	}

	return m, nil
}

// handleDeleteRemotePromptKeys handles keyboard input for remote deletion prompt.
func (m BranchViewModel) handleDeleteRemotePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h", "right", "l":
		// Toggle button selection
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "tab":
		// Tab to switch buttons
		m.confirmSelectedBtn = (m.confirmSelectedBtn + 1) % 2
		return m, nil

	case "enter":
		// Confirm selected button
		if m.confirmSelectedBtn == 1 {
			// Yes selected - delete remote too
			m.state = BranchViewManaging
			m.confirmSelectedBtn = 0
			return m, m.deleteBranch(true)
		}
		// No selected - just local deletion
		m.state = BranchViewBrowsing
		m.successMessage = fmt.Sprintf("Local branch '%s' deleted", m.selectedBranch.Name())
		m.confirmSelectedBtn = 0
		return m, m.loadBranches()

	case "esc":
		// ESC = No, keep remote
		m.state = BranchViewBrowsing
		m.successMessage = fmt.Sprintf("Local branch '%s' deleted", m.selectedBranch.Name())
		m.confirmSelectedBtn = 0
		return m, m.loadBranches()
	}

	return m, nil
}

// handleRenamingKeys handles keyboard input during branch renaming.
func (m BranchViewModel) handleRenamingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		// Confirm rename
		return m, m.renameBranch()

	case "esc":
		// Cancel rename
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.renameInput.SetValue("")
		return m, nil
	}

	// Update text input
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

// handleUpstreamKeys handles keyboard input during upstream setting.
func (m BranchViewModel) handleUpstreamKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		// Confirm upstream
		return m, m.setUpstream()

	case "esc":
		// Cancel
		m.state = BranchViewBrowsing
		m.selectedBranch = nil
		m.upstreamInput.SetValue("")
		return m, nil
	}

	// Update text input
	m.upstreamInput, cmd = m.upstreamInput.Update(msg)
	return m, cmd
}

// deleteBranch initiates branch deletion.
func (m BranchViewModel) deleteBranch(alsoDeleteRemote bool) tea.Cmd {
	branchName := m.selectedBranch.Name()

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Determine remote name (default to "origin")
		remoteName := "origin"

		req := usecase.DeleteBranchRequest{
			RepoPath:          m.repoPath,
			BranchName:        branchName,
			Force:             m.forceDelete,
			AlsoDeleteRemote:  alsoDeleteRemote,
			RemoteName:        remoteName,
			ProtectedBranches: m.config.Git.ProtectedBranches,
		}

		resp, err := m.manageBranchesUC.DeleteBranch(ctx, req)
		if err != nil {
			return branchLoadErrorMsg{err}
		}

		return branchDeletedMsg{resp}
	}
}

// renameBranch initiates branch renaming.
func (m BranchViewModel) renameBranch() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := usecase.RenameBranchRequest{
			RepoPath: m.repoPath,
			OldName:  m.selectedBranch.Name(),
			NewName:  m.renameInput.Value(),
		}

		resp, err := m.manageBranchesUC.RenameBranch(ctx, req)
		if err != nil {
			return branchLoadErrorMsg{err}
		}

		return branchRenamedMsg{resp}
	}
}

// setUpstream initiates upstream setting.
func (m BranchViewModel) setUpstream() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := usecase.SetUpstreamRequest{
			RepoPath:   m.repoPath,
			BranchName: m.selectedBranch.Name(),
			Upstream:   m.upstreamInput.Value(),
		}

		resp, err := m.manageBranchesUC.SetUpstream(ctx, req)
		if err != nil {
			return branchLoadErrorMsg{err}
		}

		return upstreamSetMsg{resp}
	}
}

// updateViewportContent updates the viewport content based on current state.
func (m *BranchViewModel) updateViewportContent() {
	if m.state == BranchViewExpanded {
		// Update both viewports for split view
		m.viewport.SetContent(m.renderBranchTable(true))
		m.detailViewport.SetContent(m.renderDetailPanel())
	} else {
		// Update single viewport for full table
		m.viewport.SetContent(m.renderBranchTable(false))
	}
}

// View renders the branch view.
func (m BranchViewModel) View() string {
	styles := GetGlobalThemeManager().GetStyles()

	// Render based on state
	switch m.state {
	case BranchViewDeleting:
		return m.renderDeleteConfirmation()
	case BranchViewForceDeletePrompt:
		return m.renderForceDeletePrompt()
	case BranchViewDeleteRemotePrompt:
		return m.renderDeleteRemotePrompt()
	case BranchViewRenaming:
		return m.renderRenameModal()
	case BranchViewSettingUpstream:
		return m.renderUpstreamModal()
	case BranchViewManaging:
		// Show loading overlay
		return m.renderLoadingOverlay("Deleting branch...")
	}

	// Render logo
	logo := m.renderLogo()

	// Render messages
	messages := m.renderMessages()

	// Render content
	var content string
	if m.state == BranchViewExpanded {
		// Split view (35/65)
		leftPanel := styles.ViewportStyle.Render(m.viewport.View())
		rightPanel := styles.ViewportStyle.Render(m.detailViewport.View())

		divider := lipgloss.NewStyle().
			Foreground(styles.ColorBorder).
			Render(strings.Repeat("│\n", m.viewport.Height))

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanel,
			divider,
			rightPanel,
		)
	} else {
		// Full width table
		content = styles.ViewportStyle.Render(m.viewport.View())
	}

	// Render footer
	footer := m.renderFooter()

	// Combine sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		messages,
		"",
		content,
		"",
		footer,
	)
}

// renderLogo renders the branch view logo.
func (m BranchViewModel) renderLogo() string {
	styles := GetGlobalThemeManager().GetStyles()
	logo := styles.Header.Render("BRANCH MANAGEMENT")
	repoInfo := styles.RepoLabel.Render("Repository: ") + styles.RepoValue.Render(m.repoPath)
	return lipgloss.JoinVertical(lipgloss.Left, logo, repoInfo)
}

// renderMessages renders success/error messages.
func (m BranchViewModel) renderMessages() string {
	if m.errorMessage != "" {
		styles := GetGlobalThemeManager().GetStyles()
		return styles.StatusError.Render("✗ " + m.errorMessage)
	}
	if m.successMessage != "" {
		styles := GetGlobalThemeManager().GetStyles()
		return styles.StatusOk.Render("✓ " + m.successMessage)
	}
	return ""
}

// renderBranchTable renders the branch list table.
func (m BranchViewModel) renderBranchTable(isCompact bool) string {
	if len(m.branches) == 0 {
		return "\n\n      No branches found\n\n      Loading branches...\n      If this persists, press 'R' to refresh or check repository status."
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Table header
	headerStyle := styles.StatusInfo.Bold(true)
	var header string
	if isCompact {
		header = fmt.Sprintf("%-25s %-10s", "Branch", "Status")
	} else {
		header = fmt.Sprintf("%-30s %-12s %-15s %-15s %-12s",
			"Branch", "Type", "Ahead/Behind", "Upstream", "Commits")
	}
	lines = append(lines, headerStyle.Render(header))

	// Use viewport width, but ensure minimum
	dividerWidth := m.viewport.Width
	if dividerWidth < 60 {
		dividerWidth = 60
	}
	lines = append(lines, strings.Repeat("─", dividerWidth))

	// Branch rows
	for i, branch := range m.branches {
		// Determine style based on selection
		var rowStyle lipgloss.Style
		if i == m.selectedIndex {
			rowStyle = styles.ListItemSelected
		} else {
			rowStyle = styles.ListItemNormal
		}

		// Status icon
		statusIcon := m.getBranchStatusIcon(branch)

		// Format branch name
		branchName := branch.Name()
		if branch.Name() == m.currentBranch {
			branchName = "✓ " + branchName
		}

		// Build row
		var row string
		if isCompact {
			status := fmt.Sprintf("%s %s", statusIcon, m.getBranchTypeString(branch))
			row = fmt.Sprintf("%-25s %-10s", truncate(branchName, 23), status)
		} else {
			typeStr := m.getBranchTypeString(branch)
			divergence := m.getDivergenceString(branch)
			upstream := branch.Upstream()
			if upstream == "" {
				upstream = "-"
			}
			commits := fmt.Sprintf("%d", branch.CommitCount())

			row = fmt.Sprintf("%-30s %-12s %-15s %-15s %-12s",
				truncate(branchName, 28),
				typeStr,
				divergence,
				truncate(upstream, 13),
				commits,
			)
		}

		lines = append(lines, rowStyle.Render(row))
	}

	return strings.Join(lines, "\n")
}

// renderDetailPanel renders the detail panel for the selected branch.
func (m BranchViewModel) renderDetailPanel() string {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.branches) {
		return "No branch selected"
	}

	branch := m.branches[m.selectedIndex]
	styles := GetGlobalThemeManager().GetStyles()

	var lines []string

	// Branch name
	lines = append(lines, styles.Header.Bold(true).Render(branch.Name()))
	lines = append(lines, "")

	// Metadata
	lines = append(lines, styles.StatusInfo.Render("Branch Information:"))
	lines = append(lines, fmt.Sprintf("  Type: %s", m.getBranchTypeString(branch)))
	lines = append(lines, fmt.Sprintf("  Parent: %s", getOrDefault(branch.Parent(), "-")))
	lines = append(lines, fmt.Sprintf("  Upstream: %s", getOrDefault(branch.Upstream(), "-")))
	lines = append(lines, "")

	// Divergence
	if branch.AheadBy() > 0 || branch.BehindBy() > 0 {
		lines = append(lines, styles.StatusInfo.Render("Sync Status:"))
		if branch.AheadBy() > 0 {
			lines = append(lines, fmt.Sprintf("  ↑ %d commit(s) ahead", branch.AheadBy()))
		}
		if branch.BehindBy() > 0 {
			lines = append(lines, fmt.Sprintf("  ↓ %d commit(s) behind", branch.BehindBy()))
		}
		lines = append(lines, "")
	}

	// Commit count
	if branch.CommitCount() > 0 {
		lines = append(lines, styles.StatusInfo.Render(fmt.Sprintf("Commits: %d", branch.CommitCount())))
		lines = append(lines, "")
	}

	// Available actions
	lines = append(lines, styles.StatusInfo.Render("Available Actions:"))
	if branch.Name() != m.currentBranch {
		lines = append(lines, "  [d] Delete branch")
	}
	lines = append(lines, "  [r] Rename branch")
	lines = append(lines, "  [u] Set upstream tracking")
	lines = append(lines, "")
	lines = append(lines, "  [enter] Collapse detail view")

	return strings.Join(lines, "\n")
}

// renderDeleteConfirmation renders the delete confirmation modal.
func (m BranchViewModel) renderDeleteConfirmation() string {
	if m.selectedBranch == nil {
		return ""
	}

	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Render("⚠ Delete Branch")

	// Message
	message := fmt.Sprintf("Are you sure you want to delete branch '%s'?", m.selectedBranch.Name())
	if m.selectedBranch.Type() == domain.BranchTypeProtected {
		message += "\n\n⚠️  This is a protected branch!"
	}
	message += "\n\nThis action cannot be undone."

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(message)

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

	if m.confirmSelectedBtn == 0 {
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
		messageStyle,
		"",
		"",
		buttons,
		"",
		helpText,
	)

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Confirmation)).
		Width(60)

	return "\n\n" + lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderForceDeletePrompt renders the force delete confirmation prompt.
func (m BranchViewModel) renderForceDeletePrompt() string {
	if m.selectedBranch == nil {
		return ""
	}

	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorWarning).
		Render("⚠ Force Delete Required")

	// Message
	message := fmt.Sprintf("Branch '%s' is not fully merged.\n\n", m.selectedBranch.Name())
	message += "This branch contains commits that haven't been merged into its parent branch.\n\n"
	message += "Are you sure you want to force delete it?\n\n"
	message += "⚠️  This will permanently lose any unmerged changes!"

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(message)

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
		Background(styles.ColorWarning).
		Foreground(lipgloss.Color("#000000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorWarning)

	// Render buttons
	noBtn := "Cancel"
	yesBtn := "Force Delete"

	if m.confirmSelectedBtn == 0 {
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
		messageStyle,
		"",
		"",
		buttons,
		"",
		helpText,
	)

	// Create modal box with warning color
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorWarning).
		Background(lipgloss.Color(theme.Backgrounds.Confirmation)).
		Width(70)

	return "\n\n" + lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderDeleteRemotePrompt renders the remote deletion prompt.
func (m BranchViewModel) renderDeleteRemotePrompt() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorText).
		Render("ℹ Delete Remote Branch?")

	// Message
	message := fmt.Sprintf("Local branch '%s' has been deleted.\n\nDo you also want to delete the remote branch?", m.selectedBranch.Name())
	messageStyle := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(message)

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

	if m.confirmSelectedBtn == 0 {
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
		messageStyle,
		"",
		"",
		buttons,
		"",
		helpText,
	)

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Background(lipgloss.Color(theme.Backgrounds.Confirmation)).
		Width(60)

	return "\n\n" + lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderRenameModal renders the rename branch modal.
func (m BranchViewModel) renderRenameModal() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	title := "Rename Branch"
	message := fmt.Sprintf("Renaming: %s", m.selectedBranch.Name())

	// Build modal content
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		message,
		"",
		m.renameInput.View(),
		"",
		"[enter] Confirm    [esc] Cancel",
	)

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(lipgloss.Color(theme.Backgrounds.FormInput)).
		Padding(layout.SpacingMD).
		Width(layout.ModalWidthMD).
		Height(layout.ModalHeightSM)

	modal := modalStyle.Render(content)

	// Center modal
	return lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

// renderUpstreamModal renders the set upstream modal.
func (m BranchViewModel) renderUpstreamModal() string {
	styles := GetGlobalThemeManager().GetStyles()
	theme := GetGlobalThemeManager().GetCurrentTheme()

	title := "Set Upstream Branch"
	message := fmt.Sprintf("Branch: %s", m.selectedBranch.Name())

	// Build modal content
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		message,
		"",
		m.upstreamInput.View(),
		"",
		"[enter] Confirm    [esc] Cancel",
	)

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Background(lipgloss.Color(theme.Backgrounds.FormInput)).
		Padding(layout.SpacingMD).
		Width(layout.ModalWidthMD).
		Height(layout.ModalHeightSM)

	modal := modalStyle.Render(content)

	// Center modal
	return lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

// renderFooter renders the footer with keyboard shortcuts.
func (m BranchViewModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()

	var help string
	switch m.state {
	case BranchViewBrowsing:
		help = "↑↓: navigate • enter: expand • d: delete • r: rename • u: set upstream • R: refresh • esc: back"
	case BranchViewExpanded:
		help = "↑↓: navigate • enter: collapse • d: delete • r: rename • u: set upstream • esc: back"
	default:
		help = "See modal for options"
	}

	var metadata string
	if len(m.branches) == 0 {
		metadata = "No branches loaded - Press 'R' to refresh"
	} else if m.currentBranch == "" {
		metadata = fmt.Sprintf("%d branch(es) loaded", len(m.branches))
	} else {
		metadata = fmt.Sprintf("%d branch(es) • Current: %s", len(m.branches), m.currentBranch)
	}

	footer := styles.Footer.Render(help)
	if metadata != "" {
		footer = footer + " " + styles.Metadata.Render(metadata)
	}

	return footer
}

// getBranchStatusIcon returns the status icon for a branch.
func (m BranchViewModel) getBranchStatusIcon(branch *domain.BranchInfo) string {
	if branch.Type() == domain.BranchTypeProtected {
		return "🔒"
	}
	if branch.Name() == m.currentBranch {
		return "✓"
	}
	if branch.AheadBy() > 0 || branch.BehindBy() > 0 {
		return "↕"
	}
	return "•"
}

// getBranchTypeString returns the string representation of branch type.
func (m BranchViewModel) getBranchTypeString(branch *domain.BranchInfo) string {
	switch branch.Type() {
	case domain.BranchTypeProtected:
		return "Protected"
	case domain.BranchTypeFeature:
		return "Feature"
	case domain.BranchTypeHotfix:
		return "Hotfix"
	case domain.BranchTypeBugfix:
		return "Bugfix"
	case domain.BranchTypeRelease:
		return "Release"
	case domain.BranchTypeRefactor:
		return "Refactor"
	default:
		return "Other"
	}
}

// getDivergenceString returns the ahead/behind string for a branch.
func (m BranchViewModel) getDivergenceString(branch *domain.BranchInfo) string {
	ahead := branch.AheadBy()
	behind := branch.BehindBy()

	if ahead == 0 && behind == 0 {
		return "up to date"
	}

	parts := []string{}
	if ahead > 0 {
		parts = append(parts, fmt.Sprintf("↑%d", ahead))
	}
	if behind > 0 {
		parts = append(parts, fmt.Sprintf("↓%d", behind))
	}

	return strings.Join(parts, " ")
}

// ShouldReturnToDashboard returns whether the view wants to return to dashboard.
func (m BranchViewModel) ShouldReturnToDashboard() bool {
	return m.returnToDashboard
}

// renderLoadingOverlay renders a loading message.
func (m BranchViewModel) renderLoadingOverlay(message string) string {
	styles := GetGlobalThemeManager().GetStyles()

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary).
		Render("Processing...")

	content := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(message)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Padding(layout.SpacingLG).
		Width(50).
		Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, title, "", content))

	return lipgloss.Place(
		m.windowWidth,
		m.windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// isCurrentBranch checks if a branch name matches the current branch.
func (m BranchViewModel) isCurrentBranch(branchName string) bool {
	// The current branch should be first in the sorted list from GetAllBranches
	return len(m.branches) > 0 && branchName == m.branches[0].Name()
}

// scrollToSelected ensures the selected item is visible in the viewport.
func (m *BranchViewModel) scrollToSelected() {
	// Calculate which line the selected item is on
	// Header takes 2 lines (header + divider)
	selectedLine := m.selectedIndex + 2

	// Get viewport visible range
	viewportTop := m.viewport.YOffset
	viewportBottom := viewportTop + m.viewport.Height

	// Scroll up if selected is above visible area
	if selectedLine < viewportTop {
		m.viewport.YOffset = selectedLine
		if m.viewport.YOffset < 0 {
			m.viewport.YOffset = 0
		}
	}

	// Scroll down if selected is below visible area
	if selectedLine >= viewportBottom {
		m.viewport.YOffset = selectedLine - m.viewport.Height + 1
	}
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
