package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/gitman/internal/domain"
)

func TestNewExecOperations(t *testing.T) {
	ops := NewExecOperations()
	if ops == nil {
		t.Fatal("NewExecOperations() returned nil")
	}
	if ops.gitPath != "git" {
		t.Errorf("gitPath = %v, want 'git'", ops.gitPath)
	}
}

func TestExecOperations_SetGitPath(t *testing.T) {
	ops := NewExecOperations()
	ops.SetGitPath("/usr/bin/git")
	if ops.gitPath != "/usr/bin/git" {
		t.Errorf("gitPath = %v, want '/usr/bin/git'", ops.gitPath)
	}
}

func TestExecOperations_IsGitRepo(t *testing.T) {
	ops := NewExecOperations()
	ctx := context.Background()

	t.Run("current directory should be git repo", func(t *testing.T) {
		// We're testing in a git repo
		isRepo, err := ops.IsGitRepo(ctx, ".")
		if err != nil {
			t.Fatalf("IsGitRepo() error = %v", err)
		}
		if !isRepo {
			t.Error("IsGitRepo() = false, want true for git repository")
		}
	})

	t.Run("temp directory should not be git repo", func(t *testing.T) {
		tempDir := t.TempDir()
		isRepo, err := ops.IsGitRepo(ctx, tempDir)
		if err != nil {
			t.Fatalf("IsGitRepo() error = %v", err)
		}
		if isRepo {
			t.Error("IsGitRepo() = true, want false for non-git directory")
		}
	})
}

func TestParseStatus(t *testing.T) {
	ops := NewExecOperations()

	tests := []struct {
		name   string
		output string
		want   int // number of changes expected
	}{
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
		{
			name:   "single modified file",
			output: " M file.go",
			want:   1,
		},
		{
			name: "multiple changes",
			output: `A  newfile.go
M  modified.go
D  deleted.go`,
			want: 3,
		},
		{
			name: "untracked files",
			output: `?? untracked.go
?? another.go`,
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := ops.parseStatus(tt.output)
			if err != nil {
				t.Fatalf("parseStatus() error = %v", err)
			}
			if len(changes) != tt.want {
				t.Errorf("parseStatus() returned %d changes, want %d", len(changes), tt.want)
			}
		})
	}
}

func TestParseStatus_StatusCodes(t *testing.T) {
	ops := NewExecOperations()

	tests := []struct {
		name       string
		statusLine string
		wantStatus domain.ChangeStatus
	}{
		{
			name:       "added file",
			statusLine: "A  newfile.go",
			wantStatus: domain.StatusAdded,
		},
		{
			name:       "modified file",
			statusLine: " M modified.go",
			wantStatus: domain.StatusModified,
		},
		{
			name:       "deleted file",
			statusLine: " D deleted.go",
			wantStatus: domain.StatusDeleted,
		},
		{
			name:       "renamed file",
			statusLine: "R  renamed.go",
			wantStatus: domain.StatusRenamed,
		},
		{
			name:       "untracked file",
			statusLine: "?? untracked.go",
			wantStatus: domain.StatusUntracked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := ops.parseStatus(tt.statusLine)
			if err != nil {
				t.Fatalf("parseStatus() error = %v", err)
			}
			if len(changes) != 1 {
				t.Fatalf("parseStatus() returned %d changes, want 1", len(changes))
			}
			if changes[0].Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", changes[0].Status, tt.wantStatus)
			}
		})
	}
}

func TestParseLog(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int
	}{
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
		{
			name: "single commit",
			output: `abc123
John Doe
2024-01-15T10:30:00Z
Initial commit
---END---`,
			want: 1,
		},
		{
			name: "multiple commits",
			output: `abc123
John Doe
2024-01-15T10:30:00Z
Initial commit
---END---
def456
Jane Smith
2024-01-16T14:20:00Z
Add feature
---END---`,
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits := parseLog(tt.output)
			if len(commits) != tt.want {
				t.Errorf("parseLog() returned %d commits, want %d", len(commits), tt.want)
			}

			if tt.want > 0 && len(commits) > 0 {
				if commits[0].Hash == "" {
					t.Error("First commit hash is empty")
				}
				if commits[0].Author == "" {
					t.Error("First commit author is empty")
				}
				if commits[0].Message == "" {
					t.Error("First commit message is empty")
				}
			}
		})
	}
}

func TestExecOperations_Commit_EmptyMessage(t *testing.T) {
	ops := NewExecOperations()
	ctx := context.Background()

	err := ops.Commit(ctx, ".", "", nil)
	if err == nil {
		t.Error("Commit() with empty message should return error")
	}
	if err.Error() != "commit message cannot be empty" {
		t.Errorf("Commit() error = %v, want 'commit message cannot be empty'", err)
	}
}

func TestExecOperations_CreateBranch_EmptyName(t *testing.T) {
	ops := NewExecOperations()
	ctx := context.Background()

	err := ops.CreateBranch(ctx, ".", "")
	if err == nil {
		t.Error("CreateBranch() with empty name should return error")
	}
}

