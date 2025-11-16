package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

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
	args := []string{"push"}

	if force {
		args = append(args, "--force")
	}

	if branch != "" {
		args = append(args, "origin", branch)
	}

	_, stderr, err := e.execGit(ctx, repoPath, args...)
	if err != nil {
		if strings.Contains(stderr, "no upstream branch") {
			return fmt.Errorf("no upstream branch configured. Use: git push -u origin %s", branch)
		}
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
