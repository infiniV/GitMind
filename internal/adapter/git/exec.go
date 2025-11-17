package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/gitman/internal/domain"
)

// ExecOperations implements Operations using os/exec to call git commands.
type ExecOperations struct {
	gitPath string // Path to git executable (defaults to "git")
}

// NewExecOperations creates a new ExecOperations instance.
func NewExecOperations() *ExecOperations {
	return &ExecOperations{
		gitPath: "git",
	}
}

// SetGitPath sets the path to the git executable (useful for testing or custom installations).
func (e *ExecOperations) SetGitPath(path string) {
	e.gitPath = path
}

// execGit executes a git command and returns stdout, stderr, and error.
func (e *ExecOperations) execGit(ctx context.Context, repoPath string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, e.gitPath, args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

// IsGitRepo returns true if the path is a valid git repository.
func (e *ExecOperations) IsGitRepo(ctx context.Context, path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	stdout, _, err := e.execGit(ctx, absPath, "rev-parse", "--git-dir")
	if err != nil {
		return false, nil // Not a git repo
	}

	return stdout != "", nil
}

// GetCurrentBranch returns the name of the current branch.
func (e *ExecOperations) GetCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	stdout, stderr, err := e.execGit(ctx, repoPath, "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %s: %w", stderr, err)
	}

	if stdout == "" {
		// Might be in detached HEAD state
		return "HEAD", nil
	}

	return stdout, nil
}

// HasRemote returns true if the repository has a remote configured.
func (e *ExecOperations) HasRemote(ctx context.Context, repoPath string) (bool, error) {
	stdout, _, err := e.execGit(ctx, repoPath, "remote")
	if err != nil {
		return false, fmt.Errorf("failed to check remotes: %w", err)
	}

	return len(strings.TrimSpace(stdout)) > 0, nil
}

// GetStatus returns the current repository status including changes and branch info.
func (e *ExecOperations) GetStatus(ctx context.Context, repoPath string) (*domain.Repository, error) {
	repo, err := domain.NewRepository(repoPath)
	if err != nil {
		return nil, err
	}

	// Get current branch
	branch, err := e.GetCurrentBranch(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	repo.SetCurrentBranch(branch)

	// Check for remote
	hasRemote, err := e.HasRemote(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	repo.SetHasRemote(hasRemote)

	// If has remote, get remote information
	if hasRemote {
		// Get remote name (usually "origin")
		remoteName, err := e.GetRemoteName(ctx, repoPath)
		if err == nil {
			repo.SetRemoteName(remoteName)

			// Get remote URL
			remoteURL, err := e.GetRemoteURL(ctx, repoPath, remoteName)
			if err == nil {
				repo.SetRemoteURL(remoteURL)

				// Check if it's a GitHub remote
				isGitHub := IsGitHubRemote(remoteURL)
				repo.SetIsGitHubRemote(isGitHub)
			}

			// Get ahead/behind counts
			ahead, behind, err := e.GetRemoteSyncStatus(ctx, repoPath, branch)
			if err == nil {
				repo.SetCommitsAhead(ahead)
				repo.SetCommitsBehind(behind)
			}
		}
	}

	// Get status in porcelain format
	stdout, stderr, err := e.execGit(ctx, repoPath, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %s: %w", stderr, err)
	}

	// Parse status output
	changes, err := e.parseStatus(stdout)
	if err != nil {
		return nil, err
	}

	// Get line stats for each file (non-fatal if it fails)
	// This can fail with untracked files or binary files
	_ = e.populateLineStats(ctx, repoPath, changes)

	repo.SetChanges(changes)
	repo.SetIsClean(len(changes) == 0)

	return repo, nil
}

// parseStatus parses git status --porcelain output.
func (e *ExecOperations) parseStatus(output string) ([]domain.FileChange, error) {
	if output == "" {
		return []domain.FileChange{}, nil
	}

	lines := strings.Split(output, "\n")
	changes := make([]domain.FileChange, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 4 {
			continue // Invalid line
		}

		statusCode := line[:2]
		filePath := strings.TrimSpace(line[3:])

		change := domain.FileChange{
			Path: filePath,
		}

		// Parse status code
		switch {
		case strings.Contains(statusCode, "A"):
			change.Status = domain.StatusAdded
		case strings.Contains(statusCode, "M"):
			change.Status = domain.StatusModified
		case strings.Contains(statusCode, "D"):
			change.Status = domain.StatusDeleted
		case strings.Contains(statusCode, "R"):
			change.Status = domain.StatusRenamed
		case strings.Contains(statusCode, "?"):
			change.Status = domain.StatusUntracked
		default:
			change.Status = domain.StatusModified
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// populateLineStats populates additions/deletions for each file change.
func (e *ExecOperations) populateLineStats(ctx context.Context, repoPath string, changes []domain.FileChange) error {
	if len(changes) == 0 {
		return nil
	}

	// Get stats for staged changes
	stagedStats, _ := e.getDiffStats(ctx, repoPath, true)

	// Get stats for unstaged changes
	unstagedStats, _ := e.getDiffStats(ctx, repoPath, false)

	// Merge stats (unstaged takes precedence since it's more recent)
	allStats := make(map[string]struct{ added, deleted int })
	for path, stats := range stagedStats {
		allStats[path] = stats
	}
	for path, stats := range unstagedStats {
		allStats[path] = stats
	}

	// Apply stats to changes
	for i := range changes {
		if stats, ok := allStats[changes[i].Path]; ok {
			changes[i].Additions = stats.added
			changes[i].Deletions = stats.deleted
		} else if changes[i].Status == domain.StatusUntracked {
			// For untracked files, count lines in the file
			changes[i].Additions = e.countFileLines(ctx, repoPath, changes[i].Path)
			changes[i].Deletions = 0
		}
	}

	return nil
}

// getDiffStats runs git diff --numstat and parses the output.
func (e *ExecOperations) getDiffStats(ctx context.Context, repoPath string, staged bool) (map[string]struct{ added, deleted int }, error) {
	args := []string{"diff", "--numstat"}
	if staged {
		args = append(args, "--cached")
	}

	stdout, _, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]struct{ added, deleted int })
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added := 0
		deleted := 0

		// Parse added/deleted (can be "-" for binary files)
		if parts[0] != "-" {
			_, _ = fmt.Sscanf(parts[0], "%d", &added)
		}
		if parts[1] != "-" {
			_, _ = fmt.Sscanf(parts[1], "%d", &deleted)
		}

		filePath := parts[2]
		stats[filePath] = struct{ added, deleted int }{added, deleted}
	}

	return stats, nil
}

// countFileLines counts the number of lines in a file (for untracked files).
func (e *ExecOperations) countFileLines(ctx context.Context, repoPath, filePath string) int {
	fullPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return 0
	}

	// Check if binary
	if len(content) > 0 && strings.Contains(string(content[:min(512, len(content))]), "\x00") {
		return 0 // Binary file
	}

	lines := strings.Split(string(content), "\n")
	// Don't count empty trailing newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return len(lines) - 1
	}
	return len(lines)
}

