package domain

import (
	"fmt"
	"regexp"
)

// Theme represents a visual theme for the TUI.
type Theme struct {
	Name        string
	Description string
	Colors      ThemeColors
	Backgrounds ThemeBackgrounds
}

// ThemeColors defines the primary color palette for a theme.
type ThemeColors struct {
	// Primary accent color (used for selected elements, borders, etc.)
	Primary string

	// Secondary accent color (darker shade of primary)
	Secondary string

	// Success indicator color (green tones)
	Success string

	// Warning indicator color (amber/orange tones)
	Warning string

	// Error indicator color (red tones)
	Error string

	// Muted text color (for less important content)
	Muted string

	// Border color for UI elements
	Border string

	// Selected element color (usually same as Primary)
	Selected string

	// Main text color
	Text string

	// Confidence level colors
	HighConfidence   string
	MediumConfidence string
	LowConfidence    string
}

// ThemeBackgrounds defines background colors for various UI elements.
type ThemeBackgrounds struct {
	// Badge backgrounds
	BadgeHigh   string
	BadgeMedium string
	BadgeLow    string

	// Form element backgrounds
	FormInput   string
	FormFocused string

	// Modal and overlay backgrounds
	Modal         string
	Submenu       string
	Dashboard     string
	Confirmation  string
	ErrorModal    string
}

// hexColorRegex matches valid hex color codes (#RGB or #RRGGBB).
var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// Validate checks if the theme has valid color values.
func (t Theme) Validate() error {
	colors := map[string]string{
		"Primary":          t.Colors.Primary,
		"Secondary":        t.Colors.Secondary,
		"Success":          t.Colors.Success,
		"Warning":          t.Colors.Warning,
		"Error":            t.Colors.Error,
		"Muted":            t.Colors.Muted,
		"Border":           t.Colors.Border,
		"Selected":         t.Colors.Selected,
		"Text":             t.Colors.Text,
		"HighConfidence":   t.Colors.HighConfidence,
		"MediumConfidence": t.Colors.MediumConfidence,
		"LowConfidence":    t.Colors.LowConfidence,
		"BadgeHigh":        t.Backgrounds.BadgeHigh,
		"BadgeMedium":      t.Backgrounds.BadgeMedium,
		"BadgeLow":         t.Backgrounds.BadgeLow,
		"FormInput":        t.Backgrounds.FormInput,
		"FormFocused":      t.Backgrounds.FormFocused,
		"Modal":            t.Backgrounds.Modal,
		"Submenu":          t.Backgrounds.Submenu,
		"Dashboard":        t.Backgrounds.Dashboard,
		"Confirmation":     t.Backgrounds.Confirmation,
		"ErrorModal":       t.Backgrounds.ErrorModal,
	}

	for name, color := range colors {
		if !hexColorRegex.MatchString(color) {
			return fmt.Errorf("invalid hex color for %s: %s", name, color)
		}
	}

	if t.Name == "" {
		return fmt.Errorf("theme name cannot be empty")
	}

	return nil
}

// GetName returns the theme's name.
func (t Theme) GetName() string {
	return t.Name
}

// GetDescription returns the theme's description.
func (t Theme) GetDescription() string {
	return t.Description
}
