package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	successPrefix = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true).
			Render("[SUCCESS]")

	errorPrefix = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			Render("[ERROR]")

	infoPrefix = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Render("[INFO]")

	warningPrefix = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true).
			Render("[WARNING]")
)

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("%s %s\n", successPrefix, message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("%s %s\n", errorPrefix, message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("%s %s\n", infoPrefix, message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("%s %s\n", warningPrefix, message)
}

// PrintSubtle prints a muted/subtle message
func PrintSubtle(message string) {
	fmt.Println(lipgloss.NewStyle().Foreground(colorMuted).Render(message))
}

// FormatValue highlights a value in output
func FormatValue(value string) string {
	return lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render(value)
}

// FormatLabel formats a label
func FormatLabel(label string) string {
	return lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(label)
}