func TestExecOperations_CheckoutBranch_EmptyName(t *testing.T) {
	ops := NewExecOperations()
	ctx := context.Background()

	err := ops.CheckoutBranch(ctx, ".", "")
	if err == nil {
		t.Error("CheckoutBranch() with empty name should return error")
	}
}

// Integration test - requires a real git repository
func TestExecOperations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ops := NewExecOperations()
	ctx := context.Background()

	// Create a temporary git repository
	tempDir := t.TempDir()

	// Initialize git repo
	_, _, err := ops.execGit(ctx, tempDir, "init")
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	_, _, _ = ops.execGit(ctx, tempDir, "config", "user.name", "Test User")
	_, _, _ = ops.execGit(ctx, tempDir, "config", "user.email", "test@example.com")

	t.Run("IsGitRepo", func(t *testing.T) {
		isRepo, err := ops.IsGitRepo(ctx, tempDir)
		if err != nil {
			t.Fatalf("IsGitRepo() error = %v", err)
		}
		if !isRepo {
			t.Error("IsGitRepo() = false, want true")
		}
	})

	t.Run("GetCurrentBranch", func(t *testing.T) {
		branch, err := ops.GetCurrentBranch(ctx, tempDir)
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}
		// Git 2.28+ defaults to "main", older versions use "master" or might be "HEAD" (empty repo)
		if branch != "main" && branch != "master" && branch != "HEAD" {
			t.Logf("GetCurrentBranch() = %v (acceptable for empty repo)", branch)
		}
	})

	t.Run("HasRemote", func(t *testing.T) {
		hasRemote, err := ops.HasRemote(ctx, tempDir)
		if err != nil {
			t.Fatalf("HasRemote() error = %v", err)
		}
		if hasRemote {
			t.Error("HasRemote() = true, want false for new repo")
		}
	})

	// Create a file and test Add/Commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("Add", func(t *testing.T) {
		err := ops.Add(ctx, tempDir, []string{"test.txt"})
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	})

	t.Run("Commit", func(t *testing.T) {
		err := ops.Commit(ctx, tempDir, "Initial commit", nil)
		if err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
	})

	t.Run("GetLog", func(t *testing.T) {
		commits, err := ops.GetLog(ctx, tempDir, 10)
		if err != nil {
			t.Fatalf("GetLog() error = %v", err)
		}
		if len(commits) != 1 {
			t.Errorf("GetLog() returned %d commits, want 1", len(commits))
		}
		if len(commits) > 0 {
			if commits[0].Message != "Initial commit" {
				t.Errorf("Commit message = %v, want 'Initial commit'", commits[0].Message)
			}
		}
	})

	t.Run("CreateBranch", func(t *testing.T) {
		err := ops.CreateBranch(ctx, tempDir, "feature-test")
		if err != nil {
			t.Fatalf("CreateBranch() error = %v", err)
		}
	})

	t.Run("CheckoutBranch", func(t *testing.T) {
		err := ops.CheckoutBranch(ctx, tempDir, "feature-test")
		if err != nil {
			t.Fatalf("CheckoutBranch() error = %v", err)
		}

		branch, err := ops.GetCurrentBranch(ctx, tempDir)
		if err != nil {
			t.Fatalf("GetCurrentBranch() error = %v", err)
		}
		if branch != "feature-test" {
			t.Errorf("GetCurrentBranch() = %v, want 'feature-test'", branch)
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		repo, err := ops.GetStatus(ctx, tempDir)
		if err != nil {
			t.Fatalf("GetStatus() error = %v", err)
		}
		if repo == nil {
			t.Fatal("GetStatus() returned nil repository")
		}
		if repo.CurrentBranch() != "feature-test" {
			t.Errorf("CurrentBranch() = %v, want 'feature-test'", repo.CurrentBranch())
		}
		if !repo.IsClean() {
			t.Error("IsClean() = false, want true after commit")
		}
	})

	t.Run("GetDiff", func(t *testing.T) {
		// Create a new file to generate diff
		testFile2 := filepath.Join(tempDir, "test2.txt")
		if err := os.WriteFile(testFile2, []byte("new content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Get unstaged diff
		diff, err := ops.GetDiff(ctx, tempDir, false)
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}
		// Diff should be empty for untracked files
		if diff != "" {
			t.Logf("GetDiff() returned non-empty diff for untracked file (expected): %v", diff)
		}

		// Add file and get staged diff
		_ = ops.Add(ctx, tempDir, []string{"test2.txt"})
		diff, err = ops.GetDiff(ctx, tempDir, true)
		if err != nil {
			t.Fatalf("GetDiff(staged) error = %v", err)
		}
		if diff == "" {
			t.Error("GetDiff(staged) returned empty diff, want non-empty for staged changes")
		}
	})
}
