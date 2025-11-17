package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/gitman/internal/domain"
)

// TestAllCheckboxGroups_ToggleFunctionality tests all checkbox groups in all screens
func TestAllCheckboxGroups_ToggleFunctionality(t *testing.T) {
	t.Run("OnboardingCommits - CommitTypes", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingCommitsScreen(5, 8, cfg)
		screen.focusedField = 1 // Commit types field
		screen.commitTypes.FocusedIdx = 0
		initialChecked := screen.commitTypes.Items[0].Checked

		t.Logf("Initial state: focusedField=%d, FocusedIdx=%d, Checked=%v",
			screen.focusedField, screen.commitTypes.FocusedIdx, initialChecked)

		// Simulate space key - use string matching like the actual code does
		msg := tea.KeyMsg{}
		msg.Type = tea.KeyRunes
		msg.Runes = []rune(" ")
		updated, _ := screen.Update(msg)

		t.Logf("After update: focusedField=%d, FocusedIdx=%d, Checked=%v",
			updated.focusedField, updated.commitTypes.FocusedIdx, updated.commitTypes.Items[0].Checked)

		if updated.commitTypes.Items[0].Checked == initialChecked {
			t.Errorf("CommitTypes checkbox did not toggle. Initial: %v, After: %v",
				initialChecked, updated.commitTypes.Items[0].Checked)
		}
	})

	t.Run("OnboardingBranches - ProtectedBranches", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingBranchesScreen(4, 8, cfg)
		screen.focusedField = 1 // Protected branches field
		screen.protectedBranches.FocusedIdx = 0
		initialChecked := screen.protectedBranches.Items[0].Checked

		t.Logf("Initial state: focusedField=%d, FocusedIdx=%d, Checked=%v",
			screen.focusedField, screen.protectedBranches.FocusedIdx, initialChecked)

		// Simulate space key - use string matching
		msg := tea.KeyMsg{}
		msg.Type = tea.KeyRunes
		msg.Runes = []rune(" ")
		updated, _ := screen.Update(msg)

		t.Logf("After update: focusedField=%d, FocusedIdx=%d, Checked=%v",
			updated.focusedField, updated.protectedBranches.FocusedIdx, updated.protectedBranches.Items[0].Checked)

		if updated.protectedBranches.Items[0].Checked == initialChecked {
			t.Errorf("ProtectedBranches checkbox did not toggle. Initial: %v, After: %v",
				initialChecked, updated.protectedBranches.Items[0].Checked)
		}
	})

	t.Run("OnboardingNaming - AllowedPrefixes", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingNamingScreen(6, 8, cfg)
		screen.focusedField = 2 // Allowed prefixes field
		screen.allowedPrefixes.FocusedIdx = 0
		initialChecked := screen.allowedPrefixes.Items[0].Checked

		t.Logf("Initial state: focusedField=%d, FocusedIdx=%d, Checked=%v",
			screen.focusedField, screen.allowedPrefixes.FocusedIdx, initialChecked)

		// Simulate space key - use string matching
		msg := tea.KeyMsg{}
		msg.Type = tea.KeyRunes
		msg.Runes = []rune(" ")
		updated, _ := screen.Update(msg)

		t.Logf("After update: focusedField=%d, FocusedIdx=%d, Checked=%v",
			updated.focusedField, updated.allowedPrefixes.FocusedIdx, updated.allowedPrefixes.Items[0].Checked)

		if updated.allowedPrefixes.Items[0].Checked == initialChecked {
			t.Errorf("AllowedPrefixes checkbox did not toggle. Initial: %v, After: %v",
				initialChecked, updated.allowedPrefixes.Items[0].Checked)
		}
	})
}

// TestAllCheckboxGroups_FocusIndicator tests that focus indicator (>) appears
func TestAllCheckboxGroups_FocusIndicator(t *testing.T) {
	t.Run("CommitTypes shows focus indicator", func(t *testing.T) {
		types := []string{"feat", "fix", "docs"}
		checked := []bool{true, true, true}
		group := NewCheckboxGroup("Test", types, checked)
		group.FocusedIdx = 1 // Focus second item

		view := group.View()

		// View should contain "> " for focused item
		if !contains(view, "> [x] fix") {
			t.Errorf("Focus indicator not found in view:\n%s", view)
		}
	})

	t.Run("ProtectedBranches shows focus indicator", func(t *testing.T) {
		branches := []string{"main", "master", "develop"}
		checked := []bool{true, true, false}
		group := NewCheckboxGroup("Protected", branches, checked)
		group.FocusedIdx = 0 // Focus first item

		view := group.View()

		// View should contain "> " for focused item
		if !contains(view, "> [x] main") {
			t.Errorf("Focus indicator not found in view:\n%s", view)
		}
	})
}

// TestAllCheckboxGroups_Navigation tests left/right navigation
func TestAllCheckboxGroups_Navigation(t *testing.T) {
	t.Run("CommitTypes navigation", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingCommitsScreen(5, 8, cfg)
		screen.focusedField = 1
		screen.commitTypes.FocusedIdx = 0

		// Navigate right
		msg := tea.KeyMsg{Type: tea.KeyRight}
		updated, _ := screen.Update(msg)

		if updated.commitTypes.FocusedIdx != 1 {
			t.Errorf("Navigation failed. Expected FocusedIdx=1, got %d", updated.commitTypes.FocusedIdx)
		}
	})

	t.Run("ProtectedBranches navigation", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingBranchesScreen(4, 8, cfg)
		screen.focusedField = 1
		screen.protectedBranches.FocusedIdx = 0

		// Navigate right
		msg := tea.KeyMsg{Type: tea.KeyRight}
		updated, _ := screen.Update(msg)

		if updated.protectedBranches.FocusedIdx != 1 {
			t.Errorf("Navigation failed. Expected FocusedIdx=1, got %d", updated.protectedBranches.FocusedIdx)
		}
	})

	t.Run("AllowedPrefixes navigation", func(t *testing.T) {
		cfg := &domain.Config{}
		screen := NewOnboardingNamingScreen(6, 8, cfg)
		screen.focusedField = 2
		screen.allowedPrefixes.FocusedIdx = 0

		// Navigate right
		msg := tea.KeyMsg{Type: tea.KeyRight}
		updated, _ := screen.Update(msg)

		if updated.allowedPrefixes.FocusedIdx != 1 {
			t.Errorf("Navigation failed. Expected FocusedIdx=1, got %d", updated.allowedPrefixes.FocusedIdx)
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
