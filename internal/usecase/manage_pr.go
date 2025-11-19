package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
)

// ManagePRUseCase handles PR management operations (update, close, merge).
type ManagePRUseCase struct{}

// NewManagePRUseCase creates a new ManagePRUseCase.
func NewManagePRUseCase() *ManagePRUseCase {
	return &ManagePRUseCase{}
}

// ManagePRRequest contains the parameters for managing a PR.
type ManagePRRequest struct {
	RepoPath string
	PRNumber int
	Action   domain.PRAction
	Updates  map[string]string // For update action
	MergeMethod string         // For merge action: "merge", "squash", "rebase"
}

// ManagePRResponse contains the result of the management operation.
type ManagePRResponse struct {
	Success bool
	Message string
	PRInfo  *domain.PRInfo
}

// Execute performs the PR management operation.
func (uc *ManagePRUseCase) Execute(ctx context.Context, req ManagePRRequest) (*ManagePRResponse, error) {
	if req.PRNumber <= 0 {
		return nil, fmt.Errorf("invalid PR number: %d", req.PRNumber)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp := &ManagePRResponse{
		Success: false,
	}

	switch req.Action {
	case domain.PRActionUpdate:
		if err := github.UpdatePR(ctx, req.RepoPath, req.PRNumber, req.Updates); err != nil {
			return nil, fmt.Errorf("failed to update PR: %w", err)
		}
		resp.Message = fmt.Sprintf("Pull request #%d updated successfully", req.PRNumber)
		resp.Success = true

	case domain.PRActionClose:
		if err := github.ClosePR(ctx, req.RepoPath, req.PRNumber); err != nil {
			return nil, fmt.Errorf("failed to close PR: %w", err)
		}
		resp.Message = fmt.Sprintf("Pull request #%d closed", req.PRNumber)
		resp.Success = true

	case domain.PRActionMerge:
		if req.MergeMethod == "" {
			req.MergeMethod = "merge"
		}
		if err := github.MergePRRemote(ctx, req.RepoPath, req.PRNumber, req.MergeMethod); err != nil {
			return nil, fmt.Errorf("failed to merge PR: %w", err)
		}
		resp.Message = fmt.Sprintf("Pull request #%d merged using %s method", req.PRNumber, req.MergeMethod)
		resp.Success = true

	case domain.PRActionConvertToDraft:
		if err := github.ConvertPRToDraft(ctx, req.RepoPath, req.PRNumber); err != nil {
			return nil, fmt.Errorf("failed to convert PR to draft: %w", err)
		}
		resp.Message = fmt.Sprintf("Pull request #%d converted to draft", req.PRNumber)
		resp.Success = true

	case domain.PRActionMarkReady:
		if err := github.MarkPRReady(ctx, req.RepoPath, req.PRNumber); err != nil {
			return nil, fmt.Errorf("failed to mark PR as ready: %w", err)
		}
		resp.Message = fmt.Sprintf("Pull request #%d marked as ready for review", req.PRNumber)
		resp.Success = true

	default:
		return nil, fmt.Errorf("unsupported PR action: %s", req.Action)
	}

	// Get updated PR info
	prInfo, err := github.GetPR(ctx, req.RepoPath, req.PRNumber)
	if err == nil {
		resp.PRInfo = prInfo
	}

	return resp, nil
}
