package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// OnboardingAIScreen handles AI provider configuration
type OnboardingAIScreen struct {
	step       int
	totalSteps int
	config     *domain.Config

	// Form fields
	focusedField   int
	provider       Dropdown
	apiKey         TextInput
	apiTier        RadioGroup
	defaultModel   Dropdown
	fallbackModel  Dropdown
	maxDiffSize    TextInput
	includeContext Checkbox

	shouldContinue bool
	shouldGoBack   bool
	error          string
	
	width  int
	height int
}

// NewOnboardingAIScreen creates a new AI screen
func NewOnboardingAIScreen(step, totalSteps int, config *domain.Config) OnboardingAIScreen {
	// Provider options
	providers := []string{"cerebras"}
	providerIdx := 0

	// Model options
	models := []string{"llama-3.3-70b", "llama3.1-8b"}
	defaultModelIdx := 0
	fallbackModelIdx := 1

	// Find current selections
	for i, p := range providers {
		if p == config.AI.Provider {
			providerIdx = i
			break
		}
	}
	for i, m := range models {
		if m == config.AI.DefaultModel {
			defaultModelIdx = i
		}
		if m == config.AI.FallbackModel {
			fallbackModelIdx = i
		}
	}

	// Tier index
	tierIdx := 0 // Free
	if config.AI.APITier == "pro" {
		tierIdx = 1
	}

	screen := OnboardingAIScreen{
		step:       step,
		totalSteps: totalSteps,
		config:     config,

		provider:      NewDropdown("AI Provider", providers, providerIdx),
		apiKey:        NewTextInput("API Key", ""),
		apiTier:       NewRadioGroup("API Tier", []string{"Free", "Pro"}, tierIdx),
		defaultModel:  NewDropdown("Default Model", models, defaultModelIdx),
		fallbackModel: NewDropdown("Fallback Model", models, fallbackModelIdx),
		maxDiffSize:   NewTextInput("Max Diff Size (bytes)", "100000"),
		includeContext: NewCheckbox("Include branch context in AI analysis", config.AI.IncludeContext),

		focusedField: 0,
		width:        100,
		height:       40,
	}

	// Set current values
	screen.apiKey.Password = true
	if config.AI.APIKey != "" {
		screen.apiKey.Value = config.AI.APIKey
	}
	if config.AI.MaxDiffSize > 0 {
		screen.maxDiffSize.Value = fmt.Sprintf("%d", config.AI.MaxDiffSize)
	}

	return screen
}

