package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/domain"
)

// ThemeManager manages the current theme and provides styled components.
type ThemeManager struct {
	currentTheme domain.Theme
	styles       *ThemeStyles
}

// ThemeStyles contains all lipgloss styles for the TUI.
type ThemeStyles struct {
	// Color values (as lipgloss.Color)
	ColorPrimary          lipgloss.Color
	ColorSecondary        lipgloss.Color
	ColorSuccess          lipgloss.Color
	ColorWarning          lipgloss.Color
	ColorError            lipgloss.Color
	ColorMuted            lipgloss.Color
	ColorBorder           lipgloss.Color
	ColorSelected         lipgloss.Color
	ColorText             lipgloss.Color
	ColorHighConfidence   lipgloss.Color
	ColorMediumConfidence lipgloss.Color
	ColorLowConfidence    lipgloss.Color

	// Header styles
	Header        lipgloss.Style
	SectionTitle  lipgloss.Style
	RepoLabel     lipgloss.Style
	RepoValue     lipgloss.Style
	Warning       lipgloss.Style
	CommitBox     lipgloss.Style

	// Option styles
	OptionSelected lipgloss.Style
	OptionNormal   lipgloss.Style
	OptionCursor   lipgloss.Style
	Description    lipgloss.Style

	// Footer styles
	Footer       lipgloss.Style
	ShortcutKey  lipgloss.Style
	ShortcutDesc lipgloss.Style
	Metadata     lipgloss.Style

	// Dashboard card styles
	DashboardCard       lipgloss.Style
	DashboardCardActive lipgloss.Style
	CardTitle           lipgloss.Style

	// Status indicator styles
	StatusOk      lipgloss.Style
	StatusWarning lipgloss.Style
	StatusError   lipgloss.Style
	StatusInfo    lipgloss.Style

	// Submenu styles
	Submenu             lipgloss.Style
	SubmenuOption       lipgloss.Style
	SubmenuOptionActive lipgloss.Style
	Checkbox            lipgloss.Style

	// Badge styles
	BadgeHigh   lipgloss.Style
	BadgeMedium lipgloss.Style
	BadgeLow    lipgloss.Style

	// Separator style
	Separator lipgloss.Style

	// Option box styles
	SelectedOptionBox lipgloss.Style
	NormalOptionBox   lipgloss.Style
	OptionLabel       lipgloss.Style
	OptionDesc        lipgloss.Style

	// Loading style
	Loading lipgloss.Style

	// Tab styles
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabBar      lipgloss.Style

	// Form component styles
	FormLabel           lipgloss.Style
	FormInput           lipgloss.Style
	FormInputFocused    lipgloss.Style
	FormHelp            lipgloss.Style
	FormButton          lipgloss.Style
	FormButtonInactive  lipgloss.Style

	// Menu styles (for vertical action menu)
	MenuSelected lipgloss.Style
	MenuNormal   lipgloss.Style

	// Git tree styles
	TreeBranch lipgloss.Style
	TreeCommit lipgloss.Style
	TreeMerge  lipgloss.Style
}

// NewThemeManager creates a new theme manager with the specified theme.
func NewThemeManager(theme domain.Theme) *ThemeManager {
	tm := &ThemeManager{
		currentTheme: theme,
		styles:       &ThemeStyles{},
	}
	tm.regenerateStyles()
	return tm
}

// GetCurrentTheme returns the current theme.
func (tm *ThemeManager) GetCurrentTheme() domain.Theme {
	return tm.currentTheme
}

// SetTheme changes the current theme and regenerates all styles.
func (tm *ThemeManager) SetTheme(theme domain.Theme) {
	tm.currentTheme = theme
	tm.regenerateStyles()
}

// GetStyles returns the current theme styles.
func (tm *ThemeManager) GetStyles() *ThemeStyles {
	return tm.styles
}

