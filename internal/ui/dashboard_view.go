package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// ActiveSubmenu represents which submenu is currently open
type ActiveSubmenu int

const (
	NoSubmenu ActiveSubmenu = iota
	CommitOptionsMenu
	MergeOptionsMenu
	CommitListMenu
	BranchListMenu
	QuickStatusMenu
	HelpMenu
	RepositoryDetailsMenu
)

// Dashboard actions that can be returned
type DashboardAction int

const (
	ActionNone DashboardAction = iota
	ActionCommit
	ActionMerge
	ActionSwitchBranch
	ActionRefresh
	ActionFetch
	ActionPull
	ActionPush
	ActionViewGitHub
	ActionShowGitHubInfo
	ActionSetupRemote
)

// DashboardModel represents the state of the dashboard view
type DashboardModel struct {
	gitOps        git.Operations
	repoPath      string
	repo          *domain.Repository
	branchInfo    *domain.BranchInfo
	branches      []string
	recentCommits []git.CommitInfo
	selectedCard  int
	activeSubmenu ActiveSubmenu
	submenuIndex  int

	// Submenu options
	useConventional bool
	customMessage   string
	sourceBranch    string
	targetBranch    string

	// State
	loading   bool
	err       error
	cancelled bool

	// Action to return
	action       DashboardAction
	actionParams map[string]interface{}
}

// Message types for async updates
type repoStatusMsg struct {
	repo       *domain.Repository
	branchInfo *domain.BranchInfo
}

type branchesMsg []string
type commitsMsg []git.CommitInfo
type errorMsg struct{ err error }

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(gitOps git.Operations, repoPath string) DashboardModel {
	return DashboardModel{
		gitOps:        gitOps,
		repoPath:      repoPath,
		selectedCard:  0,
		activeSubmenu: NoSubmenu,
		loading:       true,
		actionParams:  make(map[string]interface{}),
	}
}

// Init initializes the model and starts data fetching
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		fetchRepoStatus(m.gitOps, m.repoPath),
		fetchBranches(m.gitOps, m.repoPath),
		fetchRecentCommits(m.gitOps, m.repoPath),
	)
}

// Update handles messages
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repoStatusMsg:
		m.repo = msg.repo
		m.branchInfo = msg.branchInfo
		m.checkLoading()
		return m, nil

	case branchesMsg:
		m.branches = msg
		m.checkLoading()
		return m, nil

	case commitsMsg:
		m.recentCommits = msg
		m.checkLoading()
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		// Submenu navigation
		if m.activeSubmenu != NoSubmenu {
			return m.handleSubmenuKey(msg)
		}

		// Main dashboard navigation
		switch msg.String() {
		case "q":
			// Close any active submenu, or do nothing if at top level
			if m.activeSubmenu != NoSubmenu {
				m.activeSubmenu = NoSubmenu
				m.submenuIndex = 0
			}
			return m, nil

		case "up", "k":
			if m.selectedCard >= 3 {
				m.selectedCard -= 3
			}

		case "down", "j":
			if m.selectedCard < 3 {
				m.selectedCard += 3
			}

		case "left", "h":
			if m.selectedCard%3 > 0 {
				m.selectedCard--
			}

		case "right", "l":
			if m.selectedCard%3 < 2 {
				m.selectedCard++
			}

		case "tab":
			m.selectedCard = (m.selectedCard + 1) % 6

		case "shift+tab":
			m.selectedCard = (m.selectedCard - 1 + 6) % 6

		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchRepoStatus(m.gitOps, m.repoPath),
				fetchBranches(m.gitOps, m.repoPath),
				fetchRecentCommits(m.gitOps, m.repoPath),
			)

		case "enter":
			return m.handleCardActivation()
		}
	}

	return m, nil
}

// handleSubmenuKey handles keyboard input in submenus
func (m DashboardModel) handleSubmenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.activeSubmenu = NoSubmenu
		m.submenuIndex = 0
		return m, nil

	case "up", "k":
		if m.submenuIndex > 0 {
			m.submenuIndex--
		}

	case "down", "j":
		maxIndex := m.getSubmenuMaxIndex()
		if m.submenuIndex < maxIndex {
			m.submenuIndex++
		}

	case "enter", " ":
		return m.handleSubmenuSelection()
	}

	return m, nil
}