// GetDiff returns the diff for staged/unstaged changes.
func (e *ExecOperations) GetDiff(ctx context.Context, repoPath string, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}

	stdout, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %s: %w", stderr, err)
	}

	return stdout, nil
}

// Add stages files for commit.
func (e *ExecOperations) Add(ctx context.Context, repoPath string, files []string) error {
	args := []string{"add"}

	if len(files) == 0 {
		args = append(args, "-A") // Add all changes
	} else {
		args = append(args, files...)
	}

	_, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to add files: %s: %w", stderr, err)
	}

	return nil
}

// Commit creates a commit with the given message.
func (e *ExecOperations) Commit(ctx context.Context, repoPath string, message string, files []string) error {
	if message == "" {
		return errors.New("commit message cannot be empty")
	}

	// Stage files if specified
	if len(files) > 0 {
		if err := e.Add(ctx, repoPath, files); err != nil {
			return err
		}
	}

	args := []string{"commit", "-m", message}

	_, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		// Check if the error is because there's nothing to commit
		if strings.Contains(stderr, "nothing to commit") {
			return errors.New("no changes to commit")
		}
		return fmt.Errorf("failed to commit: %s: %w", stderr, err)
	}

	return nil
}

// CreateBranch creates a new branch with the given name.
func (e *ExecOperations) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	if branchName == "" {
		return errors.New("branch name cannot be empty")
	}

	_, stderr, err := e.execGit(ctx, repoPath, "branch", branchName)
	if err != nil {
		if strings.Contains(stderr, "already exists") {
			return fmt.Errorf("branch '%s' already exists", branchName)
		}
		return fmt.Errorf("failed to create branch: %s: %w", stderr, err)
	}

	return nil
}

// CheckoutBranch switches to the specified branch.
func (e *ExecOperations) CheckoutBranch(ctx context.Context, repoPath, branchName string) error {
	if branchName == "" {
		return errors.New("branch name cannot be empty")
	}

	_, stderr, err := e.execGit(ctx, repoPath, "checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %s: %w", stderr, err)
	}

	return nil
}

