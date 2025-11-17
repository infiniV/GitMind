package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/adapter/config"
	"github.com/yourusername/gitman/internal/adapter/github"
	"github.com/yourusername/gitman/internal/domain"
)

// SettingsTab represents a settings category tab
type SettingsTab int

const (
	SettingsGit SettingsTab = iota
	SettingsGitHub
	SettingsCommits
	SettingsNaming
	SettingsAI
	SettingsUI
)

// SettingsView represents the settings tab view
type SettingsView struct {
	cfg        *domain.Config
	cfgManager *config.Manager

	currentTab   SettingsTab
	focusedField int

	// Git settings fields
	gitMainBranch       TextInput
	gitProtectedBranches CheckboxGroup
	gitCustomProtected  TextInput
	gitAutoPush         Checkbox
	gitAutoPull         Checkbox

	// GitHub settings fields
	ghEnabled           Checkbox
	ghDefaultVisibility RadioGroup
	ghDefaultLicense    Dropdown
	ghDefaultGitIgnore  Dropdown
	ghEnableIssues      Checkbox
	ghEnableWiki        Checkbox
	ghEnableProjects    Checkbox

	// Commits settings fields
	commitConvention      RadioGroup
	commitTypes           CheckboxGroup
	commitRequireScope    Checkbox
	commitRequireBreaking Checkbox
	commitCustomTemplate  TextInput

	// Naming settings fields
	namingEnforce        Checkbox
	namingPattern        TextInput
	namingAllowedPrefixes CheckboxGroup
	namingCustomPrefix   TextInput

	// AI settings fields
	aiProvider       Dropdown
	aiAPIKey         TextInput
	aiAPITier        RadioGroup
	aiDefaultModel   Dropdown
	aiFallbackModel  Dropdown
	aiMaxDiffSize    TextInput
	aiIncludeContext Checkbox

	// UI settings fields
	uiTheme         Dropdown
	originalTheme   string // Track original theme for preview/revert

	// State
	hasChanges bool
	saveStatus string
}