// handleCardActivation opens submenu or performs action when card is selected
func (m DashboardModel) handleCardActivation() (tea.Model, tea.Cmd) {
	switch m.selectedCard {
	case 0: // Repository Status - show repository details menu
		m.activeSubmenu = RepositoryDetailsMenu
		m.submenuIndex = 0

	case 1: // AI Commit - show commit options
		m.activeSubmenu = CommitOptionsMenu
		m.submenuIndex = 0

	case 2: // AI Merge - show merge options
		m.activeSubmenu = MergeOptionsMenu
		m.submenuIndex = 0

	case 3: // Recent Commits - show commit list
		m.activeSubmenu = CommitListMenu
		m.submenuIndex = 0

	case 4: // Branch Switcher - show branch list
		m.activeSubmenu = BranchListMenu
		m.submenuIndex = 0

	case 5: // Quick Actions - show help
		m.activeSubmenu = HelpMenu
		m.submenuIndex = 0
	}

	return m, nil
}

// handleSubmenuSelection handles Enter key in submenus
func (m DashboardModel) handleSubmenuSelection() (tea.Model, tea.Cmd) {
	switch m.activeSubmenu {
	case CommitOptionsMenu:
		switch m.submenuIndex {
		case 0:
			// Toggle conventional commits
			m.useConventional = !m.useConventional
		case 2:
			// Execute commit
			m.action = ActionCommit
			m.actionParams["conventional"] = m.useConventional
			m.actionParams["message"] = m.customMessage
			m.activeSubmenu = NoSubmenu
			m.submenuIndex = 0
			return m, nil
		}

	case MergeOptionsMenu:
		if m.submenuIndex == 2 {
			// Execute merge
			m.action = ActionMerge
			m.actionParams["source"] = m.sourceBranch
			m.actionParams["target"] = m.targetBranch
			m.activeSubmenu = NoSubmenu
			m.submenuIndex = 0
			return m, nil
		}

	case BranchListMenu:
		if m.submenuIndex < len(m.branches) {
			// Switch to selected branch
			m.action = ActionSwitchBranch
			m.actionParams["branch"] = m.branches[m.submenuIndex]
			m.activeSubmenu = NoSubmenu
			m.submenuIndex = 0
			return m, nil
		}

	case RepositoryDetailsMenu:
		// Build the action list dynamically to match rendering
		actionIndex := 0
		if m.repo != nil && m.repo.HasRemote() {
			// Fetch is always first
			if actionIndex == m.submenuIndex {
				m.action = ActionFetch
				m.activeSubmenu = NoSubmenu
				return m, nil
			}
			actionIndex++

			// Pull if behind
			if m.repo.CommitsBehind() > 0 {
				if actionIndex == m.submenuIndex {
					m.action = ActionPull
					m.activeSubmenu = NoSubmenu
					return m, nil
				}
				actionIndex++
			}

			// Push if ahead
			if m.repo.CommitsAhead() > 0 {
				if actionIndex == m.submenuIndex {
					m.action = ActionPush
					m.activeSubmenu = NoSubmenu
					return m, nil
				}
				actionIndex++
			}

			// GitHub actions if GitHub remote
			if m.repo.IsGitHubRemote() {
				// View on GitHub (web)
				if actionIndex == m.submenuIndex {
					m.action = ActionViewGitHub
					m.activeSubmenu = NoSubmenu
					return m, nil
				}
				actionIndex++

				// Show GitHub info
				if actionIndex == m.submenuIndex {
					m.action = ActionShowGitHubInfo
					m.activeSubmenu = NoSubmenu
					return m, nil
				}
				actionIndex++
			}
		} else {
			// Setup remote if no remote
			if actionIndex == m.submenuIndex {
				m.action = ActionSetupRemote
				m.activeSubmenu = NoSubmenu
				return m, nil
			}
			actionIndex++
		}

		// Refresh is always last
		if actionIndex == m.submenuIndex {
			m.action = ActionRefresh
			m.activeSubmenu = NoSubmenu
			return m, nil
		}

	case QuickStatusMenu, CommitListMenu, HelpMenu:
		// These are read-only, just close on enter
		m.activeSubmenu = NoSubmenu
		m.submenuIndex = 0
	}

	return m, nil
}

