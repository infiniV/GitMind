# GitMind UI Components Library

This package provides reusable, consistent UI components for the GitMind terminal user interface. All components follow the same design patterns and integrate with the global theme system.

## Table of Contents

- [Layout System](#layout-system)
- [Components](#components)
  - [Modal](#modal)
  - [Card](#card)
  - [Error Components](#error-components)
  - [Logo & Branding](#logo--branding)
  - [Footer](#footer)
  - [Text Utilities](#text-utilities)
- [Form Components](#form-components)
- [Usage Examples](#usage-examples)

---

## Layout System

Location: `internal/ui/layout/constants.go`

Provides standardized spacing, sizing, and layout calculations.

### Constants

```go
// Spacing
layout.SpacingXS  // 1
layout.SpacingSM  // 2
layout.SpacingMD  // 3
layout.SpacingLG  // 4
layout.SpacingXL  // 6

// Modal Widths
layout.ModalWidthSM  // 50
layout.ModalWidthMD  // 60
layout.ModalWidthLG  // 70
layout.ModalWidthXL  // 80

// Modal Heights
layout.ModalHeightSM  // 10
layout.ModalHeightMD  // 15
layout.ModalHeightLG  // 20

// Split Ratios
layout.SplitRatio35_65  // 0.35
layout.SplitRatio40_60  // 0.40
layout.SplitRatio50_50  // 0.50
```

### Helper Functions

```go
// Calculate content height after headers/footers
height := layout.CalculateContentHeight(windowHeight)

// Calculate content height with tabs
height := layout.CalculateContentHeightWithTabs(windowHeight)

// Split view into left/right widths
leftWidth, rightWidth := layout.CalculateSplitWidths(totalWidth, layout.SplitRatio35_65)

// Calculate viewport height for scrollable content
viewportHeight := layout.CalculateViewportHeight(windowHeight)

// Center content horizontally or vertically
x := layout.CenterHorizontal(windowWidth, contentWidth)
y := layout.CenterVertical(windowHeight, contentHeight)
```

---

## Components

### Modal

Location: `internal/ui/components/modal.go`

Reusable modal dialog component with multiple types and full keyboard navigation.

#### Modal Types

- `ModalInfo` - Informational messages
- `ModalError` - Error displays
- `ModalWarning` - Warning messages
- `ModalConfirm` - Yes/No confirmations
- `ModalForm` - Forms with inputs and buttons

#### Quick Start

```go
// Simple error modal
modal := components.NewErrorModal("Something went wrong!")
content := modal.RenderCentered(windowWidth, windowHeight)

// Confirmation modal
modal := components.NewConfirmModal(
    "Delete Branch?",
    "This action cannot be undone.",
    func() { /* on yes */ },
    func() { /* on no */ },
)

// Form modal
inputs := []components.ModalInput{
    components.NewModalInput("Name", "Enter name...", ""),
    components.NewModalInput("Email", "Enter email...", ""),
}

buttons := []components.ModalButton{
    {Label: "Submit", Primary: true, OnSelect: func() { /* submit */ }},
    {Label: "Cancel", Primary: false, OnSelect: func() { /* cancel */ }},
}

modal := components.NewFormModal("User Info", "Please fill out the form:", inputs, buttons)
```

#### Navigation

- **Tab/Shift+Tab**: Switch between inputs and buttons
- **Up/Down/Left/Right**: Navigate options
- **Enter**: Confirm/Select
- **Esc**: Cancel/Dismiss

#### Methods

```go
modal.Update(keyString)              // Handle keyboard input
modal.UpdateInput(msg)               // Forward input to text fields
modal.GetSelectedButton()            // Get currently selected button
modal.GetInputValues()               // Get all input values as map
modal.Render()                       // Render modal
modal.RenderCentered(width, height)  // Render centered on screen
```

---

### Card

Location: `internal/ui/components/card.go`

Reusable card component for displaying content with borders and styling.

#### Card Types

- `CardDefault` - Default styling
- `CardPrimary` - Primary color border
- `CardSuccess` - Success color border
- `CardWarning` - Warning color border
- `CardError` - Error color border
- `CardInfo` - Info color border

#### Usage

```go
// Basic card
card := components.NewCard("Title", "Content goes here")
rendered := card.Render()

// Dashboard card with size
card := components.NewDashboardCard("Status", statusContent, 30, 10)
card.SetActive(true)  // Highlight as active
rendered := card.Render()

// Styled card
card := components.NewCard("Warning", "This is important")
card.SetType(components.CardWarning)
card.SetSize(50, 15)
rendered := card.Render()

// Status card with icon
statusCard := components.NewStatusCard(
    "Build Status",
    "✓",
    "Passing",
    "Last run: 2 minutes ago",
)
rendered := statusCard.Render()

// Info card with key-value pairs
infoCard := components.NewInfoCard("Repository Info", []components.InfoItem{
    {Label: "Branch", Value: "main", Icon: ""},
    {Label: "Commits", Value: "42", Icon: ""},
    {Label: "Status", Value: "Clean", Icon: "✓"},
})
rendered := infoCard.Render()
```

---

### Error Components

Location: `internal/ui/components/error.go`

Consistent error, warning, and info banners with suggested actions.

#### Error Severity

- `SeverityError` - Red error styling
- `SeverityWarning` - Yellow warning styling
- `SeverityInfo` - Blue info styling

#### Usage

```go
// Simple error banner
banner := components.NewErrorBanner("Failed to connect to repository")
rendered := banner.Render()

// Error with title and actions
banner := components.NewErrorBanner("Connection timeout").
    WithTitle("Network Error").
    WithActions(
        "Check your internet connection",
        "Verify repository URL",
        "Try again in a few minutes",
    ).
    WithWidth(60)

// Warning banner
banner := components.NewWarningBanner("Uncommitted changes detected")

// Info banner
banner := components.NewInfoBanner("Remote repository is 3 commits ahead")

// Validation errors for forms
errors := components.NewValidationErrors()
errors.Add("email", "Email is required")
errors.Add("password", "Password must be at least 8 characters")

// Get errors for specific field
if err := errors.GetFirst("email"); err != nil {
    fmt.Println(err.Render())  // "✗ Email is required"
}

// Render all errors
if errors.Has() {
    fmt.Println(errors.Render())
}
```

---

### Logo & Branding

Location: `internal/ui/components/logo.go`

Centralized ASCII art logos and branding elements.

#### Logo Variants

- `LogoDashboard` - GM logo for dashboard
- `LogoCommit` - COMMIT logo
- `LogoMerge` - MERGE logo
- `LogoSettings` - Settings variant
- `LogoOnboard` - Onboarding variant

#### Usage

```go
// Render logo with repo info
logo := components.RenderLogo(components.LogoDashboard, "my-project (main)")

// Compact logo for limited space
logo := components.RenderLogoCompact(components.LogoCommit)

// Branding text
branding := components.RenderBranding()  // "GitMind - AI-Powered Git Manager"

// Custom header
header := components.RenderHeader("Settings", "Configure your preferences")

// Divider line
divider := components.RenderDivider(80)  // 80-character wide divider
```

---

### Footer

Location: `internal/ui/components/footer.go`

Consistent footer with keyboard shortcuts and metadata.

#### Usage

```go
// Custom footer
footer := components.NewFooter([]components.Shortcut{
    {Key: "enter", Description: "confirm"},
    {Key: "esc", Description: "cancel"},
})
footer.WithMetadata("Step 3 of 5")
footer.WithWidth(windowWidth)
rendered := footer.Render()

// Pre-built footers
dashboardFooter := components.DashboardFooter("my-repo", windowWidth)
commitFooter := components.CommitFooter("editing", "main branch", windowWidth)
mergeFooter := components.MergeFooter("selecting", "2 commits", windowWidth)
settingsFooter := components.SettingsFooter(hasUnsavedChanges, windowWidth)
onboardFooter := components.OnboardFooter(currentStep, totalSteps, windowWidth)

// Common shortcuts (reusable)
shortcuts := []components.Shortcut{
    components.ShortcutQuit,
    components.ShortcutBack,
    components.ShortcutConfirm,
    components.ShortcutNavigate,
    components.ShortcutTab,
    components.ShortcutRefresh,
}

// Helper functions
helpText := components.HelpText("Tab to navigate", "Enter to confirm", "Esc to cancel")
statusLine := components.StatusLine("✓", "Build passed", "success")
```

---

### Text Utilities

Location: `internal/ui/components/text_utils.go`

Text manipulation and formatting utilities.

#### Usage

```go
// Word wrapping
wrapped := components.WrapText(longText, 80)
manualWrapped := components.WrapTextManual(longText, 80)

// Truncation
short := components.TruncateText("Very long text here", 20)  // "Very long text he..."
middle := components.TruncateMiddle("path/to/very/long/file.txt", 20)

// Padding and alignment
padded := components.PadRight("Left", 20)
padded := components.PadLeft("Right", 20)
centered := components.Center("Centered", 40)

// Indentation
indented := components.IndentText(multilineText, 4)  // 4 spaces

// ANSI handling
plainText := components.StripANSI(styledText)
charCount := components.CountVisibleChars(styledText)

// Pluralization
word := components.Pluralize(count, "file", "files")
formatted := components.FormatCount(5, "commit", "commits")  // "5 commits"

// Line splitting
lines := components.SplitIntoLines(longText, 80)

// Case conversion
title := components.TitleCase("hello world")  // "Hello World"

// Highlighting
highlighted := components.HighlightText(
    "Find this word in text",
    "word",
    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("yellow")),
)

// List formatting
list := components.FormatList([]string{"Item 1", "Item 2", "Item 3"}, "•")
numbered := components.FormatNumberedList([]string{"First", "Second", "Third"})

// Path ellipsizing
short := components.EllipsizeMiddle("/very/long/path/to/file.txt", 20, "...")
```

---

## Form Components

Location: `internal/ui/form_components.go`

Enhanced with inline error display support.

### TextInput

```go
input := ui.NewTextInput("Email", "user@example.com")
input.Focused = true
input.Value = "test@email.com"

// Set validation error
input.SetError("Invalid email format")

// Clear error
input.ClearError()

rendered := input.View()
```

### Dropdown

```go
dropdown := ui.NewDropdown("Theme", []string{"light", "dark", "auto"}, 0)
dropdown.Focused = true

// Set validation error
dropdown.SetError("Theme selection is required")

// Clear error
dropdown.ClearError()

rendered := dropdown.View()
```

### Other Form Components

- `Checkbox` - Single checkbox
- `RadioGroup` - Radio button group
- `CheckboxGroup` - Multiple checkboxes
- `Button` - Clickable button
- `HelpText` - Helper text

All form components integrate with the global theme system and support focus states.

---

## Usage Examples

### Creating a Custom Modal with Validation

```go
// Create modal with form inputs
inputs := []components.ModalInput{
    components.NewModalInput("Branch Name", "feature/...", ""),
}

buttons := []components.ModalButton{
    {
        Label: "Create",
        Primary: true,
        OnSelect: func() {
            values := modal.GetInputValues()
            branchName := values["Branch Name"]

            // Validate
            if branchName == "" {
                // Show error in modal
                return
            }

            // Create branch
            createBranch(branchName)
        },
    },
    {Label: "Cancel", Primary: false, OnSelect: func() { /* cancel */ }},
}

modal := components.NewFormModal(
    "Create Branch",
    "Enter a name for the new branch:",
    inputs,
    buttons,
)

// In Update() method
modal.Update(keyPress)
if inputMode {
    modal.UpdateInput(msg)
}

// In View() method
return modal.RenderCentered(windowWidth, windowHeight)
```

### Building a Dashboard Card

```go
// Create status card
changes := []string{
    "2 files modified",
    "1 file added",
}

card := components.NewDashboardCard(
    "Repository Status",
    strings.Join(changes, "\n"),
    30,
    10,
)

if isActive {
    card.SetActive(true)
}

rendered := card.Render()
```

### Error Handling Pattern

```go
// Show error banner at top of view
if err != nil {
    banner := components.NewErrorBanner(err.Error()).
        WithTitle("Operation Failed").
        WithActions(
            "Check your network connection",
            "Verify repository permissions",
        ).
        WithWidth(windowWidth - 4)

    content := banner.Render() + "\n\n" + mainContent
    return content
}

// Or use modal for blocking errors
if criticalError != nil {
    modal := components.NewErrorModal(criticalError.Error())
    return modal.RenderCentered(windowWidth, windowHeight)
}
```

### Consistent Layout

```go
// Use layout constants for spacing
content := lipgloss.NewStyle().
    Padding(layout.SpacingMD).
    Margin(layout.SpacingSM).
    Render(innerContent)

// Calculate split widths
leftWidth, rightWidth := layout.CalculateSplitWidths(windowWidth, layout.SplitRatio35_65)

leftPanel := renderLeftPanel(leftWidth)
rightPanel := renderRightPanel(rightWidth)

view := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
```

---

## Design Principles

1. **Consistency**: All components use the same styling patterns and theme system
2. **Reusability**: Components are self-contained and can be used anywhere
3. **Accessibility**: Keyboard navigation works consistently across all components
4. **Theming**: All components respect the global theme and update instantly
5. **Composability**: Components can be nested and combined
6. **Validation**: Form components support inline error display
7. **Documentation**: All public APIs are documented with examples

---

## Migration Guide

### Replacing Old Modal Implementations

**Before:**
```go
// Custom modal in view file
func renderCustomModal() string {
    // 50+ lines of duplicate code
    modalStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("#1a1a1a")).  // Hardcoded!
        // ... more styling
}
```

**After:**
```go
modal := components.NewConfirmModal(title, message, onYes, onNo)
return modal.RenderCentered(width, height)
```

### Replacing Magic Numbers

**Before:**
```go
viewportHeight := msg.Height - 15  // What is 15?
```

**After:**
```go
viewportHeight := layout.CalculateViewportHeight(msg.Height)
```

### Replacing Hardcoded Colors

**Before:**
```go
style := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a"))
```

**After:**
```go
styles := ui.GetGlobalThemeManager().GetStyles()
style := lipgloss.NewStyle().Background(styles.BackgroundModal)
```

---

## Contributing

When adding new components:

1. Follow existing patterns and naming conventions
2. Integrate with the global theme system
3. Support keyboard navigation
4. Add inline documentation
5. Include usage examples in this README
6. Write tests for component logic

---

## Questions?

See the individual component files for more detailed documentation and implementation details.