// NewSettingsView creates a new settings view
func NewSettingsView(cfg *domain.Config, cfgManager *config.Manager) *SettingsView {
	// Initialize Git fields
	protectedBranches := []string{"main", "master", "develop", "production"}
	protectedChecked := make([]bool, len(protectedBranches))
	for i, branch := range protectedBranches {
		for _, protected := range cfg.Git.ProtectedBranches {
			if branch == protected {
				protectedChecked[i] = true
				break
			}
		}
	}

	// Initialize Commits fields
	commitTypes := []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"}
	commitTypesChecked := make([]bool, len(commitTypes))
	if cfg.Commits.Convention == "conventional" {
		for i, cType := range commitTypes {
			for _, enabled := range cfg.Commits.Types {
				if cType == enabled {
					commitTypesChecked[i] = true
					break
				}
			}
		}
	} else {
		for i := range commitTypesChecked {
			commitTypesChecked[i] = true
		}
	}

	conventionIdx := 0
	switch cfg.Commits.Convention {
	case "custom":
		conventionIdx = 1
	case "none":
		conventionIdx = 2
	}

	// Initialize Naming fields
	allowedPrefixes := []string{"feature", "bugfix", "hotfix", "release", "refactor"}
	prefixesChecked := make([]bool, len(allowedPrefixes))
	for i, prefix := range allowedPrefixes {
		for _, allowed := range cfg.Naming.AllowedPrefixes {
			if prefix == allowed {
				prefixesChecked[i] = true
				break
			}
		}
	}

	// Initialize AI fields
	providers := []string{"cerebras", "openai", "anthropic", "ollama"}
	providerIdx := 0
	for i, p := range providers {
		if p == cfg.AI.Provider {
			providerIdx = i
			break
		}
	}

	models := []string{"llama-3.3-70b", "llama-3.1-8b", "gpt-4", "claude-3-sonnet"}
	defaultModelIdx := 0
	fallbackModelIdx := 0
	for i, m := range models {
		if m == cfg.AI.DefaultModel {
			defaultModelIdx = i
		}
		if m == cfg.AI.FallbackModel {
			fallbackModelIdx = i
		}
	}

	tierIdx := 0
	if cfg.AI.APITier == "pro" {
		tierIdx = 1
	}

	// Initialize text inputs with actual values
	gitMainBranchInput := NewTextInput("Main Branch", "main")
	gitMainBranchInput.Value = cfg.Git.MainBranch
	if gitMainBranchInput.Value == "" {
		gitMainBranchInput.Value = "main"
	}

	commitCustomTemplateInput := NewTextInput("Custom Template", "{type}({scope}): {description}")
	if cfg.Commits.CustomTemplate != "" {
		commitCustomTemplateInput.Value = cfg.Commits.CustomTemplate
	}

	namingPatternInput := NewTextInput("Branch Pattern", "feature/{description}")
	if cfg.Naming.Pattern != "" {
		namingPatternInput.Value = cfg.Naming.Pattern
	}

	aiAPIKeyInput := NewTextInput("API Key", "Enter API key")
	if cfg.AI.APIKey != "" {
		aiAPIKeyInput.Value = cfg.AI.APIKey
	}

	aiMaxDiffSizeInput := NewTextInput("Max Diff Size (KB)", "50")
	if cfg.AI.MaxDiffSize > 0 {
		aiMaxDiffSizeInput.Value = fmt.Sprintf("%d", cfg.AI.MaxDiffSize)
	}

	return &SettingsView{
		cfg:        cfg,
		cfgManager: cfgManager,
		currentTab: SettingsGit,

		// Git
		gitMainBranch:        gitMainBranchInput,
		gitProtectedBranches: NewCheckboxGroup("Protected Branches", protectedBranches, protectedChecked),
		gitCustomProtected:   NewTextInput("Custom Protected Branch", "staging"),
		gitAutoPush:          NewCheckbox("Auto-push commits", cfg.Git.AutoPush),
		gitAutoPull:          NewCheckbox("Auto-pull on checkout", cfg.Git.AutoPull),

		// GitHub
		ghEnabled:           NewCheckbox("Enable GitHub integration", cfg.GitHub.Enabled),
		ghDefaultVisibility: NewRadioGroup("Default Visibility", []string{"Public", "Private"}, map[string]int{"public": 0, "private": 1}[cfg.GitHub.DefaultVisibility]),
		ghDefaultLicense:    NewDropdown("Default License", github.GetLicenseTemplates(), 0),
		ghDefaultGitIgnore:  NewDropdown("Default .gitignore", github.GetGitIgnoreTemplates(), 0),
		ghEnableIssues:      NewCheckbox("Enable Issues by default", cfg.GitHub.EnableIssues),
		ghEnableWiki:        NewCheckbox("Enable Wiki by default", cfg.GitHub.EnableWiki),
		ghEnableProjects:    NewCheckbox("Enable Projects by default", cfg.GitHub.EnableProjects),

		// Commits
		commitConvention: NewRadioGroup("Convention", []string{
			"Conventional Commits",
			"Custom Template",
			"None (freeform)",
		}, conventionIdx),
		commitTypes:           NewCheckboxGroup("Allowed Types", commitTypes, commitTypesChecked),
		commitRequireScope:    NewCheckbox("Require scope", cfg.Commits.RequireScope),
		commitRequireBreaking: NewCheckbox("Require breaking change marker", cfg.Commits.RequireBreaking),
		commitCustomTemplate:  commitCustomTemplateInput,

		// Naming
		namingEnforce:         NewCheckbox("Enforce naming patterns", cfg.Naming.Enforce),
		namingPattern:         namingPatternInput,
		namingAllowedPrefixes: NewCheckboxGroup("Allowed Prefixes", allowedPrefixes, prefixesChecked),
		namingCustomPrefix:    NewTextInput("Custom Prefix", ""),

		// AI
		aiProvider:       NewDropdown("Provider", providers, providerIdx),
		aiAPIKey:         aiAPIKeyInput,
		aiAPITier:        NewRadioGroup("API Tier", []string{"Free", "Pro"}, tierIdx),
		aiDefaultModel:   NewDropdown("Default Model", models, defaultModelIdx),
		aiFallbackModel:  NewDropdown("Fallback Model", models, fallbackModelIdx),
		aiMaxDiffSize:    aiMaxDiffSizeInput,
		aiIncludeContext: NewCheckbox("Include commit history context", cfg.AI.IncludeContext),

		// UI
		uiTheme:       NewDropdown("Theme", GetThemeNames(), findThemeIndex(cfg.UI.Theme)),
		originalTheme: cfg.UI.Theme,
	}
}

