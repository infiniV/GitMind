package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/gitman/internal/domain"
)

// Legacy Config structure for migration
type LegacyConfig struct {
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

// Load loads the configuration from disk with automatic migration.
func (m *Manager) Load() (*domain.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Return default config
		return domain.NewDefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Try parsing as new JSON format first
	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		// Successfully parsed as new format
		return &cfg, nil
	}

	// Try parsing as old key=value format
	oldCfg, err := parseSimpleConfig(string(data))
	if err == nil {
		// Successfully parsed as old format, migrate
		newCfg := m.migrateFromLegacy(oldCfg)

		// Backup old config
		backupPath := m.configPath + ".backup"
		if err := os.WriteFile(backupPath, data, 0600); err == nil {
			fmt.Printf("Backed up old config to: %s\n", backupPath)
		}

		// Save new format
		if err := m.Save(newCfg); err != nil {
			return nil, fmt.Errorf("failed to save migrated config: %w", err)
		}

		fmt.Println("Configuration migrated to new format")
		return newCfg, nil
	}

	// Failed to parse both formats
	return nil, fmt.Errorf("failed to parse config file (tried both new and old formats)")
}

// Save saves the configuration to disk in JSON format.
func (m *Manager) Save(config *domain.Config) error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the configured API key as a domain object.
func (m *Manager) GetAPIKey(config *domain.Config) (*domain.APIKey, error) {
	if config.AI.APIKey == "" {
		return nil, fmt.Errorf("API key not configured. Run 'gm config' or 'gm onboard' to set up")
	}

	apiKey, err := domain.NewAPIKey(config.AI.APIKey, config.AI.Provider)
	if err != nil {
		return nil, err
	}

	// Set tier from config
	tier, err := domain.ParseAPITier(config.AI.APITier)
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

// migrateFromLegacy converts old config format to new domain.Config
func (m *Manager) migrateFromLegacy(old *LegacyConfig) *domain.Config {
	cfg := domain.NewDefaultConfig()

	// Migrate AI settings
	cfg.AI.Provider = old.AIProvider
	if cfg.AI.Provider == "" {
		cfg.AI.Provider = "cerebras"
	}
	cfg.AI.APIKey = old.APIKey
	cfg.AI.APITier = old.APITier
	if cfg.AI.APITier == "" {
		cfg.AI.APITier = "free"
	}
	cfg.AI.DefaultModel = old.DefaultModel
	if cfg.AI.DefaultModel == "" {
		cfg.AI.DefaultModel = "llama-3.3-70b"
	}

	// Migrate git settings
	if len(old.ProtectedBranches) > 0 {
		cfg.Git.ProtectedBranches = old.ProtectedBranches
	}

	// Migrate commit settings
	if old.UseConventionalCommits {
		cfg.Commits.Convention = "conventional"
	} else {
		cfg.Commits.Convention = "none"
	}

	return cfg
}

// parseSimpleConfig parses the legacy key=value config format.
func parseSimpleConfig(content string) (*LegacyConfig, error) {
	config := &LegacyConfig{
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