// Push pushes commits to the remote repository.
func (e *ExecOperations) Push(ctx context.Context, repoPath, branch string, force bool) error {
	if branch == "" {
		var err error
		branch, err = e.GetCurrentBranch(ctx, repoPath)
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	args := []string{"push"}

	// Check if upstream is configured
	_, _, upstreamErr := e.execGit(ctx, repoPath, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if upstreamErr != nil {
		// No upstream configured - set it with -u flag
		args = append(args, "-u")
	}

	if force {
		args = append(args, "--force")
	}

	args = append(args, "origin", branch)

	_, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to push: %s: %w", stderr, err)
	}

	return nil
}

// Pull pulls changes from the remote repository.
func (e *ExecOperations) Pull(ctx context.Context, repoPath string) error {
	_, stderr, err := e.execGit(ctx, repoPath, "pull")
	if err != nil {
		if strings.Contains(stderr, "no tracking information") {
			return errors.New("no tracking information for the current branch")
		}
		return fmt.Errorf("failed to pull: %s: %w", stderr, err)
	}

	return nil
}

// Fetch fetches updates from the remote repository without merging.
func (e *ExecOperations) Fetch(ctx context.Context, repoPath string) error {
	_, stderr, err := e.execGit(ctx, repoPath, "fetch")
	if err != nil {
		return fmt.Errorf("failed to fetch: %s: %w", stderr, err)
	}

	return nil
}

// GetRemoteURL returns the URL for the specified remote.
func (e *ExecOperations) GetRemoteURL(ctx context.Context, repoPath, remoteName string) (string, error) {
	if remoteName == "" {
		remoteName = "origin"
	}

	stdout, stderr, err := e.execGit(ctx, repoPath, "remote", "get-url", remoteName)
	if err != nil {
		if strings.Contains(stderr, "No such remote") {
			return "", fmt.Errorf("remote '%s' not found", remoteName)
		}
		return "", fmt.Errorf("failed to get remote URL: %s: %w", stderr, err)
	}

	return stdout, nil
}

// GetRemoteName returns the primary remote name (defaults to "origin").
func (e *ExecOperations) GetRemoteName(ctx context.Context, repoPath string) (string, error) {
	stdout, _, err := e.execGit(ctx, repoPath, "remote")
	if err != nil {
		return "", fmt.Errorf("failed to get remotes: %w", err)
	}

	remotes := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(remotes) == 0 || remotes[0] == "" {
		return "", errors.New("no remotes configured")
	}

	// Prefer "origin" if it exists
	for _, remote := range remotes {
		if remote == "origin" {
			return "origin", nil
		}
	}

	// Otherwise return the first remote
	return remotes[0], nil
}

// GetRemoteSyncStatus returns commits ahead/behind relative to remote tracking branch.
func (e *ExecOperations) GetRemoteSyncStatus(ctx context.Context, repoPath, branch string) (ahead, behind int, err error) {
	if branch == "" {
		branch, err = e.GetCurrentBranch(ctx, repoPath)
		if err != nil {
			return 0, 0, err
		}
	}

	// Get the remote tracking branch
	remoteBranch, _, err := e.execGit(ctx, repoPath, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		// No upstream branch configured - try to compare with origin/<branch>
		remoteName, err := e.GetRemoteName(ctx, repoPath)
		if err != nil {
			// No remote at all
			return 0, 0, nil
		}

		// Check if remote branch exists
		remoteBranch = remoteName + "/" + branch
		_, _, err = e.execGit(ctx, repoPath, "rev-parse", "--verify", remoteBranch)
		if err != nil {
			// Remote branch doesn't exist - count all local commits as "ahead"
			stdout, _, err := e.execGit(ctx, repoPath, "rev-list", "--count", branch)
			if err != nil {
				return 0, 0, nil
			}
			_, err = fmt.Sscanf(stdout, "%d", &ahead)
			if err != nil {
				return 0, 0, nil
			}
			// If we have commits, we're ahead (need to push to create remote branch)
			return ahead, 0, nil
		}
	}

	// Get ahead/behind counts
	stdout, stderr, err := e.execGit(ctx, repoPath, "rev-list", "--left-right", "--count", branch+"..."+remoteBranch)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get divergence: %s: %w", stderr, err)
	}

	// Parse output: "ahead\tbehind"
	parts := strings.Split(stdout, "\t")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected git output format: %s", stdout)
	}

	_, err = fmt.Sscanf(parts[0], "%d", &ahead)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse ahead count: %w", err)
	}

	_, err = fmt.Sscanf(parts[1], "%d", &behind)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse behind count: %w", err)
	}

	return ahead, behind, nil
}

