package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestOnboardingCommitsScreen_ConventionRadioButton tests convention selection
func TestOnboardingCommitsScreen_ConventionRadioButton(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int
		keyString     string
		expectedValue int
	}{
		{"Conventional to Custom with right", 0, "right", 1},
		{"Custom to None with right", 1, "right", 2},
		{"None to Conventional with right (wrap)", 2, "right", 0},
		{"Custom to Conventional with left", 1, "left", 0},
		{"Conventional to None with left (wrap)", 0, "left", 2},
		{"Conventional to Custom with space", 0, "space", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
			screen.convention.Selected = tt.initialValue
			screen.focusedField = 0 // Convention field

			// Simulate key press
			msg := tea.KeyMsg{}
			if tt.keyString == "space" {
				msg.Type = tea.KeySpace
			} else if tt.keyString == "right" {
				msg.Type = tea.KeyRight
			} else if tt.keyString == "left" {
				msg.Type = tea.KeyLeft
			}

			updated, _ := screen.Update(msg)

			if updated.convention.Selected != tt.expectedValue {
				t.Errorf("expected %d, got %d", tt.expectedValue, updated.convention.Selected)
			}
		})
	}
}

// TestOnboardingCommitsScreen_CommitTypesCheckboxGroup tests commit types selection
func TestOnboardingCommitsScreen_CommitTypesCheckboxGroup(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingCommitsScreen(5, 8, cfg)
	screen.focusedField = 1 // Commit types field

	// Set focused index to first item
	screen.commitTypes.FocusedIdx = 0
	initialChecked := screen.commitTypes.Items[0].Checked

	// Simulate space key to toggle
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, _ := screen.Update(msg)

	expectedChecked := !initialChecked
	if updated.commitTypes.Items[0].Checked != expectedChecked {
		t.Errorf("expected %v, got %v", expectedChecked, updated.commitTypes.Items[0].Checked)
	}
}

// TestOnboardingCommitsScreen_RequireScopeCheckbox tests require scope toggle
func TestOnboardingCommitsScreen_RequireScopeCheckbox(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle require scope from false to true", false, true},
		{"Toggle require scope from true to false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
			screen.requireScope.Checked = tt.initialValue
			screen.focusedField = 2 // Require scope field

			// Simulate space key
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			if updated.requireScope.Checked != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, updated.requireScope.Checked)
			}
		})
	}
}

// TestOnboardingCommitsScreen_RequireBreakingCheckbox tests require breaking toggle
func TestOnboardingCommitsScreen_RequireBreakingCheckbox(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle require breaking from false to true", false, true},
		{"Toggle require breaking from true to false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
			screen.requireBreaking.Checked = tt.initialValue
			screen.focusedField = 3 // Require breaking field

			// Simulate space key
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			if updated.requireBreaking.Checked != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, updated.requireBreaking.Checked)
			}
		})
	}
}

// TestOnboardingCommitsScreen_CustomTemplateInput tests custom template text input
func TestOnboardingCommitsScreen_CustomTemplateInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to template", "[type]", ":", "[type]:"},
		{"Add space to template", "[type]:", " ", "[type]: "},
		{"Backspace from template", "[type]: ", "backspace", "[type]: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
			screen.convention.Selected = 1 // Custom convention
			screen.customTemplate.Value = tt.initialValue
			screen.focusedField = 4 // Custom template field

			// Simulate key press
			msg := tea.KeyMsg{}
			if tt.keyString == "backspace" {
				msg.Type = tea.KeyBackspace
			} else if tt.keyString == " " {
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(" ")
			} else {
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(tt.keyString)
			}

			updated, _ := screen.Update(msg)

			if updated.customTemplate.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.customTemplate.Value)
			}
		})
	}
}

// TestOnboardingCommitsScreen_CustomTemplateOnlyForCustomConvention tests template is only editable for custom
func TestOnboardingCommitsScreen_CustomTemplateOnlyForCustomConvention(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingCommitsScreen(5, 8, cfg)
	screen.convention.Selected = 0 // Conventional (not custom)
	screen.customTemplate.Value = "test"
	screen.focusedField = 4 // Custom template field

	// Simulate character key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updated, _ := screen.Update(msg)

	// Should not have changed because convention is not Custom
	if updated.customTemplate.Value != "test" {
		t.Errorf("template should not change when convention is not Custom, got '%s'", updated.customTemplate.Value)
	}
}

// TestOnboardingCommitsScreen_Navigation tests field navigation
func TestOnboardingCommitsScreen_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		initialField  int
		keyType       tea.KeyType
		expectedField int
	}{
		{"Tab from field 0 to 1", 0, tea.KeyTab, 1},
		{"Tab from field 5 wraps to 0", 5, tea.KeyTab, 0},
		{"Down from field 0 to 1", 0, tea.KeyDown, 1},
		{"Up from field 1 to 0", 1, tea.KeyUp, 0},
		{"Up from field 0 wraps to 5", 0, tea.KeyUp, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
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

// TestOnboardingCommitsScreen_SubmitButton tests continue button
func TestOnboardingCommitsScreen_SubmitButton(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingCommitsScreen(5, 8, cfg)
	screen.focusedField = 5 // Continue button

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	if !updated.ShouldContinue() {
		t.Error("pressing Enter on continue button should set shouldContinue flag")
	}
}

// TestOnboardingCommitsScreen_EscapeGoesBack tests escape key navigation
func TestOnboardingCommitsScreen_EscapeGoesBack(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingCommitsScreen(5, 8, cfg)

	// Simulate Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := screen.Update(msg)

	if !updated.ShouldGoBack() {
		t.Error("pressing Esc should set shouldGoBack flag")
	}
}

// TestOnboardingCommitsScreen_CommitTypesNavigation tests navigation within commit types
func TestOnboardingCommitsScreen_CommitTypesNavigation(t *testing.T) {
	tests := []struct {
		name        string
		initialIdx  int
		keyType     tea.KeyType
		expectedIdx int
	}{
		{"Move right to next type", 0, tea.KeyRight, 1},
		{"Move left to previous type", 1, tea.KeyLeft, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingCommitsScreen(5, 8, cfg)
			screen.focusedField = 1 // Commit types field
			screen.commitTypes.FocusedIdx = tt.initialIdx

			// Simulate key press
			msg := tea.KeyMsg{Type: tt.keyType}
			updated, _ := screen.Update(msg)

			if updated.commitTypes.FocusedIdx != tt.expectedIdx {
				t.Errorf("expected focused index %d, got %d", tt.expectedIdx, updated.commitTypes.FocusedIdx)
			}
		})
	}
}
