package usecase

import (
	"context"
	"fmt"

	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// ManageBranchesUseCase handles branch management operations with validation.
type ManageBranchesUseCase struct {
	gitOps git.Operations
}

// NewManageBranchesUseCase creates a new ManageBranchesUseCase.
func NewManageBranchesUseCase(gitOps git.Operations) *ManageBranchesUseCase {
	return &ManageBranchesUseCase{
		gitOps: gitOps,
	}
}

// DeleteBranchRequest contains parameters for deleting a branch.
type DeleteBranchRequest struct {
	RepoPath          string
	BranchName        string
	Force             bool
	AlsoDeleteRemote  bool
	RemoteName        string
	ProtectedBranches []string
}

// DeleteBranchResponse contains the result of branch deletion.
type DeleteBranchResponse struct {
	Success              bool
	LocalDeleted         bool
	RemoteDeleted        bool
	Message              string
	RemoteDeletionError  error
}

// RenameBranchRequest contains parameters for renaming a branch.
type RenameBranchRequest struct {
	RepoPath string
	OldName  string
	NewName  string
}

// RenameBranchResponse contains the result of branch rename.
type RenameBranchResponse struct {
	Success bool
	Message string
}

// SetUpstreamRequest contains parameters for setting upstream branch.
type SetUpstreamRequest struct {
	RepoPath   string
	BranchName string
	Upstream   string
}

// SetUpstreamResponse contains the result of setting upstream.
type SetUpstreamResponse struct {
	Success bool
	Message string
}

// DeleteBranch deletes a branch with validation and optional remote deletion.
func (uc *ManageBranchesUseCase) DeleteBranch(ctx context.Context, req DeleteBranchRequest) (*DeleteBranchResponse, error) {
	if req.BranchName == "" {
		return nil, fmt.Errorf("branch name is required")
	}

	// Get current branch to prevent deletion
	currentBranch, err := uc.gitOps.GetCurrentBranch(ctx, req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Prevent deleting current branch
	if req.BranchName == currentBranch {
		return nil, fmt.Errorf("cannot delete currently checked out branch '%s'", req.BranchName)
	}

	// Check if branch is protected
	for _, protected := range req.ProtectedBranches {
		if req.BranchName == protected {
			return nil, fmt.Errorf("cannot delete protected branch '%s'", req.BranchName)
		}
	}

	resp := &DeleteBranchResponse{
		Success: true,
	}

	// Delete local branch
	if err := uc.gitOps.DeleteBranch(ctx, req.RepoPath, req.BranchName, req.Force); err != nil {
		return nil, fmt.Errorf("failed to delete local branch: %w", err)
	}

	resp.LocalDeleted = true
	resp.Message = fmt.Sprintf("Local branch '%s' deleted successfully", req.BranchName)

	// Delete remote branch if requested
	if req.AlsoDeleteRemote && req.RemoteName != "" {
		if err := uc.gitOps.DeleteRemoteBranch(ctx, req.RepoPath, req.RemoteName, req.BranchName); err != nil {
			// Don't fail the whole operation if remote deletion fails
			resp.RemoteDeletionError = err
			resp.Message = fmt.Sprintf("Local branch deleted, but remote deletion failed: %v", err)
		} else {
			resp.RemoteDeleted = true
			resp.Message = fmt.Sprintf("Branch '%s' deleted locally and remotely", req.BranchName)
		}
	}

	return resp, nil
}

// RenameBranch renames a branch with validation.
func (uc *ManageBranchesUseCase) RenameBranch(ctx context.Context, req RenameBranchRequest) (*RenameBranchResponse, error) {
	if req.OldName == "" || req.NewName == "" {
		return nil, fmt.Errorf("both old and new branch names are required")
	}

	if req.OldName == req.NewName {
		return nil, fmt.Errorf("new branch name must be different from old name")
	}

	// Check if new name already exists
	branches, err := uc.gitOps.ListBranches(ctx, req.RepoPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branch := range branches {
		if branch == req.NewName {
			return nil, fmt.Errorf("branch '%s' already exists", req.NewName)
		}
	}

	// Perform rename
	if err := uc.gitOps.RenameBranch(ctx, req.RepoPath, req.OldName, req.NewName); err != nil {
		return nil, fmt.Errorf("failed to rename branch: %w", err)
	}

	return &RenameBranchResponse{
		Success: true,
		Message: fmt.Sprintf("Branch renamed from '%s' to '%s'", req.OldName, req.NewName),
	}, nil
}

// SetUpstream sets the upstream tracking branch with validation.
func (uc *ManageBranchesUseCase) SetUpstream(ctx context.Context, req SetUpstreamRequest) (*SetUpstreamResponse, error) {
	if req.BranchName == "" {
		return nil, fmt.Errorf("branch name is required")
	}

	if req.Upstream == "" {
		return nil, fmt.Errorf("upstream branch is required")
	}

	// Perform set upstream
	if err := uc.gitOps.SetUpstreamBranch(ctx, req.RepoPath, req.BranchName, req.Upstream); err != nil {
		return nil, fmt.Errorf("failed to set upstream: %w", err)
	}

	return &SetUpstreamResponse{
		Success: true,
		Message: fmt.Sprintf("Upstream set to '%s' for branch '%s'", req.Upstream, req.BranchName),
	}, nil
}

// GetAllBranches retrieves all branches with detailed information.
func (uc *ManageBranchesUseCase) GetAllBranches(ctx context.Context, repoPath string, protectedBranches []string) ([]*domain.BranchInfo, error) {
	// Get current branch first
	currentBranch, err := uc.gitOps.GetCurrentBranch(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get list of all local branches
	branches, err := uc.gitOps.ListBranches(ctx, repoPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// Build detailed info for each branch
	branchInfos := make([]*domain.BranchInfo, 0, len(branches))
	for _, branchName := range branches {
		branchInfo, err := domain.NewBranchInfo(branchName)
		if err != nil {
			continue // Skip invalid branch names
		}

		// Set branch type (protected, feature, etc.)
		branchType := domain.DetectBranchType(branchName, protectedBranches)
		branchInfo.SetType(branchType)

		// Get parent branch from git config
		parent, _ := uc.gitOps.GetParentBranch(ctx, repoPath, branchName)
		if parent != "" {
			branchInfo.SetParent(parent)
		}

		// Get upstream tracking branch
		hasUpstream, _ := uc.gitOps.HasUpstream(ctx, repoPath, branchName)
		if hasUpstream {
			// Try to get the actual upstream branch name
			// This is safe to fail - we'll just not have upstream info
			ahead, behind, err := uc.gitOps.GetRemoteSyncStatus(ctx, repoPath, branchName)
			if err == nil {
				branchInfo.SetAheadBy(ahead)
				branchInfo.SetBehindBy(behind)
			}
		}

		// Get commit count relative to parent (if parent exists)
		if parent != "" && parent != branchName {
			commits, err := uc.gitOps.GetBranchCommits(ctx, repoPath, branchName, parent)
			if err == nil {
				branchInfo.SetCommitCount(len(commits))
			}
		}

		branchInfos = append(branchInfos, branchInfo)
	}

	// Sort branches - current branch first, then protected, then by name
	sortedBranches := sortBranches(branchInfos, currentBranch)

	return sortedBranches, nil
}

// sortBranches sorts branches with current first, then protected, then alphabetically.
func sortBranches(branches []*domain.BranchInfo, currentBranch string) []*domain.BranchInfo {
	var current, protected, other []*domain.BranchInfo

	for _, branch := range branches {
		if branch.Name() == currentBranch {
			current = append(current, branch)
		} else if branch.Type() == domain.BranchTypeProtected {
			protected = append(protected, branch)
		} else {
			other = append(other, branch)
		}
	}

	// Combine: current + protected + others
	result := make([]*domain.BranchInfo, 0, len(branches))
	result = append(result, current...)
	result = append(result, protected...)
	result = append(result, other...)

	return result
}
