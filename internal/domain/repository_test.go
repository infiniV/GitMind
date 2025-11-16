package domain

import (
	"path/filepath"
	"testing"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid path",
			path:    "/home/user/project",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    ".",
			wantErr: false,
		},
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errContains: "repository path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewRepository(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRepository() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewRepository() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRepository() unexpected error = %v", err)
				return
			}

			if repo == nil {
				t.Error("NewRepository() returned nil repository")
				return
			}

			// Path should be converted to absolute
			if !filepath.IsAbs(repo.Path()) {
				t.Errorf("NewRepository() path = %v, want absolute path", repo.Path())
			}

			// Should initialize with empty changes
			if len(repo.Changes()) != 0 {
				t.Errorf("NewRepository() changes = %v, want empty slice", repo.Changes())
			}
		})
	}
}

func TestRepository_SettersAndGetters(t *testing.T) {
	repo, _ := NewRepository(".")

	t.Run("CurrentBranch", func(t *testing.T) {
		repo.SetCurrentBranch("main")
		if got := repo.CurrentBranch(); got != "main" {
			t.Errorf("CurrentBranch() = %v, want %v", got, "main")
		}
	})

	t.Run("HasRemote", func(t *testing.T) {
		repo.SetHasRemote(true)
		if got := repo.HasRemote(); got != true {
			t.Errorf("HasRemote() = %v, want %v", got, true)
		}
	})

	t.Run("IsClean", func(t *testing.T) {
		repo.SetIsClean(true)
		if got := repo.IsClean(); got != true {
			t.Errorf("IsClean() = %v, want %v", got, true)
		}
	})
}

func TestRepository_Changes(t *testing.T) {
	repo, _ := NewRepository(".")

	changes := []FileChange{
		{Path: "file1.go", Status: StatusModified, Additions: 10, Deletions: 5},
		{Path: "file2.go", Status: StatusAdded, Additions: 20, Deletions: 0},
	}

	repo.SetChanges(changes)

	if len(repo.Changes()) != 2 {
		t.Errorf("Changes() length = %v, want %v", len(repo.Changes()), 2)
	}

	repo.AddChange(FileChange{Path: "file3.go", Status: StatusDeleted, Additions: 0, Deletions: 15})

	if len(repo.Changes()) != 3 {
		t.Errorf("After AddChange(), Changes() length = %v, want %v", len(repo.Changes()), 3)
	}
}

func TestRepository_TotalChanges(t *testing.T) {
	repo, _ := NewRepository(".")

	if got := repo.TotalChanges(); got != 0 {
		t.Errorf("TotalChanges() = %v, want %v", got, 0)
	}

	repo.SetChanges([]FileChange{
		{Path: "file1.go", Status: StatusModified},
		{Path: "file2.go", Status: StatusAdded},
		{Path: "file3.go", Status: StatusDeleted},
	})

	if got := repo.TotalChanges(); got != 3 {
		t.Errorf("TotalChanges() = %v, want %v", got, 3)
	}
}