// getSubmenuMaxIndex returns the maximum index for current submenu
func (m DashboardModel) getSubmenuMaxIndex() int {
	switch m.activeSubmenu {
	case CommitOptionsMenu:
		return 2 // 3 options: conventional, message, execute
	case MergeOptionsMenu:
		return 2 // 3 options: source, target, execute
	case CommitListMenu:
		return len(m.recentCommits) - 1
	case BranchListMenu:
		return len(m.branches) - 1
	case QuickStatusMenu:
		return 0 // Read-only
	case HelpMenu:
		return 0 // Read-only
	case RepositoryDetailsMenu:
		// Count available actions dynamically
		count := 0
		if m.repo != nil && m.repo.HasRemote() {
			count++ // Fetch
			if m.repo.CommitsBehind() > 0 {
				count++ // Pull
			}
			if m.repo.CommitsAhead() > 0 {
				count++ // Push
			}
			if m.repo.IsGitHubRemote() {
				count += 2 // View on GitHub + Show GitHub info
			}
		} else {
			count++ // Setup remote
		}
		count++ // Refresh
		return count - 1 // Return max index (count - 1)
	}
	return 0
}

// checkLoading checks if all data is loaded
func (m *DashboardModel) checkLoading() {
	if m.repo != nil && m.branches != nil && m.recentCommits != nil {
		m.loading = false
	}
}

// View renders the dashboard
func (m DashboardModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Render(fmt.Sprintf("ERROR: %v\n", m.err))
	}

	if m.loading {
		return lipgloss.NewStyle().
			Foreground(colorPrimary).
			Render("Loading dashboard...")
	}

	var sections []string

	// Header
	header := headerStyle.Render("GitMind Dashboard")
	sections = append(sections, header)

	// Card grid (2x3)
	topRow := m.renderTopRow()
	bottomRow := m.renderBottomRow()
	sections = append(sections, topRow, bottomRow)

	// Submenu overlay (if active)
	if m.activeSubmenu != NoSubmenu {
		submenu := m.renderSubmenu()
		sections = append(sections, submenu)
	}

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTopRow renders the top 3 cards
func (m DashboardModel) renderTopRow() string {
	card0 := m.renderCard(0, "REPOSITORY", m.renderRepoStatusCard())
	card1 := m.renderCard(1, "COMMIT", m.renderCommitCard())
	card2 := m.renderCard(2, "MERGE", m.renderMergeCard())

	return lipgloss.JoinHorizontal(lipgloss.Top, card0, " ", card1, " ", card2)
}

// renderBottomRow renders the bottom 3 cards
func (m DashboardModel) renderBottomRow() string {
	card3 := m.renderCard(3, "RECENT COMMITS", m.renderCommitsCard())
	card4 := m.renderCard(4, "BRANCHES", m.renderBranchesCard())
	card5 := m.renderCard(5, "QUICK ACTIONS", m.renderActionsCard())

	return lipgloss.JoinHorizontal(lipgloss.Top, card3, " ", card4, " ", card5)
}

// renderCard wraps content in a card with title
func (m DashboardModel) renderCard(index int, title, content string) string {
	style := dashboardCardStyle
	isActive := index == m.selectedCard && m.activeSubmenu == NoSubmenu
	if isActive {
		style = dashboardCardActiveStyle
	}

	// Title at top (no enter symbol)
	titleLine := cardTitleStyle.Render(title)

	// Content with muted style
	contentStyle := lipgloss.NewStyle().Foreground(colorMuted)
	contentStr := contentStyle.Render(content)

	// Build fixed-size interior: title at top, content at bottom with spacing
	// Use Place to enforce exact dimensions: 36 width x 8 height
	titleBlock := lipgloss.Place(36, 1, lipgloss.Left, lipgloss.Top, titleLine)
	contentBlock := lipgloss.Place(36, 4, lipgloss.Left, lipgloss.Bottom, contentStr)

	// Join with 3 blank lines in between to total 8 lines
	spacer := strings.Repeat("\n", 2)
	interior := titleBlock + spacer + contentBlock

	return style.Render(interior)
}

