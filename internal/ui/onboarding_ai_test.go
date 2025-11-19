package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestOnboardingAIScreen_APITierRadioButton tests API tier selection
func TestOnboardingAIScreen_APITierRadioButton(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int
		keyString     string
		expectedValue int
	}{
		{"Free to Pro with right arrow", 0, "right", 1},
		{"Pro to Free with left arrow", 1, "left", 0},
		{"Free to Pro with space", 0, "space", 1},
		{"Wrap forward: Pro to Free", 1, "right", 0},
		{"Wrap backward: Free to Pro", 0, "left", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.apiTier.Selected = tt.initialValue
			screen.focusedField = 2 // API Tier field

			// Simulate key press
			msg := tea.KeyMsg{}
			switch tt.keyString {
			case "space":
				msg.Type = tea.KeySpace
			case "right":
				msg.Type = tea.KeyRight
			case "left":
				msg.Type = tea.KeyLeft
			}

			updated, _ := screen.Update(msg)

			if updated.apiTier.Selected != tt.expectedValue {
				t.Errorf("expected %d, got %d", tt.expectedValue, updated.apiTier.Selected)
			}
		})
	}
}

// TestOnboardingAIScreen_APIKeyInput tests API key text input
func TestOnboardingAIScreen_APIKeyInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to API key", "csk-123", "4", "csk-1234"},
		{"Add space to API key", "csk", " ", "csk "},
		{"Backspace from API key", "csk-123", "backspace", "csk-12"},
		{"Backspace from empty", "", "backspace", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.apiKey.Value = tt.initialValue
			screen.focusedField = 1 // API Key field

			// Simulate key press
			msg := tea.KeyMsg{}
			switch tt.keyString {
			case "backspace":
				msg.Type = tea.KeyBackspace
			case " ":
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(" ")
			default:
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(tt.keyString)
			}

			updated, _ := screen.Update(msg)

			if updated.apiKey.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.apiKey.Value)
			}

			// Verify error is cleared when typing
			if updated.error != "" && tt.keyString != "backspace" {
				t.Error("error should be cleared when typing")
			}
		})
	}
}

// TestOnboardingAIScreen_IncludeContextCheckbox tests checkbox toggle
func TestOnboardingAIScreen_IncludeContextCheckbox(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle include context from false to true", false, true},
		{"Toggle include context from true to false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.includeContext.Checked = tt.initialValue
			screen.focusedField = 6 // Include Context field

			// Simulate space key
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			if updated.includeContext.Checked != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, updated.includeContext.Checked)
			}
		})
	}
}

// TestOnboardingAIScreen_ValidationError tests empty API key validation
func TestOnboardingAIScreen_ValidationError(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingAIScreen(7, 8, cfg)
	screen.apiKey.Value = "" // Empty API key
	screen.focusedField = 7  // Continue button

	// Simulate Enter key on Continue button
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	if updated.error == "" {
		t.Error("expected validation error for empty API key")
	}

	if updated.ShouldContinue() {
		t.Error("should not continue with empty API key")
	}
}

// TestOnboardingAIScreen_SuccessfulSubmit tests successful form submission
func TestOnboardingAIScreen_SuccessfulSubmit(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingAIScreen(7, 8, cfg)
	screen.apiKey.Value = "csk-test-key-123"
	screen.focusedField = 7 // Continue button

	// Simulate Enter key on Continue button
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	if !updated.ShouldContinue() {
		t.Error("should continue with valid API key")
	}

	if updated.error != "" {
		t.Errorf("expected no error, got: %s", updated.error)
	}
}

// TestOnboardingAIScreen_MaxDiffSizeInput tests max diff size input
func TestOnboardingAIScreen_MaxDiffSizeInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add digit to max diff size", "100", "0", "1000"},
		{"Backspace from max diff size", "1000", "backspace", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.maxDiffSize.Value = tt.initialValue
			screen.focusedField = 5 // Max Diff Size field

			// Simulate key press
			msg := tea.KeyMsg{}
			if tt.keyString == "backspace" {
				msg.Type = tea.KeyBackspace
			} else {
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(tt.keyString)
			}

			updated, _ := screen.Update(msg)

			if updated.maxDiffSize.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.maxDiffSize.Value)
			}
		})
	}
}

// TestOnboardingAIScreen_Navigation tests field navigation
func TestOnboardingAIScreen_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		initialField  int
		keyType       tea.KeyType
		expectedField int
	}{
		{"Tab from field 0 to 1", 0, tea.KeyTab, 1},
		{"Tab from field 7 wraps to 0", 7, tea.KeyTab, 0},
		{"Down from field 0 to 1", 0, tea.KeyDown, 1},
		{"Up from field 1 to 0", 1, tea.KeyUp, 0},
		{"Up from field 0 wraps to 7", 0, tea.KeyUp, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.focusedField = tt.initialField

			// Simulate key press
			msg := tea.KeyMsg{Type: tt.keyType}
			updated, _ := screen.Update(msg)

			if updated.focusedField != tt.expectedField {
				t.Errorf("expected field %d, got %d", tt.expectedField, updated.focusedField)
			}
		})
	}
}

// TestOnboardingAIScreen_DropdownToggle tests dropdown open/close
func TestOnboardingAIScreen_DropdownToggle(t *testing.T) {
	tests := []struct {
		name         string
		fieldIndex   int
		initialOpen  bool
		expectedOpen bool
	}{
		{"Toggle provider dropdown open", 0, false, true},
		{"Toggle provider dropdown closed", 0, true, false},
		{"Toggle default model dropdown open", 3, false, true},
		{"Toggle fallback model dropdown open", 4, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingAIScreen(7, 8, cfg)
			screen.focusedField = tt.fieldIndex

			// Set initial dropdown state
			switch tt.fieldIndex {
			case 0:
				screen.provider.Open = tt.initialOpen
			case 3:
				screen.defaultModel.Open = tt.initialOpen
			case 4:
				screen.fallbackModel.Open = tt.initialOpen
			}

			// Simulate Enter key to toggle
			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updated, _ := screen.Update(msg)

			// Check updated state
			var result bool
			switch tt.fieldIndex {
			case 0:
				result = updated.provider.Open
			case 3:
				result = updated.defaultModel.Open
			case 4:
				result = updated.fallbackModel.Open
			}

			if result != tt.expectedOpen {
				t.Errorf("expected %v, got %v", tt.expectedOpen, result)
			}
		})
	}
}

// TestOnboardingAIScreen_EscapeGoesBack tests escape key navigation
func TestOnboardingAIScreen_EscapeGoesBack(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingAIScreen(7, 8, cfg)

	// Simulate Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := screen.Update(msg)

	if !updated.ShouldGoBack() {
		t.Error("pressing Esc should set shouldGoBack flag")
	}
}
