package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
)

// ListPRUseCase lists pull requests for a repository.
type ListPRUseCase struct{}

// NewListPRUseCase creates a new ListPRUseCase.
func NewListPRUseCase() *ListPRUseCase {
	return &ListPRUseCase{}
}

// ListPRRequest contains the parameters for listing PRs.
type ListPRRequest struct {
	RepoPath string
	State    string // "open", "closed", "merged", "all"
}

// ListPRResponse contains the list of PRs.
type ListPRResponse struct {
	PRs     []*domain.PRInfo
	Count   int
	Message string
}

// Execute lists pull requests.
func (uc *ListPRUseCase) Execute(ctx context.Context, req ListPRRequest) (*ListPRResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// List PRs via GitHub CLI
	prs, err := github.ListPRs(ctx, req.RepoPath, req.State)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	return &ListPRResponse{
		PRs:     prs,
		Count:   len(prs),
		Message: fmt.Sprintf("Found %d pull request(s)", len(prs)),
	}, nil
}