// renderRepoStatusCard renders repository status content
func (m DashboardModel) renderRepoStatusCard() string {
	var lines []string

	if m.repo == nil {
		lines = append(lines, "Loading...")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	// Line 1: Branch name (shortened if too long)
	branch := m.repo.CurrentBranch()
	if len(branch) > 28 {
		branch = branch[:25] + "..."
	}
	lines = append(lines, branch)

	// Line 2: Clean or file count
	if m.repo.HasChanges() {
		fileCount := m.repo.TotalChanges()
		statusText := fmt.Sprintf("%d file(s) changed", fileCount)
		lines = append(lines, statusWarningStyle.Render(statusText))
	} else {
		lines = append(lines, statusOkStyle.Render("Clean"))
	}

	// Line 3: +additions -deletions (only if has changes)
	if m.repo.HasChanges() {
		additions := m.repo.TotalAdditions()
		deletions := m.repo.TotalDeletions()
		diffText := fmt.Sprintf("+%d -%d", additions, deletions)
		lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(diffText))
	} else {
		lines = append(lines, "")
	}

	// Line 4: Remote sync status
	if m.repo.HasRemote() {
		syncStatus := m.repo.SyncStatusSummary()

		// Add remote type indicator
		var remoteIndicator string
		if m.repo.IsGitHubRemote() {
			remoteIndicator = statusInfoStyle.Render("→ GitHub")
		} else {
			remoteIndicator = lipgloss.NewStyle().Foreground(colorMuted).Render("→ Remote")
		}

		// Colorize sync status
		var syncStatusStyled string
		if syncStatus == "synced" {
			syncStatusStyled = statusOkStyle.Render("synced")
		} else if syncStatus == "no remote" {
			syncStatusStyled = lipgloss.NewStyle().Foreground(colorMuted).Render("-")
		} else {
			// Has ahead/behind indicators
			syncStatusStyled = lipgloss.NewStyle().Foreground(colorMuted).Render(syncStatus)
		}

		lines = append(lines, syncStatusStyled+" "+remoteIndicator)
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render("no remote"))
	}

	return strings.Join(lines, "\n")
}

