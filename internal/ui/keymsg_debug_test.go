package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestKeyMsg_String tests what msg.String() returns for different keys
func TestKeyMsg_String(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() tea.KeyMsg
		expected string
	}{
		{
			name: "KeySpace type",
			setup: func() tea.KeyMsg {
				return tea.KeyMsg{Type: tea.KeySpace}
			},
			expected: " ",
		},
		{
			name: "Runes with space",
			setup: func() tea.KeyMsg {
				msg := tea.KeyMsg{}
				msg.Type = tea.KeyRunes
				msg.Runes = []rune(" ")
				return msg
			},
			expected: " ",
		},
		{
			name: "String 'space'",
			setup: func() tea.KeyMsg {
				msg := tea.KeyMsg{}
				msg.Type = -1
				// This won't work, just for testing
				return msg
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.setup()
			result := msg.String()
			t.Logf("msg.String() returned: %q (expected: %q)", result, tt.expected)

			// Also log what the switch would match
			switch result {
			case "space":
				t.Logf("Would match case 'space'")
			case " ":
				t.Logf("Would match case ' '")
			default:
				t.Logf("Would NOT match 'space' or ' ', got: %q", result)
			}
		})
	}
}
