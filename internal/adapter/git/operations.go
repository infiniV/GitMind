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

	// Fetch fetches updates from the remote repository without merging.
	Fetch(ctx context.Context, repoPath string) error

	// HasUpstream checks if the specified branch has an upstream tracking branch.
	// If branch is empty, checks the current branch.
	HasUpstream(ctx context.Context, repoPath, branch string) (bool, error)

	// GetUnpushedCommits returns the number of commits that haven't been pushed to the remote.
	// If branch is empty, uses the current branch.
	GetUnpushedCommits(ctx context.Context, repoPath, branch string) (int, error)

	// GetCommitRange returns commits between base and head branches.
	// This is useful for PR descriptions to see what commits would be included.
	GetCommitRange(ctx context.Context, repoPath, baseBranch, headBranch string) ([]CommitInfo, error)

	// GetRemoteURL returns the URL for the specified remote (usually "origin").
	GetRemoteURL(ctx context.Context, repoPath, remoteName string) (string, error)

	// GetRemoteName returns the primary remote name (defaults to "origin").
	GetRemoteName(ctx context.Context, repoPath string) (string, error)

	// GetRemoteSyncStatus returns commits ahead/behind relative to remote tracking branch.
	GetRemoteSyncStatus(ctx context.Context, repoPath, branch string) (ahead, behind int, err error)

	// IsGitRepo returns true if the path is a valid git repository.
	IsGitRepo(ctx context.Context, path string) (bool, error)

	// GetLog returns recent commit history (limited to count).
	GetLog(ctx context.Context, repoPath string, count int) ([]CommitInfo, error)

	// Branch Intelligence Operations

	// GetBranchInfo returns detailed information about the current branch.
	GetBranchInfo(ctx context.Context, repoPath string, protectedBranches []string) (*domain.BranchInfo, error)

	// GetMergeBase returns the common ancestor commit hash between two branches.
	GetMergeBase(ctx context.Context, repoPath, branch1, branch2 string) (string, error)

	// GetBranchCommits returns commits unique to a branch (not in excludeBranch).
	GetBranchCommits(ctx context.Context, repoPath, branch, excludeBranch string) ([]CommitInfo, error)

	// ListBranches returns all local and optionally remote branches.
	ListBranches(ctx context.Context, repoPath string, includeRemote bool) ([]string, error)

	// GetDivergence returns how many commits ahead/behind branch1 is compared to branch2.
	GetDivergence(ctx context.Context, repoPath, branch1, branch2 string) (ahead, behind int, err error)

	// Parent Branch Tracking (via git config)

	// GetParentBranch returns the parent branch for the given branch.
	GetParentBranch(ctx context.Context, repoPath, branch string) (string, error)

	// SetParentBranch sets the parent branch for the given branch in git config.
	SetParentBranch(ctx context.Context, repoPath, branch, parent string) error

	// Merge Operations

	// Merge merges sourceBranch into the current branch using the specified strategy.
	Merge(ctx context.Context, repoPath, sourceBranch, strategy, message string) error

	// CanMerge checks if sourceBranch can be merged into targetBranch without conflicts.
	// Returns true if merge is clean, false + conflict list if there are conflicts.
	CanMerge(ctx context.Context, repoPath, sourceBranch, targetBranch string) (bool, []string, error)

	// AbortMerge aborts an in-progress merge.
	AbortMerge(ctx context.Context, repoPath string) error
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

// GitHubRepo represents parsed GitHub repository information from a git URL.
type GitHubRepo struct {
	Owner string
	Repo  string
}