// renderCommitCard renders commit card content
func (m DashboardModel) renderCommitCard() string {
	var lines []string

	if m.repo == nil {
		lines = append(lines, "Loading...")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	if m.repo.HasChanges() {
		lines = append(lines, "Analyze changes with")
		lines = append(lines, "AI assistance")
		lines = append(lines, "")
		lines = append(lines, "")
	} else {
		lines = append(lines, statusOkStyle.Render("No changes"))
		lines = append(lines, "")
		lines = append(lines, "Make changes first")
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderMergeCard renders merge card content
func (m DashboardModel) renderMergeCard() string {
	var lines []string

	if m.branchInfo == nil {
		lines = append(lines, "Loading...")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	if m.branchInfo.Parent() != "" {
		parent := m.branchInfo.Parent()
		if len(parent) > 25 {
			parent = parent[:22] + "..."
		}
		lines = append(lines, fmt.Sprintf("Target: %s", parent))
		if m.branchInfo.CommitCount() > 0 {
			lines = append(lines, fmt.Sprintf("%d commits ready", m.branchInfo.CommitCount()))
		} else {
			lines = append(lines, "")
		}
		lines = append(lines, "")
		lines = append(lines, "")
	} else {
		lines = append(lines, "No parent branch")
		lines = append(lines, "configured")
		lines = append(lines, "")
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderCommitsCard renders recent commits card content
func (m DashboardModel) renderCommitsCard() string {
	var lines []string

	if m.recentCommits == nil {
		lines = append(lines, "Loading...")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	if len(m.recentCommits) == 0 {
		lines = append(lines, "No commits yet")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	maxCommits := 3
	if len(m.recentCommits) < maxCommits {
		maxCommits = len(m.recentCommits)
	}

	for i := 0; i < maxCommits; i++ {
		commit := m.recentCommits[i]
		hash := commit.Hash[:7]
		msg := commit.Message
		if len(msg) > 25 {
			msg = msg[:22] + "..."
		}
		lines = append(lines, fmt.Sprintf("%s %s", hash, msg))
	}

	if len(m.recentCommits) > maxCommits {
		lines = append(lines, fmt.Sprintf("... +%d more", len(m.recentCommits)-maxCommits))
	}

	// Pad to exactly 4 lines
	for len(lines) < 4 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderBranchesCard renders branches card content
func (m DashboardModel) renderBranchesCard() string {
	var lines []string

	if m.branches == nil {
		lines = append(lines, "Loading...")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	if len(m.branches) == 0 {
		lines = append(lines, "No branches")
		lines = append(lines, "")
		lines = append(lines, "")
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	lines = append(lines, fmt.Sprintf("%d branches", len(m.branches)))
	lines = append(lines, "")

	maxBranches := 2
	if len(m.branches) < maxBranches {
		maxBranches = len(m.branches)
	}

	for i := 0; i < maxBranches; i++ {
		branch := m.branches[i]
		if len(branch) > 30 {
			branch = branch[:27] + "..."
		}
		lines = append(lines, "• "+branch)
	}

	// Pad to exactly 4 lines
	for len(lines) < 4 {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderActionsCard renders quick actions card content
func (m DashboardModel) renderActionsCard() string {
	var lines []string
	lines = append(lines, "Help & Shortcuts")
	lines = append(lines, "")
	lines = append(lines, "r - Refresh")
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// renderSubmenu renders the active submenu as an overlay
func (m DashboardModel) renderSubmenu() string {
	var content string

	switch m.activeSubmenu {
	case CommitOptionsMenu:
		content = m.renderCommitOptionsMenu()
	case MergeOptionsMenu:
		content = m.renderMergeOptionsMenu()
	case CommitListMenu:
		content = m.renderCommitListMenu()
	case BranchListMenu:
		content = m.renderBranchListMenu()
	case QuickStatusMenu:
		content = m.renderQuickStatusMenu()
	case HelpMenu:
		content = m.renderHelpMenu()
	case RepositoryDetailsMenu:
		content = m.renderRepositoryDetailsMenu()
	}

	return "\n" + submenuStyle.Render(content)
}

// renderCommitOptionsMenu renders commit options submenu
func (m DashboardModel) renderCommitOptionsMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Commit Options"))
	lines = append(lines, "")

	// Option 0: Conventional commits
	checkbox := "[ ]"
	if m.useConventional {
		checkbox = checkboxStyle.Render("[x]")
	}
	opt0 := fmt.Sprintf("%s Conventional commits format", checkbox)
	if m.submenuIndex == 0 {
		opt0 = submenuOptionActiveStyle.Render("▶ " + opt0)
	} else {
		opt0 = submenuOptionStyle.Render("  " + opt0)
	}
	lines = append(lines, opt0)

	// Option 1: Custom message (placeholder)
	opt1 := "  Add custom context (not implemented)"
	if m.submenuIndex == 1 {
		opt1 = submenuOptionActiveStyle.Render("▶ Add custom context (not implemented)")
	} else {
		opt1 = submenuOptionStyle.Render(opt1)
	}
	lines = append(lines, opt1)

	// Option 2: Execute
	opt2 := "  Analyze and commit"
	if m.submenuIndex == 2 {
		opt2 = submenuOptionActiveStyle.Render("▶ " + statusInfoStyle.Render("Analyze and commit"))
	} else {
		opt2 = submenuOptionStyle.Render(opt2)
	}
	lines = append(lines, opt2)

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("Space: toggle  •  Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderMergeOptionsMenu renders merge options submenu
func (m DashboardModel) renderMergeOptionsMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Merge Options"))
	lines = append(lines, "")

	// Option 0: Source branch (placeholder)
	opt0 := "  Specify source branch (not implemented)"
	if m.submenuIndex == 0 {
		opt0 = submenuOptionActiveStyle.Render("▶ Specify source branch (not implemented)")
	} else {
		opt0 = submenuOptionStyle.Render(opt0)
	}
	lines = append(lines, opt0)

	// Option 1: Target branch (placeholder)
	opt1 := "  Specify target branch (not implemented)"
	if m.submenuIndex == 1 {
		opt1 = submenuOptionActiveStyle.Render("▶ Specify target branch (not implemented)")
	} else {
		opt1 = submenuOptionStyle.Render(opt1)
	}
	lines = append(lines, opt1)

	// Option 2: Execute
	opt2 := "  Auto-detect and merge"
	if m.submenuIndex == 2 {
		opt2 = submenuOptionActiveStyle.Render("▶ " + statusInfoStyle.Render("Auto-detect and merge"))
	} else {
		opt2 = submenuOptionStyle.Render(opt2)
	}
	lines = append(lines, opt2)

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderCommitListMenu renders scrollable commit list
func (m DashboardModel) renderCommitListMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Recent Commits"))
	lines = append(lines, "")

	if len(m.recentCommits) == 0 {
		lines = append(lines, submenuOptionStyle.Render("No commits yet"))
	} else {
		maxDisplay := 10
		if len(m.recentCommits) < maxDisplay {
			maxDisplay = len(m.recentCommits)
		}

		for i := 0; i < maxDisplay; i++ {
			commit := m.recentCommits[i]
			hash := statusInfoStyle.Render(commit.Hash[:7])
			msg := commit.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}

			line := fmt.Sprintf("%s  %s", hash, msg)
			if i == m.submenuIndex {
				line = submenuOptionActiveStyle.Render("▶ " + line)
			} else {
				line = submenuOptionStyle.Render("  " + line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("↑/↓: navigate  •  Esc: close"))

	return strings.Join(lines, "\n")
}

// renderBranchListMenu renders scrollable branch list
func (m DashboardModel) renderBranchListMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Switch Branch"))
	lines = append(lines, "")

	if len(m.branches) == 0 {
		lines = append(lines, submenuOptionStyle.Render("No branches"))
	} else {
		maxDisplay := 10
		if len(m.branches) < maxDisplay {
			maxDisplay = len(m.branches)
		}

		for i := 0; i < maxDisplay; i++ {
			branch := m.branches[i]
			isCurrent := m.repo != nil && branch == m.repo.CurrentBranch()

			indicator := "  "
			if isCurrent {
				indicator = statusOkStyle.Render("* ")
			}

			line := indicator + branch
			if i == m.submenuIndex {
				line = submenuOptionActiveStyle.Render("▶ " + line)
			} else {
				line = submenuOptionStyle.Render("  " + line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("↑/↓: navigate  •  Enter: switch  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderQuickStatusMenu renders detailed status
func (m DashboardModel) renderQuickStatusMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Repository Status"))
	lines = append(lines, "")

	if m.repo == nil {
		lines = append(lines, submenuOptionStyle.Render("Loading..."))
	} else {
		lines = append(lines, repoLabelStyle.Render("Path:")+" "+repoValueStyle.Render(m.repo.Path()))
		lines = append(lines, repoLabelStyle.Render("Branch:")+" "+repoValueStyle.Render(m.repo.CurrentBranch()))

		if m.branchInfo != nil {
			lines = append(lines, repoLabelStyle.Render("Type:")+" "+repoValueStyle.Render(string(m.branchInfo.Type())))
			if m.branchInfo.Parent() != "" {
				lines = append(lines, repoLabelStyle.Render("Parent:")+" "+repoValueStyle.Render(m.branchInfo.Parent()))
			}
		}

		lines = append(lines, "")
		lines = append(lines, repoLabelStyle.Render("Changes:")+" "+repoValueStyle.Render(m.repo.ChangeSummary()))

		if m.repo.HasChanges() {
			lines = append(lines, "")
			lines = append(lines, submenuOptionStyle.Render("Modified files:"))
			changes := m.repo.Changes()
			maxFiles := 5
			if len(changes) < maxFiles {
				maxFiles = len(changes)
			}
			for i := 0; i < maxFiles; i++ {
				change := changes[i]
				lines = append(lines, submenuOptionStyle.Render(fmt.Sprintf("  %s (+%d -%d)", change.Path, change.Additions, change.Deletions)))
			}
			if len(changes) > maxFiles {
				lines = append(lines, submenuOptionStyle.Render(fmt.Sprintf("  ... and %d more files", len(changes)-maxFiles)))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("Esc: close"))

	return strings.Join(lines, "\n")
}

// renderHelpMenu renders help and shortcuts
func (m DashboardModel) renderHelpMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("Help & Shortcuts"))
	lines = append(lines, "")

	lines = append(lines, statusInfoStyle.Render("Navigation:"))
	lines = append(lines, submenuOptionStyle.Render("  ↑↓←→ / hjkl   Navigate cards"))
	lines = append(lines, submenuOptionStyle.Render("  Tab / ⇧Tab   Cycle through cards"))
	lines = append(lines, submenuOptionStyle.Render("  Enter         Activate card"))
	lines = append(lines, "")

	lines = append(lines, statusInfoStyle.Render("Actions:"))
	lines = append(lines, submenuOptionStyle.Render("  r             Refresh dashboard"))
	lines = append(lines, submenuOptionStyle.Render("  q / Esc       Quit"))
	lines = append(lines, "")

	lines = append(lines, statusInfoStyle.Render("Cards:"))
	lines = append(lines, submenuOptionStyle.Render("  Repository       View current status"))
	lines = append(lines, submenuOptionStyle.Render("  Commit           Analyze & commit changes"))
	lines = append(lines, submenuOptionStyle.Render("  Merge            Merge to parent branch"))
	lines = append(lines, submenuOptionStyle.Render("  Recent Commits   Browse commit history"))
	lines = append(lines, submenuOptionStyle.Render("  Branches         Switch branches"))
	lines = append(lines, submenuOptionStyle.Render("  Quick Actions    This help menu"))

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("Esc: close"))

	return strings.Join(lines, "\n")
}

// renderRepositoryDetailsMenu renders repository details and actions submenu
func (m DashboardModel) renderRepositoryDetailsMenu() string {
	var lines []string
	lines = append(lines, cardTitleStyle.Render("REPOSITORY DETAILS"))
	lines = append(lines, "")

	if m.repo == nil {
		lines = append(lines, "Loading repository information...")
		lines = append(lines, "")
		lines = append(lines, shortcutDescStyle.Render("Esc: close"))
		return strings.Join(lines, "\n")
	}

	// Repository path
	lines = append(lines, statusInfoStyle.Render("Path:"))
	lines = append(lines, "  "+lipgloss.NewStyle().Foreground(colorMuted).Render(m.repo.Path()))
	lines = append(lines, "")

	// Branch information
	lines = append(lines, statusInfoStyle.Render("Branch:"))
	branchLine := "  " + m.repo.CurrentBranch()
	if m.branchInfo != nil {
		branchLine += " (" + string(m.branchInfo.Type()) + ")"
		if m.branchInfo.Parent() != "" {
			branchLine += " ← " + m.branchInfo.Parent()
		}
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(branchLine))
	lines = append(lines, "")

	// Remote information
	if m.repo.HasRemote() {
		lines = append(lines, statusInfoStyle.Render("Remote:"))
		remoteURL := m.repo.RemoteURL()
		if len(remoteURL) > 60 {
			remoteURL = remoteURL[:57] + "..."
		}
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(colorMuted).Render(remoteURL))

		// Sync status
		statusLine := "  Status: "
		syncStatus := m.repo.SyncStatusSummary()
		if syncStatus == "synced" {
			statusLine += statusOkStyle.Render("synced")
		} else {
			ahead := m.repo.CommitsAhead()
			behind := m.repo.CommitsBehind()
			if ahead > 0 {
				statusLine += statusInfoStyle.Render(fmt.Sprintf("↑%d ahead", ahead))
			}
			if behind > 0 {
				if ahead > 0 {
					statusLine += "  "
				}
				statusLine += statusWarningStyle.Render(fmt.Sprintf("↓%d behind", behind))
			}
		}
		lines = append(lines, statusLine)
		lines = append(lines, "")
	} else {
		lines = append(lines, statusWarningStyle.Render("Remote:"))
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(colorMuted).Render("No remote configured"))
		lines = append(lines, "")
	}

	// Changes summary
	lines = append(lines, statusInfoStyle.Render("Changes:"))
	if m.repo.HasChanges() {
		changeSummary := fmt.Sprintf("  %d files (+%d -%d)",
			m.repo.TotalChanges(),
			m.repo.TotalAdditions(),
			m.repo.TotalDeletions())
		lines = append(lines, statusWarningStyle.Render(changeSummary))

		// Show modified files (up to 3)
		changes := m.repo.Changes()
		displayCount := len(changes)
		if displayCount > 3 {
			displayCount = 3
		}
		for i := 0; i < displayCount; i++ {
			change := changes[i]
			changeLine := fmt.Sprintf("    • %s (+%d -%d)",
				change.Path,
				change.Additions,
				change.Deletions)
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(changeLine))
		}
		if len(changes) > 3 {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorMuted).Render(
				fmt.Sprintf("    ... and %d more", len(changes)-3)))
		}
	} else {
		lines = append(lines, "  "+statusOkStyle.Render("Clean"))
	}
	lines = append(lines, "")

	// Separator
	lines = append(lines, renderSeparator(70))
	lines = append(lines, "")

	// Actions section
	lines = append(lines, statusInfoStyle.Render("Actions:"))
	lines = append(lines, "")

	// Build actions dynamically
	actionIndex := 0
	if m.repo.HasRemote() {
		// Fetch
		fetchLine := "Fetch from remote"
		if actionIndex == m.submenuIndex {
			fetchLine = submenuOptionActiveStyle.Render("▶ " + fetchLine)
		} else {
			fetchLine = submenuOptionStyle.Render("  " + fetchLine)
		}
		lines = append(lines, fetchLine)
		actionIndex++

		// Pull if behind
		if m.repo.CommitsBehind() > 0 {
			pullLine := fmt.Sprintf("Pull from remote (↓%d available)", m.repo.CommitsBehind())
			if actionIndex == m.submenuIndex {
				pullLine = submenuOptionActiveStyle.Render("▶ " + pullLine)
			} else {
				pullLine = submenuOptionStyle.Render("  " + pullLine)
			}
			lines = append(lines, pullLine)
			actionIndex++
		}

		// Push if ahead
		if m.repo.CommitsAhead() > 0 {
			pushLine := fmt.Sprintf("Push to remote (↑%d commits)", m.repo.CommitsAhead())
			if actionIndex == m.submenuIndex {
				pushLine = submenuOptionActiveStyle.Render("▶ " + pushLine)
			} else {
				pushLine = submenuOptionStyle.Render("  " + pushLine)
			}
			lines = append(lines, pushLine)
			actionIndex++
		}

		// GitHub actions
		if m.repo.IsGitHubRemote() {
			// View on GitHub (web)
			githubLine := "View on GitHub (web)"
			if actionIndex == m.submenuIndex {
				githubLine = submenuOptionActiveStyle.Render("▶ " + githubLine)
			} else {
				githubLine = submenuOptionStyle.Render("  " + githubLine)
			}
			lines = append(lines, githubLine)
			actionIndex++

			// Show GitHub info
			infoLine := "Show GitHub info"
			if actionIndex == m.submenuIndex {
				infoLine = submenuOptionActiveStyle.Render("▶ " + infoLine)
			} else {
				infoLine = submenuOptionStyle.Render("  " + infoLine)
			}
			lines = append(lines, infoLine)
			actionIndex++
		}
	} else {
		// Setup remote
		setupLine := "Set up remote"
		if actionIndex == m.submenuIndex {
			setupLine = submenuOptionActiveStyle.Render("▶ " + setupLine)
		} else {
			setupLine = submenuOptionStyle.Render("  " + setupLine)
		}
		lines = append(lines, setupLine)
		actionIndex++
	}

	// Refresh (always last)
	refreshLine := "Refresh status"
	if actionIndex == m.submenuIndex {
		refreshLine = submenuOptionActiveStyle.Render("▶ " + refreshLine)
	} else {
		refreshLine = submenuOptionStyle.Render("  " + refreshLine)
	}
	lines = append(lines, refreshLine)

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("↑/↓: navigate  •  Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderFooter renders dashboard footer
func (m DashboardModel) renderFooter() string {
	shortcuts := []string{
		shortcutKeyStyle.Render("↑↓←→/hjkl") + " " + shortcutDescStyle.Render("navigate"),
		shortcutKeyStyle.Render("enter") + " " + shortcutDescStyle.Render("select"),
		shortcutKeyStyle.Render("r") + " " + shortcutDescStyle.Render("refresh"),
		shortcutKeyStyle.Render("q/esc") + " " + shortcutDescStyle.Render("quit"),
	}

	return footerStyle.Render(strings.Join(shortcuts, "  •  "))
}

// Getters for action results
func (m DashboardModel) GetAction() DashboardAction {
	return m.action
}

func (m DashboardModel) GetActionParams() map[string]interface{} {
	return m.actionParams
}

func (m DashboardModel) IsCancelled() bool {
	return m.cancelled
}

// Async data fetching commands

func fetchRepoStatus(gitOps git.Operations, repoPath string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		repo, err := gitOps.GetStatus(ctx, repoPath)
		if err != nil {
			return errorMsg{err}
		}

		branchInfo, err := gitOps.GetBranchInfo(ctx, repoPath, []string{"main", "master", "develop"})
		if err != nil {
			return errorMsg{err}
		}

		return repoStatusMsg{repo: repo, branchInfo: branchInfo}
	}
}

func fetchBranches(gitOps git.Operations, repoPath string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		branches, err := gitOps.ListBranches(ctx, repoPath, false)
		if err != nil {
			return errorMsg{err}
		}

		return branchesMsg(branches)
	}
}

func fetchRecentCommits(gitOps git.Operations, repoPath string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		commits, err := gitOps.GetLog(ctx, repoPath, 10)
		if err != nil {
			return errorMsg{err}
		}

		return commitsMsg(commits)
	}
}
