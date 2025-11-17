package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color scheme (Claude Code inspired)
var (
	// Primary colors - warm, professional palette
	colorPrimary   = lipgloss.Color("#C15F3C") // Claude Orange (rust/warm)
	colorSecondary = lipgloss.Color("#A14A2F") // Dark Orange
	colorSuccess   = lipgloss.Color("#7A9A6E") // Warm green
	colorWarning   = lipgloss.Color("#D4945A") // Warm amber
	colorError     = lipgloss.Color("#C16B6B") // Warm red
	colorMuted     = lipgloss.Color("#B1ADA1") // Cloudy (warm gray)

	// Confidence level colors
	colorHighConfidence   = lipgloss.Color("#7A9A6E") // Warm green
	colorMediumConfidence = lipgloss.Color("#D4945A") // Warm amber
	colorLowConfidence    = lipgloss.Color("#C16B6B") // Warm red

	// UI element colors
	colorBorder   = lipgloss.Color("#3A3631") // Warm dark gray
	colorSelected = lipgloss.Color("#C15F3C") // Claude Orange
	colorText     = lipgloss.Color("#e8e6e3") // Light warm text
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
			Foreground(colorText)

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
				Foreground(colorMuted)

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

	// Dashboard card styles
	dashboardCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(1).
				Width(38)

	dashboardCardActiveStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(colorSelected).
					Padding(1).
					Width(38)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	cardContentStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	cardIconStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	// Status indicator styles
	statusOkStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	statusWarningStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	statusInfoStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Submenu styles
	submenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Background(lipgloss.Color("#1F2937"))

	submenuOptionStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	submenuOptionActiveStyle = lipgloss.NewStyle().
					Foreground(colorSelected).
					Bold(true)

	checkboxStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Badge styles (for confidence levels, status indicators)
	badgeHighStyle = lipgloss.NewStyle().
				Foreground(colorHighConfidence).
				Background(lipgloss.Color("#1F3A2C")).
				Padding(0, 1).
				Bold(true)

	badgeMediumStyle = lipgloss.NewStyle().
					Foreground(colorMediumConfidence).
					Background(lipgloss.Color("#3A2F1F")).
					Padding(0, 1).
					Bold(true)

	badgeLowStyle = lipgloss.NewStyle().
				Foreground(colorLowConfidence).
				Background(lipgloss.Color("#3A1F1F")).
				Padding(0, 1).
				Bold(true)

	badgeInfoStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Background(lipgloss.Color("#2F2A1F")).
				Padding(0, 1).
				Bold(true)

	// Separator styles
	separatorStyle = lipgloss.NewStyle().
				Foreground(colorBorder).
				MarginTop(1).
				MarginBottom(1)

	// Option box styles (for better visual hierarchy)
	selectedOptionBoxStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(colorSelected).
					Padding(1).
					MarginBottom(1)

	normalOptionBoxStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(colorBorder).
					Padding(1).
					MarginBottom(1)

	// Option label and description styles
	optionLabelStyle = lipgloss.NewStyle().
					Foreground(colorText).
					Bold(true)

	optionDescStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				PaddingTop(0).
				PaddingLeft(2)

	// Loading styles
	loadingStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	// Tab styles
	tabActiveStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				Underline(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	tabBarStyle = lipgloss.NewStyle().
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorBorder).
				MarginBottom(1).
				PaddingLeft(1)

	// Form component styles
	formLabelStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true).
				Width(20)

	formInputStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(lipgloss.Color("#2F2A1F")).
				Padding(0, 1)

	formInputFocusedStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(lipgloss.Color("#3A2F1F")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	formHelpStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true).
				PaddingLeft(2)

	formButtonStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorPrimary).
				Padding(0, 2).
				Bold(true)

	formButtonInactiveStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Background(lipgloss.Color("#2F2A1F")).
				Padding(0, 2)
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

// Helper function for confidence badges
func getConfidenceBadge(confidence float64) string {
	label := getConfidenceLabel(confidence)
	percent := fmt.Sprintf("%.0f%%", confidence*100)

	var badgeStyle lipgloss.Style
	if confidence >= 0.7 {
		badgeStyle = badgeHighStyle
	} else if confidence >= 0.5 {
		badgeStyle = badgeMediumStyle
	} else {
		badgeStyle = badgeLowStyle
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		badgeStyle.Render(label),
		" ",
		statusInfoStyle.Render(percent),
	)
}

// Helper function for horizontal separators
func renderSeparator(width int) string {
	if width <= 0 {
		width = 60
	}
	return separatorStyle.Render(strings.Repeat("─", width))
}
