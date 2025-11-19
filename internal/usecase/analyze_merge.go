package usecase

import (
	"context"
	"fmt"

	"github.com/yourusername/gitman/internal/adapter/ai"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// AnalyzeMergeUseCase analyzes a merge operation and provides AI recommendations.
type AnalyzeMergeUseCase struct {
	gitOps     git.Operations
	aiProvider ai.Provider
}

// NewAnalyzeMergeUseCase creates a new AnalyzeMergeUseCase.
func NewAnalyzeMergeUseCase(gitOps git.Operations, aiProvider ai.Provider) *AnalyzeMergeUseCase {
	return &AnalyzeMergeUseCase{
		gitOps:     gitOps,
		aiProvider: aiProvider,
	}
}

// AnalyzeMergeRequest contains the input for merge analysis.
type AnalyzeMergeRequest struct {
	RepoPath          string
	SourceBranch      string   // Optional, defaults to current branch
	TargetBranch      string   // Optional, defaults to parent branch
	ProtectedBranches []string
	APIKey            *domain.APIKey
}

// AnalyzeMergeResponse contains the result of merge analysis.
type AnalyzeMergeResponse struct {
	SourceBranchInfo  *domain.BranchInfo
	TargetBranch      string
	CommitCount       int
	Commits           []git.CommitInfo
	CanMerge          bool
	Conflicts         []string
	SuggestedStrategy string
	MergeMessage      *domain.CommitMessage
	SuggestedPR       *domain.PROptions // AI-suggested PR options (alternative to direct merge)
	Reasoning         string
	TokensUsed        int
	Model             string
}

// Execute performs the merge analysis.
func (uc *AnalyzeMergeUseCase) Execute(ctx context.Context, req AnalyzeMergeRequest) (*AnalyzeMergeResponse, error) {
	// Validate repository
	isRepo, err := uc.gitOps.IsGitRepo(ctx, req.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository: %s", req.RepoPath)
	}

	// Get source branch (current or specified)
	sourceBranch := req.SourceBranch
	if sourceBranch == "" {
		sourceBranch, err = uc.gitOps.GetCurrentBranch(ctx, req.RepoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	// Get source branch info
	sourceBranchInfo, err := uc.gitOps.GetBranchInfo(ctx, req.RepoPath, req.ProtectedBranches)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch info: %w", err)
	}

	// Get list of existing branches first
	branches, err := uc.gitOps.ListBranches(ctx, req.RepoPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	// Helper to check if branch exists
	branchExists := func(name string) bool {
		for _, branch := range branches {
			if branch == name {
				return true
			}
		}
		return false
	}

	// Determine target branch (specified, parent, or fallback to common branches)
	targetBranch := req.TargetBranch
	if targetBranch == "" {
		// Try to get configured parent branch
		parentBranch := sourceBranchInfo.Parent()
		if parentBranch != "" && branchExists(parentBranch) {
			targetBranch = parentBranch
		} else {
			// Parent doesn't exist or not configured, try common branch names
			commonBranches := []string{"main", "master", "develop", "development"}
			for _, branch := range commonBranches {
				if branch != sourceBranch && branchExists(branch) {
					targetBranch = branch
					break
				}
			}

			// Still no target? Use suggested merge target
			if targetBranch == "" {
				targetBranch = sourceBranchInfo.SuggestedMergeTarget()
			}
		}
	}

	// Validate target branch exists
	if !branchExists(targetBranch) {
		// Provide helpful error with available branches
		availableBranches := []string{}
		for _, branch := range branches {
			if branch != sourceBranch {
				availableBranches = append(availableBranches, branch)
			}
		}

		if len(availableBranches) == 0 {
			return nil, fmt.Errorf("no other branches available to merge into")
		}

		return nil, fmt.Errorf("target branch '%s' does not exist. Available branches: %v. Use -t flag to specify target", targetBranch, availableBranches)
	}

	// Check if source and target are the same
	if sourceBranch == targetBranch {
		return nil, fmt.Errorf("cannot merge branch into itself")
	}

	// Get commits to be merged (commits in source but not in target)
	commits, err := uc.gitOps.GetBranchCommits(ctx, req.RepoPath, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits to merge: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits to merge (branch '%s' is up to date with '%s')", sourceBranch, targetBranch)
	}

	// Check if merge is possible (detect conflicts)
	canMerge, conflicts, err := uc.gitOps.CanMerge(ctx, req.RepoPath, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to check merge possibility: %w", err)
	}

	// Get AI recommendation for merge message and strategy
	commitMessages := make([]string, len(commits))
	for i, commit := range commits {
		commitMessages[i] = commit.Message
	}

	mergeMessageReq := ai.MergeMessageRequest{
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		Commits:      commitMessages,
		CommitCount:  len(commits),
		APIKey:       req.APIKey,
	}

	mergeMessageResp, err := uc.aiProvider.GenerateMergeMessage(ctx, mergeMessageReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate merge message: %w", err)
	}

	// Generate suggested PR options as an alternative to direct merge
	// AI logic: Suggest PR when changes are significant, need review, or involve protected branches
	var suggestedPR *domain.PROptions

	// Heuristics for when to suggest PR:
	// 1. Many commits (>3)
	// 2. Merging to protected branch
	// 3. Has conflicts
	// 4. Complex changes (detected by AI or commit count)
	shouldSuggestPR := len(commits) > 3 || len(conflicts) > 0 || isProtectedBranch(targetBranch, req.ProtectedBranches)

	if shouldSuggestPR {
		// Build PR title from merge message
		prTitle := fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch)
		if mergeMessageResp.MergeMessage != nil && mergeMessageResp.MergeMessage.Title() != "" {
			prTitle = mergeMessageResp.MergeMessage.Title()
		}

		// Build PR body from commits
		prBody := fmt.Sprintf("This PR merges %d commit(s) from %s into %s.\n\n", len(commits), sourceBranch, targetBranch)
		if mergeMessageResp.MergeMessage != nil && mergeMessageResp.MergeMessage.Body() != "" {
			prBody += mergeMessageResp.MergeMessage.Body() + "\n\n"
		}
		prBody += "## Commits\n"
		for i, commit := range commits {
			if i < 10 { // Limit to first 10 commits
				prBody += fmt.Sprintf("- %s\n", commit.Message)
			}
		}
		if len(commits) > 10 {
			prBody += fmt.Sprintf("\n...and %d more commits\n", len(commits)-10)
		}

		prOpts, err := domain.NewPROptions(prTitle, targetBranch, sourceBranch)
		if err == nil {
			prOpts.SetBody(prBody)
			prOpts.SetUseTemplate(true)

			// Suggest draft if there are conflicts
			if len(conflicts) > 0 {
				prOpts.SetIsDraft(true)
			}

			suggestedPR = prOpts
		}
	}

	return &AnalyzeMergeResponse{
		SourceBranchInfo:  sourceBranchInfo,
		TargetBranch:      targetBranch,
		CommitCount:       len(commits),
		Commits:           commits,
		CanMerge:          canMerge,
		Conflicts:         conflicts,
		SuggestedStrategy: mergeMessageResp.SuggestedStrategy,
		MergeMessage:      mergeMessageResp.MergeMessage,
		SuggestedPR:       suggestedPR,
		Reasoning:         mergeMessageResp.Reasoning,
		TokensUsed:        mergeMessageResp.TokensUsed,
		Model:             mergeMessageResp.Model,
	}, nil
}

// isProtectedBranch checks if a branch is in the protected branches list.
func isProtectedBranch(branch string, protectedBranches []string) bool {
	for _, protected := range protectedBranches {
		if branch == protected {
			return true
		}
	}
	return false
}
