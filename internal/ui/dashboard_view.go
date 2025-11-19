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
	config        *domain.Config
	repo          *domain.Repository
	branchInfo    *domain.BranchInfo
	branches      []string
	recentCommits []git.CommitInfo
	selectedCard  int
	activeSubmenu ActiveSubmenu
	submenuIndex  int

	// Submenu options
	customMessage string
	sourceBranch  string
	targetBranch  string

	// State
	loading   bool
	err       error
	cancelled bool

	// Action to return
	action       DashboardAction
	actionParams map[string]interface{}

	// App info
	version string
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
func NewDashboardModel(gitOps git.Operations, repoPath string, config *domain.Config) DashboardModel {
	return DashboardModel{
		gitOps:        gitOps,
		repoPath:      repoPath,
		config:        config,
		selectedCard:  0,
		activeSubmenu: NoSubmenu,
		loading:       true,
		actionParams:  make(map[string]interface{}),
		version:       "0.1.0", // Default version
	}
}

// SetVersion sets the application version
func (m *DashboardModel) SetVersion(version string) {
	m.version = version
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
		if m.submenuIndex == 0 {
			// Execute commit
			m.action = ActionCommit
			m.actionParams["conventional"] = m.config.Commits.Convention == "conventional"
			m.activeSubmenu = NoSubmenu
			m.submenuIndex = 0
			return m, nil
		}

	case MergeOptionsMenu:
		if m.submenuIndex == 0 {
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
		return 0 // 1 option: execute
	case MergeOptionsMenu:
		return 0 // 1 option: execute
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
		count++          // Refresh
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

// renderHeader renders the header with ASCII art logo and repo info
func (m DashboardModel) renderHeader() string {
	styles := GetGlobalThemeManager().GetStyles()

	// ASCII art logo for "GM"
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	logo := logoStyle.Render(
		`  ██████╗ ███╗   ███╗
  ██╔════╝ ████╗ ████║
  ██║  ███╗██╔████╔██║
  ██║   ██║██║╚██╔╝██║
  ╚██████╔╝██║ ╚═╝ ██║
   ╚═════╝ ╚═╝     ╚═╝`)

	// Build info section (right side)
	var infoLines []string

	// Line 1: GitMind + version
	versionLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary).
		Render("GitMind") + " " +
		lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Render("v"+m.version)
	infoLines = append(infoLines, versionLine)

	// Line 2: Repository path
	repoPath := m.repoPath
	if len(repoPath) > 60 {
		repoPath = "..." + repoPath[len(repoPath)-57:]
	}
	repoLine := lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Render(repoPath)
	infoLines = append(infoLines, repoLine)

	// Line 3: Branch and status
	if m.repo != nil {
		branchName := m.repo.CurrentBranch()
		if len(branchName) > 40 {
			branchName = branchName[:37] + "..."
		}

		statusText := ""
		statusColor := styles.ColorSuccess
		if m.repo.HasChanges() {
			statusText = "Modified"
			statusColor = styles.ColorWarning
		} else {
			statusText = "Clean"
		}

		branchLine := fmt.Sprintf("%s %s %s",
			lipgloss.NewStyle().Foreground(styles.ColorSecondary).Render("Branch: "+branchName),
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("•"),
			lipgloss.NewStyle().Foreground(statusColor).Render(statusText))

		infoLines = append(infoLines, branchLine)
	} else {
		infoLines = append(infoLines, "")
	}

	// Combine logo and info sections
	infoSection := strings.Join(infoLines, "\n")

	// Center the info section vertically relative to the logo (6 lines)
	// Info has 3 lines, so add padding
	infoBlock := lipgloss.NewStyle().
		PaddingLeft(4).
		PaddingTop(1).
		Render(infoSection)

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		logo,
		infoBlock,
	)

	return header
}

