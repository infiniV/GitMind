package domain

import (
	"fmt"
)

// Config represents the complete GitMind configuration
type Config struct {
	Version string        `json:"version"`
	Git     GitConfig     `json:"git"`
	GitHub  GitHubConfig  `json:"github"`
	Commits CommitsConfig `json:"commits"`
	Naming  NamingConfig  `json:"naming"`
	AI      AIConfig      `json:"ai"`
	UI      UIConfig      `json:"ui"`
}

// GitConfig holds git-related configuration
type GitConfig struct {
	MainBranch        string   `json:"main_branch"`
	ProtectedBranches []string `json:"protected_branches"`
	AutoPush          bool     `json:"auto_push"`
	AutoPull          bool     `json:"auto_pull"`
}

// GitHubConfig holds GitHub integration settings
type GitHubConfig struct {
	Enabled           bool     `json:"enabled"`
	DefaultVisibility string   `json:"default_visibility"` // "public" or "private"
	DefaultLicense    string   `json:"default_license"`
	DefaultGitIgnore  string   `json:"default_gitignore"`
	EnableIssues      bool     `json:"enable_issues"`
	EnableWiki        bool     `json:"enable_wiki"`
	EnableProjects    bool     `json:"enable_projects"`
	// PR Configuration
	PRDefaultBase      string   `json:"pr_default_base"`       // Default base branch for PRs
	PRUseTemplate      bool     `json:"pr_use_template"`       // Load .github/PULL_REQUEST_TEMPLATE.md
	PRDefaultDraft     bool     `json:"pr_default_draft"`      // Create PRs as draft by default
	PRDefaultLabels    []string `json:"pr_default_labels"`     // Auto-apply labels to new PRs
	PRAutoDeleteBranch bool     `json:"pr_auto_delete_branch"` // Delete branch after PR merge
}

// CommitsConfig holds commit convention settings
type CommitsConfig struct {
	Convention      string   `json:"convention"`       // "conventional", "custom", or "none"
	Types           []string `json:"types"`            // Allowed commit types
	RequireScope    bool     `json:"require_scope"`    // Require scope in conventional commits
	RequireBreaking bool     `json:"require_breaking"` // Require breaking change marker
	CustomTemplate  string   `json:"custom_template"`  // Custom commit template
}

// NamingConfig holds branch naming convention settings
type NamingConfig struct {
	Enforce         bool     `json:"enforce"`
	Pattern         string   `json:"pattern"` // e.g., "feature/{description}"
	AllowedPrefixes []string `json:"allowed_prefixes"`
}

// AIConfig holds AI provider settings
type AIConfig struct {
	Provider       string `json:"provider"`
	APIKey         string `json:"api_key"`
	APITier        string `json:"api_tier"`
	DefaultModel   string `json:"default_model"`
	FallbackModel  string `json:"fallback_model"`
	MaxDiffSize    int    `json:"max_diff_size"`
	IncludeContext bool   `json:"include_context"`
}

// UIConfig holds UI/theme settings
type UIConfig struct {
	Theme string `json:"theme"` // Theme name (e.g., "claude-warm", "ocean-blue")
}

// NewDefaultConfig creates a new config with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		Version: "2.0",
		Git: GitConfig{
			MainBranch:        "main",
			ProtectedBranches: []string{"main", "master", "develop"},
			AutoPush:          false,
			AutoPull:          false,
		},
		GitHub: GitHubConfig{
			Enabled:            false,
			DefaultVisibility:  "public",
			DefaultLicense:     "MIT",
			DefaultGitIgnore:   "Go",
			EnableIssues:       true,
			EnableWiki:         false,
			EnableProjects:     false,
			PRDefaultBase:      "main",
			PRUseTemplate:      true,
			PRDefaultDraft:     false,
			PRDefaultLabels:    []string{},
			PRAutoDeleteBranch: false,
		},
		Commits: CommitsConfig{
			Convention:      "conventional",
			Types:           []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
			RequireScope:    false,
			RequireBreaking: false,
			CustomTemplate:  "",
		},
		Naming: NamingConfig{
			Enforce:         false,
			Pattern:         "feature/{description}",
			AllowedPrefixes: []string{"feature", "hotfix", "bugfix", "release", "refactor"},
		},
		AI: AIConfig{
			Provider:       "cerebras",
			APIKey:         "",
			APITier:        "free",
			DefaultModel:   "llama-3.3-70b",
			FallbackModel:  "llama3.1-8b",
			MaxDiffSize:    100000,
			IncludeContext: true,
		},
		UI: UIConfig{
			Theme: "claude-warm",
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Git config
	if c.Git.MainBranch == "" {
		return fmt.Errorf("git.main_branch cannot be empty")
	}

	// Validate GitHub config
	if c.GitHub.Enabled {
		if c.GitHub.DefaultVisibility != "public" && c.GitHub.DefaultVisibility != "private" {
			return fmt.Errorf("github.default_visibility must be 'public' or 'private'")
		}
	}

	// Validate Commits config
	if c.Commits.Convention != "conventional" && c.Commits.Convention != "custom" && c.Commits.Convention != "none" {
		return fmt.Errorf("commits.convention must be 'conventional', 'custom', or 'none'")
	}
	if c.Commits.Convention == "conventional" && len(c.Commits.Types) == 0 {
		return fmt.Errorf("commits.types cannot be empty when using conventional commits")
	}
	if c.Commits.Convention == "custom" && c.Commits.CustomTemplate == "" {
		return fmt.Errorf("commits.custom_template cannot be empty when using custom convention")
	}

	// Validate AI config
	if c.AI.Provider == "" {
		return fmt.Errorf("ai.provider cannot be empty")
	}
	if c.AI.APIKey == "" {
		return fmt.Errorf("ai.api_key cannot be empty")
	}
	if c.AI.DefaultModel == "" {
		return fmt.Errorf("ai.default_model cannot be empty")
	}

	return nil
}

// GetProtectedBranches returns the list of protected branches
func (c *Config) GetProtectedBranches() []string {
	return c.Git.ProtectedBranches
}

// IsProtectedBranch checks if a branch is protected
func (c *Config) IsProtectedBranch(branch string) bool {
	for _, protected := range c.Git.ProtectedBranches {
		if protected == branch {
			return true
		}
	}
	return false
}

// GetCommitTypes returns the allowed commit types
func (c *Config) GetCommitTypes() []string {
	return c.Commits.Types
}

// IsValidCommitType checks if a commit type is allowed
func (c *Config) IsValidCommitType(commitType string) bool {
	for _, allowed := range c.Commits.Types {
		if allowed == commitType {
			return true
		}
	}
	return false
}

// IsValidBranchPrefix checks if a branch prefix is allowed
func (c *Config) IsValidBranchPrefix(prefix string) bool {
	if !c.Naming.Enforce {
		return true // Not enforcing, so all prefixes are valid
	}

	for _, allowed := range c.Naming.AllowedPrefixes {
		if allowed == prefix {
			return true
		}
	}
	return false
}
