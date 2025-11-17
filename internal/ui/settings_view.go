package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/domain"
)

// SettingsView represents the settings tab view
type SettingsView struct {
	cfg        *domain.Config
	cfgManager *config.Manager

	// Will be used later for nested tabs
	currentTab int // 0=Git, 1=GitHub, 2=Commits, 3=Naming, 4=AI
}

// NewSettingsView creates a new settings view
func NewSettingsView(cfg *domain.Config, cfgManager *config.Manager) *SettingsView {
	return &SettingsView{
		cfg:        cfg,
		cfgManager: cfgManager,
		currentTab: 0,
	}
}

// Init initializes the settings view
func (m SettingsView) Init() tea.Cmd {
	return nil
}

// Update handles messages for the settings view
func (m SettingsView) Update(msg tea.Msg) (SettingsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			// Future: Switch between nested tabs
			// For now, do nothing
		}
	}

	return m, nil
}

// View renders the settings view
func (m SettingsView) View() string {
	var sections []string

	// Header
	header := headerStyle.Render("GitMind - Settings")
	sections = append(sections, header)

	// Placeholder content
	content := lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(2).
		Render("Settings interface coming soon...\n\n" +
			"This will include:\n" +
			"  - Git configuration\n" +
			"  - GitHub integration\n" +
			"  - Commit conventions\n" +
			"  - Branch naming patterns\n" +
			"  - AI provider settings\n\n" +
			"Press 1 to return to Dashboard")

	sections = append(sections, content)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
