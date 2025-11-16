package domain

import (
	"errors"
	"strings"
)

// BranchType represents the type of a git branch.
type BranchType string

const (
	// BranchTypeProtected indicates a protected branch (main, master, develop, etc.)
	BranchTypeProtected BranchType = "protected"
	// BranchTypeFeature indicates a feature branch (feature/*, feat/*)
	BranchTypeFeature BranchType = "feature"
	// BranchTypeHotfix indicates a hotfix branch (hotfix/*)
	BranchTypeHotfix BranchType = "hotfix"
	// BranchTypeBugfix indicates a bugfix branch (bugfix/*, fix/*)
	BranchTypeBugfix BranchType = "bugfix"
	// BranchTypeRelease indicates a release branch (release/*)
	BranchTypeRelease BranchType = "release"
	// BranchTypeRefactor indicates a refactor branch (refactor/*)
	BranchTypeRefactor BranchType = "refactor"
	// BranchTypeOther indicates other branch types
	BranchTypeOther BranchType = "other"
)

// String returns the string representation of the branch type.
func (bt BranchType) String() string {
	return string(bt)
}

// MergeStrategy represents the strategy to use when merging a branch.
type MergeStrategy string

const (
	// MergeStrategyRegular performs a regular merge (creates merge commit)
	MergeStrategyRegular MergeStrategy = "regular"
	// MergeStrategySquash squashes all commits into one
	MergeStrategySquash MergeStrategy = "squash"
	// MergeStrategyFastForward performs a fast-forward merge if possible
	MergeStrategyFastForward MergeStrategy = "fast-forward"
	// MergeStrategyRebase rebases the branch onto the target
	MergeStrategyRebase MergeStrategy = "rebase"
	// MergeStrategyAsk prompts the user to choose
	MergeStrategyAsk MergeStrategy = "ask"
)

// String returns the string representation of the merge strategy.
func (ms MergeStrategy) String() string {
	return string(ms)
}

// BranchInfo contains metadata about a git branch.
type BranchInfo struct {
	name        string
	branchType  BranchType
	parent      string // Parent/base branch
	upstream    string // Upstream tracking branch
	aheadBy     int    // Commits ahead of upstream
	behindBy    int    // Commits behind of upstream
	commitCount int    // Number of commits on this branch (relative to parent)
	isProtected bool   // Whether this is a protected branch
}

// NewBranchInfo creates a new BranchInfo instance.
func NewBranchInfo(name string) (*BranchInfo, error) {
	if name == "" {
		return nil, errors.New("branch name cannot be empty")
	}

	branchType := DetectBranchType(name, []string{})

	return &BranchInfo{
		name:        name,
		branchType:  branchType,
		isProtected: branchType == BranchTypeProtected,
	}, nil
}

// Name returns the branch name.
func (bi *BranchInfo) Name() string {
	return bi.name
}

// Type returns the branch type.
func (bi *BranchInfo) Type() BranchType {
	return bi.branchType
}

// Parent returns the parent/base branch.
func (bi *BranchInfo) Parent() string {
	return bi.parent
}

// SetParent sets the parent/base branch.
func (bi *BranchInfo) SetParent(parent string) {
	bi.parent = parent
}

// Upstream returns the upstream tracking branch.
func (bi *BranchInfo) Upstream() string {
	return bi.upstream
}

// SetUpstream sets the upstream tracking branch.
func (bi *BranchInfo) SetUpstream(upstream string) {
	bi.upstream = upstream
}

// AheadBy returns commits ahead of upstream.
func (bi *BranchInfo) AheadBy() int {
	return bi.aheadBy
}

// SetAheadBy sets commits ahead of upstream.
func (bi *BranchInfo) SetAheadBy(count int) {
	bi.aheadBy = count
}

// BehindBy returns commits behind upstream.
func (bi *BranchInfo) BehindBy() int {
	return bi.behindBy
}

// SetBehindBy sets commits behind upstream.
func (bi *BranchInfo) SetBehindBy(count int) {
	bi.behindBy = count
}

// CommitCount returns the number of commits on this branch.
func (bi *BranchInfo) CommitCount() int {
	return bi.commitCount
}

// SetCommitCount sets the number of commits.
func (bi *BranchInfo) SetCommitCount(count int) {
	bi.commitCount = count
}

// IsProtected returns true if this is a protected branch.
func (bi *BranchInfo) IsProtected() bool {
	return bi.isProtected
}

// SetIsProtected sets whether this is a protected branch.
func (bi *BranchInfo) SetIsProtected(protected bool) {
	bi.isProtected = protected
	if protected {
		bi.branchType = BranchTypeProtected
	}
}

// SetType sets the branch type.
func (bi *BranchInfo) SetType(branchType BranchType) {
	bi.branchType = branchType
	bi.isProtected = branchType == BranchTypeProtected
}

// DetectBranchType detects the type of branch based on naming patterns and protected list.
func DetectBranchType(name string, protectedBranches []string) BranchType {
	// Check if in protected list first
	for _, protected := range protectedBranches {
		if name == protected {
			return BranchTypeProtected
		}
	}

	// Common protected branch names (fallback)
	commonProtected := []string{"main", "master", "develop", "development", "production", "prod"}
	for _, protected := range commonProtected {
		if name == protected {
			return BranchTypeProtected
		}
	}

	// Check naming patterns
	lowerName := strings.ToLower(name)

	if strings.HasPrefix(lowerName, "feature/") || strings.HasPrefix(lowerName, "feat/") {
		return BranchTypeFeature
	}

	if strings.HasPrefix(lowerName, "hotfix/") {
		return BranchTypeHotfix
	}

	if strings.HasPrefix(lowerName, "bugfix/") || strings.HasPrefix(lowerName, "fix/") {
		return BranchTypeBugfix
	}

	if strings.HasPrefix(lowerName, "release/") {
		return BranchTypeRelease
	}

	if strings.HasPrefix(lowerName, "refactor/") {
		return BranchTypeRefactor
	}

	return BranchTypeOther
}

// ShouldCreateBranch returns true if changes on this branch should create a sub-branch.
func (bi *BranchInfo) ShouldCreateBranch() bool {
	// Protected branches should always create a branch
	return bi.isProtected
}

// CanMergeTo returns true if this branch can be merged to the target branch.
func (bi *BranchInfo) CanMergeTo(targetBranch string) bool {
	// Can't merge to self
	if bi.name == targetBranch {
		return false
	}

	// Protected branches typically don't get merged (they're the merge target)
	if bi.isProtected {
		return false
	}

	return true
}

// SuggestedMergeTarget returns the suggested branch to merge into.
func (bi *BranchInfo) SuggestedMergeTarget() string {
	// If we have a parent, that's the target
	if bi.parent != "" {
		return bi.parent
	}

	// Feature branches typically merge to main/develop
	if bi.branchType == BranchTypeFeature || bi.branchType == BranchTypeRefactor {
		return "main" // or could be "develop" depending on workflow
	}

	// Hotfix branches merge to main
	if bi.branchType == BranchTypeHotfix {
		return "main"
	}

	// Bugfix branches merge to main or develop
	if bi.branchType == BranchTypeBugfix {
		return "main"
	}

	// Default to main
	return "main"
}