// findThemeIndex finds the index of a theme by name
func findThemeIndex(themeName string) int {
	themes := GetThemeNames()
	for i, name := range themes {
		if name == themeName {
			return i
		}
	}
	return 0 // Default to first theme (claude-warm)
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
		case "g", "G":
			// Switch to Git tab
			m.currentTab = SettingsGit
			m.focusedField = 0
			return m, nil

		case "h", "H":
			// Switch to GitHub tab
			m.currentTab = SettingsGitHub
			m.focusedField = 0
			return m, nil

		case "c", "C":
			// Switch to Commits tab
			m.currentTab = SettingsCommits
			m.focusedField = 0
			return m, nil

		case "n", "N":
			// Switch to Naming tab
			m.currentTab = SettingsNaming
			m.focusedField = 0
			return m, nil

		case "a", "A":
			// Switch to AI tab
			m.currentTab = SettingsAI
			m.focusedField = 0
			return m, nil

		case "u", "U":
			// Switch to UI tab
			m.currentTab = SettingsUI
			m.focusedField = 0
			return m, nil

		case "s", "S":
			// Save settings
			return m, m.saveSettings()

		case "tab", "down":
			maxFields := m.getMaxFields()
			m.focusedField = (m.focusedField + 1) % maxFields
			return m, nil

		case "shift+tab", "up":
			maxFields := m.getMaxFields()
			m.focusedField = (m.focusedField - 1 + maxFields) % maxFields
			return m, nil

		case "enter", "space":
			m.handleFieldInteraction()
			m.hasChanges = true
			return m, nil

		case "left":
			m.handleLeftKey()
			m.hasChanges = true
			return m, nil

		case "right":
			m.handleRightKey()
			m.hasChanges = true
			return m, nil

		default:
			m.handleTextInput(msg)
			m.hasChanges = true
			return m, nil
		}
	}

	return m, nil
}

// getMaxFields returns the number of fields for the current tab
func (m SettingsView) getMaxFields() int {
	switch m.currentTab {
	case SettingsGit:
		return 6 // 5 fields + save button
	case SettingsGitHub:
		return 8
	case SettingsCommits:
		return 6
	case SettingsNaming:
		return 5
	case SettingsAI:
		return 8
	case SettingsUI:
		return 1 // theme dropdown only (auto-saves)
	default:
		return 1
	}
}

// handleFieldInteraction handles enter/space on focused field
func (m *SettingsView) handleFieldInteraction() {
	switch m.currentTab {
	case SettingsGit:
		switch m.focusedField {
		case 1:
			// Toggle focused checkbox in protected branches group
			if m.gitProtectedBranches.FocusedIdx >= 0 && m.gitProtectedBranches.FocusedIdx < len(m.gitProtectedBranches.Items) {
				m.gitProtectedBranches.Items[m.gitProtectedBranches.FocusedIdx].Checked = !m.gitProtectedBranches.Items[m.gitProtectedBranches.FocusedIdx].Checked
			}
		case 3:
			m.gitAutoPush.Checked = !m.gitAutoPush.Checked
		case 4:
			m.gitAutoPull.Checked = !m.gitAutoPull.Checked
		case 5:
			// Save button - handled by saveSettings()
		}

	case SettingsGitHub:
		switch m.focusedField {
		case 0:
			m.ghEnabled.Checked = !m.ghEnabled.Checked
		case 2:
			m.ghDefaultLicense.Toggle()
		case 3:
			m.ghDefaultGitIgnore.Toggle()
		case 4:
			m.ghEnableIssues.Checked = !m.ghEnableIssues.Checked
		case 5:
			m.ghEnableWiki.Checked = !m.ghEnableWiki.Checked
		case 6:
			m.ghEnableProjects.Checked = !m.ghEnableProjects.Checked
		}

	case SettingsCommits:
		switch m.focusedField {
		case 1:
			// Toggle focused checkbox in commit types group
			if m.commitTypes.FocusedIdx >= 0 && m.commitTypes.FocusedIdx < len(m.commitTypes.Items) {
				m.commitTypes.Items[m.commitTypes.FocusedIdx].Checked = !m.commitTypes.Items[m.commitTypes.FocusedIdx].Checked
			}
		case 2:
			m.commitRequireScope.Checked = !m.commitRequireScope.Checked
		case 3:
			m.commitRequireBreaking.Checked = !m.commitRequireBreaking.Checked
		}

	case SettingsNaming:
		switch m.focusedField {
		case 0:
			m.namingEnforce.Checked = !m.namingEnforce.Checked
		case 2:
			// Toggle focused checkbox in allowed prefixes group
			if m.namingAllowedPrefixes.FocusedIdx >= 0 && m.namingAllowedPrefixes.FocusedIdx < len(m.namingAllowedPrefixes.Items) {
				m.namingAllowedPrefixes.Items[m.namingAllowedPrefixes.FocusedIdx].Checked = !m.namingAllowedPrefixes.Items[m.namingAllowedPrefixes.FocusedIdx].Checked
			}
		}

	case SettingsAI:
		switch m.focusedField {
		case 0:
			m.aiProvider.Toggle()
		case 3:
			m.aiDefaultModel.Toggle()
		case 4:
			m.aiFallbackModel.Toggle()
		case 6:
			m.aiIncludeContext.Checked = !m.aiIncludeContext.Checked
		}

	case SettingsUI:
		switch m.focusedField {
		case 0:
			m.uiTheme.Toggle()
		}
	}
}