// GetLog returns recent commit history.
func (e *ExecOperations) GetLog(ctx context.Context, repoPath string, count int) ([]CommitInfo, error) {
	if count <= 0 {
		count = 10 // Default to 10 commits
	}

	format := "--pretty=format:%H%n%an%n%aI%n%s%n---END---"
	args := []string{"log", fmt.Sprintf("-%d", count), format}

	stdout, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %s: %w", stderr, err)
	}

	return parseLog(stdout), nil
}

// parseLog parses git log output.
func parseLog(output string) []CommitInfo {
	if output == "" {
		return []CommitInfo{}
	}

	commits := []CommitInfo{}
	entries := strings.Split(output, "---END---")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		lines := strings.Split(entry, "\n")
		if len(lines) < 4 {
			continue
		}

		commit := CommitInfo{
			Hash:    lines[0],
			Author:  lines[1],
			Date:    lines[2],
			Message: lines[3],
		}

		commits = append(commits, commit)
	}

	return commits
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetBranchInfo returns detailed information about the current branch.
func (e *ExecOperations) GetBranchInfo(ctx context.Context, repoPath string, protectedBranches []string) (*domain.BranchInfo, error) {
	// Get current branch name
	branchName, err := e.GetCurrentBranch(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	// Create branch info
	branchInfo, err := domain.NewBranchInfo(branchName)
	if err != nil {
		return nil, err
	}

	// Set protected status based on config
	branchType := domain.DetectBranchType(branchName, protectedBranches)
	branchInfo.SetType(branchType)

	// Get parent branch from git config
	parent, err := e.GetParentBranch(ctx, repoPath, branchName)
	if err == nil && parent != "" {
		branchInfo.SetParent(parent)
	}

	// Get upstream tracking branch
	upstream, _ := e.getUpstreamBranch(ctx, repoPath, branchName)
	if upstream != "" {
		branchInfo.SetUpstream(upstream)

		// Get divergence from upstream
		ahead, behind, err := e.GetDivergence(ctx, repoPath, branchName, upstream)
		if err == nil {
			branchInfo.SetAheadBy(ahead)
			branchInfo.SetBehindBy(behind)
		}
	}

	// Get commit count relative to parent
	if parent != "" {
		commits, err := e.GetBranchCommits(ctx, repoPath, branchName, parent)
		if err == nil {
			branchInfo.SetCommitCount(len(commits))
		}
	}

	return branchInfo, nil
}

// getUpstreamBranch returns the upstream tracking branch.
func (e *ExecOperations) getUpstreamBranch(ctx context.Context, repoPath, branch string) (string, error) {
	stdout, _, err := e.execGit(ctx, repoPath, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// GetMergeBase returns the common ancestor commit hash between two branches.
func (e *ExecOperations) GetMergeBase(ctx context.Context, repoPath, branch1, branch2 string) (string, error) {
	if branch1 == "" || branch2 == "" {
		return "", errors.New("branch names cannot be empty")
	}

	stdout, stderr, err := e.execGit(ctx, repoPath, "merge-base", branch1, branch2)
	if err != nil {
		return "", fmt.Errorf("failed to get merge base: %s: %w", stderr, err)
	}

	return strings.TrimSpace(stdout), nil
}

// GetBranchCommits returns commits unique to a branch (not in excludeBranch).
func (e *ExecOperations) GetBranchCommits(ctx context.Context, repoPath, branch, excludeBranch string) ([]CommitInfo, error) {
	if branch == "" || excludeBranch == "" {
		return nil, errors.New("branch names cannot be empty")
	}

	// Use git log <excludeBranch>..<branch> to get commits only on branch
	format := "--pretty=format:%H%n%an%n%aI%n%s%n---END---"
	revRange := fmt.Sprintf("%s..%s", excludeBranch, branch)

	stdout, stderr, err := e.execGit(ctx, repoPath, "log", revRange, format)
	if err != nil {
		// If error is because branches don't have common ancestor, return empty list
		if strings.Contains(stderr, "Invalid symmetric difference expression") ||
		   strings.Contains(stderr, "unknown revision") {
			return []CommitInfo{}, nil
		}
		return nil, fmt.Errorf("failed to get branch commits: %s: %w", stderr, err)
	}

	return parseLog(stdout), nil
}

// ListBranches returns all local and optionally remote branches.
func (e *ExecOperations) ListBranches(ctx context.Context, repoPath string, includeRemote bool) ([]string, error) {
	args := []string{"branch", "--list"}
	if includeRemote {
		args = append(args, "-a")
	}

	stdout, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %s: %w", stderr, err)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	branches := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove * marker for current branch
		line = strings.TrimPrefix(line, "* ")

		// Remove remotes/ prefix if present
		line = strings.TrimPrefix(line, "remotes/")

		branches = append(branches, line)
	}

	return branches, nil
}

// GetDivergence returns how many commits ahead/behind branch1 is compared to branch2.
func (e *ExecOperations) GetDivergence(ctx context.Context, repoPath, branch1, branch2 string) (ahead, behind int, err error) {
	if branch1 == "" || branch2 == "" {
		return 0, 0, errors.New("branch names cannot be empty")
	}

	// Use git rev-list --left-right --count to get divergence
	revRange := fmt.Sprintf("%s...%s", branch2, branch1)
	stdout, stderr, gitErr := e.execGit(ctx, repoPath, "rev-list", "--left-right", "--count", revRange)
	if gitErr != nil {
		return 0, 0, fmt.Errorf("failed to get divergence: %s: %w", stderr, gitErr)
	}

	// Output format: "<behind>\t<ahead>"
	parts := strings.Fields(strings.TrimSpace(stdout))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected output format: %s", stdout)
	}

	_, _ = fmt.Sscanf(parts[0], "%d", &behind)
	_, _ = fmt.Sscanf(parts[1], "%d", &ahead)

	return ahead, behind, nil
}

// GetParentBranch returns the parent branch for the given branch from git config.
func (e *ExecOperations) GetParentBranch(ctx context.Context, repoPath, branch string) (string, error) {
	if branch == "" {
		return "", errors.New("branch name cannot be empty")
	}

	configKey := fmt.Sprintf("branch.%s.parent", branch)
	stdout, _, err := e.execGit(ctx, repoPath, "config", "--get", configKey)
	if err != nil {
		// Config key not found is not an error, just means no parent set
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
}

// SetParentBranch sets the parent branch for the given branch in git config.
func (e *ExecOperations) SetParentBranch(ctx context.Context, repoPath, branch, parent string) error {
	if branch == "" || parent == "" {
		return errors.New("branch and parent names cannot be empty")
	}

	configKey := fmt.Sprintf("branch.%s.parent", branch)
	_, stderr, err := e.execGit(ctx, repoPath, "config", configKey, parent)
	if err != nil {
		return fmt.Errorf("failed to set parent branch: %s: %w", stderr, err)
	}

	return nil
}

// Merge merges sourceBranch into the current branch using the specified strategy.
func (e *ExecOperations) Merge(ctx context.Context, repoPath, sourceBranch, strategy, message string) error {
	if sourceBranch == "" {
		return errors.New("source branch cannot be empty")
	}

	args := []string{"merge"}

	// Apply strategy
	switch strategy {
	case "squash":
		args = append(args, "--squash")
	case "fast-forward":
		args = append(args, "--ff-only")
	case "regular":
		args = append(args, "--no-ff")
	case "rebase":
		// Rebase is different from merge, handle separately
		return e.rebaseBranch(ctx, repoPath, sourceBranch)
	default:
		// Default merge (allows fast-forward if possible)
	}

	// Add message if provided (and not squash, as squash requires separate commit)
	if message != "" && strategy != "squash" {
		args = append(args, "-m", message)
	}

	// Add source branch
	args = append(args, sourceBranch)

	_, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		if strings.Contains(stderr, "CONFLICT") {
			return fmt.Errorf("merge conflict: %s", stderr)
		}
		return fmt.Errorf("merge failed: %s: %w", stderr, err)
	}

	// For squash merge, we need to commit separately
	if strategy == "squash" {
		if message == "" {
			message = fmt.Sprintf("Merge branch '%s' (squashed)", sourceBranch)
		}
		if err := e.Commit(ctx, repoPath, message, nil); err != nil {
			return fmt.Errorf("failed to commit squashed merge: %w", err)
		}
	}

	return nil
}

// rebaseBranch rebases the current branch onto the source branch.
func (e *ExecOperations) rebaseBranch(ctx context.Context, repoPath, sourceBranch string) error {
	_, stderr, err := e.execGit(ctx, repoPath, "rebase", sourceBranch)
	if err != nil {
		if strings.Contains(stderr, "CONFLICT") {
			return fmt.Errorf("rebase conflict: %s", stderr)
		}
		return fmt.Errorf("rebase failed: %s: %w", stderr, err)
	}
	return nil
}

// CanMerge checks if sourceBranch can be merged into targetBranch without conflicts.
func (e *ExecOperations) CanMerge(ctx context.Context, repoPath, sourceBranch, targetBranch string) (bool, []string, error) {
	if sourceBranch == "" || targetBranch == "" {
		return false, nil, errors.New("branch names cannot be empty")
	}

	// Save current branch
	currentBranch, err := e.GetCurrentBranch(ctx, repoPath)
	if err != nil {
		return false, nil, err
	}

	// Checkout target branch
	if currentBranch != targetBranch {
		if err := e.CheckoutBranch(ctx, repoPath, targetBranch); err != nil {
			return false, nil, fmt.Errorf("failed to checkout target branch: %w", err)
		}
		// Ensure we return to original branch
		defer func() { _ = e.CheckoutBranch(ctx, repoPath, currentBranch) }()
	}

	// Try merge with --no-commit --no-ff to preview
	_, stderr, err := e.execGit(ctx, repoPath, "merge", "--no-commit", "--no-ff", sourceBranch)

	// Always abort the merge preview
	defer func() { _ = e.AbortMerge(ctx, repoPath) }()

	if err != nil {
		if strings.Contains(stderr, "CONFLICT") {
			// Parse conflict files from stderr
			conflicts := parseConflictFiles(stderr)
			return false, conflicts, nil
		}
		return false, nil, fmt.Errorf("merge preview failed: %s: %w", stderr, err)
	}

	return true, nil, nil
}

// parseConflictFiles extracts conflicting file paths from git merge stderr.
func parseConflictFiles(stderr string) []string {
	var conflicts []string
	lines := strings.Split(stderr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CONFLICT") {
			// Extract filename from "CONFLICT (content): Merge conflict in <filename>"
			parts := strings.Split(line, "in ")
			if len(parts) >= 2 {
				filename := strings.TrimSpace(parts[len(parts)-1])
				conflicts = append(conflicts, filename)
			}
		}
	}

	return conflicts
}

