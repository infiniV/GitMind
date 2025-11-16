package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/gitman/internal/domain"
)

// Config represents the application configuration.
type Config struct {
	AIProvider             string   `json:"ai_provider"`
	APIKey                 string   `json:"api_key"`
	APITier                string   `json:"api_tier"`
	UseConventionalCommits bool     `json:"use_conventional_commits"`
	DefaultModel           string   `json:"default_model"`
	ProtectedBranches      []string `json:"protected_branches"`
	DefaultMergeStrategy   string   `json:"default_merge_strategy"`
}

// Manager handles configuration persistence.
type Manager struct {
	configPath string
}

// NewManager creates a new config manager.
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".gitman.json")
	return &Manager{
		configPath: configPath,
	}, nil
}

// Load loads the configuration from disk.
func (m *Manager) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Return default config
		return &Config{
			AIProvider:             "cerebras",
			UseConventionalCommits: false,
			DefaultModel:           "llama-3.3-70b",
			APITier:                "free",
			ProtectedBranches:      []string{"main", "master", "develop"},
			DefaultMergeStrategy:   "ask",
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON - simple manual parsing since we only have a few fields
	config := &Config{}

	// For MVP, we'll use a simple format
	// In production, use encoding/json
	lines := string(data)
	if len(lines) == 0 {
		return config, nil
	}

	// Simple key=value parsing (JSON parsing would be better but this is quicker for MVP)
	return parseSimpleConfig(lines)
}

// Save saves the configuration to disk.
func (m *Manager) Save(config *Config) error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Format config as simple key=value
	protectedBranches := joinStrings(config.ProtectedBranches, ",")
	if protectedBranches == "" {
		protectedBranches = "main,master,develop"
	}

	mergeStrategy := config.DefaultMergeStrategy
	if mergeStrategy == "" {
		mergeStrategy = "ask"
	}

	content := fmt.Sprintf(`ai_provider=%s
api_key=%s
api_tier=%s
use_conventional_commits=%v
default_model=%s
protected_branches=%s
default_merge_strategy=%s
`, config.AIProvider, config.APIKey, config.APITier, config.UseConventionalCommits, config.DefaultModel, protectedBranches, mergeStrategy)

	// Write config file
	if err := os.WriteFile(m.configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the configured API key as a domain object.
func (m *Manager) GetAPIKey(config *Config) (*domain.APIKey, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key not configured. Run 'gm config' to set up")
	}

	apiKey, err := domain.NewAPIKey(config.APIKey, config.AIProvider)
	if err != nil {
		return nil, err
	}

	// Set tier from config
	tier, err := domain.ParseAPITier(config.APITier)
	if err != nil {
		tier = domain.TierFree // Default to free
	}
	apiKey.SetTier(tier)

	return apiKey, nil
}

// ConfigPath returns the path to the config file.
func (m *Manager) ConfigPath() string {
	return m.configPath
}

// parseSimpleConfig parses a simple key=value config format.
func parseSimpleConfig(content string) (*Config, error) {
	config := &Config{
		AIProvider:             "cerebras",
		UseConventionalCommits: false,
		DefaultModel:           "llama-3.3-70b",
		APITier:                "free",
		ProtectedBranches:      []string{"main", "master", "develop"},
		DefaultMergeStrategy:   "ask",
	}

	lines := splitLines(content)
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}

		parts := splitKeyValue(line)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "ai_provider":
			config.AIProvider = value
		case "api_key":
			config.APIKey = value
		case "api_tier":
			config.APITier = value
		case "use_conventional_commits":
			config.UseConventionalCommits = value == "true"
		case "default_model":
			config.DefaultModel = value
		case "protected_branches":
			config.ProtectedBranches = splitString(value, ",")
		case "default_merge_strategy":
			config.DefaultMergeStrategy = value
		}
	}

	return config, nil
}

func splitLines(s string) []string {
	var lines []string
	var current string

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func splitKeyValue(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// splitString splits a string by delimiter
func splitString(s, delim string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	var current string

	for i := 0; i < len(s); i++ {
		if string(s[i]) == delim {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(s[i])
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// joinStrings joins a slice of strings with delimiter
func joinStrings(strs []string, delim string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += delim + strs[i]
	}

	return result
}
