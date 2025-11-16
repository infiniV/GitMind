package usecase

import (
	"context"
	"fmt"

	"github.com/yourusername/gitman/internal/adapter/ai"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// AnalyzeCommitUseCase orchestrates the commit analysis workflow.
type AnalyzeCommitUseCase struct {
	gitOps   git.Operations
	aiProvider ai.Provider
}

// NewAnalyzeCommitUseCase creates a new AnalyzeCommitUseCase.
func NewAnalyzeCommitUseCase(gitOps git.Operations, aiProvider ai.Provider) *AnalyzeCommitUseCase {
	return &AnalyzeCommitUseCase{
		gitOps:     gitOps,
		aiProvider: aiProvider,
	}
}

// AnalyzeCommitRequest contains the input for commit analysis.
type AnalyzeCommitRequest struct {
	RepoPath               string
	UserPrompt             string
	UseConventionalCommits bool
	APIKey                 *domain.APIKey
}

// AnalyzeCommitResponse contains the result of commit analysis.
type AnalyzeCommitResponse struct {
	Repository *domain.Repository
	Decision   *domain.Decision
	Diff       string
	TokensUsed int
	Model      string
}

// Execute performs the commit analysis.
func (uc *AnalyzeCommitUseCase) Execute(ctx context.Context, req AnalyzeCommitRequest) (*AnalyzeCommitResponse, error) {
	// Validate repository
	isRepo, err := uc.gitOps.IsGitRepo(ctx, req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", req.RepoPath)
	}

	// Get repository status
	repo, err := uc.gitOps.GetStatus(ctx, req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status: %w", err)
	}

	// Check if there are changes to commit
	if !repo.HasChanges() {
		return nil, fmt.Errorf("no changes to commit")
	}

	// Get diff (check both staged and unstaged)
	stagedDiff, err := uc.gitOps.GetDiff(ctx, req.RepoPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff: %w", err)
	}

	unstagedDiff, err := uc.gitOps.GetDiff(ctx, req.RepoPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get unstaged diff: %w", err)
	}

	// Combine diffs
	diff := stagedDiff
	if diff == "" {
		diff = unstagedDiff
	}

	// If no diff available, we likely have untracked files
	// Read them directly from filesystem WITHOUT staging (to preserve clean state for branching)
	if diff == "" && repo.HasChanges() {
		// Build a synthetic diff from file contents
		fileDiff, err := uc.buildUntrackedFilesDiff(req.RepoPath, repo)
		if err != nil {
			// Fallback to simple file listing if we can't read files
			diff = fmt.Sprintf("New files to be added:\n%s", repo.ChangeSummary())
		} else {
			diff = fileDiff
		}
	}

	// Get recent commit log for context
	recentCommits, err := uc.gitOps.GetLog(ctx, req.RepoPath, 5)
	if err != nil {
		// Non-fatal, continue without log context
		recentCommits = []git.CommitInfo{}
	}

	recentLog := make([]string, len(recentCommits))
	for i, commit := range recentCommits {
		recentLog[i] = commit.Message
	}

	// Prepare AI analysis request
	aiReq := ai.AnalysisRequest{
		Repository:             repo,
		Diff:                   diff,
		RecentLog:              recentLog,
		UserPrompt:             req.UserPrompt,
		APIKey:                 req.APIKey,
		UseConventionalCommits: req.UseConventionalCommits,
	}

	// Analyze with AI
	aiResp, err := uc.aiProvider.Analyze(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	return &AnalyzeCommitResponse{
		Repository: repo,
		Decision:   aiResp.Decision,
		Diff:       diff,
		TokensUsed: aiResp.TokensUsed,
		Model:      aiResp.Model,
	}, nil
}