// relativeTime returns a human-readable relative time string
func relativeTime(tStr string) string {
	t, err := time.Parse(time.RFC3339, tStr)
	if err != nil {
		return tStr
	}

	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 48*time.Hour {
		return "yesterday"
	} else {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

// View renders the dashboard
func (m DashboardModel) View() string {
	styles := GetGlobalThemeManager().GetStyles()

	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(styles.ColorError).
			Bold(true).
			Render(fmt.Sprintf("ERROR: %v\n", m.err))
	}

	if m.loading {
		return lipgloss.NewStyle().
			Foreground(styles.ColorPrimary).
			Render("Loading dashboard...")
	}

	var sections []string

	// Header with ASCII art
	header := m.renderHeader()
	sections = append(sections, header)
	sections = append(sections, "") // Blank line after header

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
	styles := GetGlobalThemeManager().GetStyles()
	style := styles.DashboardCard
	isActive := index == m.selectedCard && m.activeSubmenu == NoSubmenu
	if isActive {
		style = styles.DashboardCardActive
	}

	// Title at top
	titleLine := styles.CardTitle.Render(title)

	// Content
	// We don't need to force height here as the style handles it,
	// but we should ensure content doesn't overflow or look empty.
	contentStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	contentStr := contentStyle.Render(content)

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, titleLine, contentStr))
}

// renderRepoStatusCard renders repository status content
func (m DashboardModel) renderRepoStatusCard() string {
	if m.repo == nil {
		return "Loading..."
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Branch
	branch := m.repo.CurrentBranch()
	if len(branch) > 25 {
		branch = branch[:22] + "..."
	}
	lines = append(lines, fmt.Sprintf("%s %s",
		lipgloss.NewStyle().Foreground(styles.ColorSecondary).Render("Branch:"),
		lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true).Render(branch)))

	// Changes
	if m.repo.HasChanges() {
		stats := fmt.Sprintf("+%d -%d", m.repo.TotalAdditions(), m.repo.TotalDeletions())
		lines = append(lines, fmt.Sprintf("%s %s",
			styles.StatusWarning.Render("ℹ"),
			fmt.Sprintf("%d files changed (%s)", m.repo.TotalChanges(), stats)))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s",
			styles.StatusOk.Render("✓"),
			"Working directory clean"))
	}

	// Remote
	if m.repo.HasRemote() {
		syncStatus := m.repo.SyncStatusSummary()
		icon := "☁"
		if m.repo.IsGitHubRemote() {
			icon = "☁"
		}

		statusColor := styles.ColorMuted
		if syncStatus == "synced" {
			statusColor = styles.ColorSuccess
		} else if strings.Contains(syncStatus, "ahead") || strings.Contains(syncStatus, "behind") {
			statusColor = styles.ColorWarning
		}

		lines = append(lines, fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render(icon),
			lipgloss.NewStyle().Foreground(statusColor).Render(syncStatus)))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("☁"),
			"No remote configured"))
	}

	return strings.Join(lines, "\n\n")
}

