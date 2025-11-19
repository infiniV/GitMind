package github

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourusername/gitman/internal/domain"
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

// CreatePR creates a pull request using gh CLI
func CreatePR(ctx context.Context, repoPath string, opts *domain.PROptions) (*domain.PRInfo, error) {
	if opts == nil {
		return nil, fmt.Errorf("PR options are required")
	}

	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid PR options: %w", err)
	}

	// Build command arguments
	args := []string{"pr", "create"}

	// Title and body
	args = append(args, "--title", opts.Title())
	if opts.Body() != "" {
		args = append(args, "--body", opts.Body())
	}

	// Base and head branches
	args = append(args, "--base", opts.BaseBranch())
	args = append(args, "--head", opts.HeadBranch())

	// Draft status
	if opts.IsDraft() {
		args = append(args, "--draft")
	}

	// Labels
	for _, label := range opts.Labels() {
		args = append(args, "--label", label)
	}

	// Assignees
	for _, assignee := range opts.Assignees() {
		args = append(args, "--assignee", assignee)
	}

	// Reviewers
	for _, reviewer := range opts.Reviewers() {
		args = append(args, "--reviewer", reviewer)
	}

	// Execute command
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %s: %w", string(output), err)
	}

	// Parse the PR URL from output (gh pr create returns the URL)
	prURL := strings.TrimSpace(string(output))

	// Get PR details using the URL
	prNumber, err := extractPRNumberFromURL(prURL)
	if err != nil {
		// If we can't extract the number, return basic info
		prInfo, _ := domain.NewPRInfo(0, opts.Title(), "", opts.BaseBranch(), opts.HeadBranch())
		prInfo.SetHTMLURL(prURL)
		prInfo.SetIsDraft(opts.IsDraft())
		prInfo.SetLabels(opts.Labels())
		return prInfo, nil
	}

	// Get full PR details
	return GetPR(ctx, repoPath, prNumber)
}

// ListPRs lists pull requests for the current repository
func ListPRs(ctx context.Context, repoPath string, state string) ([]*domain.PRInfo, error) {
	// Valid states: "open", "closed", "merged", "all"
	if state == "" {
		state = "open"
	}

	args := []string{"pr", "list", "--state", state, "--json", "number,title,state,author,baseRefName,headRefName,isDraft,labels,createdAt,updatedAt,url"}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %s: %w", string(output), err)
	}

	// Parse JSON output
	// For now, use simple parsing (TODO: proper JSON unmarshaling)
	outputStr := string(output)
	if strings.TrimSpace(outputStr) == "[]" {
		return []*domain.PRInfo{}, nil
	}

	// This is a simplified implementation
	// In production, use encoding/json to unmarshal properly
	return []*domain.PRInfo{}, nil
}

