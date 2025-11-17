package domain

import (
	"time"
)

// BranchStatus represents the status of a branch relative to its parent or merge target.
type BranchStatus int

const (
	BranchStatusUnknown BranchStatus = iota
	BranchStatusUpToDate
	BranchStatusAhead
	BranchStatusBehind
	BranchStatusDiverged
	BranchStatusConflict
)

func (s BranchStatus) String() string {
	switch s {
	case BranchStatusUpToDate:
		return "up-to-date"
	case BranchStatusAhead:
		return "ahead"
	case BranchStatusBehind:
		return "behind"
	case BranchStatusDiverged:
		return "diverged"
	case BranchStatusConflict:
		return "conflict"
	default:
		return "unknown"
	}
}

// CommitNode represents a single commit in the git graph visualization.
type CommitNode struct {
	Hash      string    // Full commit hash
	ShortHash string    // Abbreviated hash (7 chars)
	Author    string    // Commit author name
	Date      time.Time // Commit timestamp
	Message   string    // Commit message (first line)
	FullMessage string  // Full commit message (multi-line)
	Branch    string    // Branch this commit belongs to
	Parents   []string  // Parent commit hashes
	Children  []string  // Child commit hashes (for traversal)
	GraphLine string    // Pre-rendered ASCII graph line (e.g., "* ─┬─")
	IsMerge   bool      // True if this is a merge commit
	IsHead    bool      // True if this is the HEAD commit
	Tags      []string  // Git tags on this commit
}

// BranchNode represents a branch in the repository with its relationships.
type BranchNode struct {
	Name          string       // Branch name (e.g., "feature/dual-pane")
	FullName      string       // Full ref name (e.g., "refs/heads/feature/dual-pane")
	Parent        string       // Parent branch name (from git config or inferred)
	Type          BranchType   // Branch type (feature, hotfix, etc.)
	Head          string       // Hash of the HEAD commit on this branch
	Commits       []string     // Commit hashes unique to this branch (not in parent)
	Status        BranchStatus // Status relative to parent (ahead, behind, etc.)
	AheadCount    int          // Number of commits ahead of parent
	BehindCount   int          // Number of commits behind parent
	IsProtected   bool         // True if this is a protected branch
	IsCurrent     bool         // True if this is the currently checked out branch
	IsRemote      bool         // True if this is a remote tracking branch
	LastCommitDate time.Time   // Date of the most recent commit on this branch
}

// CommitGraph represents the full git graph structure.
type CommitGraph struct {
	Commits  []CommitNode       // All commits in chronological order (newest first)
	Branches []BranchNode       // All branches in the repository
	BranchMap map[string]*BranchNode // Map of branch name to BranchNode
	CommitMap map[string]*CommitNode // Map of commit hash to CommitNode
}

// GetBranchByName returns the BranchNode for the given branch name, or nil if not found.
func (cg *CommitGraph) GetBranchByName(name string) *BranchNode {
	if cg.BranchMap == nil {
		return nil
	}
	return cg.BranchMap[name]
}

// GetCommitByHash returns the CommitNode for the given hash, or nil if not found.
func (cg *CommitGraph) GetCommitByHash(hash string) *CommitNode {
	if cg.CommitMap == nil {
		return nil
	}
	return cg.CommitMap[hash]
}

// GetCommitsForBranch returns all commits that are unique to the given branch (not in parent).
func (cg *CommitGraph) GetCommitsForBranch(branchName string) []CommitNode {
	branch := cg.GetBranchByName(branchName)
	if branch == nil {
		return nil
	}

	var commits []CommitNode
	for _, hash := range branch.Commits {
		if commit := cg.GetCommitByHash(hash); commit != nil {
			commits = append(commits, *commit)
		}
	}
	return commits
}

// MergeStatus represents the merge status between two branches.
type MergeStatus struct {
	SourceBranch string       // Source branch (to be merged from)
	TargetBranch string       // Target branch (to be merged into)
	Status       BranchStatus // Overall merge status
	Conflicts    []string     // Files with merge conflicts (if any)
	CanFastForward bool       // True if merge can be fast-forwarded
	AheadCount   int          // Commits in source not in target
	BehindCount  int          // Commits in target not in source
	CommonAncestor string     // Hash of the common ancestor commit
}
