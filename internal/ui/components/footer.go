package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/ui"
)

// Shortcut represents a keyboard shortcut
type Shortcut struct {
	Key         string
	Description string
}

// Footer represents a footer component
type Footer struct {
	Shortcuts []Shortcut
	Metadata  string // Optional metadata to display on the right
	Width     int
}

// NewFooter creates a new footer
func NewFooter(shortcuts []Shortcut) *Footer {
	return &Footer{
		Shortcuts: shortcuts,
		Width:     0, // Auto width
	}
}

// WithMetadata adds metadata to the footer
func (f *Footer) WithMetadata(metadata string) *Footer {
	f.Metadata = metadata
	return f
}

// WithWidth sets the footer width
func (f *Footer) WithWidth(width int) *Footer {
	f.Width = width
	return f
}

// Render renders the footer
func (f *Footer) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	primaryStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	var parts []string
	for _, shortcut := range f.Shortcuts {
		key := primaryStyle.Render(shortcut.Key)
		desc := mutedStyle.Render(shortcut.Description)
		parts = append(parts, key+" "+desc)
	}

	shortcuts := strings.Join(parts, " • ")

	if f.Metadata != "" {
		// If metadata exists, align it to the right
		if f.Width > 0 {
			metaStyle := mutedStyle.Italic(true)
			meta := metaStyle.Render(f.Metadata)

			// Calculate spacing
			shortcutsLen := lipgloss.Width(shortcuts)
			metaLen := lipgloss.Width(meta)
			spacing := f.Width - shortcutsLen - metaLen

			if spacing > 0 {
				return shortcuts + strings.Repeat(" ", spacing) + meta
			}
		}
		// Fallback: just append metadata
		return shortcuts + " " + mutedStyle.Render(f.Metadata)
	}

	return shortcuts
}

// Common footer shortcuts for reuse
var (
	ShortcutQuit = Shortcut{
		Key:         "q",
		Description: "quit",
	}
	ShortcutBack = Shortcut{
		Key:         "esc",
		Description: "back",
	}
	ShortcutConfirm = Shortcut{
		Key:         "enter",
		Description: "confirm",
	}
	ShortcutCancel = Shortcut{
		Key:         "esc",
		Description: "cancel",
	}
	ShortcutNavigate = Shortcut{
		Key:         "↑↓/hjkl",
		Description: "navigate",
	}
	ShortcutTab = Shortcut{
		Key:         "tab",
		Description: "switch",
	}
	ShortcutRefresh = Shortcut{
		Key:         "r",
		Description: "refresh",
	}
	ShortcutHelp = Shortcut{
		Key:         "?",
		Description: "help",
	}
)

// DashboardFooter creates a footer for the dashboard
func DashboardFooter(metadata string, width int) string {
	footer := NewFooter([]Shortcut{
		ShortcutNavigate,
		ShortcutTab,
		ShortcutConfirm,
		ShortcutRefresh,
		ShortcutQuit,
	})
	return footer.WithMetadata(metadata).WithWidth(width).Render()
}

// CommitFooter creates a footer for the commit view
func CommitFooter(state string, metadata string, width int) string {
	var shortcuts []Shortcut

	switch state {
	case "analyzing":
		shortcuts = []Shortcut{
			{Key: "esc", Description: "back to dashboard"},
		}
	case "editing":
		shortcuts = []Shortcut{
			{Key: "enter", Description: "confirm"},
			{Key: "esc", Description: "cancel"},
		}
	case "confirming":
		shortcuts = []Shortcut{
			{Key: "tab", Description: "navigate"},
			{Key: "enter", Description: "confirm"},
			{Key: "esc", Description: "cancel"},
		}
	default:
		shortcuts = []Shortcut{
			{Key: "↑↓", Description: "select option"},
			{Key: "enter", Description: "choose"},
			{Key: "e", Description: "edit message"},
			{Key: "esc", Description: "back"},
		}
	}

	footer := NewFooter(shortcuts)
	return footer.WithMetadata(metadata).WithWidth(width).Render()
}

// MergeFooter creates a footer for the merge/PR view
func MergeFooter(state string, metadata string, width int) string {
	var shortcuts []Shortcut

	switch state {
	case "analyzing":
		shortcuts = []Shortcut{
			{Key: "esc", Description: "back to dashboard"},
		}
	case "selecting":
		shortcuts = []Shortcut{
			{Key: "↑↓", Description: "navigate"},
			{Key: "enter", Description: "select"},
			{Key: "l", Description: "list PRs"},
			{Key: "c", Description: "create PR"},
			{Key: "esc", Description: "cancel"},
		}
	default:
		shortcuts = []Shortcut{
			{Key: "↑↓", Description: "navigate"},
			{Key: "enter", Description: "select"},
			{Key: "l", Description: "list PRs"},
			{Key: "c", Description: "create PR"},
			{Key: "esc", Description: "back"},
		}
	}

	footer := NewFooter(shortcuts)
	return footer.WithMetadata(metadata).WithWidth(width).Render()
}

// SettingsFooter creates a footer for the settings view
func SettingsFooter(unsavedChanges bool, width int) string {
	shortcuts := []Shortcut{
		{Key: "←→", Description: "switch tab"},
		{Key: "↑↓", Description: "navigate"},
		{Key: "enter", Description: "toggle/edit"},
		{Key: "ctrl+s", Description: "save"},
		{Key: "esc", Description: "back"},
	}

	metadata := ""
	if unsavedChanges {
		metadata = "● Unsaved changes"
	}

	footer := NewFooter(shortcuts)
	return footer.WithMetadata(metadata).WithWidth(width).Render()
}

// OnboardFooter creates a footer for the onboarding wizard
func OnboardFooter(step, totalSteps int, width int) string {
	shortcuts := []Shortcut{
		{Key: "enter", Description: "next"},
		{Key: "esc", Description: "skip"},
	}

	if step > 1 {
		shortcuts = append([]Shortcut{
			{Key: "←", Description: "back"},
		}, shortcuts...)
	}

	metadata := ""
	if totalSteps > 0 {
		metadata = lipgloss.NewStyle().Render(
			"Step " + string(rune('0'+step)) + " of " + string(rune('0'+totalSteps)),
		)
	}

	footer := NewFooter(shortcuts)
	return footer.WithMetadata(metadata).WithWidth(width).Render()
}

// HelpText renders help text in a consistent format
func HelpText(parts ...string) string {
	styles := ui.GetGlobalThemeManager().GetStyles()
	mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	return mutedStyle.Render(strings.Join(parts, " • "))
}

// StatusLine renders a status line with icon
func StatusLine(icon, message string, statusType string) string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	var style lipgloss.Style
	switch statusType {
	case "success":
		style = styles.StatusOk
	case "error":
		style = styles.StatusError
	case "warning":
		style = styles.StatusWarning
	case "info":
		style = styles.StatusInfo
	default:
		style = lipgloss.NewStyle().Foreground(styles.ColorText)
	}

	return style.Render(icon + " " + message)
}
