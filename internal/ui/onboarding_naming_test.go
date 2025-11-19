package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestOnboardingNamingScreen_EnforceCheckbox tests enforce toggle
func TestOnboardingNamingScreen_EnforceCheckbox(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle enforce from false to true", false, true},
		{"Toggle enforce from true to false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingNamingScreen(6, 8, cfg)
			screen.enforce.Checked = tt.initialValue
			screen.focusedField = 0 // Enforce field

			// Simulate space key
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			if updated.enforce.Checked != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, updated.enforce.Checked)
			}
		})
	}
}

// TestOnboardingNamingScreen_PatternInput tests pattern text input
func TestOnboardingNamingScreen_PatternInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to pattern", "feature/", "*", "feature/*"},
		{"Add space to pattern", "feature", " ", "feature "},
		{"Backspace from pattern", "feature/*", "backspace", "feature/"},
		{"Backspace from empty", "", "backspace", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingNamingScreen(6, 8, cfg)
			screen.pattern.Value = tt.initialValue
			screen.focusedField = 1 // Pattern field

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

			if updated.pattern.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.pattern.Value)
			}
		})
	}
}

// TestOnboardingNamingScreen_AllowedPrefixesCheckboxGroup tests prefix selection
func TestOnboardingNamingScreen_AllowedPrefixesCheckboxGroup(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingNamingScreen(6, 8, cfg)
	screen.focusedField = 2 // Allowed prefixes field

	// Set focused index to first item
	screen.allowedPrefixes.FocusedIdx = 0
	initialChecked := screen.allowedPrefixes.Items[0].Checked

	// Simulate space key to toggle
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, _ := screen.Update(msg)

	expectedChecked := !initialChecked
	if updated.allowedPrefixes.Items[0].Checked != expectedChecked {
		t.Errorf("expected %v, got %v", expectedChecked, updated.allowedPrefixes.Items[0].Checked)
	}
}

// TestOnboardingNamingScreen_AllowedPrefixesNavigation tests navigation within prefixes
func TestOnboardingNamingScreen_AllowedPrefixesNavigation(t *testing.T) {
	tests := []struct {
		name        string
		initialIdx  int
		keyType     tea.KeyType
		expectedIdx int
	}{
		{"Move right to next prefix", 0, tea.KeyRight, 1},
		{"Move left to previous prefix", 1, tea.KeyLeft, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingNamingScreen(6, 8, cfg)
			screen.focusedField = 2 // Allowed prefixes field
			screen.allowedPrefixes.FocusedIdx = tt.initialIdx

			// Simulate key press
			msg := tea.KeyMsg{Type: tt.keyType}
			updated, _ := screen.Update(msg)

			if updated.allowedPrefixes.FocusedIdx != tt.expectedIdx {
				t.Errorf("expected focused index %d, got %d", tt.expectedIdx, updated.allowedPrefixes.FocusedIdx)
			}
		})
	}
}

// TestOnboardingNamingScreen_CustomPrefixInput tests custom prefix text input
func TestOnboardingNamingScreen_CustomPrefixInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to custom prefix", "exp", "t", "expt"},
		{"Backspace from custom prefix", "expt", "backspace", "exp"},
		{"Add space to custom prefix", "exp", " ", "exp "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingNamingScreen(6, 8, cfg)
			screen.customPrefix.Value = tt.initialValue
			screen.focusedField = 3 // Custom prefix field

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

			if updated.customPrefix.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.customPrefix.Value)
			}
		})
	}
}

// TestOnboardingNamingScreen_Navigation tests field navigation
func TestOnboardingNamingScreen_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		initialField  int
		keyType       tea.KeyType
		expectedField int
	}{
		{"Tab from field 0 to 1", 0, tea.KeyTab, 1},
		{"Tab from field 4 wraps to 0", 4, tea.KeyTab, 0},
		{"Down from field 0 to 1", 0, tea.KeyDown, 1},
		{"Up from field 1 to 0", 1, tea.KeyUp, 0},
		{"Up from field 0 wraps to 4", 0, tea.KeyUp, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingNamingScreen(6, 8, cfg)
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

// TestOnboardingNamingScreen_SubmitButton tests continue button
func TestOnboardingNamingScreen_SubmitButton(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingNamingScreen(6, 8, cfg)
	screen.focusedField = 4 // Continue button

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	if !updated.ShouldContinue() {
		t.Error("pressing Enter on continue button should set shouldContinue flag")
	}
}

// TestOnboardingNamingScreen_EscapeGoesBack tests escape key navigation
func TestOnboardingNamingScreen_EscapeGoesBack(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingNamingScreen(6, 8, cfg)

	// Simulate Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := screen.Update(msg)

	if !updated.ShouldGoBack() {
		t.Error("pressing Esc should set shouldGoBack flag")
	}
}

// TestOnboardingNamingScreen_PreviewUpdatesOnChange tests preview generation
func TestOnboardingNamingScreen_PreviewUpdatesOnChange(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingNamingScreen(6, 8, cfg)
	screen.enforce.Checked = true
	screen.pattern.Value = "feature/*"
	screen.focusedField = 1 // Pattern field

	// Get initial preview
	initialPreview := screen.previewExample

	// Change pattern
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, _ := screen.Update(msg)

	// Preview should have updated (test that it's different or at least exists)
	if updated.previewExample == initialPreview && updated.pattern.Value != screen.pattern.Value {
		t.Error("preview should update when pattern changes")
	}
}

// TestOnboardingNamingScreen_EnforceDisabledNoPreview tests no preview when disabled
func TestOnboardingNamingScreen_EnforceDisabledNoPreview(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingNamingScreen(6, 8, cfg)
	screen.enforce.Checked = false
	screen.focusedField = 0 // Enforce field

	// Check that preview indicates enforcement is disabled
	if screen.previewExample != "Enforcement disabled" {
		t.Errorf("expected 'Enforcement disabled', got '%s'", screen.previewExample)
	}
}
