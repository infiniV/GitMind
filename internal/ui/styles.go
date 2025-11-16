package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color scheme
var (
	// Primary colors
	colorPrimary   = lipgloss.Color("#00D9FF") // Cyan
	colorSecondary = lipgloss.Color("#7C3AED") // Purple
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorError     = lipgloss.Color("#EF4444") // Red
	colorMuted     = lipgloss.Color("#6B7280") // Gray

	// Confidence level colors
	colorHighConfidence   = lipgloss.Color("#10B981") // Green
	colorMediumConfidence = lipgloss.Color("#F59E0B") // Amber
	colorLowConfidence    = lipgloss.Color("#EF4444") // Red

	// UI element colors
	colorBorder   = lipgloss.Color("#374151") // Dark gray
	colorSelected = lipgloss.Color("#00D9FF") // Cyan (same as primary)
)

// Style definitions
var (
	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	// Section title styles
	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				MarginTop(1)

	// Repository info styles
	repoLabelStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(12)

	repoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3F4F6"))

	// Warning style
	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	// Commit message box
	commitBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

	// Option styles
	optionSelectedStyle = lipgloss.NewStyle().
				Foreground(colorSelected).
				Bold(true)

	optionNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))

	optionCursorStyle = lipgloss.NewStyle().
				Foreground(colorSelected).
				Bold(true)

	// Confidence badge styles
	highConfidenceStyle = lipgloss.NewStyle().
				Foreground(colorHighConfidence).
				Bold(true)

	mediumConfidenceStyle = lipgloss.NewStyle().
				Foreground(colorMediumConfidence).
				Bold(true)

	lowConfidenceStyle = lipgloss.NewStyle().
				Foreground(colorLowConfidence).
				Bold(true)

	// Description style
	descriptionStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true).
				PaddingLeft(3)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			PaddingTop(1)

	shortcutKeyStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	shortcutDescStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	// Metadata style
	metadataStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)
)

// Helper functions for confidence levels
func getConfidenceStyle(confidence float64) lipgloss.Style {
	if confidence >= 0.7 {
		return highConfidenceStyle
	} else if confidence >= 0.5 {
		return mediumConfidenceStyle
	}
	return lowConfidenceStyle
}

func getConfidenceLabel(confidence float64) string {
	if confidence >= 0.7 {
		return "HIGH"
	} else if confidence >= 0.5 {
		return "MEDIUM"
	}
	return "LOW"
}