// handleLeftKey handles left arrow key
func (m *SettingsView) handleLeftKey() {
	switch m.currentTab {
	case SettingsGit:
		if m.focusedField == 1 {
			// Navigate within protected branches checkbox group
			m.gitProtectedBranches.FocusedIdx = (m.gitProtectedBranches.FocusedIdx - 1 + len(m.gitProtectedBranches.Items)) % len(m.gitProtectedBranches.Items)
		}

	case SettingsGitHub:
		if m.focusedField == 1 {
			m.ghDefaultVisibility.Selected = (m.ghDefaultVisibility.Selected - 1 + len(m.ghDefaultVisibility.Options)) % len(m.ghDefaultVisibility.Options)
		} else if m.focusedField == 2 && m.ghDefaultLicense.Open {
			m.ghDefaultLicense.Previous()
		} else if m.focusedField == 3 && m.ghDefaultGitIgnore.Open {
			m.ghDefaultGitIgnore.Previous()
		}

	case SettingsCommits:
		switch m.focusedField {
		case 0:
			m.commitConvention.Selected = (m.commitConvention.Selected - 1 + len(m.commitConvention.Options)) % len(m.commitConvention.Options)
		case 1:
			// Navigate within commit types checkbox group
			m.commitTypes.FocusedIdx = (m.commitTypes.FocusedIdx - 1 + len(m.commitTypes.Items)) % len(m.commitTypes.Items)
		}

	case SettingsNaming:
		if m.focusedField == 2 {
			// Navigate within allowed prefixes checkbox group
			m.namingAllowedPrefixes.FocusedIdx = (m.namingAllowedPrefixes.FocusedIdx - 1 + len(m.namingAllowedPrefixes.Items)) % len(m.namingAllowedPrefixes.Items)
		}

	case SettingsAI:
		if m.focusedField == 0 && m.aiProvider.Open {
			m.aiProvider.Previous()
		} else if m.focusedField == 2 {
			m.aiAPITier.Previous()
		} else if m.focusedField == 3 && m.aiDefaultModel.Open {
			m.aiDefaultModel.Previous()
		} else if m.focusedField == 4 && m.aiFallbackModel.Open {
			m.aiFallbackModel.Previous()
		}

	case SettingsUI:
		if m.focusedField == 0 && m.uiTheme.Open {
			m.uiTheme.Previous()
			// Apply theme immediately and save to config
			selectedTheme := m.uiTheme.GetSelected()
			SetGlobalTheme(selectedTheme)
			m.cfg.UI.Theme = selectedTheme
			m.originalTheme = selectedTheme
			// Auto-save config
			_ = m.cfgManager.Save(m.cfg)
		}
	}
}

