package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestOnboardingBranchesScreen_MainBranchInput tests main branch text input
func TestOnboardingBranchesScreen_MainBranchInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to main branch", "mai", "n", "main"},
		{"Backspace from main branch", "main", "backspace", "mai"},
		{"Add space to main branch", "main", " ", "main "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingBranchesScreen(4, 8, cfg)
			screen.mainBranch.Value = tt.initialValue
			screen.focusedField = 0 // Main branch field

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

			if updated.mainBranch.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.mainBranch.Value)
			}
		})
	}
}

// TestOnboardingBranchesScreen_ProtectedBranchesCheckboxGroup tests checkbox group
func TestOnboardingBranchesScreen_ProtectedBranchesCheckboxGroup(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingBranchesScreen(4, 8, cfg)
	screen.focusedField = 1 // Protected branches field

	// Initially first item should be unchecked
	initialChecked := screen.protectedBranches.Items[0].Checked

	// Set focused index
	screen.protectedBranches.FocusedIdx = 0

	// Simulate space key to toggle
	msg := tea.KeyMsg{Type: tea.KeySpace}
	updated, _ := screen.Update(msg)

	expectedChecked := !initialChecked
	if updated.protectedBranches.Items[0].Checked != expectedChecked {
		t.Errorf("expected %v, got %v", expectedChecked, updated.protectedBranches.Items[0].Checked)
	}
}

// TestOnboardingBranchesScreen_ProtectedBranchesNavigation tests navigation within checkbox group
func TestOnboardingBranchesScreen_ProtectedBranchesNavigation(t *testing.T) {
	tests := []struct {
		name          string
		initialIdx    int
		keyType       tea.KeyType
		expectedIdx   int
	}{
		{"Move right to next item", 0, tea.KeyRight, 1},
		{"Move left to previous item", 1, tea.KeyLeft, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingBranchesScreen(4, 8, cfg)
			screen.focusedField = 1 // Protected branches field
			screen.protectedBranches.FocusedIdx = tt.initialIdx

			// Simulate key press
			msg := tea.KeyMsg{Type: tt.keyType}
			updated, _ := screen.Update(msg)

			if updated.protectedBranches.FocusedIdx != tt.expectedIdx {
				t.Errorf("expected focused index %d, got %d", tt.expectedIdx, updated.protectedBranches.FocusedIdx)
			}
		})
	}
}

// TestOnboardingBranchesScreen_CustomProtectedBranchInput tests custom branch input
func TestOnboardingBranchesScreen_CustomProtectedBranchInput(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to custom branch", "feat", "s", "feats"},
		{"Backspace from custom branch", "feats", "backspace", "feat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingBranchesScreen(4, 8, cfg)
			screen.customProtected.Value = tt.initialValue
			screen.focusedField = 2 // Custom protected field

			// Simulate key press
			msg := tea.KeyMsg{}
			if tt.keyString == "backspace" {
				msg.Type = tea.KeyBackspace
			} else {
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(tt.keyString)
			}

			updated, _ := screen.Update(msg)

			if updated.customProtected.Value != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, updated.customProtected.Value)
			}
		})
	}
}

// TestOnboardingBranchesScreen_CustomBranchAddition tests adding custom protected branch
func TestOnboardingBranchesScreen_CustomBranchAddition(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingBranchesScreen(4, 8, cfg)
	screen.customProtected.Value = "staging"
	screen.focusedField = 2 // Custom protected field

	initialCount := len(screen.protectedBranches.Items)

	// Simulate Enter key to add
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	// Should have added one more item
	if len(updated.protectedBranches.Items) != initialCount+1 {
		t.Errorf("expected %d items, got %d", initialCount+1, len(updated.protectedBranches.Items))
	}

	// Custom field should be cleared
	if updated.customProtected.Value != "" {
		t.Errorf("expected empty custom field, got '%s'", updated.customProtected.Value)
	}
}

// TestOnboardingBranchesScreen_AutoPushCheckbox tests auto-push toggle
func TestOnboardingBranchesScreen_AutoPushCheckbox(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle auto-push from false to true", false, true},
		{"Toggle auto-push from true to false", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingBranchesScreen(4, 8, cfg)
			screen.autoPush.Checked = tt.initialValue
			screen.focusedField = 3 // Auto-push field

			// Simulate space key
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			if updated.autoPush.Checked != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, updated.autoPush.Checked)
			}
		})
	}
}

// TestOnboardingBranchesScreen_AutoPullQuickToggle tests 'P' key for auto-pull
func TestOnboardingBranchesScreen_AutoPullQuickToggle(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingBranchesScreen(4, 8, cfg)
	screen.autoPull.Checked = false

	// Simulate 'p' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}
	updated, _ := screen.Update(msg)

	if !updated.autoPull.Checked {
		t.Error("pressing 'p' should toggle auto-pull")
	}

	// Test 'P' (uppercase)
	msg2 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("P")}
	updated2, _ := updated.Update(msg2)

	if updated2.autoPull.Checked {
		t.Error("pressing 'P' should toggle auto-pull off")
	}
}

// TestOnboardingBranchesScreen_Navigation tests field navigation
func TestOnboardingBranchesScreen_Navigation(t *testing.T) {
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
			screen := NewOnboardingBranchesScreen(4, 8, cfg)
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

// TestOnboardingBranchesScreen_SubmitButton tests continue button
func TestOnboardingBranchesScreen_SubmitButton(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingBranchesScreen(4, 8, cfg)
	screen.focusedField = 4 // Continue button

	// Simulate Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := screen.Update(msg)

	if !updated.ShouldContinue() {
		t.Error("pressing Enter on continue button should set shouldContinue flag")
	}
}

// TestOnboardingBranchesScreen_BackNavigation tests back navigation
func TestOnboardingBranchesScreen_BackNavigation(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingBranchesScreen(4, 8, cfg)
	screen.focusedField = 0 // Main branch field

	// Simulate left key on field 0
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := screen.Update(msg)

	if !updated.ShouldGoBack() {
		t.Error("pressing left on field 0 should set shouldGoBack flag")
	}
}