func TestRepository_TotalAdditionsAndDeletions(t *testing.T) {
	tests := []struct {
		name             string
		changes          []FileChange
		wantAdditions    int
		wantDeletions    int
		wantHasChanges   bool
		wantIsLargeChangeset bool
	}{
		{
			name:             "no changes",
			changes:          []FileChange{},
			wantAdditions:    0,
			wantDeletions:    0,
			wantHasChanges:   false,
			wantIsLargeChangeset: false,
		},
		{
			name: "small changeset",
			changes: []FileChange{
				{Path: "file1.go", Additions: 10, Deletions: 5},
				{Path: "file2.go", Additions: 20, Deletions: 3},
			},
			wantAdditions:    30,
			wantDeletions:    8,
			wantHasChanges:   true,
			wantIsLargeChangeset: false,
		},
		{
			name: "large changeset by line count",
			changes: []FileChange{
				{Path: "file1.go", Additions: 300, Deletions: 250},
			},
			wantAdditions:    300,
			wantDeletions:    250,
			wantHasChanges:   true,
			wantIsLargeChangeset: true,
		},
		{
			name: "large changeset by file count",
			changes: func() []FileChange {
				changes := make([]FileChange, 25)
				for i := range changes {
					changes[i] = FileChange{Path: "file.go", Additions: 1, Deletions: 1}
				}
				return changes
			}(),
			wantAdditions:    25,
			wantDeletions:    25,
			wantHasChanges:   true,
			wantIsLargeChangeset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _ := NewRepository(".")
			repo.SetChanges(tt.changes)

			if got := repo.TotalAdditions(); got != tt.wantAdditions {
				t.Errorf("TotalAdditions() = %v, want %v", got, tt.wantAdditions)
			}

			if got := repo.TotalDeletions(); got != tt.wantDeletions {
				t.Errorf("TotalDeletions() = %v, want %v", got, tt.wantDeletions)
			}

			if got := repo.HasChanges(); got != tt.wantHasChanges {
				t.Errorf("HasChanges() = %v, want %v", got, tt.wantHasChanges)
			}

			if got := repo.IsLargeChangeset(); got != tt.wantIsLargeChangeset {
				t.Errorf("IsLargeChangeset() = %v, want %v", got, tt.wantIsLargeChangeset)
			}
		})
	}
}

func TestRepository_ChangeSummary(t *testing.T) {
	tests := []struct {
		name    string
		changes []FileChange
		want    string
	}{
		{
			name:    "no changes",
			changes: []FileChange{},
			want:    "no changes",
		},
		{
			name: "single file",
			changes: []FileChange{
				{Path: "file.go", Additions: 10, Deletions: 5},
			},
			want: "1 file(s) changed, +10 -5",
		},
		{
			name: "multiple files",
			changes: []FileChange{
				{Path: "file1.go", Additions: 10, Deletions: 5},
				{Path: "file2.go", Additions: 20, Deletions: 3},
			},
			want: "2 file(s) changed, +30 -8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _ := NewRepository(".")
			repo.SetChanges(tt.changes)

			if got := repo.ChangeSummary(); got != tt.want {
				t.Errorf("ChangeSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetChangesByStatus(t *testing.T) {
	repo, _ := NewRepository(".")
	repo.SetChanges([]FileChange{
		{Path: "file1.go", Status: StatusModified},
		{Path: "file2.go", Status: StatusAdded},
		{Path: "file3.go", Status: StatusModified},
		{Path: "file4.go", Status: StatusDeleted},
	})

	tests := []struct {
		name   string
		status ChangeStatus
		want   int
	}{
		{"get modified", StatusModified, 2},
		{"get added", StatusAdded, 1},
		{"get deleted", StatusDeleted, 1},
		{"get renamed", StatusRenamed, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.GetChangesByStatus(tt.status)
			if len(got) != tt.want {
				t.Errorf("GetChangesByStatus(%v) returned %d changes, want %d", tt.status, len(got), tt.want)
			}
		})
	}
}

func TestRepository_String(t *testing.T) {
	repo, _ := NewRepository(".")
	repo.SetCurrentBranch("main")
	repo.SetChanges([]FileChange{
		{Path: "file.go", Additions: 10, Deletions: 5},
	})

	str := repo.String()
	if !contains(str, "main") {
		t.Errorf("String() = %v, want to contain 'main'", str)
	}
	if !contains(str, "1 file") {
		t.Errorf("String() = %v, want to contain '1 file'", str)
	}
}

func TestChangeStatus_String(t *testing.T) {
	tests := []struct {
		status ChangeStatus
		want   string
	}{
		{StatusAdded, "added"},
		{StatusModified, "modified"},
		{StatusDeleted, "deleted"},
		{StatusRenamed, "renamed"},
		{StatusUntracked, "untracked"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("ChangeStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
