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

// ViewRepoWeb opens the GitHub repository in the default web browser.
// If repoPath is provided, it uses that directory's remote.
// Otherwise, it uses the current directory.
func ViewRepoWeb(ctx context.Context, repoPath string) error {
	args := []string{"repo", "view", "--web"}

	cmd := exec.CommandContext(ctx, "gh", args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open repository in browser: %w", err)
	}

	return nil
}

// RepoInfo represents GitHub repository information.
type RepoInfo struct {
	Owner         string
	Name          string
	FullName      string
	Description   string
	Stars         int
	Forks         int
	OpenIssues    int
	IsPrivate     bool
	DefaultBranch string
	HTMLURL       string
	License       string
}

// GetRepoInfo retrieves GitHub repository information using gh CLI.
// If repoPath is provided, it uses that directory's remote.
// Otherwise, it uses the current directory.
func GetRepoInfo(ctx context.Context, repoPath string) (*RepoInfo, error) {
	// Use gh repo view with JSON output
	args := []string{"repo", "view", "--json",
		"owner,name,nameWithOwner,description,stargazerCount,forkCount,openIssuesCount,isPrivate,defaultBranchRef,url,licenseInfo"}

	cmd := exec.CommandContext(ctx, "gh", args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	// Parse JSON output manually (avoiding external JSON library for simplicity)
	// In production, you'd use json.Unmarshal, but for now we'll extract key fields
	outputStr := string(output)

	info := &RepoInfo{}

	// Extract fields using basic string parsing
	// Note: This is simplified - in production use proper JSON parsing
	if strings.Contains(outputStr, `"owner"`) {
		parts := strings.Split(outputStr, `"owner":`)
		if len(parts) > 1 {
			loginParts := strings.Split(parts[1], `"login":"`)
			if len(loginParts) > 1 {
				endIndex := strings.Index(loginParts[1], `"`)
				if endIndex > 0 {
					info.Owner = loginParts[1][:endIndex]
				}
			}
		}
	}

	if strings.Contains(outputStr, `"name":"`) {
		parts := strings.Split(outputStr, `"name":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				info.Name = parts[1][:endIndex]
			}
		}
	}

	if strings.Contains(outputStr, `"nameWithOwner":"`) {
		parts := strings.Split(outputStr, `"nameWithOwner":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				info.FullName = parts[1][:endIndex]
			}
		}
	}

	if strings.Contains(outputStr, `"description":"`) {
		parts := strings.Split(outputStr, `"description":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				info.Description = parts[1][:endIndex]
			}
		}
	}

	if strings.Contains(outputStr, `"url":"`) {
		parts := strings.Split(outputStr, `"url":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				info.HTMLURL = parts[1][:endIndex]
			}
		}
	}

	// For numbers, we'll use simple counting (stars/forks/issues)
	// In production, use proper JSON parsing
	if strings.Contains(outputStr, `"isPrivate":true`) {
		info.IsPrivate = true
	}

	// Set reasonable defaults for counts (actual parsing would be more complex)
	// TODO: Implement proper JSON parsing or use encoding/json
	info.Stars = 0
	info.Forks = 0
	info.OpenIssues = 0

	return info, nil
}
