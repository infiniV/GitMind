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
)

// Dashboard actions that can be returned
type DashboardAction int

const (
	ActionNone DashboardAction = iota
	ActionCommit
	ActionMerge
	ActionSwitchBranch
	ActionRefresh
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
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit

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
	case 0: // Repository Status - show quick status
		m.activeSubmenu = QuickStatusMenu
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
		if m.submenuIndex == 0 {
			// Toggle conventional commits
			m.useConventional = !m.useConventional
		} else if m.submenuIndex == 2 {
			// Execute commit
			m.action = ActionCommit
			m.actionParams["conventional"] = m.useConventional
			m.actionParams["message"] = m.customMessage
			return m, tea.Quit
		}

	case MergeOptionsMenu:
		if m.submenuIndex == 2 {
			// Execute merge
			m.action = ActionMerge
			m.actionParams["source"] = m.sourceBranch
			m.actionParams["target"] = m.targetBranch
			return m, tea.Quit
		}

	case BranchListMenu:
		if m.submenuIndex < len(m.branches) {
			// Switch to selected branch
			m.action = ActionSwitchBranch
			m.actionParams["branch"] = m.branches[m.submenuIndex]
			return m, tea.Quit
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
	card0 := m.renderCard(0, "🔍 Repository", m.renderRepoStatusCard())
	card1 := m.renderCard(1, "📝 AI Commit", m.renderCommitCard())
	card2 := m.renderCard(2, "🔀 AI Merge", m.renderMergeCard())

	return lipgloss.JoinHorizontal(lipgloss.Top, card0, " ", card1, " ", card2)
}

// renderBottomRow renders the bottom 3 cards
func (m DashboardModel) renderBottomRow() string {
	card3 := m.renderCard(3, "📜 Recent Commits", m.renderCommitsCard())
	card4 := m.renderCard(4, "🌿 Branches", m.renderBranchesCard())
	card5 := m.renderCard(5, "⚡ Quick Actions", m.renderActionsCard())

	return lipgloss.JoinHorizontal(lipgloss.Top, card3, " ", card4, " ", card5)
}

// renderCard wraps content in a card with title
func (m DashboardModel) renderCard(index int, title, content string) string {
	style := dashboardCardStyle
	if index == m.selectedCard && m.activeSubmenu == NoSubmenu {
		style = dashboardCardActiveStyle
	}

	titleStr := cardTitleStyle.Render(title)
	contentStr := cardContentStyle.Render(content)

	cardContent := lipgloss.JoinVertical(lipgloss.Left, titleStr, contentStr)
	return style.Render(cardContent)
}

// renderRepoStatusCard renders repository status content
func (m DashboardModel) renderRepoStatusCard() string {
	if m.repo == nil {
		return "Loading..."
	}

	var lines []string

	// Branch
	branchLine := fmt.Sprintf("Branch: %s", m.repo.CurrentBranch())
	if m.branchInfo != nil {
		branchLine += fmt.Sprintf(" [%s]", m.branchInfo.Type())
	}
	lines = append(lines, branchLine)

	// Changes
	status := statusOkStyle.Render("✓ Clean")
	if m.repo.HasChanges() {
		status = statusWarningStyle.Render(fmt.Sprintf("⚠ %s", m.repo.ChangeSummary()))
	}
	lines = append(lines, status)

	// Remote
	if m.repo.HasRemote() {
		lines = append(lines, statusInfoStyle.Render("ℹ Remote configured"))
	} else {
		lines = append(lines, statusWarningStyle.Render("⚠ No remote"))
	}

	return strings.Join(lines, "\n")
}

// renderCommitCard renders AI commit card content
func (m DashboardModel) renderCommitCard() string {
	if m.repo == nil {
		return "Loading..."
	}

	var lines []string

	if m.repo.HasChanges() {
		lines = append(lines, "Analyze changes")
		lines = append(lines, "with AI assistance")
		lines = append(lines, "")
		lines = append(lines, statusInfoStyle.Render("Press Enter"))
	} else {
		lines = append(lines, statusOkStyle.Render("✓ No changes"))
		lines = append(lines, "")
		lines = append(lines, "Make changes to")
		lines = append(lines, "enable AI commit")
	}

	return strings.Join(lines, "\n")
}

// renderMergeCard renders AI merge card content
func (m DashboardModel) renderMergeCard() string {
	if m.branchInfo == nil {
		return "Loading..."
	}

	var lines []string

	if m.branchInfo.Parent() != "" {
		lines = append(lines, fmt.Sprintf("Merge to: %s", m.branchInfo.Parent()))
		lines = append(lines, "")
		if m.branchInfo.CommitCount() > 0 {
			lines = append(lines, fmt.Sprintf("%d commits ready", m.branchInfo.CommitCount()))
		}
		lines = append(lines, statusInfoStyle.Render("Press Enter"))
	} else {
		lines = append(lines, "No parent branch")
		lines = append(lines, "configured")
	}

	return strings.Join(lines, "\n")
}

// renderCommitsCard renders recent commits card content
func (m DashboardModel) renderCommitsCard() string {
	if m.recentCommits == nil {
		return "Loading..."
	}

	if len(m.recentCommits) == 0 {
		return "No commits yet"
	}

	var lines []string
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

	return strings.Join(lines, "\n")
}

// renderBranchesCard renders branches card content
func (m DashboardModel) renderBranchesCard() string {
	if m.branches == nil {
		return "Loading..."
	}

	if len(m.branches) == 0 {
		return "No branches"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%d branches", len(m.branches)))
	lines = append(lines, "")

	maxBranches := 3
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

	return strings.Join(lines, "\n")
}

// renderActionsCard renders quick actions card content
func (m DashboardModel) renderActionsCard() string {
	var lines []string
	lines = append(lines, "Press Enter for:")
	lines = append(lines, "• Help & Shortcuts")
	lines = append(lines, "• Command Guide")
	lines = append(lines, "")
	lines = append(lines, "r - Refresh")
	lines = append(lines, "q - Quit")

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
		checkbox = checkboxStyle.Render("[✓]")
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
				indicator = statusOkStyle.Render("✓ ")
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
	lines = append(lines, submenuOptionStyle.Render("  🔍 Repository  View current status"))
	lines = append(lines, submenuOptionStyle.Render("  📝 AI Commit   Analyze & commit changes"))
	lines = append(lines, submenuOptionStyle.Render("  🔀 AI Merge    Merge to parent branch"))
	lines = append(lines, submenuOptionStyle.Render("  📜 Commits     Browse commit history"))
	lines = append(lines, submenuOptionStyle.Render("  🌿 Branches    Switch branches"))
	lines = append(lines, submenuOptionStyle.Render("  ⚡ Actions     This help menu"))

	lines = append(lines, "")
	lines = append(lines, shortcutDescStyle.Render("Esc: close"))

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
