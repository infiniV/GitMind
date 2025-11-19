package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
)

// ExecutePRUseCase executes pull request operations.
type ExecutePRUseCase struct {
	gitOps git.Operations
	ghOps  GitHubOperations
}

// GitHubOperations defines GitHub-specific operations needed for PR management.
type GitHubOperations interface {
	CreatePR(ctx context.Context, repoPath string, opts *domain.PROptions) (*domain.PRInfo, error)
	ListPRs(ctx context.Context, repoPath string, state string) ([]*domain.PRInfo, error)
	GetPR(ctx context.Context, repoPath string, number int) (*domain.PRInfo, error)
}

// NewExecutePRUseCase creates a new ExecutePRUseCase.
func NewExecutePRUseCase(gitOps git.Operations) *ExecutePRUseCase {
	return &ExecutePRUseCase{
		gitOps: gitOps,
	}
}

// SetGitHubOps sets the GitHub operations (for dependency injection).
func (uc *ExecutePRUseCase) SetGitHubOps(ghOps GitHubOperations) {
	uc.ghOps = ghOps
}

// ExecutePRRequest contains the parameters for creating a pull request.
type ExecutePRRequest struct {
	RepoPath string
	PROptions *domain.PROptions
	AutoPush bool
	LoadTemplate bool
}

// ExecutePRResponse contains the result of PR creation.
type ExecutePRResponse struct {
	Success bool
	PRInfo  *domain.PRInfo
	Message string
	HTMLURL string
	Pushed  bool
}

// Execute creates a pull request.
func (uc *ExecutePRUseCase) Execute(ctx context.Context, req ExecutePRRequest) (*ExecutePRResponse, error) {
	if req.PROptions == nil {
		return nil, fmt.Errorf("PR options are required")
	}

	// Validate PR options
	if err := req.PROptions.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PR options: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	resp := &ExecutePRResponse{
		Success: false,
	}

	// Step 1: Check if we need to push
	headBranch := req.PROptions.HeadBranch()
	if headBranch == "" {
		currentBranch, err := uc.gitOps.GetCurrentBranch(ctx, req.RepoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
		headBranch = currentBranch
		req.PROptions.SetHeadBranch(headBranch)
	}

	// Step 2: Smart push detection
	if req.AutoPush {
		pushed, err := uc.smartPush(ctx, req.RepoPath, headBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to push branch: %w", err)
		}
		resp.Pushed = pushed
	}

	// Step 3: Load PR template if requested
	if req.LoadTemplate && req.PROptions.Body() == "" {
		template, err := uc.loadPRTemplate(req.RepoPath)
		if err == nil && template != "" {
			req.PROptions.SetBody(template)
		}
	}

	// Step 4: Create the PR via GitHub
	if uc.ghOps == nil {
		// Create a temporary GitHub operations instance
		uc.ghOps = &gitHubOpsWrapper{}
	}

	prInfo, err := uc.ghOps.CreatePR(ctx, req.RepoPath, req.PROptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Step 5: Build success response
	resp.Success = true
	resp.PRInfo = prInfo
	resp.HTMLURL = prInfo.HTMLURL()
	resp.Message = fmt.Sprintf("Pull request #%d created successfully", prInfo.Number())

	return resp, nil
}

// smartPush pushes the branch if it hasn't been pushed or has unpushed commits.
func (uc *ExecutePRUseCase) smartPush(ctx context.Context, repoPath, branch string) (bool, error) {
	// Check if branch has upstream
	hasUpstream, err := uc.gitOps.HasUpstream(ctx, repoPath, branch)
	if err != nil {
		return false, err
	}

	// If no upstream, definitely need to push
	if !hasUpstream {
		if err := uc.gitOps.Push(ctx, repoPath, branch, false); err != nil {
			return false, fmt.Errorf("failed to push branch: %w", err)
		}
		return true, nil
	}

	// If has upstream, check if there are unpushed commits
	unpushed, err := uc.gitOps.GetUnpushedCommits(ctx, repoPath, branch)
	if err != nil {
		return false, err
	}

	if unpushed > 0 {
		if err := uc.gitOps.Push(ctx, repoPath, branch, false); err != nil {
			return false, fmt.Errorf("failed to push commits: %w", err)
		}
		return true, nil
	}

	// Already up to date
	return false, nil
}

// loadPRTemplate attempts to load the PR template from .github directory.
func (uc *ExecutePRUseCase) loadPRTemplate(repoPath string) (string, error) {
	// Try common PR template locations
	templatePaths := []string{
		filepath.Join(repoPath, ".github", "PULL_REQUEST_TEMPLATE.md"),
		filepath.Join(repoPath, ".github", "pull_request_template.md"),
		filepath.Join(repoPath, "PULL_REQUEST_TEMPLATE.md"),
		filepath.Join(repoPath, "pull_request_template.md"),
	}

	for _, path := range templatePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(content)), nil
		}
	}

	return "", fmt.Errorf("no PR template found")
}

// gitHubOpsWrapper wraps the github package functions to implement GitHubOperations interface.
type gitHubOpsWrapper struct{}

func (w *gitHubOpsWrapper) CreatePR(ctx context.Context, repoPath string, opts *domain.PROptions) (*domain.PRInfo, error) {
	return github.CreatePR(ctx, repoPath, opts)
}

func (w *gitHubOpsWrapper) ListPRs(ctx context.Context, repoPath string, state string) ([]*domain.PRInfo, error) {
	return github.ListPRs(ctx, repoPath, state)
}

func (w *gitHubOpsWrapper) GetPR(ctx context.Context, repoPath string, number int) (*domain.PRInfo, error) {
	return github.GetPR(ctx, repoPath, number)
}
