package domain

import (
	"errors"
	"fmt"
	"path/filepath"
)

// FileChange represents a single file change in the repository.
type FileChange struct {
	Path         string
	Status       ChangeStatus
	Additions    int
	Deletions    int
	IsBinary     bool
	PatchPreview string // First few lines of diff for context
}

// ChangeStatus represents the type of change made to a file.
type ChangeStatus string

const (
	// StatusAdded indicates a new file was added.
	StatusAdded ChangeStatus = "added"
	// StatusModified indicates an existing file was modified.
	StatusModified ChangeStatus = "modified"
	// StatusDeleted indicates a file was deleted.
	StatusDeleted ChangeStatus = "deleted"
	// StatusRenamed indicates a file was renamed.
	StatusRenamed ChangeStatus = "renamed"
	// StatusUntracked indicates a file is untracked.
	StatusUntracked ChangeStatus = "untracked"
)

// String returns the string representation of the change status.
func (cs ChangeStatus) String() string {
	return string(cs)
}

// Repository represents the current state of a Git repository.
type Repository struct {
	path          string
	currentBranch string
	hasRemote     bool
	isClean       bool
	changes       []FileChange
}

// NewRepository creates a new Repository instance.
func NewRepository(path string) (*Repository, error) {
	if path == "" {
		return nil, errors.New("repository path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid repository path: %w", err)
	}

	return &Repository{
		path:    absPath,
		changes: make([]FileChange, 0),
	}, nil
}

// Path returns the repository path.
func (r *Repository) Path() string {
	return r.path
}

// CurrentBranch returns the current branch name.
func (r *Repository) CurrentBranch() string {
	return r.currentBranch
}

// SetCurrentBranch sets the current branch name.
func (r *Repository) SetCurrentBranch(branch string) {
	r.currentBranch = branch
}

// HasRemote returns true if the repository has a remote configured.
func (r *Repository) HasRemote() bool {
	return r.hasRemote
}

// SetHasRemote sets whether the repository has a remote.
func (r *Repository) SetHasRemote(hasRemote bool) {
	r.hasRemote = hasRemote
}

// IsClean returns true if the repository has no uncommitted changes.
func (r *Repository) IsClean() bool {
	return r.isClean
}

// SetIsClean sets whether the repository is clean.
func (r *Repository) SetIsClean(isClean bool) {
	r.isClean = isClean
}

// Changes returns the list of file changes.
func (r *Repository) Changes() []FileChange {
	return r.changes
}

// SetChanges sets the list of file changes.
func (r *Repository) SetChanges(changes []FileChange) {
	r.changes = changes
}

// AddChange adds a single file change to the repository.
func (r *Repository) AddChange(change FileChange) {
	r.changes = append(r.changes, change)
}

// TotalChanges returns the total number of changed files.
func (r *Repository) TotalChanges() int {
	return len(r.changes)
}

// TotalAdditions returns the total number of lines added across all changes.
func (r *Repository) TotalAdditions() int {
	total := 0
	for _, change := range r.changes {
		total += change.Additions
	}
	return total
}

// TotalDeletions returns the total number of lines deleted across all changes.
func (r *Repository) TotalDeletions() int {
	total := 0
	for _, change := range r.changes {
		total += change.Deletions
	}
	return total
}

// HasChanges returns true if there are any changes in the repository.
func (r *Repository) HasChanges() bool {
	return len(r.changes) > 0
}

// ChangeSummary returns a human-readable summary of changes.
func (r *Repository) ChangeSummary() string {
	if !r.HasChanges() {
		return "no changes"
	}

	return fmt.Sprintf("%d file(s) changed, +%d -%d",
		r.TotalChanges(),
		r.TotalAdditions(),
		r.TotalDeletions())
}

// IsLargeChangeset returns true if the changeset is large (many files or lines changed).
// This can be used to determine if context reduction is needed.
func (r *Repository) IsLargeChangeset() bool {
	// Consider large if more than 20 files or more than 500 total line changes
	return r.TotalChanges() > 20 || (r.TotalAdditions()+r.TotalDeletions()) > 500
}

// GetChangesByStatus returns all changes with the given status.
func (r *Repository) GetChangesByStatus(status ChangeStatus) []FileChange {
	filtered := make([]FileChange, 0)
	for _, change := range r.changes {
		if change.Status == status {
			filtered = append(filtered, change)
		}
	}
	return filtered
}

// String returns a string representation of the repository state.
func (r *Repository) String() string {
	return fmt.Sprintf("Repository{path: %s, branch: %s, changes: %s}",
		r.path, r.currentBranch, r.ChangeSummary())
}