// handleRightKey handles right arrow key
func (m *SettingsView) handleRightKey() {
	switch m.currentTab {
	case SettingsGit:
		if m.focusedField == 1 {
			// Navigate within protected branches checkbox group
			m.gitProtectedBranches.FocusedIdx = (m.gitProtectedBranches.FocusedIdx + 1) % len(m.gitProtectedBranches.Items)
		}

	case SettingsGitHub:
		if m.focusedField == 1 {
			m.ghDefaultVisibility.Selected = (m.ghDefaultVisibility.Selected + 1) % len(m.ghDefaultVisibility.Options)
		} else if m.focusedField == 2 && m.ghDefaultLicense.Open {
			m.ghDefaultLicense.Next()
		} else if m.focusedField == 3 && m.ghDefaultGitIgnore.Open {
			m.ghDefaultGitIgnore.Next()
		}

	case SettingsCommits:
		switch m.focusedField {
		case 0:
			m.commitConvention.Selected = (m.commitConvention.Selected + 1) % len(m.commitConvention.Options)
		case 1:
			// Navigate within commit types checkbox group
			m.commitTypes.FocusedIdx = (m.commitTypes.FocusedIdx + 1) % len(m.commitTypes.Items)
		}

	case SettingsNaming:
		switch m.focusedField {
		case 2:
			// Navigate within allowed prefixes checkbox group
			m.namingAllowedPrefixes.FocusedIdx = (m.namingAllowedPrefixes.FocusedIdx + 1) % len(m.namingAllowedPrefixes.Items)
		}

	case SettingsAI:
		switch m.focusedField {
		case 0:
			if m.aiProvider.Open {
				m.aiProvider.Next()
			}
		case 2:
			m.aiAPITier.Next()
		case 3:
			if m.aiDefaultModel.Open {
				m.aiDefaultModel.Next()
			}
		case 4:
			if m.aiFallbackModel.Open {
				m.aiFallbackModel.Next()
			}
		}

	case SettingsUI:
		if m.focusedField == 0 && m.uiTheme.Open {
			m.uiTheme.Next()
			// Apply theme immediately and save to config
			selectedTheme := m.uiTheme.GetSelected()
			SetGlobalTheme(selectedTheme)
			m.cfg.UI.Theme = selectedTheme
			m.originalTheme = selectedTheme
			// Auto-save config
			_ = m.cfgManager.Save(m.cfg)
		}
	}
}

// handleTextInput handles text input for focused text fields
func (m *SettingsView) handleTextInput(msg tea.KeyMsg) {
	switch m.currentTab {
	case SettingsGit:
		switch m.focusedField {
		case 0:
			m.gitMainBranch.Update(msg)
		case 2:
			m.gitCustomProtected.Update(msg)
		}

	case SettingsCommits:
		switch m.focusedField {
		case 4:
			if m.commitConvention.Selected == 1 {
				m.commitCustomTemplate.Update(msg)
			}
		}

	case SettingsNaming:
		switch m.focusedField {
		case 1:
			m.namingPattern.Update(msg)
		case 3:
			m.namingCustomPrefix.Update(msg)
		}

	case SettingsAI:
		switch m.focusedField {
		case 1:
			m.aiAPIKey.Update(msg)
		case 5:
			m.aiMaxDiffSize.Update(msg)
		}
	}
}

// saveSettings saves the current settings to config
func (m *SettingsView) saveSettings() tea.Cmd {
	return func() tea.Msg {
		// Update config from form fields
		m.updateConfigFromFields()

		// Save to file
		if err := m.cfgManager.Save(m.cfg); err != nil {
			m.saveStatus = "Error: " + err.Error()
			return nil
		}

		m.saveStatus = "Settings saved successfully"
		m.hasChanges = false
		return nil
	}
}