// regenerateStyles rebuilds all lipgloss styles based on the current theme.
func (tm *ThemeManager) regenerateStyles() {
	c := tm.currentTheme.Colors
	bg := tm.currentTheme.Backgrounds

	// Convert theme colors to lipgloss.Color
	colorPrimary := lipgloss.Color(c.Primary)
	colorSecondary := lipgloss.Color(c.Secondary)
	colorSuccess := lipgloss.Color(c.Success)
	colorWarning := lipgloss.Color(c.Warning)
	colorError := lipgloss.Color(c.Error)
	colorMuted := lipgloss.Color(c.Muted)
	colorBorder := lipgloss.Color(c.Border)
	colorSelected := lipgloss.Color(c.Selected)
	colorText := lipgloss.Color(c.Text)
	colorHighConfidence := lipgloss.Color(c.HighConfidence)
	colorMediumConfidence := lipgloss.Color(c.MediumConfidence)
	colorLowConfidence := lipgloss.Color(c.LowConfidence)

	// Store colors
	tm.styles.ColorPrimary = colorPrimary
	tm.styles.ColorSecondary = colorSecondary
	tm.styles.ColorSuccess = colorSuccess
	tm.styles.ColorWarning = colorWarning
	tm.styles.ColorError = colorError
	tm.styles.ColorMuted = colorMuted
	tm.styles.ColorBorder = colorBorder
	tm.styles.ColorSelected = colorSelected
	tm.styles.ColorText = colorText
	tm.styles.ColorHighConfidence = colorHighConfidence
	tm.styles.ColorMediumConfidence = colorMediumConfidence
	tm.styles.ColorLowConfidence = colorLowConfidence

	// Header styles
	tm.styles.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder)

	tm.styles.SectionTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorSecondary).
		MarginTop(1)

	tm.styles.RepoLabel = lipgloss.NewStyle().
		Foreground(colorMuted).
		Width(12)

	tm.styles.RepoValue = lipgloss.NewStyle().
		Foreground(colorText)

	tm.styles.Warning = lipgloss.NewStyle().
		Foreground(colorWarning).
		Bold(true)

	tm.styles.CommitBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	// Option styles
	tm.styles.OptionSelected = lipgloss.NewStyle().
		Foreground(colorSelected).
		Bold(true)

	tm.styles.OptionNormal = lipgloss.NewStyle().
		Foreground(colorMuted)

	tm.styles.OptionCursor = lipgloss.NewStyle().
		Foreground(colorSelected).
		Bold(true)

	tm.styles.Description = lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		PaddingLeft(3)

	// Footer styles
	tm.styles.Footer = lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		PaddingTop(1)

	tm.styles.ShortcutKey = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	tm.styles.ShortcutDesc = lipgloss.NewStyle().
		Foreground(colorMuted)

	tm.styles.Metadata = lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true)

	// Dashboard card styles
	tm.styles.DashboardCard = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1).
		Width(38)

	tm.styles.DashboardCardActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSelected).
		Padding(1).
		Width(38)

	tm.styles.CardTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary)

	// Status indicator styles
	tm.styles.StatusOk = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	tm.styles.StatusWarning = lipgloss.NewStyle().
		Foreground(colorWarning).
		Bold(true)

	tm.styles.StatusError = lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	tm.styles.StatusInfo = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	// Submenu styles
	tm.styles.Submenu = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Background(lipgloss.Color(bg.Submenu))

	tm.styles.SubmenuOption = lipgloss.NewStyle().
		Foreground(colorMuted)

	tm.styles.SubmenuOptionActive = lipgloss.NewStyle().
		Foreground(colorSelected).
		Bold(true)

	tm.styles.Checkbox = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	// Badge styles
	tm.styles.BadgeHigh = lipgloss.NewStyle().
		Foreground(colorHighConfidence).
		Background(lipgloss.Color(bg.BadgeHigh)).
		Padding(0, 1).
		Bold(true)

	tm.styles.BadgeMedium = lipgloss.NewStyle().
		Foreground(colorMediumConfidence).
		Background(lipgloss.Color(bg.BadgeMedium)).
		Padding(0, 1).
		Bold(true)

	tm.styles.BadgeLow = lipgloss.NewStyle().
		Foreground(colorLowConfidence).
		Background(lipgloss.Color(bg.BadgeLow)).
		Padding(0, 1).
		Bold(true)

	// Separator style
	tm.styles.Separator = lipgloss.NewStyle().
		Foreground(colorBorder).
		MarginTop(1).
		MarginBottom(1)

	// Option box styles
	tm.styles.SelectedOptionBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSelected).
		Padding(1).
		MarginBottom(1)

	tm.styles.NormalOptionBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1).
		MarginBottom(1)

	tm.styles.OptionLabel = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	tm.styles.OptionDesc = lipgloss.NewStyle().
		Foreground(colorMuted).
		PaddingTop(0).
		PaddingLeft(2)

	// Loading style
	tm.styles.Loading = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	// Tab styles
	tm.styles.TabActive = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Underline(true)

	tm.styles.TabInactive = lipgloss.NewStyle().
		Foreground(colorMuted)

	tm.styles.TabBar = lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		MarginBottom(1).
		PaddingLeft(1)

	// Form component styles
	tm.styles.FormLabel = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true).
		Width(20)

	tm.styles.FormInput = lipgloss.NewStyle().
		Foreground(colorText).
		Background(lipgloss.Color(bg.FormInput)).
		Padding(0, 1)

	tm.styles.FormInputFocused = lipgloss.NewStyle().
		Foreground(colorText).
		Background(lipgloss.Color(bg.FormFocused)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	tm.styles.FormHelp = lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		PaddingLeft(2)

	tm.styles.FormButton = lipgloss.NewStyle().
		Foreground(colorText).
		Background(colorPrimary).
		Padding(0, 2).
		Bold(true)

	tm.styles.FormButtonInactive = lipgloss.NewStyle().
		Foreground(colorMuted).
		Background(lipgloss.Color(bg.FormInput)).
		Padding(0, 2)

	// Menu styles (for vertical action menu)
	tm.styles.MenuSelected = lipgloss.NewStyle().
		Foreground(colorText).
		Background(colorPrimary).
		Bold(true).
		Padding(0, 2)

	tm.styles.MenuNormal = lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 2)

	// Git tree styles
	tm.styles.TreeBranch = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	tm.styles.TreeCommit = lipgloss.NewStyle().
		Foreground(colorText)

	tm.styles.TreeMerge = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)
}

// GetConfidenceBadge returns a styled confidence badge.
func (tm *ThemeManager) GetConfidenceBadge(confidence float64) string {
	label := getConfidenceLabel(confidence)
	percent := fmt.Sprintf("%.0f%%", confidence*100)

	var badgeStyle lipgloss.Style
	if confidence >= 0.7 {
		badgeStyle = tm.styles.BadgeHigh
	} else if confidence >= 0.5 {
		badgeStyle = tm.styles.BadgeMedium
	} else {
		badgeStyle = tm.styles.BadgeLow
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		badgeStyle.Render(label),
		" ",
		tm.styles.StatusInfo.Render(percent),
	)
}

// RenderSeparator returns a styled horizontal separator.
func (tm *ThemeManager) RenderSeparator(width int) string {
	if width <= 0 {
		width = 60
	}
	return tm.styles.Separator.Render(strings.Repeat("─", width))
}
