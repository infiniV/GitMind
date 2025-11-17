package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestOnboardingGitHubScreen_VisibilityRadioButton tests radio button navigation
func TestOnboardingGitHubScreen_VisibilityRadioButton(t *testing.T) {
	tests := []struct {
		name          string
		initialValue  int
		keyString     string
		expectedValue int
	}{
		{"Public to Private with right arrow", 0, "right", 1},
		{"Private to Public with left arrow", 1, "left", 0},
		{"Public to Private with space", 0, " ", 1},
		{"Wrap forward: Private to Public", 1, "right", 0},
		{"Wrap backward: Public to Private", 0, "left", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
			screen.ghAvailable = true
			screen.ghAuthenticated = true
			screen.checkComplete = true
			screen.visibility.Selected = tt.initialValue
			screen.focusedField = 2 // Visibility field

			// Simulate key press
			msg := tea.KeyMsg{}
			msg.Type = tea.KeyRunes
			if tt.keyString == " " {
				msg.Type = tea.KeySpace
			} else {
				msg.Type = -1 // Regular key
			}

			updated, _ := screen.Update(msg)

			if updated.visibility.Selected != tt.expectedValue {
				t.Errorf("expected %d, got %d", tt.expectedValue, updated.visibility.Selected)
			}
		})
	}
}

// TestOnboardingGitHubScreen_CheckboxToggle tests checkbox toggling
func TestOnboardingGitHubScreen_CheckboxToggle(t *testing.T) {
	tests := []struct {
		name          string
		fieldIndex    int
		initialValue  bool
		expectedValue bool
	}{
		{"Toggle Add README from false to true", 5, false, true},
		{"Toggle Add README from true to false", 5, true, false},
		{"Toggle Enable Issues from false to true", 6, false, true},
		{"Toggle Enable Wiki from false to true", 7, false, true},
		{"Toggle Enable Projects from false to true", 8, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
			screen.ghAvailable = true
			screen.ghAuthenticated = true
			screen.checkComplete = true
			screen.focusedField = tt.fieldIndex

			// Set initial checkbox value
			switch tt.fieldIndex {
			case 5:
				screen.addReadme.Checked = tt.initialValue
			case 6:
				screen.enableIssues.Checked = tt.initialValue
			case 7:
				screen.enableWiki.Checked = tt.initialValue
			case 8:
				screen.enableProjects.Checked = tt.initialValue
			}

			// Simulate space key press
			msg := tea.KeyMsg{Type: tea.KeySpace}
			updated, _ := screen.Update(msg)

			// Check updated value
			var result bool
			switch tt.fieldIndex {
			case 5:
				result = updated.addReadme.Checked
			case 6:
				result = updated.enableIssues.Checked
			case 7:
				result = updated.enableWiki.Checked
			case 8:
				result = updated.enableProjects.Checked
			}

			if result != tt.expectedValue {
				t.Errorf("expected %v, got %v", tt.expectedValue, result)
			}
		})
	}
}

// TestOnboardingGitHubScreen_TextInput tests text input functionality
func TestOnboardingGitHubScreen_TextInput(t *testing.T) {
	tests := []struct {
		name          string
		fieldIndex    int
		initialValue  string
		keyString     string
		expectedValue string
	}{
		{"Add character to repo name", 0, "test", "x", "testx"},
		{"Add space to description", 1, "Created", " ", "Created "},
		{"Backspace from repo name", 0, "test", "backspace", "tes"},
		{"Backspace from empty", 0, "", "backspace", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
			screen.ghAvailable = true
			screen.ghAuthenticated = true
			screen.checkComplete = true
			screen.focusedField = tt.fieldIndex

			// Set initial value
			if tt.fieldIndex == 0 {
				screen.repoName.Value = tt.initialValue
			} else {
				screen.description.Value = tt.initialValue
			}

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

			// Check updated value
			var result string
			if tt.fieldIndex == 0 {
				result = updated.repoName.Value
			} else {
				result = updated.description.Value
			}

			if result != tt.expectedValue {
				t.Errorf("expected '%s', got '%s'", tt.expectedValue, result)
			}
		})
	}
}

// TestOnboardingGitHubScreen_Navigation tests field navigation
func TestOnboardingGitHubScreen_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		initialField  int
		keyType       tea.KeyType
		expectedField int
	}{
		{"Tab from field 0 to 1", 0, tea.KeyTab, 1},
		{"Tab from field 9 wraps to 0", 9, tea.KeyTab, 0},
		{"Down from field 0 to 1", 0, tea.KeyDown, 1},
		{"Up from field 1 to 0", 1, tea.KeyUp, 0},
		{"Up from field 0 wraps to 9", 0, tea.KeyUp, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
			screen.ghAvailable = true
			screen.ghAuthenticated = true
			screen.checkComplete = true
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

// TestOnboardingGitHubScreen_EscapeGoesBack tests escape key navigation
func TestOnboardingGitHubScreen_EscapeGoesBack(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
	screen.ghAvailable = true
	screen.ghAuthenticated = true
	screen.checkComplete = true

	// Simulate Esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := screen.Update(msg)

	if !updated.ShouldGoBack() {
		t.Error("pressing Esc should set shouldGoBack flag")
	}
}

// TestOnboardingGitHubScreen_DropdownToggle tests dropdown open/close
func TestOnboardingGitHubScreen_DropdownToggle(t *testing.T) {
	tests := []struct {
		name           string
		fieldIndex     int
		initialOpen    bool
		expectedOpen   bool
	}{
		{"Toggle license dropdown open", 3, false, true},
		{"Toggle license dropdown closed", 3, true, false},
		{"Toggle gitignore dropdown open", 4, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &domain.Config{}
			screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
			screen.ghAvailable = true
			screen.ghAuthenticated = true
			screen.checkComplete = true
			screen.focusedField = tt.fieldIndex

			// Set initial dropdown state
			if tt.fieldIndex == 3 {
				screen.license.Open = tt.initialOpen
			} else {
				screen.gitignore.Open = tt.initialOpen
			}

			// Simulate Enter key to toggle
			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updated, _ := screen.Update(msg)

			// Check updated state
			var result bool
			if tt.fieldIndex == 3 {
				result = updated.license.Open
			} else {
				result = updated.gitignore.Open
			}

			if result != tt.expectedOpen {
				t.Errorf("expected %v, got %v", tt.expectedOpen, result)
			}
		})
	}
}

// TestOnboardingGitHubScreen_SkipKey tests skip functionality
func TestOnboardingGitHubScreen_SkipKey(t *testing.T) {
	cfg := &domain.Config{}
	screen := NewOnboardingGitHubScreen(3, 8, cfg, "D:\\test")
	screen.ghAvailable = true
	screen.ghAuthenticated = true
	screen.checkComplete = true

	// Simulate 's' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	updated, _ := screen.Update(msg)

	if !updated.ShouldContinue() {
		t.Error("pressing 's' should set shouldContinue flag")
	}
	if !updated.shouldSkip {
		t.Error("pressing 's' should set shouldSkip flag")
	}
}
