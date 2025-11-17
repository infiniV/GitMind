package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/gitman/internal/domain"
)

// buildUntrackedFilesDiff creates a diff-like representation of untracked files
// by reading their content directly from the filesystem.
// This avoids staging files before the user makes a decision.
func (uc *AnalyzeCommitUseCase) buildUntrackedFilesDiff(repoPath string, repo *domain.Repository) (string, error) {
	var sb strings.Builder

	sb.WriteString("New files to be added:\n\n")

	for _, change := range repo.Changes() {
		// Only process untracked or new files
		if change.Status != domain.StatusUntracked && change.Status != domain.StatusAdded {
			continue
		}

		filePath := filepath.Join(repoPath, change.Path)

		// Check if file exists
		info, err := os.Stat(filePath)
		if err != nil {
			// Skip files we can't read
			sb.WriteString(fmt.Sprintf("--- %s (unreadable)\n\n", change.Path))
			continue
		}

		// Skip directories
		if info.IsDir() {
			sb.WriteString(fmt.Sprintf("--- %s/ (directory)\n\n", change.Path))
			continue
		}

		// Skip very large files (> 100KB)
		if info.Size() > 100*1024 {
			sb.WriteString(fmt.Sprintf("--- %s (large file: %d bytes)\n\n", change.Path, info.Size()))
			continue
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- %s (read error)\n\n", change.Path))
			continue
		}

		// Check if binary
		if isBinary(content) {
			sb.WriteString(fmt.Sprintf("--- %s (binary file)\n\n", change.Path))
			continue
		}

		// Add diff-like output
		sb.WriteString("--- /dev/null\n")
		sb.WriteString(fmt.Sprintf("+++ %s\n", change.Path))

		// Add file content with + prefix (like a diff)
		lines := strings.Split(string(content), "\n")

		// Limit to first 100 lines to avoid token explosion
		maxLines := 100
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			sb.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@ (truncated, showing first %d lines)\n", len(lines), maxLines))
		} else {
			sb.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
		}

		for _, line := range lines {
			sb.WriteString("+")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if sb.Len() == len("New files to be added:\n\n") {
		return "", fmt.Errorf("no readable files found")
	}

	return sb.String(), nil
}

// isBinary checks if content appears to be binary
func isBinary(content []byte) bool {
	// Check first 8KB for null bytes
	sampleSize := 8192
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	for i := 0; i < sampleSize; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

