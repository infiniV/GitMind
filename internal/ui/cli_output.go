package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Helper functions to get styled prefixes (these use the current theme)
func getSuccessPrefix() string {
	return lipgloss.NewStyle().
		Foreground(GetGlobalThemeManager().GetStyles().ColorSuccess).
		Bold(true).
		Render("[SUCCESS]")
}

func getErrorPrefix() string {
	return lipgloss.NewStyle().
		Foreground(GetGlobalThemeManager().GetStyles().ColorError).
		Bold(true).
		Render("[ERROR]")
}

func getInfoPrefix() string {
	return lipgloss.NewStyle().
		Foreground(GetGlobalThemeManager().GetStyles().ColorPrimary).
		Bold(true).
		Render("[INFO]")
}

func getWarningPrefix() string {
	return lipgloss.NewStyle().
		Foreground(GetGlobalThemeManager().GetStyles().ColorWarning).
		Bold(true).
		Render("[WARNING]")
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", getSuccessPrefix(), message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("%s %s\n", getErrorPrefix(), message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", getInfoPrefix(), message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", getWarningPrefix(), message)
}

// PrintSubtle prints a muted/subtle message
func PrintSubtle(message string) {
	styles := GetGlobalThemeManager().GetStyles()
	fmt.Println(lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(message))
}

// FormatValue highlights a value in output
func FormatValue(value string) string {
	styles := GetGlobalThemeManager().GetStyles()
	return lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Render(value)
}

// FormatLabel formats a label
func FormatLabel(label string) string {
	styles := GetGlobalThemeManager().GetStyles()
	return lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render(label)
}
