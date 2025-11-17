package github

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CreateRepoOptions contains options for creating a GitHub repository
type CreateRepoOptions struct {
	Name        string
	Description string
	Visibility  string // "public" or "private"
	License     string
	GitIgnore   string
	AddReadme   bool
	EnableIssues   bool
	EnableWiki     bool
	EnableProjects bool
}

// CheckGHAvailable checks if gh CLI is installed
func CheckGHAvailable() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// CheckGHAuthenticated checks if gh is authenticated
func CheckGHAuthenticated(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	err := cmd.Run()
	if err != nil {
		return false, nil // Not authenticated, but not an error
	}
	return true, nil
}

// CreateRepository creates a new GitHub repository using gh CLI
func CreateRepository(ctx context.Context, opts CreateRepoOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("repository name is required")
	}

	// Build command arguments
	args := []string{"repo", "create", opts.Name}

	// Visibility
	if opts.Visibility == "private" {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	// Description
	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}

	// License
	if opts.License != "" && opts.License != "None" {
		args = append(args, "--license", opts.License)
	}

	// GitIgnore template
	if opts.GitIgnore != "" && opts.GitIgnore != "None" {
		args = append(args, "--gitignore", opts.GitIgnore)
	}

	// README
	if opts.AddReadme {
		args = append(args, "--add-readme")
	}

	// Issues
	if !opts.EnableIssues {
		args = append(args, "--disable-issues")
	}

	// Wiki
	if !opts.EnableWiki {
		args = append(args, "--disable-wiki")
	}

	// Note: gh CLI doesn't support --enable/disable-projects in create command
	// Projects would need to be configured separately via web UI or API

	// Execute command
	cmd := exec.CommandContext(ctx, "gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create repository: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetGitIgnoreTemplates returns available .gitignore templates
// Note: This is a static list. In production, you might fetch from GitHub API
func GetGitIgnoreTemplates() []string {
	return []string{
		"None",
		"Go",
		"Node",
		"Python",
		"Java",
		"Rust",
		"C",
		"C++",
		"Ruby",
		"PHP",
		"Swift",
		"Kotlin",
		"VisualStudio",
		"JetBrains",
	}
}

// GetLicenseTemplates returns available license templates
func GetLicenseTemplates() []string {
	return []string{
		"None",
		"MIT",
		"Apache-2.0",
		"GPL-3.0",
		"BSD-3-Clause",
		"BSD-2-Clause",
		"ISC",
		"MPL-2.0",
		"LGPL-3.0",
		"AGPL-3.0",
	}
}

// AuthenticateGH launches gh auth login flow
func AuthenticateGH(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gh", "auth", "login")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// SetRemote sets the git remote origin to the GitHub repository
func SetRemote(ctx context.Context, repoPath, repoURL string) error {
	// Check if remote already exists
	checkCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "get-url", "origin")
	err := checkCmd.Run()

	if err == nil {
		// Remote exists, update it
		cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "set-url", "origin", repoURL)
		return cmd.Run()
	}

	// Remote doesn't exist, add it
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "add", "origin", repoURL)
	return cmd.Run()
}

// GetRepoURL extracts the repository URL from gh create output or constructs it
func GetRepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// GetCurrentUser returns the authenticated GitHub username
func GetCurrentUser(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	username := strings.TrimSpace(string(output))
	return username, nil
}
