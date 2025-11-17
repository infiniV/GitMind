package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// HighlightMode determines what to highlight in the git tree
type HighlightMode int

const (
	HighlightNone HighlightMode = iota
	HighlightHead                         // Highlight HEAD commit
	HighlightBranch                       // Highlight all commits on current branch
	HighlightMergeTarget                  // Highlight merge target and related commits
	HighlightAll                          // Highlight all branch heads
)

// GitTreeViewModel manages the git tree visualization in the right pane
type GitTreeViewModel struct {
	viewport       viewport.Model
	commitGraph    *domain.CommitGraph
	selectedCommit string // Hash of selected commit
	highlightMode  HighlightMode
	width          int
	height         int
	ready          bool
}

// NewGitTreeViewModel creates a new git tree view model
func NewGitTreeViewModel(width, height int) GitTreeViewModel {
	vp := viewport.New(width, height)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	return GitTreeViewModel{
		viewport:      vp,
		width:         width,
		height:        height,
		highlightMode: HighlightNone,
		ready:         false,
	}
}

// SetSize updates the viewport dimensions
func (m *GitTreeViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
}

// SetCommitGraph updates the commit graph data and regenerates the view content
func (m *GitTreeViewModel) SetCommitGraph(graph *domain.CommitGraph) {
	m.commitGraph = graph
	m.ready = true
	m.regenerateContent()
}

// SetHighlightMode changes the highlight mode and regenerates content
func (m *GitTreeViewModel) SetHighlightMode(mode HighlightMode, commitHash string) {
	m.highlightMode = mode
	m.selectedCommit = commitHash
	m.regenerateContent()
}