// renderCommitCard renders commit card content
func (m DashboardModel) renderCommitCard() string {
	if m.repo == nil {
		return "Loading..."
	}

	styles := GetGlobalThemeManager().GetStyles()

	if m.repo.HasChanges() {
		return fmt.Sprintf("%s\n\n%s\n%s",
			styles.StatusInfo.Render("✓ Ready to commit"),
			fmt.Sprintf("%d files staged", m.repo.TotalChanges()),
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Press Enter to start"))
	}

	return fmt.Sprintf("%s\n\n%s",
		styles.StatusOk.Render("✓ Nothing to commit"),
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Working tree clean"))
}

// renderMergeCard renders merge card content
func (m DashboardModel) renderMergeCard() string {
	if m.branchInfo == nil {
		return "Loading..."
	}

	styles := GetGlobalThemeManager().GetStyles()

	if m.branchInfo.Parent() != "" {
		parent := m.branchInfo.Parent()
		if len(parent) > 20 {
			parent = parent[:17] + "..."
		}

		status := "Up to date"
		if m.branchInfo.CommitCount() > 0 {
			status = fmt.Sprintf("%d commits ahead", m.branchInfo.CommitCount())
		}

		return fmt.Sprintf("Target: %s\n\n%s\n%s",
			lipgloss.NewStyle().Foreground(styles.ColorSecondary).Bold(true).Render(parent),
			status,
			lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Press Enter to merge"))
	}

	return fmt.Sprintf("%s\n\n%s",
		"✗ No parent branch",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("Configure in settings"))
}

// renderCommitsCard renders recent commits card content
func (m DashboardModel) renderCommitsCard() string {
	if m.recentCommits == nil {
		return "Loading..."
	}

	if len(m.recentCommits) == 0 {
		return "No commits yet"
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	maxCommits := 3
	if len(m.recentCommits) < maxCommits {
		maxCommits = len(m.recentCommits)
	}

	for i := 0; i < maxCommits; i++ {
		commit := m.recentCommits[i]
		hash := styles.StatusInfo.Render(commit.Hash[:7])
		msg := commit.Message
		if len(msg) > 20 {
			msg = msg[:17] + "..."
		}

		timeStr := relativeTime(commit.Date)

		lines = append(lines, fmt.Sprintf("%s %s", hash, msg))
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("  "+timeStr))
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

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Show total count
	lines = append(lines, fmt.Sprintf("%d local branches", len(m.branches)))
	lines = append(lines, "")

	maxBranches := 3
	if len(m.branches) < maxBranches {
		maxBranches = len(m.branches)
	}

	for i := 0; i < maxBranches; i++ {
		branch := m.branches[i]
		isCurrent := m.repo != nil && branch == m.repo.CurrentBranch()

		prefix := "  "
		style := lipgloss.NewStyle()

		if isCurrent {
			prefix = styles.StatusOk.Render("✓ ")
		}

		lines = append(lines, style.Render(prefix+branch))
	}

	return strings.Join(lines, "\n")
}

// renderActionsCard renders quick actions card content
func (m DashboardModel) renderActionsCard() string {
	styles := GetGlobalThemeManager().GetStyles()

	return fmt.Sprintf("%s\n\n%s\n%s",
		"Shortcuts:",
		styles.ShortcutKey.Render("r")+" Refresh",
		styles.ShortcutKey.Render("?")+" Help Menu")
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

	styles := GetGlobalThemeManager().GetStyles()
	return "\n" + styles.Submenu.Render(content)
}

// renderCommitOptionsMenu renders commit options submenu
func (m DashboardModel) renderCommitOptionsMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Commit Options"))
	lines = append(lines, "")

	// Show current mode (informational)
	mode := "Standard"
	if m.config.Commits.Convention == "conventional" {
		mode = "Conventional"
	}
	info := fmt.Sprintf("Format: %s (configured in settings)", mode)
	lines = append(lines, styles.Description.Render(info))
	lines = append(lines, "")

	// Option 0: Execute
	opt0 := "  Analyze and commit"
	if m.submenuIndex == 0 {
		opt0 = styles.SubmenuOptionActive.Render("> " + styles.StatusInfo.Render("Analyze and commit"))
	} else {
		opt0 = styles.SubmenuOption.Render(opt0)
	}
	lines = append(lines, opt0)

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderMergeOptionsMenu renders merge options submenu
func (m DashboardModel) renderMergeOptionsMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Merge Options"))
	lines = append(lines, "")

	// Option 0: Execute
	opt0 := "  Auto-detect and merge"
	if m.submenuIndex == 0 {
		opt0 = styles.SubmenuOptionActive.Render("> " + styles.StatusInfo.Render("Auto-detect and merge"))
	} else {
		opt0 = styles.SubmenuOption.Render(opt0)
	}
	lines = append(lines, opt0)

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderCommitListMenu renders scrollable commit list
func (m DashboardModel) renderCommitListMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Recent Commits"))
	lines = append(lines, "")

	if len(m.recentCommits) == 0 {
		lines = append(lines, styles.SubmenuOption.Render("No commits yet"))
	} else {
		maxDisplay := 10
		if len(m.recentCommits) < maxDisplay {
			maxDisplay = len(m.recentCommits)
		}

		for i := 0; i < maxDisplay; i++ {
			commit := m.recentCommits[i]
			hash := styles.StatusInfo.Render(commit.Hash[:7])
			msg := commit.Message
			if len(msg) > 50 {
				msg = msg[:47] + "..."
			}

			line := fmt.Sprintf("%s  %s", hash, msg)
			if i == m.submenuIndex {
				line = styles.SubmenuOptionActive.Render("> " + line)
			} else {
				line = styles.SubmenuOption.Render("  " + line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("↑/↓: navigate  •  Esc: close"))

	return strings.Join(lines, "\n")
}

// renderBranchListMenu renders scrollable branch list
func (m DashboardModel) renderBranchListMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Switch Branch"))
	lines = append(lines, "")

	if len(m.branches) == 0 {
		lines = append(lines, styles.SubmenuOption.Render("No branches"))
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
				indicator = styles.StatusOk.Render("✓ ")
			}

			line := indicator + branch
			if i == m.submenuIndex {
				line = styles.SubmenuOptionActive.Render("> " + line)
			} else {
				line = styles.SubmenuOption.Render("  " + line)
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("↑/↓: navigate  •  Enter: switch  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderQuickStatusMenu renders detailed status
func (m DashboardModel) renderQuickStatusMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Repository Status"))
	lines = append(lines, "")

	if m.repo == nil {
		lines = append(lines, styles.SubmenuOption.Render("Loading..."))
	} else {
		lines = append(lines, styles.RepoLabel.Render("Path:")+" "+styles.RepoValue.Render(m.repo.Path()))
		lines = append(lines, styles.RepoLabel.Render("Branch:")+" "+styles.RepoValue.Render(m.repo.CurrentBranch()))

		if m.branchInfo != nil {
			lines = append(lines, styles.RepoLabel.Render("Type:")+" "+styles.RepoValue.Render(string(m.branchInfo.Type())))
			if m.branchInfo.Parent() != "" {
				lines = append(lines, styles.RepoLabel.Render("Parent:")+" "+styles.RepoValue.Render(m.branchInfo.Parent()))
			}
		}

		lines = append(lines, "")
		lines = append(lines, styles.RepoLabel.Render("Changes:")+" "+styles.RepoValue.Render(m.repo.ChangeSummary()))

		if m.repo.HasChanges() {
			lines = append(lines, "")
			lines = append(lines, styles.SubmenuOption.Render("Modified files:"))
			changes := m.repo.Changes()
			maxFiles := 5
			if len(changes) < maxFiles {
				maxFiles = len(changes)
			}
			for i := 0; i < maxFiles; i++ {
				change := changes[i]
				lines = append(lines, styles.SubmenuOption.Render(fmt.Sprintf("  %s (+%d -%d)", change.Path, change.Additions, change.Deletions)))
			}
			if len(changes) > maxFiles {
				lines = append(lines, styles.SubmenuOption.Render(fmt.Sprintf("  ... and %d more files", len(changes)-maxFiles)))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("Esc: close"))

	return strings.Join(lines, "\n")
}

// renderHelpMenu renders help and shortcuts
func (m DashboardModel) renderHelpMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("Help & Shortcuts"))
	lines = append(lines, "")

	lines = append(lines, styles.StatusInfo.Render("Navigation:"))
	lines = append(lines, styles.SubmenuOption.Render("  ↑↓←→ / hjkl   Navigate cards"))
	lines = append(lines, styles.SubmenuOption.Render("  Tab / ⇧Tab   Cycle through cards"))
	lines = append(lines, styles.SubmenuOption.Render("  Enter         Activate card"))
	lines = append(lines, "")

	lines = append(lines, styles.StatusInfo.Render("Actions:"))
	lines = append(lines, styles.SubmenuOption.Render("  r             Refresh dashboard"))
	lines = append(lines, styles.SubmenuOption.Render("  q / Esc       Quit"))
	lines = append(lines, "")

	lines = append(lines, styles.StatusInfo.Render("Cards:"))
	lines = append(lines, styles.SubmenuOption.Render("  Repository       View current status"))
	lines = append(lines, styles.SubmenuOption.Render("  Commit           Analyze & commit changes"))
	lines = append(lines, styles.SubmenuOption.Render("  Merge            Merge to parent branch"))
	lines = append(lines, styles.SubmenuOption.Render("  Recent Commits   Browse commit history"))
	lines = append(lines, styles.SubmenuOption.Render("  Branches         Switch branches"))
	lines = append(lines, styles.SubmenuOption.Render("  Quick Actions    This help menu"))

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("Esc: close"))

	return strings.Join(lines, "\n")
}

// renderRepositoryDetailsMenu renders repository details and actions submenu
func (m DashboardModel) renderRepositoryDetailsMenu() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string
	lines = append(lines, styles.CardTitle.Render("REPOSITORY DETAILS"))
	lines = append(lines, "")

	if m.repo == nil {
		lines = append(lines, "Loading repository information...")
		lines = append(lines, "")
		lines = append(lines, styles.ShortcutDesc.Render("Esc: close"))
		return strings.Join(lines, "\n")
	}

	// Repository path
	lines = append(lines, styles.StatusInfo.Render("Path:"))
	lines = append(lines, "  "+lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(m.repo.Path()))
	lines = append(lines, "")

	// Branch information
	lines = append(lines, styles.StatusInfo.Render("Branch:"))
	branchLine := "  " + m.repo.CurrentBranch()
	if m.branchInfo != nil {
		branchLine += " (" + string(m.branchInfo.Type()) + ")"
		if m.branchInfo.Parent() != "" {
			branchLine += " ← " + m.branchInfo.Parent()
		}
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(branchLine))
	lines = append(lines, "")

	// Remote information
	if m.repo.HasRemote() {
		lines = append(lines, styles.StatusInfo.Render("Remote:"))
		remoteURL := m.repo.RemoteURL()
		if len(remoteURL) > 60 {
			remoteURL = remoteURL[:57] + "..."
		}
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(remoteURL))

		// Sync status
		statusLine := "  Status: "
		syncStatus := m.repo.SyncStatusSummary()
		if syncStatus == "synced" {
			statusLine += styles.StatusOk.Render("synced")
		} else {
			ahead := m.repo.CommitsAhead()
			behind := m.repo.CommitsBehind()
			if ahead > 0 {
				statusLine += styles.StatusInfo.Render(fmt.Sprintf("↑%d ahead", ahead))
			}
			if behind > 0 {
				if ahead > 0 {
					statusLine += "  "
				}
				statusLine += styles.StatusWarning.Render(fmt.Sprintf("↓%d behind", behind))
			}
		}
		lines = append(lines, statusLine)
		lines = append(lines, "")
	} else {
		lines = append(lines, styles.StatusWarning.Render("Remote:"))
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("No remote configured"))
		lines = append(lines, "")
	}

	// Changes summary
	lines = append(lines, styles.StatusInfo.Render("Changes:"))
	if m.repo.HasChanges() {
		changeSummary := fmt.Sprintf("  %d files (+%d -%d)",
			m.repo.TotalChanges(),
			m.repo.TotalAdditions(),
			m.repo.TotalDeletions())
		lines = append(lines, styles.StatusWarning.Render(changeSummary))

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
			lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(changeLine))
		}
		if len(changes) > 3 {
			lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
				fmt.Sprintf("    ... and %d more", len(changes)-3)))
		}
	} else {
		lines = append(lines, "  "+styles.StatusOk.Render("Clean"))
	}
	lines = append(lines, "")

	// Separator
	lines = append(lines, renderSeparator(70))
	lines = append(lines, "")

	// Actions section
	lines = append(lines, styles.StatusInfo.Render("Actions:"))
	lines = append(lines, "")

	// Build actions dynamically
	actionIndex := 0
	if m.repo.HasRemote() {
		// Fetch
		fetchLine := "Fetch from remote"
		if actionIndex == m.submenuIndex {
			fetchLine = styles.SubmenuOptionActive.Render("> " + fetchLine)
		} else {
			fetchLine = styles.SubmenuOption.Render("  " + fetchLine)
		}
		lines = append(lines, fetchLine)
		actionIndex++

		// Pull if behind
		if m.repo.CommitsBehind() > 0 {
			pullLine := fmt.Sprintf("Pull from remote (↓%d available)", m.repo.CommitsBehind())
			if actionIndex == m.submenuIndex {
				pullLine = styles.SubmenuOptionActive.Render("> " + pullLine)
			} else {
				pullLine = styles.SubmenuOption.Render("  " + pullLine)
			}
			lines = append(lines, pullLine)
			actionIndex++
		}

		// Push if ahead
		if m.repo.CommitsAhead() > 0 {
			pushLine := fmt.Sprintf("Push to remote (↑%d commits)", m.repo.CommitsAhead())
			if actionIndex == m.submenuIndex {
				pushLine = styles.SubmenuOptionActive.Render("> " + pushLine)
			} else {
				pushLine = styles.SubmenuOption.Render("  " + pushLine)
			}
			lines = append(lines, pushLine)
			actionIndex++
		}

		// GitHub actions
		if m.repo.IsGitHubRemote() {
			// View on GitHub (web)
			githubLine := "View on GitHub (web)"
			if actionIndex == m.submenuIndex {
				githubLine = styles.SubmenuOptionActive.Render("> " + githubLine)
			} else {
				githubLine = styles.SubmenuOption.Render("  " + githubLine)
			}
			lines = append(lines, githubLine)
			actionIndex++

			// Show GitHub info
			infoLine := "Show GitHub info"
			if actionIndex == m.submenuIndex {
				infoLine = styles.SubmenuOptionActive.Render("> " + infoLine)
			} else {
				infoLine = styles.SubmenuOption.Render("  " + infoLine)
			}
			lines = append(lines, infoLine)
			actionIndex++
		}
	} else {
		// Setup remote
		setupLine := "Set up remote"
		if actionIndex == m.submenuIndex {
			setupLine = styles.SubmenuOptionActive.Render("> " + setupLine)
		} else {
			setupLine = styles.SubmenuOption.Render("  " + setupLine)
		}
		lines = append(lines, setupLine)
		actionIndex++
	}

	// Refresh (always last)
	refreshLine := "Refresh status"
	if actionIndex == m.submenuIndex {
		refreshLine = styles.SubmenuOptionActive.Render("> " + refreshLine)
	} else {
		refreshLine = styles.SubmenuOption.Render("  " + refreshLine)
	}
	lines = append(lines, refreshLine)

	lines = append(lines, "")
	lines = append(lines, styles.ShortcutDesc.Render("↑/↓: navigate  •  Enter: select  •  Esc: cancel"))

	return strings.Join(lines, "\n")
}

// renderFooter renders dashboard footer
func (m DashboardModel) renderFooter() string {
	styles := GetGlobalThemeManager().GetStyles()

	// Minimal footer
	return styles.Footer.Render(
		fmt.Sprintf("%s navigate  •  %s select  •  %s quit",
			styles.ShortcutKey.Render("arrows"),
			styles.ShortcutKey.Render("enter"),
			styles.ShortcutKey.Render("q"),
		),
	)
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