// AbortMerge aborts an in-progress merge.
func (e *ExecOperations) AbortMerge(ctx context.Context, repoPath string) error {
	_, stderr, err := e.execGit(ctx, repoPath, "merge", "--abort")
	if err != nil {
		// It's okay if there's no merge in progress
		if strings.Contains(stderr, "no merge") || strings.Contains(stderr, "No merge") {
			return nil
		}
		return fmt.Errorf("failed to abort merge: %s: %w", stderr, err)
	}
	return nil
}

// IsGitHubRemote returns true if the remote URL is a GitHub repository.
func IsGitHubRemote(remoteURL string) bool {
	if remoteURL == "" {
		return false
	}

	// Check for github.com in the URL
	return strings.Contains(strings.ToLower(remoteURL), "github.com")
}

// ParseGitHubRepo extracts owner and repo name from a GitHub remote URL.
// Supports both HTTPS and SSH formats:
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
func ParseGitHubRepo(remoteURL string) (*GitHubRepo, error) {
	if remoteURL == "" {
		return nil, errors.New("remote URL is empty")
	}

	// Remove .git suffix if present
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	var path string

	// Handle HTTPS URLs
	if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
		// Extract path after github.com/
		parts := strings.Split(remoteURL, "github.com/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid GitHub URL format: %s", remoteURL)
		}
		path = parts[1]
	} else if strings.HasPrefix(remoteURL, "git@") {
		// Handle SSH URLs: git@github.com:owner/repo
		parts := strings.Split(remoteURL, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid GitHub SSH URL format: %s", remoteURL)
		}
		path = parts[1]
	} else {
		return nil, fmt.Errorf("unsupported URL format: %s", remoteURL)
	}

	// Split path into owner/repo
	pathParts := strings.SplitN(path, "/", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("could not parse owner/repo from: %s", path)
	}

	return &GitHubRepo{
		Owner: pathParts[0],
		Repo:  pathParts[1],
	}, nil
}

