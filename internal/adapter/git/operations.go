package git

import (
	"context"

	"github.com/yourusername/gitman/internal/domain"
)

// Operations defines the interface for Git operations.
// This abstraction allows for different implementations (os/exec, go-git, git2go, etc.)
// and makes the code testable by allowing mock implementations.
type Operations interface {
	// GetStatus returns the current repository status including changes and branch info.
	GetStatus(ctx context.Context, repoPath string) (*domain.Repository, error)

	// GetDiff returns the diff for staged/unstaged changes.
	// If staged is true, returns diff for staged changes; otherwise unstaged changes.
	GetDiff(ctx context.Context, repoPath string, staged bool) (string, error)

	// GetCurrentBranch returns the name of the current branch.
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)

	// HasRemote returns true if the repository has a remote configured.
	HasRemote(ctx context.Context, repoPath string) (bool, error)

	// CreateBranch creates a new branch with the given name.
	CreateBranch(ctx context.Context, repoPath, branchName string) error

	// CheckoutBranch switches to the specified branch.
	CheckoutBranch(ctx context.Context, repoPath, branchName string) error

	// Commit creates a commit with the given message.
	// If files is empty, commits all staged changes.
	Commit(ctx context.Context, repoPath string, message string, files []string) error

	// Add stages files for commit.
	// If files is empty, stages all changes (git add -A).
	Add(ctx context.Context, repoPath string, files []string) error

	// Push pushes commits to the remote repository.
	// If branch is empty, pushes the current branch.
	Push(ctx context.Context, repoPath, branch string, force bool) error

	// Pull pulls changes from the remote repository.
	Pull(ctx context.Context, repoPath string) error

	// IsGitRepo returns true if the path is a valid git repository.
	IsGitRepo(ctx context.Context, path string) (bool, error)

	// GetLog returns recent commit history (limited to count).
	GetLog(ctx context.Context, repoPath string, count int) ([]CommitInfo, error)
}

// CommitInfo represents information about a commit.
type CommitInfo struct {
	Hash    string
	Author  string
	Date    string
	Message string
}

// DiffStats represents statistics about a diff.
type DiffStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}
