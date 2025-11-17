package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/git"
	"github.com/yourusername/gitman/internal/domain"
)

// ShouldRunOnboarding determines if onboarding should run
func ShouldRunOnboarding(cfg *domain.Config, gitOps git.Operations, repoPath string) bool {
	// Check if API key is configured
	if cfg.AI.APIKey == "" {
		return true
	}

	// Check if git repository exists
	ctx := context.Background()
	isRepo, _ := gitOps.IsGitRepo(ctx, repoPath)
	if !isRepo {
		return true
	}

	// If config has defaults (newly created), suggest onboarding
	if cfg.Version == "" {
		return true
	}

	return false
}

// RunOnboarding launches the onboarding wizard
func RunOnboarding(gitOps git.Operations, cfg *domain.Config, cfgManager *config.Manager, repoPath string) error {
	model := NewAppModelWithOnboarding(gitOps, cfg, cfgManager, repoPath)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