// GetCommitGraph returns the full commit graph with ASCII visualization.
func (e *ExecOperations) GetCommitGraph(ctx context.Context, repoPath string, limit int) (*domain.CommitGraph, error) {
	// Format: hash|shortHash|author|date|message|parents|refNames
	format := "%H|%h|%an|%at|%s|%P|%D"

	args := []string{"log", "--all", "--graph", "--date-order", fmt.Sprintf("--format=%s", format)}
	if limit > 0 {
		args = append(args, fmt.Sprintf("-%d", limit))
	}

	stdout, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit graph: %s: %w", stderr, err)
	}

	// Parse the output
	graph := &domain.CommitGraph{
		Commits:   make([]domain.CommitNode, 0),
		Branches:  make([]domain.BranchNode, 0),
		BranchMap: make(map[string]*domain.BranchNode),
		CommitMap: make(map[string]*domain.CommitNode),
	}

	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Extract graph visualization and commit data
		graphPart, commitData := parseGraphLine(line)
		if commitData == "" {
			continue
		}

		// Parse commit fields
		fields := strings.Split(commitData, "|")
		if len(fields) < 7 {
			continue
		}

		hash := fields[0]
		shortHash := fields[1]
		author := fields[2]
		timestamp := fields[3]
		message := fields[4]
		parentsStr := fields[5]
		refNames := fields[6]

		// Parse date
		var commitDate time.Time
		if ts, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
			commitDate = time.Unix(ts, 0)
		}

		// Parse parents
		parents := []string{}
		if parentsStr != "" {
			parents = strings.Fields(parentsStr)
		}

		// Parse branch and tags from ref names
		branch := ""
		tags := []string{}
		isHead := false
		if refNames != "" {
			refs := strings.Split(refNames, ", ")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				if strings.HasPrefix(ref, "HEAD -> ") {
					isHead = true
					branch = strings.TrimPrefix(ref, "HEAD -> ")
				} else if strings.HasPrefix(ref, "tag: ") {
					tags = append(tags, strings.TrimPrefix(ref, "tag: "))
				} else if !strings.Contains(ref, "/") {
					// Local branch
					branch = ref
				}
			}
		}

		commit := domain.CommitNode{
			Hash:        hash,
			ShortHash:   shortHash,
			Author:      author,
			Date:        commitDate,
			Message:     message,
			FullMessage: message, // Could fetch full message separately if needed
			Branch:      branch,
			Parents:     parents,
			Children:    []string{}, // Will be populated in second pass
			GraphLine:   graphPart,
			IsMerge:     len(parents) > 1,
			IsHead:      isHead,
			Tags:        tags,
		}

		graph.Commits = append(graph.Commits, commit)
		graph.CommitMap[hash] = &graph.Commits[len(graph.Commits)-1]
	}

	// Second pass: populate children relationships
	for i := range graph.Commits {
		commit := &graph.Commits[i]
		for _, parentHash := range commit.Parents {
			if parent, exists := graph.CommitMap[parentHash]; exists {
				parent.Children = append(parent.Children, commit.Hash)
			}
		}
	}

	// Get branch tree
	branches, err := e.GetBranchTree(ctx, repoPath, []string{"main", "master", "develop"})
	if err == nil {
		graph.Branches = branches
		for i := range branches {
			graph.BranchMap[branches[i].Name] = &graph.Branches[i]
		}
	}

	return graph, nil
}