// Init initializes the screen
func (m OnboardingAIScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m OnboardingAIScreen) Update(msg tea.Msg) (OnboardingAIScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// For "Continue" button (field 7), validate and save
			if m.focusedField == 7 {
				if m.apiKey.Value == "" {
					m.error = "API key is required"
					return m, nil
				}
				m.saveToConfig()
				m.shouldContinue = true
				return m, nil
			}
			// For dropdowns, toggle them
			switch m.focusedField {
			case 0:
				m.provider.Toggle()
				return m, nil
			case 3:
				m.defaultModel.Toggle()
				return m, nil
			case 4:
				m.fallbackModel.Toggle()
				return m, nil
			}
			// For all other fields, move to next
			m.focusedField = (m.focusedField + 1) % 8
			return m, nil

		case "tab", "down":
			m.focusedField = (m.focusedField + 1) % 8
			return m, nil

		case "shift+tab", "up":
			m.focusedField = (m.focusedField - 1 + 8) % 8
			return m, nil

		case "esc":
			// Esc always goes back
			m.shouldGoBack = true
			return m, nil

		case "left":
			// ONLY for navigating within radio/dropdown - NO back navigation
			if m.focusedField == 2 {
				m.apiTier.Selected = (m.apiTier.Selected - 1 + len(m.apiTier.Options)) % len(m.apiTier.Options)
			} else if m.focusedField == 0 && m.provider.Open {
				m.provider.Previous()
			} else if m.focusedField == 3 && m.defaultModel.Open {
				m.defaultModel.Previous()
			} else if m.focusedField == 4 && m.fallbackModel.Open {
				m.fallbackModel.Previous()
			}
			return m, nil

		case "right":
			// ONLY for navigating within radio/dropdown
			if m.focusedField == 2 {
				m.apiTier.Selected = (m.apiTier.Selected + 1) % len(m.apiTier.Options)
			} else if m.focusedField == 0 && m.provider.Open {
				m.provider.Next()
			} else if m.focusedField == 3 && m.defaultModel.Open {
				m.defaultModel.Next()
			} else if m.focusedField == 4 && m.fallbackModel.Open {
				m.fallbackModel.Next()
			}
			return m, nil

		case " ": // Space character
			switch m.focusedField {
			case 2:
				m.apiTier.Selected = (m.apiTier.Selected + 1) % len(m.apiTier.Options)
			case 6:
				m.includeContext.Checked = !m.includeContext.Checked
			}
			return m, nil

		case "backspace", "delete":
			// Handle text input deletion
			switch m.focusedField {
			case 1:
				if len(m.apiKey.Value) > 0 {
					m.apiKey.Value = m.apiKey.Value[:len(m.apiKey.Value)-1]
				}
				m.error = "" // Clear error on input
			case 5:
				if len(m.maxDiffSize.Value) > 0 {
					m.maxDiffSize.Value = m.maxDiffSize.Value[:len(m.maxDiffSize.Value)-1]
				}
			}
			return m, nil

		default:
			// Handle all other text input (alphanumeric, symbols, etc.)
			if m.focusedField == 1 || m.focusedField == 5 {
				key := msg.String()
				if key == "space" {
					if m.focusedField == 1 {
						m.apiKey.Value += " "
						m.error = "" // Clear error on input
					} else {
						m.maxDiffSize.Value += " "
					}
				} else if len(key) == 1 {
					if m.focusedField == 1 {
						m.apiKey.Value += key
						m.error = "" // Clear error on input
					} else {
						m.maxDiffSize.Value += key
					}
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// saveToConfig saves the configuration
func (m *OnboardingAIScreen) saveToConfig() {
	m.config.AI.Provider = m.provider.GetSelected()
	m.config.AI.APIKey = m.apiKey.Value
	m.config.AI.APITier = strings.ToLower(m.apiTier.GetSelected())
	m.config.AI.DefaultModel = m.defaultModel.GetSelected()
	m.config.AI.FallbackModel = m.fallbackModel.GetSelected()
	m.config.AI.IncludeContext = m.includeContext.Checked

	// Parse max diff size
	var maxDiff int
	_, _ = fmt.Sscanf(m.maxDiffSize.Value, "%d", &maxDiff)
	if maxDiff > 0 {
		m.config.AI.MaxDiffSize = maxDiff
	} else {
		m.config.AI.MaxDiffSize = 100000 // Default
	}
}

// View renders the AI screen
func (m OnboardingAIScreen) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var sections []string

	// Header
	header := styles.Header.Render("AI Provider Configuration")
	sections = append(sections, header)

	// Progress
	progress := fmt.Sprintf("Step %d of %d", m.step, m.totalSteps)
	sections = append(sections, styles.Metadata.Render(progress))

	sections = append(sections, "")

	// Description
	desc := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		"Configure your AI provider for intelligent commit and merge assistance.")
	sections = append(sections, desc)
	sections = append(sections, "")

	// Provider (currently only Cerebras)
	m.provider.Focused = (m.focusedField == 0)
	sections = append(sections, m.provider.View())

	sections = append(sections, "")

	// API Key
	m.apiKey.Focused = (m.focusedField == 1)
	sections = append(sections, m.apiKey.View())

	// Help text with link
	helpText := "Get your free API key at: "
	link := lipgloss.NewStyle().Foreground(styles.ColorPrimary).Underline(true).Render("https://cloud.cerebras.ai/")
	sections = append(sections, HelpText{Text: helpText + link}.View())

	sections = append(sections, "")

	// API Tier
	m.apiTier.Focused = (m.focusedField == 2)
	sections = append(sections, m.apiTier.View())
	sections = append(sections, HelpText{Text: "Free tier has rate limits; Pro tier has higher limits"}.View())

	sections = append(sections, "")

	// Default Model
	m.defaultModel.Focused = (m.focusedField == 3)
	sections = append(sections, m.defaultModel.View())
	sections = append(sections, HelpText{Text: "llama-3.3-70b is recommended for best results"}.View())

	sections = append(sections, "")

	// Fallback Model
	m.fallbackModel.Focused = (m.focusedField == 4)
	sections = append(sections, m.fallbackModel.View())
	sections = append(sections, HelpText{Text: "Used when default model is unavailable"}.View())

	sections = append(sections, "")

	// Max Diff Size
	m.maxDiffSize.Focused = (m.focusedField == 5)
	sections = append(sections, m.maxDiffSize.View())
	sections = append(sections, HelpText{Text: "Maximum size of diffs sent to AI (larger diffs are truncated)"}.View())

	sections = append(sections, "")

	// Include Context
	m.includeContext.Focused = (m.focusedField == 6)
	sections = append(sections, m.includeContext.View())
	sections = append(sections, HelpText{Text: "Provide branch name, parent, and commit history to AI for better analysis"}.View())

	sections = append(sections, "")

	// Error message
	if m.error != "" {
		sections = append(sections, styles.StatusError.Render("Error: "+m.error))
		sections = append(sections, "")
	}

	// Continue button
	continueBtn := NewButton("Continue")
	continueBtn.Focused = (m.focusedField == 7)
	continueBtn.Active = (m.apiKey.Value != "")
	sections = append(sections, continueBtn.View())

	// Wrap in card
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	cardStyle := styles.DashboardCard.Padding(1, 2)

	// Main view assembly
	mainView := []string{
		header,
		styles.Metadata.Render(progress),
		"",
		cardStyle.Render(content),
		"",
		renderSeparator(70),
	}

	// Footer
	footer := styles.Footer.Render(
		styles.ShortcutKey.Render("Tab/↑↓")+" "+styles.ShortcutDesc.Render("Navigate")+"  "+
			styles.ShortcutKey.Render("Space/←→")+" "+styles.ShortcutDesc.Render("Select")+"  "+
			styles.ShortcutKey.Render("←")+" "+styles.ShortcutDesc.Render("Back"))
	mainView = append(mainView, footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, mainView...),
	)
}

// ShouldContinue returns true if user wants to continue
func (m OnboardingAIScreen) ShouldContinue() bool {
	return m.shouldContinue
}

// ShouldGoBack returns true if user wants to go back
func (m OnboardingAIScreen) ShouldGoBack() bool {
	return m.shouldGoBack
}
