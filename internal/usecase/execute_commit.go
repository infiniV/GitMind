package usecase

import (
	"context"
	"fmt"

	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// ExecuteCommitUseCase executes the actual commit operation based on user decision.
type ExecuteCommitUseCase struct {
	gitOps git.Operations
}

// NewExecuteCommitUseCase creates a new ExecuteCommitUseCase.
func NewExecuteCommitUseCase(gitOps git.Operations) *ExecuteCommitUseCase {
	return &ExecuteCommitUseCase{
		gitOps: gitOps,
	}
}

// ExecuteCommitRequest contains the parameters for executing a commit.
type ExecuteCommitRequest struct {
	RepoPath      string
	Decision      *domain.Decision
	Action        domain.ActionType
	CommitMessage *domain.CommitMessage
	BranchName    string
	StageAll      bool
}

// ExecuteCommitResponse contains the result of the commit execution.
type ExecuteCommitResponse struct {
	Success      bool
	BranchCreated string
	CommitHash   string
	Message      string
}

// Execute performs the commit operation.
func (uc *ExecuteCommitUseCase) Execute(ctx context.Context, req ExecuteCommitRequest) (*ExecuteCommitResponse, error) {
	if req.CommitMessage == nil {
		return nil, fmt.Errorf("commit message is required")
	}

	resp := &ExecuteCommitResponse{
		Success: true,
	}

	switch req.Action {
	case domain.ActionCommitDirect:
		// Stage files first
		if req.StageAll {
			if err := uc.gitOps.Add(ctx, req.RepoPath, nil); err != nil {
				return nil, fmt.Errorf("failed to stage files: %w", err)
			}
		}

		// Commit directly to current branch
		if err := uc.gitOps.Commit(ctx, req.RepoPath, req.CommitMessage.FullMessage(), nil); err != nil {
			return nil, fmt.Errorf("failed to commit: %w", err)
		}
		resp.Message = "Changes committed successfully"

	case domain.ActionCreateBranch:
		// Create new branch and commit there
		if req.BranchName == "" {
			return nil, fmt.Errorf("branch name is required for create-branch action")
		}

		// For empty repos, we need to make an initial commit first
		// Check if we have any commits
		commits, err := uc.gitOps.GetLog(ctx, req.RepoPath, 1)
		if err != nil || len(commits) == 0 {
			// Empty repo - make initial commit on current branch first
			if req.StageAll {
				if err := uc.gitOps.Add(ctx, req.RepoPath, nil); err != nil {
					return nil, fmt.Errorf("failed to stage files: %w", err)
				}
			}
			if err := uc.gitOps.Commit(ctx, req.RepoPath, req.CommitMessage.FullMessage(), nil); err != nil {
				return nil, fmt.Errorf("failed to make initial commit: %w", err)
			}
			resp.Message = "Made initial commit on master (cannot create branch in empty repo)"
		} else {
			// Normal flow: create branch, checkout, then stage and commit
			// Create and checkout new branch BEFORE staging
			if err := uc.gitOps.CreateBranch(ctx, req.RepoPath, req.BranchName); err != nil {
				return nil, fmt.Errorf("failed to create branch: %w", err)
			}

			if err := uc.gitOps.CheckoutBranch(ctx, req.RepoPath, req.BranchName); err != nil {
				return nil, fmt.Errorf("failed to checkout branch: %w", err)
			}

			// NOW stage files on the new branch
			if req.StageAll {
				if err := uc.gitOps.Add(ctx, req.RepoPath, nil); err != nil {
					return nil, fmt.Errorf("failed to stage files on new branch: %w", err)
				}
			}

			// Commit on new branch
			if err := uc.gitOps.Commit(ctx, req.RepoPath, req.CommitMessage.FullMessage(), nil); err != nil {
				return nil, fmt.Errorf("failed to commit on new branch: %w", err)
			}

			resp.BranchCreated = req.BranchName
			resp.Message = fmt.Sprintf("Created branch '%s' and committed changes", req.BranchName)
		}

	default:
		return nil, fmt.Errorf("unsupported action: %s", req.Action)
	}

	return resp, nil
}