// updateConfigFromFields updates the config struct from form field values
func (m *SettingsView) updateConfigFromFields() {
	// Git
	m.cfg.Git.MainBranch = m.gitMainBranch.Value
	m.cfg.Git.ProtectedBranches = m.gitProtectedBranches.GetChecked()
	if m.gitCustomProtected.Value != "" {
		m.cfg.Git.ProtectedBranches = append(m.cfg.Git.ProtectedBranches, m.gitCustomProtected.Value)
	}
	m.cfg.Git.AutoPush = m.gitAutoPush.Checked
	m.cfg.Git.AutoPull = m.gitAutoPull.Checked

	// GitHub
	m.cfg.GitHub.Enabled = m.ghEnabled.Checked
	m.cfg.GitHub.DefaultVisibility = strings.ToLower(m.ghDefaultVisibility.GetSelected())
	m.cfg.GitHub.DefaultLicense = m.ghDefaultLicense.GetSelected()
	m.cfg.GitHub.DefaultGitIgnore = m.ghDefaultGitIgnore.GetSelected()
	m.cfg.GitHub.EnableIssues = m.ghEnableIssues.Checked
	m.cfg.GitHub.EnableWiki = m.ghEnableWiki.Checked
	m.cfg.GitHub.EnableProjects = m.ghEnableProjects.Checked

	// Commits
	switch m.commitConvention.Selected {
	case 0:
		m.cfg.Commits.Convention = "conventional"
		m.cfg.Commits.Types = m.commitTypes.GetChecked()
		m.cfg.Commits.RequireScope = m.commitRequireScope.Checked
		m.cfg.Commits.RequireBreaking = m.commitRequireBreaking.Checked
	case 1:
		m.cfg.Commits.Convention = "custom"
		m.cfg.Commits.CustomTemplate = m.commitCustomTemplate.Value
	default:
		m.cfg.Commits.Convention = "none"
	}

	// Naming
	m.cfg.Naming.Enforce = m.namingEnforce.Checked
	m.cfg.Naming.Pattern = m.namingPattern.Value
	m.cfg.Naming.AllowedPrefixes = m.namingAllowedPrefixes.GetChecked()
	if m.namingCustomPrefix.Value != "" {
		m.cfg.Naming.AllowedPrefixes = append(m.cfg.Naming.AllowedPrefixes, m.namingCustomPrefix.Value)
	}

	// AI
	m.cfg.AI.Provider = m.aiProvider.GetSelected()
	if m.aiAPIKey.Value != "" && m.aiAPIKey.Value != "Enter API key" {
		m.cfg.AI.APIKey = m.aiAPIKey.Value
	}
	m.cfg.AI.APITier = []string{"free", "pro"}[m.aiAPITier.Selected]
	m.cfg.AI.DefaultModel = m.aiDefaultModel.GetSelected()
	m.cfg.AI.FallbackModel = m.aiFallbackModel.GetSelected()
	m.cfg.AI.IncludeContext = m.aiIncludeContext.Checked

	// Parse max diff size
	if m.aiMaxDiffSize.Value != "" {
		_, _ = fmt.Sscanf(m.aiMaxDiffSize.Value, "%d", &m.cfg.AI.MaxDiffSize)
	}

	// UI
	selectedTheme := m.uiTheme.GetSelected()
	m.cfg.UI.Theme = selectedTheme

	// Hot-swap theme immediately after saving
	SetGlobalTheme(selectedTheme)

	// Update original theme so it's not reverted on tab switch
	m.originalTheme = selectedTheme
}

