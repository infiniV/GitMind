package usecase

import (
	"context"
	"fmt"

	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// ExecuteMergeUseCase executes the actual merge operation.
type ExecuteMergeUseCase struct {
	gitOps git.Operations
}

// NewExecuteMergeUseCase creates a new ExecuteMergeUseCase.
func NewExecuteMergeUseCase(gitOps git.Operations) *ExecuteMergeUseCase {
	return &ExecuteMergeUseCase{
		gitOps: gitOps,
	}
}

// ExecuteMergeRequest contains the parameters for executing a merge.
type ExecuteMergeRequest struct {
	RepoPath      string
	SourceBranch  string
	TargetBranch  string
	Strategy      string // "squash", "regular", "fast-forward", "rebase"
	MergeMessage  *domain.CommitMessage
}

// ExecuteMergeResponse contains the result of the merge execution.
type ExecuteMergeResponse struct {
	Success      bool
	MergeCommit  string
	Strategy     string
	Message      string
}

// Execute performs the merge operation.
func (uc *ExecuteMergeUseCase) Execute(ctx context.Context, req ExecuteMergeRequest) (*ExecuteMergeResponse, error) {
	if req.SourceBranch == "" || req.TargetBranch == "" {
		return nil, fmt.Errorf("source and target branches are required")
	}

	if req.SourceBranch == req.TargetBranch {
		return nil, fmt.Errorf("cannot merge branch into itself")
	}

	// Get current branch to restore later if needed
	currentBranch, err := uc.gitOps.GetCurrentBranch(ctx, req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Checkout target branch if not already on it
	if currentBranch != req.TargetBranch {
		if err := uc.gitOps.CheckoutBranch(ctx, req.RepoPath, req.TargetBranch); err != nil {
			return nil, fmt.Errorf("failed to checkout target branch '%s': %w", req.TargetBranch, err)
		}
	}

	// Prepare merge message
	mergeMsg := ""
	if req.MergeMessage != nil {
		mergeMsg = req.MergeMessage.FullMessage()
	} else if req.Strategy == "squash" || req.Strategy == "regular" {
		// Default merge message if none provided
		mergeMsg = fmt.Sprintf("Merge branch '%s' into %s", req.SourceBranch, req.TargetBranch)
	}

	// Execute merge with specified strategy
	strategy := req.Strategy
	if strategy == "" {
		strategy = "regular" // Default strategy
	}

	if err := uc.gitOps.Merge(ctx, req.RepoPath, req.SourceBranch, strategy, mergeMsg); err != nil {
		// Attempt to abort merge on failure
		uc.gitOps.AbortMerge(ctx, req.RepoPath)
		return nil, fmt.Errorf("merge failed: %w", err)
	}

	// Get the merge commit hash
	mergeCommit := ""
	log, err := uc.gitOps.GetLog(ctx, req.RepoPath, 1)
	if err == nil && len(log) > 0 {
		mergeCommit = log[0].Hash[:7] // Short hash
	}

	resp := &ExecuteMergeResponse{
		Success:     true,
		MergeCommit: mergeCommit,
		Strategy:    strategy,
		Message:     fmt.Sprintf("Successfully merged '%s' into '%s'", req.SourceBranch, req.TargetBranch),
	}

	return resp, nil
}
