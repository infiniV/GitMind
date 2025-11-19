package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/ui"
	"github.com/yourusername/gitman/internal/ui/layout"
)

// ErrorSeverity defines the severity level of an error
type ErrorSeverity int

const (
	SeverityError ErrorSeverity = iota
	SeverityWarning
	SeverityInfo
)

// ErrorBanner represents an error banner component
type ErrorBanner struct {
	Title    string
	Message  string
	Actions  []string // Suggested actions to resolve
	Severity ErrorSeverity
	Width    int
}

// NewErrorBanner creates a new error banner
func NewErrorBanner(message string) *ErrorBanner {
	return &ErrorBanner{
		Title:    "Error",
		Message:  message,
		Severity: SeverityError,
		Width:    0, // Auto width
	}
}

// NewWarningBanner creates a warning banner
func NewWarningBanner(message string) *ErrorBanner {
	return &ErrorBanner{
		Title:    "Warning",
		Message:  message,
		Severity: SeverityWarning,
		Width:    0,
	}
}

// NewInfoBanner creates an info banner
func NewInfoBanner(message string) *ErrorBanner {
	return &ErrorBanner{
		Title:    "Info",
		Message:  message,
		Severity: SeverityInfo,
		Width:    0,
	}
}

// WithTitle sets a custom title
func (eb *ErrorBanner) WithTitle(title string) *ErrorBanner {
	eb.Title = title
	return eb
}

// WithActions adds suggested actions
func (eb *ErrorBanner) WithActions(actions ...string) *ErrorBanner {
	eb.Actions = actions
	return eb
}

// WithWidth sets the width
func (eb *ErrorBanner) WithWidth(width int) *ErrorBanner {
	eb.Width = width
	return eb
}

// Render renders the error banner
func (eb *ErrorBanner) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	var bannerStyle lipgloss.Style
	var titleStyle lipgloss.Style
	var icon string

	switch eb.Severity {
	case SeverityError:
		bannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorError).
			Padding(layout.SpacingSM, layout.SpacingMD)
		titleStyle = styles.StatusError.Bold(true)
		icon = "✗"
	case SeverityWarning:
		bannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorWarning).
			Padding(layout.SpacingSM, layout.SpacingMD)
		titleStyle = styles.StatusWarning.Bold(true)
		icon = "⚠"
	case SeverityInfo:
		bannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorSecondary).
			Padding(layout.SpacingSM, layout.SpacingMD)
		titleStyle = styles.StatusInfo.Bold(true)
		icon = "ℹ"
	}

	if eb.Width > 0 {
		bannerStyle = bannerStyle.Width(eb.Width - (layout.SpacingMD * 2) - 2)
	}

	var content strings.Builder

	// Title with icon
	content.WriteString(titleStyle.Render(icon+" "+eb.Title) + "\n")

	// Message
	textStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
	content.WriteString(textStyle.Render(eb.Message))

	// Actions if provided
	if len(eb.Actions) > 0 {
		content.WriteString("\n\n")
		mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
		content.WriteString(mutedStyle.Render("Suggested actions:") + "\n")
		for _, action := range eb.Actions {
			content.WriteString(textStyle.Render("  • " + action) + "\n")
		}
	}

	return bannerStyle.Render(content.String())
}

// ValidationError represents a validation error for form fields
type ValidationError struct {
	Field   string
	Message string
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Render renders the validation error (for inline display under inputs)
func (ve *ValidationError) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()
	return styles.StatusError.Render("✗ " + ve.Message)
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// NewValidationErrors creates a new validation errors collection
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: []ValidationError{},
	}
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Has checks if there are any errors
func (ve *ValidationErrors) Has() bool {
	return len(ve.Errors) > 0
}

// Get gets errors for a specific field
func (ve *ValidationErrors) Get(field string) []ValidationError {
	var errors []ValidationError
	for _, err := range ve.Errors {
		if err.Field == field {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetFirst gets the first error for a field
func (ve *ValidationErrors) GetFirst(field string) *ValidationError {
	for _, err := range ve.Errors {
		if err.Field == field {
			return &err
		}
	}
	return nil
}

// Render renders all validation errors as a list
func (ve *ValidationErrors) Render() string {
	if !ve.Has() {
		return ""
	}

	styles := ui.GetGlobalThemeManager().GetStyles()

	var content strings.Builder
	errorStyle := styles.StatusError.Bold(true)
	content.WriteString(errorStyle.Render("✗ Validation Errors:") + "\n\n")

	textStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorText)
	for _, err := range ve.Errors {
		fieldLabel := textStyle.Render(err.Field)
		errMsg := styles.StatusError.Render(err.Message)
		content.WriteString("  " + fieldLabel + ": " + errMsg + "\n")
	}

	return content.String()
}

// Clear clears all errors
func (ve *ValidationErrors) Clear() {
	ve.Errors = []ValidationError{}
}