// View renders the settings view
func (m SettingsView) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Nested tab bar
	tabBar := m.renderNestedTabBar()
	sections = append(sections, tabBar)
	sections = append(sections, "")

	// Content based on current tab
	content := m.renderTabContent()
	sections = append(sections, content)

	// Changes indicator and save status
	if m.hasChanges {
		sections = append(sections, "")
		sections = append(sections, lipgloss.NewStyle().
			Foreground(styles.ColorWarning).
			Render("* Unsaved changes"))
	}

	if m.saveStatus != "" {
		sections = append(sections, "")
		statusStyle := lipgloss.NewStyle().Foreground(styles.ColorSuccess)
		if strings.HasPrefix(m.saveStatus, "Error") {
			statusStyle = lipgloss.NewStyle().Foreground(styles.ColorError)
		}
		sections = append(sections, statusStyle.Render(m.saveStatus))
	}

	sections = append(sections, "")
	sections = append(sections, renderSeparator(70))

	// Footer
	footer := styles.Footer.Render(
		styles.ShortcutKey.Render("G/H/C/N/A/U") + " " + styles.ShortcutDesc.Render("Switch Tab") + "  " +
			styles.ShortcutKey.Render("Tab/↑↓") + " " + styles.ShortcutDesc.Render("Navigate") + "  " +
			styles.ShortcutKey.Render("S") + " " + styles.ShortcutDesc.Render("Save"))
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderNestedTabBar renders the nested tab navigation
func (m SettingsView) renderNestedTabBar() string {
	styles := GetGlobalThemeManager().GetStyles()
	tabs := []struct {
		name string
		key  string
	}{
		{"Git", "G"},
		{"GitHub", "H"},
		{"Commits", "C"},
		{"Naming", "N"},
		{"AI", "A"},
		{"UI", "U"},
	}
	var tabViews []string

	for i, tab := range tabs {
		var style lipgloss.Style
		if SettingsTab(i) == m.currentTab {
			style = styles.TabActive
		} else {
			style = styles.TabInactive
		}
		tabViews = append(tabViews, style.Render(fmt.Sprintf(" [%s] %s ", tab.key, tab.name)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
}

// renderTabContent renders the content for the current tab
func (m SettingsView) renderTabContent() string {
	switch m.currentTab {
	case SettingsGit:
		return m.renderGitSettings()
	case SettingsGitHub:
		return m.renderGitHubSettings()
	case SettingsCommits:
		return m.renderCommitsSettings()
	case SettingsNaming:
		return m.renderNamingSettings()
	case SettingsAI:
		return m.renderAISettings()
	case SettingsUI:
		return m.renderUISettings()
	default:
		return ""
	}
}

// renderGitSettings renders Git configuration settings
func (m SettingsView) renderGitSettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("Git Configuration"))
	lines = append(lines, "")

	// Main branch
	m.gitMainBranch.Focused = (m.focusedField == 0)
	lines = append(lines, m.gitMainBranch.View())
	lines = append(lines, "")

	// Protected branches
	lines = append(lines, m.gitProtectedBranches.View())
	lines = append(lines, "")

	// Custom protected branch
	m.gitCustomProtected.Focused = (m.focusedField == 2)
	lines = append(lines, m.gitCustomProtected.View())
	lines = append(lines, "")

	// Auto push/pull
	m.gitAutoPush.Focused = (m.focusedField == 3)
	lines = append(lines, "  "+m.gitAutoPush.View())
	m.gitAutoPull.Focused = (m.focusedField == 4)
	lines = append(lines, "  "+m.gitAutoPull.View())
	lines = append(lines, "")

	// Save button
	saveBtn := NewButton("Save Changes")
	saveBtn.Focused = (m.focusedField == 5)
	lines = append(lines, saveBtn.View())

	return strings.Join(lines, "\n")
}

// renderGitHubSettings renders GitHub configuration settings
func (m SettingsView) renderGitHubSettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("GitHub Integration"))
	lines = append(lines, "")

	// Enabled checkbox
	m.ghEnabled.Focused = (m.focusedField == 0)
	lines = append(lines, "  "+m.ghEnabled.View())
	lines = append(lines, "")

	// Visibility
	m.ghDefaultVisibility.Focused = (m.focusedField == 1)
	lines = append(lines, m.ghDefaultVisibility.View())
	lines = append(lines, "")

	// License
	m.ghDefaultLicense.Focused = (m.focusedField == 2)
	lines = append(lines, m.ghDefaultLicense.View())
	lines = append(lines, "")

	// .gitignore
	m.ghDefaultGitIgnore.Focused = (m.focusedField == 3)
	lines = append(lines, m.ghDefaultGitIgnore.View())
	lines = append(lines, "")

	// Options
	lines = append(lines, styles.FormLabel.Render("Default Options:"))
	m.ghEnableIssues.Focused = (m.focusedField == 4)
	lines = append(lines, "  "+m.ghEnableIssues.View())
	m.ghEnableWiki.Focused = (m.focusedField == 5)
	lines = append(lines, "  "+m.ghEnableWiki.View())
	m.ghEnableProjects.Focused = (m.focusedField == 6)
	lines = append(lines, "  "+m.ghEnableProjects.View())
	lines = append(lines, "")

	// Save button
	saveBtn := NewButton("Save Changes")
	saveBtn.Focused = (m.focusedField == 7)
	lines = append(lines, saveBtn.View())

	return strings.Join(lines, "\n")
}

// renderCommitsSettings renders commit convention settings
func (m SettingsView) renderCommitsSettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("Commit Conventions"))
	lines = append(lines, "")

	// Convention selection
	m.commitConvention.Focused = (m.focusedField == 0)
	lines = append(lines, m.commitConvention.View())
	lines = append(lines, "")

	// Show fields based on convention
	switch m.commitConvention.Selected {
	case 0: // Conventional
		lines = append(lines, m.commitTypes.View())
		lines = append(lines, "")

		m.commitRequireScope.Focused = (m.focusedField == 2)
		lines = append(lines, "  "+m.commitRequireScope.View())
		m.commitRequireBreaking.Focused = (m.focusedField == 3)
		lines = append(lines, "  "+m.commitRequireBreaking.View())

	case 1: // Custom
		m.commitCustomTemplate.Focused = (m.focusedField == 4)
		lines = append(lines, m.commitCustomTemplate.View())
		lines = append(lines, HelpText{Text: "Placeholders: {type}, {scope}, {description}, {body}"}.View())
	}

	lines = append(lines, "")

	// Save button
	saveBtn := NewButton("Save Changes")
	saveBtn.Focused = (m.focusedField == 5)
	lines = append(lines, saveBtn.View())

	return strings.Join(lines, "\n")
}