// Init initializes the model
func (m GitTreeViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m GitTreeViewModel) Update(msg tea.Msg) (GitTreeViewModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the git tree view
func (m GitTreeViewModel) View() string {
	if !m.ready || m.commitGraph == nil {
		return m.renderPlaceholder()
	}

	styles := GetGlobalThemeManager().GetStyles()

	var sections []string

	// Branch structure overlay (top section)
	branchOverlay := m.renderBranchStructure()
	sections = append(sections, branchOverlay)

	// Separator
	separator := GetGlobalThemeManager().RenderSeparator(m.width)
	sections = append(sections, separator)

	// Commit graph viewport (scrollable middle section)
	sections = append(sections, m.viewport.View())

	// Commit details panel (bottom section, if commit selected)
	if m.selectedCommit != "" {
		detailsPanel := m.renderCommitDetails()
		if detailsPanel != "" {
			sections = append(sections, "")
			sections = append(sections, GetGlobalThemeManager().RenderSeparator(m.width))
			sections = append(sections, detailsPanel)
		}
	}

	return styles.CommitBox.
		Width(m.width - 4). // Account for padding
		Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// renderPlaceholder renders a message when no git tree data is available
func (m GitTreeViewModel) renderPlaceholder() string {
	styles := GetGlobalThemeManager().GetStyles()

	msg := "Loading git tree..."
	if m.commitGraph != nil && len(m.commitGraph.Commits) == 0 {
		msg = "No commits found in repository"
	}

	return styles.CommitBox.
		Width(m.width - 4).
		Height(m.height - 4).
		Render(lipgloss.Place(
			m.width-8,
			m.height-8,
			lipgloss.Center,
			lipgloss.Center,
			styles.Loading.Render(msg),
		))
}

// renderBranchStructure renders the branch hierarchy tree at the top
func (m GitTreeViewModel) renderBranchStructure() string {
	if m.commitGraph == nil || len(m.commitGraph.Branches) == 0 {
		return ""
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Find root branches (branches without parents)
	roots := make([]domain.BranchNode, 0)
	for _, branch := range m.commitGraph.Branches {
		if branch.Parent == "" || !m.hasBranch(branch.Parent) {
			roots = append(roots, branch)
		}
	}

	// Render each root branch tree
	for _, root := range roots {
		lines = append(lines, m.renderBranchNode(root, "", true)...)
	}

	// Limit to max 8 lines
	if len(lines) > 8 {
		lines = lines[:8]
	}

	// Pad to exactly 8 lines
	for len(lines) < 8 {
		lines = append(lines, "")
	}

	return styles.SectionTitle.Render("BRANCHES") + "\n" + strings.Join(lines, "\n")
}

// renderBranchNode renders a single branch node with its children
func (m GitTreeViewModel) renderBranchNode(branch domain.BranchNode, indent string, isLast bool) []string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Determine branch icon based on status
	icon := "├──"
	if isLast {
		icon = "└──"
	}
	if indent == "" {
		icon = ""
	}

	// Branch name with status indicator
	branchName := branch.Name
	if branch.IsCurrent {
		branchName = styles.TreeBranch.Render("* " + branchName)
	} else {
		branchName = styles.TreeCommit.Render(branchName)
	}

	// Status badge
	statusBadge := ""
	switch branch.Status {
	case domain.BranchStatusAhead:
		statusBadge = styles.StatusInfo.Render(fmt.Sprintf(" [%d ahead]", branch.AheadCount))
	case domain.BranchStatusBehind:
		statusBadge = styles.StatusWarning.Render(fmt.Sprintf(" [%d behind]", branch.BehindCount))
	case domain.BranchStatusDiverged:
		statusBadge = styles.StatusWarning.Render(fmt.Sprintf(" [%d ahead, %d behind]", branch.AheadCount, branch.BehindCount))
	case domain.BranchStatusConflict:
		statusBadge = styles.StatusError.Render(" [conflict]")
	}

	line := indent + icon + " " + branchName + statusBadge
	lines = append(lines, line)

	// Find children
	children := m.findChildBranches(branch.Name)
	for i, child := range children {
		childIndent := indent
		if indent != "" {
			if isLast {
				childIndent += "    "
			} else {
				childIndent += "│   "
			}
		}
		isChildLast := i == len(children)-1
		lines = append(lines, m.renderBranchNode(child, childIndent, isChildLast)...)
	}

	return lines
}

// findChildBranches finds all branches that have the given branch as parent
func (m GitTreeViewModel) findChildBranches(parentName string) []domain.BranchNode {
	if m.commitGraph == nil {
		return nil
	}

	var children []domain.BranchNode
	for _, branch := range m.commitGraph.Branches {
		if branch.Parent == parentName {
			children = append(children, branch)
		}
	}
	return children
}

// hasBranch checks if a branch with the given name exists
func (m GitTreeViewModel) hasBranch(name string) bool {
	if m.commitGraph == nil {
		return false
	}
	return m.commitGraph.GetBranchByName(name) != nil
}

// renderCommitDetails renders detailed information about the selected commit
func (m GitTreeViewModel) renderCommitDetails() string {
	if m.selectedCommit == "" || m.commitGraph == nil {
		return ""
	}

	commit := m.commitGraph.GetCommitByHash(m.selectedCommit)
	if commit == nil {
		return ""
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Title
	lines = append(lines, styles.SectionTitle.Render("COMMIT DETAILS"))
	lines = append(lines, "")

	// Hash
	lines = append(lines, styles.RepoLabel.Render("Hash:")+" "+styles.TreeCommit.Render(commit.ShortHash))

	// Author
	lines = append(lines, styles.RepoLabel.Render("Author:")+" "+commit.Author)

	// Date
	timeAgo := formatTimeAgo(commit.Date)
	lines = append(lines, styles.RepoLabel.Render("Date:")+" "+timeAgo)

	// Message
	lines = append(lines, styles.RepoLabel.Render("Message:")+" "+commit.Message)

	// Tags (if any)
	if len(commit.Tags) > 0 {
		tagsStr := strings.Join(commit.Tags, ", ")
		lines = append(lines, styles.RepoLabel.Render("Tags:")+" "+styles.StatusInfo.Render(tagsStr))
	}

	// Merge indicator
	if commit.IsMerge {
		lines = append(lines, styles.TreeMerge.Render("MERGE COMMIT"))
	}

	// Limit to max 10 lines
	if len(lines) > 10 {
		lines = lines[:10]
	}

	return strings.Join(lines, "\n")
}

// regenerateContent rebuilds the viewport content based on current commit graph and highlight mode
func (m *GitTreeViewModel) regenerateContent() {
	if m.commitGraph == nil || len(m.commitGraph.Commits) == 0 {
		m.viewport.SetContent("")
		return
	}

	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Render each commit in the graph
	for _, commit := range m.commitGraph.Commits {
		line := m.renderCommitLine(commit, styles)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
}

// renderCommitLine renders a single commit line in the graph
func (m *GitTreeViewModel) renderCommitLine(commit domain.CommitNode, styles *ThemeStyles) string {
	// Graph visualization (e.g., "*", "│", "─┬─")
	graphPart := commit.GraphLine
	if graphPart == "" {
		if commit.IsMerge {
			graphPart = "◆ "
		} else {
			graphPart = "* "
		}
	}

	// Hash
	hashPart := styles.TreeCommit.Render(commit.ShortHash)

	// Time ago
	timeAgo := formatTimeAgo(commit.Date)
	timePart := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(fmt.Sprintf("(%s)", timeAgo))

	// Message (truncate if too long)
	message := commit.Message
	maxMsgLen := m.width - 30 // Reserve space for graph, hash, time
	if len(message) > maxMsgLen && maxMsgLen > 3 {
		message = message[:maxMsgLen-3] + "..."
	}
	msgPart := message

	// Author (abbreviated)
	authorPart := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(fmt.Sprintf("<%s>", truncateAuthor(commit.Author)))

	// Highlight based on mode
	shouldHighlight := m.shouldHighlightCommit(commit)
	if shouldHighlight {
		graphPart = styles.TreeBranch.Render(graphPart)
		msgPart = styles.TreeBranch.Render(msgPart)
	}

	// Mark HEAD
	if commit.IsHead {
		graphPart = styles.StatusOk.Render("HEAD -> ") + graphPart
	}

	// Assemble line
	return fmt.Sprintf("%s %s - %s %s %s", graphPart, hashPart, msgPart, timePart, authorPart)
}

// shouldHighlightCommit determines if a commit should be highlighted based on current mode
func (m *GitTreeViewModel) shouldHighlightCommit(commit domain.CommitNode) bool {
	switch m.highlightMode {
	case HighlightHead:
		return commit.IsHead
	case HighlightBranch:
		return commit.Branch != "" // Highlight commits on any branch
	case HighlightMergeTarget:
		return commit.Hash == m.selectedCommit || contains(commit.Children, m.selectedCommit)
	case HighlightAll:
		return true
	default:
		return false
	}
}

// Helper functions

func formatTimeAgo(t time.Time) string {
	dur := time.Since(t)
	if dur < time.Minute {
		return "just now"
	} else if dur < time.Hour {
		mins := int(dur.Minutes())
		return fmt.Sprintf("%d min ago", mins)
	} else if dur < 24*time.Hour {
		hours := int(dur.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(dur.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func truncateAuthor(author string) string {
	// Extract name before email (if present)
	if idx := strings.Index(author, "<"); idx > 0 {
		author = strings.TrimSpace(author[:idx])
	}
	// Truncate to max 15 chars
	if len(author) > 15 {
		return author[:12] + "..."
	}
	return author
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
