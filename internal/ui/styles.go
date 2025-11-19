package ui

// defaultThemeManager is the global theme manager instance.
// This is initialized with the Claude Warm theme by default and can be
// replaced when the application loads the user's theme preference.
var defaultThemeManager *ThemeManager

// init initializes the default theme manager with Claude Warm theme.
func init() {
	defaultThemeManager = NewThemeManager(ThemeClaudeWarm)
}

// SetGlobalTheme updates the global theme manager with a new theme.
// This should be called when the application loads the user's theme preference.
// After calling this, all UI components will use the new theme colors and styles.
func SetGlobalTheme(theme string) {
	selectedTheme := GetThemeByName(theme)
	defaultThemeManager.SetTheme(selectedTheme)
}

// GetGlobalThemeManager returns the global theme manager instance.
// UI components should call GetGlobalThemeManager().GetStyles() to access
// theme styles that will automatically update when the theme changes.
func GetGlobalThemeManager() *ThemeManager {
	return defaultThemeManager
}

// Backward compatibility helpers - these delegate to the global theme manager.
// These are provided for existing code during transition, but new code should
// use GetGlobalThemeManager().GetStyles() directly.

func renderSeparator(width int) string {
	return defaultThemeManager.RenderSeparator(width)
}

func getConfidenceLabel(confidence float64) string {
	if confidence >= 0.7 {
		return "HIGH"
	} else if confidence >= 0.5 {
		return "MEDIUM"
	}
	return "LOW"
}