// parseGraphLine extracts the ASCII graph part and commit data from a git log --graph line.
func parseGraphLine(line string) (graph string, commitData string) {
	// Find the first occurrence of a hash (40 hex chars after graph symbols)
	// Graph symbols: *, |, /, \, space
	for i, ch := range line {
		if ch != '*' && ch != '|' && ch != '/' && ch != '\\' && ch != ' ' && ch != '_' && ch != '-' {
			// Found start of commit data
			graph = line[:i]
			commitData = strings.TrimSpace(line[i:])
			return
		}
	}
	return line, ""
}

// GetBranchTree returns all branches with their parent-child relationships.
func (e *ExecOperations) GetBranchTree(ctx context.Context, repoPath string, protectedBranches []string) ([]domain.BranchNode, error) {
	// Get all local branches
	stdout, stderr, err := e.execGit(ctx, repoPath, "branch", "--format=%(refname:short)|%(upstream:short)|%(HEAD)")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %s: %w", stderr, err)
	}

	currentBranch, _ := e.GetCurrentBranch(ctx, repoPath)
	branches := make([]domain.BranchNode, 0)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 3 {
			continue
		}

		branchName := strings.TrimSpace(fields[0])
		_ = strings.TrimSpace(fields[1]) // upstream (not used yet)
		isCurrentMarker := strings.TrimSpace(fields[2])

		// Get branch info
		branchInfo, err := e.GetBranchInfo(ctx, repoPath, protectedBranches)
		isCurrent := isCurrentMarker == "*" || branchName == currentBranch

		// Get parent branch from git config
		parent, _ := e.GetParentBranch(ctx, repoPath, branchName)

		// Get branch head commit
		headHash, _, _ := e.execGit(ctx, repoPath, "rev-parse", branchName)
		headHash = strings.TrimSpace(headHash)

		// Get last commit date
		lastCommitDateStr, _, _ := e.execGit(ctx, repoPath, "log", "-1", "--format=%at", branchName)
		var lastCommitDate time.Time
		if ts, err := strconv.ParseInt(strings.TrimSpace(lastCommitDateStr), 10, 64); err == nil {
			lastCommitDate = time.Unix(ts, 0)
		}

		// Determine branch type
		branchType := domain.DetectBranchType(branchName, protectedBranches)

		// Get commits unique to this branch (if has parent)
		commits := []string{}
		aheadCount := 0
		behindCount := 0
		status := domain.BranchStatusUnknown

		if parent != "" {
			// Get commits ahead
			branchCommits, err := e.GetBranchCommits(ctx, repoPath, branchName, parent)
			if err == nil {
				aheadCount = len(branchCommits)
				for _, c := range branchCommits {
					commits = append(commits, c.Hash)
				}
			}

			// Get commits behind
			parentCommits, err := e.GetBranchCommits(ctx, repoPath, parent, branchName)
			if err == nil {
				behindCount = len(parentCommits)
			}

			// Determine status
			if aheadCount == 0 && behindCount == 0 {
				status = domain.BranchStatusUpToDate
			} else if aheadCount > 0 && behindCount == 0 {
				status = domain.BranchStatusAhead
			} else if aheadCount == 0 && behindCount > 0 {
				status = domain.BranchStatusBehind
			} else {
				status = domain.BranchStatusDiverged
			}

			// Check for conflicts
			canMerge, _, _ := e.CanMerge(ctx, repoPath, branchName, parent)
			if !canMerge {
				status = domain.BranchStatusConflict
			}
		}

		// Check if protected
		isProtected := false
		for _, pb := range protectedBranches {
			if pb == branchName {
				isProtected = true
				break
			}
		}

		branch := domain.BranchNode{
			Name:           branchName,
			FullName:       "refs/heads/" + branchName,
			Parent:         parent,
			Type:           branchType,
			Head:           headHash,
			Commits:        commits,
			Status:         status,
			AheadCount:     aheadCount,
			BehindCount:    behindCount,
			IsProtected:    isProtected,
			IsCurrent:      isCurrent,
			IsRemote:       false,
			LastCommitDate: lastCommitDate,
		}

		// Update from branchInfo if this is the current branch
		if isCurrent && branchInfo != nil && err == nil {
			branch.Parent = branchInfo.Parent()
			branch.Type = branchInfo.Type()
			branch.Commits = make([]string, branchInfo.CommitCount())
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

// GetMergeStatus returns detailed merge status between source and target branches.
func (e *ExecOperations) GetMergeStatus(ctx context.Context, repoPath, sourceBranch, targetBranch string) (*domain.MergeStatus, error) {
	if sourceBranch == "" || targetBranch == "" {
		return nil, errors.New("source and target branches cannot be empty")
	}

	status := &domain.MergeStatus{
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		Status:       domain.BranchStatusUnknown,
		Conflicts:    []string{},
	}

	// Get merge base (common ancestor)
	mergeBase, err := e.GetMergeBase(ctx, repoPath, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get merge base: %w", err)
	}
	status.CommonAncestor = mergeBase

	// Get divergence
	ahead, behind, err := e.GetDivergence(ctx, repoPath, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get divergence: %w", err)
	}
	status.AheadCount = ahead
	status.BehindCount = behind

	// Determine status
	if ahead == 0 && behind == 0 {
		status.Status = domain.BranchStatusUpToDate
		status.CanFastForward = true
	} else if ahead > 0 && behind == 0 {
		status.Status = domain.BranchStatusAhead
		status.CanFastForward = true
	} else if ahead == 0 && behind > 0 {
		status.Status = domain.BranchStatusBehind
	} else {
		status.Status = domain.BranchStatusDiverged
	}

	// Check for conflicts
	canMerge, conflicts, err := e.CanMerge(ctx, repoPath, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to check merge conflicts: %w", err)
	}

	if !canMerge {
		status.Status = domain.BranchStatusConflict
		status.Conflicts = conflicts
		status.CanFastForward = false
	}

	return status, nil
}