// renderNamingSettings renders branch naming pattern settings
func (m SettingsView) renderNamingSettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("Branch Naming Patterns"))
	lines = append(lines, "")

	// Enforce checkbox
	m.namingEnforce.Focused = (m.focusedField == 0)
	lines = append(lines, "  "+m.namingEnforce.View())
	lines = append(lines, "")

	// Pattern
	m.namingPattern.Focused = (m.focusedField == 1)
	lines = append(lines, m.namingPattern.View())
	lines = append(lines, HelpText{Text: "Use {description} placeholder"}.View())
	lines = append(lines, "")

	// Allowed prefixes
	lines = append(lines, m.namingAllowedPrefixes.View())
	lines = append(lines, "")

	// Custom prefix
	m.namingCustomPrefix.Focused = (m.focusedField == 3)
	lines = append(lines, m.namingCustomPrefix.View())
	lines = append(lines, "")

	// Save button
	saveBtn := NewButton("Save Changes")
	saveBtn.Focused = (m.focusedField == 4)
	lines = append(lines, saveBtn.View())

	return strings.Join(lines, "\n")
}

// renderAISettings renders AI provider configuration settings
func (m SettingsView) renderAISettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("AI Provider Configuration"))
	lines = append(lines, "")

	// Provider
	m.aiProvider.Focused = (m.focusedField == 0)
	lines = append(lines, m.aiProvider.View())
	lines = append(lines, "")

	// API Key
	m.aiAPIKey.Focused = (m.focusedField == 1)
	m.aiAPIKey.Password = true
	lines = append(lines, m.aiAPIKey.View())
	lines = append(lines, "")

	// API Tier
	m.aiAPITier.Focused = (m.focusedField == 2)
	lines = append(lines, m.aiAPITier.View())
	lines = append(lines, "")

	// Models
	m.aiDefaultModel.Focused = (m.focusedField == 3)
	lines = append(lines, m.aiDefaultModel.View())
	lines = append(lines, "")

	m.aiFallbackModel.Focused = (m.focusedField == 4)
	lines = append(lines, m.aiFallbackModel.View())
	lines = append(lines, "")

	// Max diff size
	m.aiMaxDiffSize.Focused = (m.focusedField == 5)
	lines = append(lines, m.aiMaxDiffSize.View())
	lines = append(lines, "")

	// Include context
	m.aiIncludeContext.Focused = (m.focusedField == 6)
	lines = append(lines, "  "+m.aiIncludeContext.View())
	lines = append(lines, "")

	// Save button
	saveBtn := NewButton("Save Changes")
	saveBtn.Focused = (m.focusedField == 7)
	lines = append(lines, saveBtn.View())

	return strings.Join(lines, "\n")
}

// renderUISettings renders UI/theme configuration settings
func (m SettingsView) renderUISettings() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	lines = append(lines, styles.FormLabel.Render("User Interface Configuration"))
	lines = append(lines, "")

	// Theme dropdown
	m.uiTheme.Focused = (m.focusedField == 0)
	lines = append(lines, m.uiTheme.View())
	lines = append(lines, "")

	// Theme preview
	currentTheme := GetGlobalThemeManager().GetCurrentTheme()
	previewLines := []string{
		"",
		"Preview:",
		"  " + styles.StatusOk.Render("Success") + "  " +
			styles.StatusWarning.Render("Warning") + "  " +
			styles.StatusError.Render("Error") + "  " +
			styles.StatusInfo.Render("Info"),
		"  Primary: " + lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render("███"),
		"  " + lipgloss.NewStyle().Foreground(styles.ColorMuted).Italic(true).
			Render("Theme: "+currentTheme.Description),
		"",
	}
	lines = append(lines, strings.Join(previewLines, "\n"))

	// Help text
	helpText := lipgloss.NewStyle().Foreground(styles.ColorMuted).Italic(true).
		Render("Note: Theme changes are applied and saved automatically.")
	lines = append(lines, helpText)

	return strings.Join(lines, "\n")
}