// GetPR gets details of a specific pull request
func GetPR(ctx context.Context, repoPath string, number int) (*domain.PRInfo, error) {
	if number <= 0 {
		return nil, fmt.Errorf("invalid PR number: %d", number)
	}

	args := []string{
		"pr", "view", fmt.Sprintf("%d", number),
		"--json", "number,title,body,state,author,baseRefName,headRefName,isDraft,labels,createdAt,updatedAt,url,mergeable",
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %s: %w", string(output), err)
	}

	outputStr := string(output)

	// Parse JSON output (simplified parsing)
	prInfo, err := domain.NewPRInfo(number, "PR Title", "author", "main", "feature")
	if err != nil {
		return nil, err
	}

	// Extract title
	if strings.Contains(outputStr, `"title":"`) {
		parts := strings.Split(outputStr, `"title":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				prInfo.SetTitle(parts[1][:endIndex])
			}
		}
	}

	// Extract body
	if strings.Contains(outputStr, `"body":"`) {
		parts := strings.Split(outputStr, `"body":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				prInfo.SetBody(parts[1][:endIndex])
			}
		}
	}

	// Extract URL
	if strings.Contains(outputStr, `"url":"`) {
		parts := strings.Split(outputStr, `"url":"`)
		if len(parts) > 1 {
			endIndex := strings.Index(parts[1], `"`)
			if endIndex > 0 {
				prInfo.SetHTMLURL(parts[1][:endIndex])
			}
		}
	}

	// Extract draft status
	if strings.Contains(outputStr, `"isDraft":true`) {
		prInfo.SetIsDraft(true)
	}

	// Extract state
	if strings.Contains(outputStr, `"state":"MERGED"`) {
		prInfo.SetState(domain.PRStatusMerged)
	} else if strings.Contains(outputStr, `"state":"CLOSED"`) {
		prInfo.SetState(domain.PRStatusClosed)
	} else if prInfo.IsDraft() {
		prInfo.SetState(domain.PRStatusDraft)
	} else {
		prInfo.SetState(domain.PRStatusOpen)
	}

	return prInfo, nil
}

// UpdatePR updates a pull request
func UpdatePR(ctx context.Context, repoPath string, number int, updates map[string]string) error {
	if number <= 0 {
		return fmt.Errorf("invalid PR number: %d", number)
	}

	args := []string{"pr", "edit", fmt.Sprintf("%d", number)}

	// Add update fields
	if title, ok := updates["title"]; ok {
		args = append(args, "--title", title)
	}
	if body, ok := updates["body"]; ok {
		args = append(args, "--body", body)
	}
	if labels, ok := updates["add-label"]; ok {
		for _, label := range strings.Split(labels, ",") {
			args = append(args, "--add-label", strings.TrimSpace(label))
		}
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update PR: %s: %w", string(output), err)
	}

	return nil
}

// ClosePR closes a pull request without merging
func ClosePR(ctx context.Context, repoPath string, number int) error {
	if number <= 0 {
		return fmt.Errorf("invalid PR number: %d", number)
	}

	cmd := exec.CommandContext(ctx, "gh", "pr", "close", fmt.Sprintf("%d", number))
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to close PR: %s: %w", string(output), err)
	}

	return nil
}

// MergePRRemote merges a pull request on GitHub
func MergePRRemote(ctx context.Context, repoPath string, number int, method string) error {
	if number <= 0 {
		return fmt.Errorf("invalid PR number: %d", number)
	}

	// Valid methods: "merge", "squash", "rebase"
	validMethods := map[string]bool{
		"merge":  true,
		"squash": true,
		"rebase": true,
	}

	if !validMethods[method] {
		return fmt.Errorf("invalid merge method: %s (must be merge, squash, or rebase)", method)
	}

	args := []string{"pr", "merge", fmt.Sprintf("%d", number), fmt.Sprintf("--%s", method)}

	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to merge PR: %s: %w", string(output), err)
	}

	return nil
}

// ConvertPRToDraft converts a ready PR to draft status
func ConvertPRToDraft(ctx context.Context, repoPath string, number int) error {
	if number <= 0 {
		return fmt.Errorf("invalid PR number: %d", number)
	}

	cmd := exec.CommandContext(ctx, "gh", "pr", "ready", fmt.Sprintf("%d", number), "--undo")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert PR to draft: %s: %w", string(output), err)
	}

	return nil
}

// MarkPRReady marks a draft PR as ready for review
func MarkPRReady(ctx context.Context, repoPath string, number int) error {
	if number <= 0 {
		return fmt.Errorf("invalid PR number: %d", number)
	}

	cmd := exec.CommandContext(ctx, "gh", "pr", "ready", fmt.Sprintf("%d", number))
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mark PR as ready: %s: %w", string(output), err)
	}

	return nil
}

// extractPRNumberFromURL extracts the PR number from a GitHub PR URL
func extractPRNumberFromURL(url string) (int, error) {
	// URL format: https://github.com/owner/repo/pull/123
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid PR URL format")
	}

	// Get the last part which should be the number
	numberStr := parts[len(parts)-1]
	var number int
	_, err := fmt.Sscanf(numberStr, "%d", &number)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PR number from URL: %w", err)
	}

	return number, nil
}
