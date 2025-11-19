# GitMind UI Design System

This document defines the design system, patterns, and guidelines for building consistent UI components in GitMind.

## Table of Contents

1. [Design Principles](#design-principles)
2. [Layout System](#layout-system)
3. [Spacing](#spacing)
4. [Typography](#typography)
5. [Colors & Themes](#colors--themes)
6. [Components](#components)
7. [Patterns](#patterns)
8. [Do's and Don'ts](#dos-and-donts)

---

## Design Principles

### 1. Consistency First
- Use established components from `internal/ui/components/`
- Never duplicate implementations
- Follow existing patterns for similar functionality

### 2. Theme Integration
- Always use the global theme system
- Never hardcode colors (`#1a1a1a` ❌)
- Access colors via `GetGlobalThemeManager().GetStyles()`

### 3. Layout Standards
- Use constants from `internal/ui/layout/constants.go`
- No magic numbers
- Document any new spacing values

### 4. Accessibility
- All components must support keyboard navigation
- Focus states must be clearly visible
- Provide keyboard shortcuts for common actions

### 5. Error Handling
- Always provide clear error messages
- Show recovery actions when possible
- Use inline validation for forms
- Display errors contextually, not just as modals

---

## Layout System

### Screen Structure

```
┌─────────────────────────────────────────┐
│ Logo (LogoHeight = 6)                   │ ← Header
├─────────────────────────────────────────┤
│                                         │
│                                         │
│ Main Content Area                       │ ← Content
│                                         │
│                                         │
├─────────────────────────────────────────┤
│ Footer (FooterHeight = 2)               │ ← Footer
└─────────────────────────────────────────┘
```

### Layout Constants

**Heights:**
```go
layout.HeaderHeight       = 8   // Logo + title area
layout.FooterHeight       = 2   // Keyboard shortcuts
layout.TabBarHeight       = 3   // Tab navigation
layout.StatusBarHeight    = 1   // Status line
layout.LogoHeight         = 6   // ASCII art logo
```

**Content Calculations:**
```go
// Available height for content
contentHeight := layout.CalculateContentHeight(windowHeight)
// windowHeight - HeaderHeight - FooterHeight

// With tabs
contentHeight := layout.CalculateContentHeightWithTabs(windowHeight)
// windowHeight - HeaderHeight - FooterHeight - TabBarHeight

// Viewport for scrolling
viewportHeight := layout.CalculateViewportHeight(windowHeight)
// windowHeight - LogoHeight - HeaderHeight - FooterHeight - SpacingMD
```

### Split Layouts

For two-column layouts:

```go
// 35% / 65% split (common for detail views)
leftWidth, rightWidth := layout.CalculateSplitWidths(
    totalWidth,
    layout.SplitRatio35_65,
)

// Other ratios
layout.SplitRatio40_60  // Balanced
layout.SplitRatio50_50  // Equal
```

**Example:**
```
┌─────────────┬───────────────────────────┐
│             │                           │
│ Left Panel  │ Right Panel               │
│ (35%)       │ (65%)                     │
│             │                           │
└─────────────┴───────────────────────────┘
```

---

## Spacing

### Spacing Scale

```go
layout.SpacingXS  = 1   // Tight spacing, rarely used
layout.SpacingSM  = 2   // Between related items
layout.SpacingMD  = 3   // Standard spacing (default)
layout.SpacingLG  = 4   // Between sections
layout.SpacingXL  = 6   // Major section breaks
```

### Usage Guidelines

**Padding:**
- Cards: `SpacingMD` (3)
- Modals: `SpacingMD` to `SpacingLG` (3-4)
- Buttons: `SpacingSM` horizontal (2)
- Form fields: `SpacingSM` (2)

**Margins:**
- Between cards: `SpacingMD` (3)
- Between sections: `SpacingLG` (4)
- Between major areas: `SpacingXL` (6)
- Between lines: `SpacingSM` (2)

**Example:**
```go
// Card with proper spacing
cardStyle := lipgloss.NewStyle().
    Padding(layout.SpacingMD).           // Internal padding
    Margin(layout.SpacingSM).            // External margin
    Border(lipgloss.RoundedBorder())
```

---

## Typography

### Text Hierarchy

```go
styles := ui.GetGlobalThemeManager().GetStyles()

// Primary heading
title := styles.Primary.Bold(true).Render("Title")

// Secondary heading
subtitle := styles.Text.Bold(true).Render("Subtitle")

// Body text
body := styles.Text.Render("Content")

// Muted/helper text
help := styles.Muted.Render("Helper text")

// Emphasis
important := styles.Text.Bold(true).Render("Important")
```

### Text Styles

| Style | When to Use | Example |
|-------|-------------|---------|
| `Bold` | Headings, labels, emphasis | "Repository Status" |
| `Italic` | Metadata, secondary info | "Last updated 2h ago" |
| `Underline` | Rarely (links in future) | Avoid |
| `Primary Color` | Active items, selections | Selected card title |
| `Muted Color` | Helper text, placeholders | Keyboard shortcuts |

### Text Length Guidelines

- **Modal titles**: Max 50 characters
- **Card titles**: Max 30 characters
- **Button labels**: Max 15 characters
- **Error messages**: 1-2 sentences, suggest action
- **Help text**: Concise, under 80 characters

---

## Colors & Themes

### Theme System

All colors come from the active theme. **Never hardcode colors.**

```go
styles := ui.GetGlobalThemeManager().GetStyles()

// Foreground colors
styles.Colors.Primary      // Primary brand color
styles.Colors.Secondary    // Secondary accent
styles.Colors.Text         // Body text
styles.Colors.Muted        // Subtle text
styles.Colors.Success      // Success states (green)
styles.Colors.Warning      // Warning states (yellow)
styles.Colors.Error        // Error states (red)
styles.Colors.Info         // Info states (blue)
styles.Colors.Border       // Borders and dividers
styles.Colors.Selected     // Selected items

// Background colors
styles.Backgrounds.Modal
styles.Backgrounds.Submenu
styles.Backgrounds.Dashboard
styles.Backgrounds.Confirmation
styles.Backgrounds.ErrorModal
styles.Backgrounds.FormInput
styles.Backgrounds.FormFocused
```

### Color Usage

**Primary Color:**
- Active/focused items
- Selected items
- Call-to-action buttons
- Logo and branding

**Success (Green):**
- Successful operations
- Clean repository status
- Passing tests/checks

**Warning (Yellow):**
- Warnings and cautions
- Uncommitted changes
- Non-critical issues

**Error (Red):**
- Errors and failures
- Destructive actions
- Validation errors

**Info (Blue):**
- Informational messages
- Hints and tips
- Non-blocking notices

**Muted:**
- Helper text
- Keyboard shortcuts
- Metadata
- Placeholders

---

## Components

### Modal Sizes

```go
// Small modal (50 wide, 10 high)
layout.ModalWidthSM, layout.ModalHeightSM
// Use for: Simple messages, quick confirmations

// Medium modal (60 wide, 15 high)
layout.ModalWidthMD, layout.ModalHeightMD
// Use for: Forms, detailed confirmations

// Large modal (70 wide, 20 high)
layout.ModalWidthLG, layout.ModalHeightLG
// Use for: Complex forms, multi-step processes
```

### Card Patterns

**Dashboard Cards:**
```go
card := components.NewDashboardCard(
    "Title",
    content,
    cardWidth,
    cardHeight,
)
card.SetActive(isSelected)
```

**Status Cards:**
```go
card := components.NewStatusCard(
    "Build Status",
    "✓",           // Icon
    "Passing",     // Status
    "Last run: 2m ago",  // Details
)
```

**Info Cards:**
```go
card := components.NewInfoCard("Details", []components.InfoItem{
    {Label: "Branch", Value: "main"},
    {Label: "Author", Value: "user@email.com"},
})
```

### Button Patterns

**Primary Button:**
```go
ModalButton{
    Label: "Confirm",
    Primary: true,
    OnSelect: handleConfirm,
}
```

**Secondary Button:**
```go
ModalButton{
    Label: "Cancel",
    Primary: false,
    OnSelect: handleCancel,
}
```

**Button Order:**
- Primary action on the left
- Cancel/back on the right
- Max 3 buttons per modal

---

## Patterns

### Error Display Pattern

**Inline Validation:**
```go
// Form field error
input := ui.NewTextInput("Email", "user@example.com")
if !valid {
    input.SetError("Invalid email format")
}
```

**Error Banner:**
```go
// Non-blocking error at top
banner := components.NewErrorBanner(err.Error()).
    WithActions("Check connection", "Retry")
```

**Error Modal:**
```go
// Blocking error
modal := components.NewErrorModal(criticalError.Error())
```

**When to Use Each:**
- **Inline**: Form validation, field-specific errors
- **Banner**: Non-critical errors, can continue working
- **Modal**: Critical errors, must acknowledge before proceeding

### Loading States

```go
// Show loading overlay
if isLoading {
    spinner := renderLoadingSpinner(message)
    return spinner
}
```

**Loading Messages:**
- "Analyzing changes..." (commit)
- "Analyzing branches..." (merge)
- "Saving settings..."
- "Connecting to repository..."

### Confirmation Pattern

```go
// Always confirm destructive actions
modal := components.NewConfirmModal(
    "Delete Branch?",
    "This action cannot be undone. Branch 'feature/xyz' will be permanently deleted.",
    onConfirm,
    onCancel,
)
```

**Confirm These Actions:**
- Delete (branch, commit, file)
- Force push
- Reset hard
- Discard changes
- Overwrite

### Footer Pattern

```go
// State-specific shortcuts
footer := components.NewFooter([]components.Shortcut{
    {Key: "↑↓", Description: "navigate"},
    {Key: "enter", Description: "select"},
    {Key: "esc", Description: "back"},
})

// Add metadata
footer.WithMetadata("main branch")
footer.WithWidth(windowWidth)
```

### Logo Pattern

```go
// Full logo with repo info
logo := components.RenderLogo(
    components.LogoDashboard,
    "my-project (main)",
)

// Compact for limited space
logo := components.RenderLogoCompact(components.LogoCommit)
```

---

## Do's and Don'ts

### Layout

✅ **DO:**
- Use layout constants for all spacing
- Calculate heights with helper functions
- Center modals on screen
- Respect minimum sizes (5 lines viewport)

❌ **DON'T:**
- Hardcode numbers like `msg.Height - 15`
- Use random padding values
- Assume terminal size
- Create layouts that break at small sizes

### Colors

✅ **DO:**
- Use theme colors for all styling
- Test with all available themes
- Use semantic colors (error = red)
- Provide good contrast

❌ **DON'T:**
- Hardcode hex colors (#1a1a1a)
- Use colors inconsistently
- Assume dark/light background
- Make text unreadable

### Components

✅ **DO:**
- Use existing components from `/components`
- Follow established patterns
- Support keyboard navigation
- Show focus states

❌ **DON'T:**
- Create duplicate components
- Ignore existing patterns
- Make components mouse-only
- Hide focus indicators

### Error Handling

✅ **DO:**
- Show clear error messages
- Suggest recovery actions
- Use inline validation
- Provide context

❌ **DON'T:**
- Show generic "Error occurred"
- Leave users stuck
- Validate only on submit
- Hide errors in logs

### Text

✅ **DO:**
- Keep text concise
- Use sentence case
- Wrap long text properly
- Show full errors

❌ **DON'T:**
- Write novels in modals
- Use ALL CAPS everywhere
- Let text overflow
- Truncate error messages

### Navigation

✅ **DO:**
- Support arrow keys AND hjkl
- Show keyboard shortcuts
- Allow Esc to go back
- Tab through inputs

❌ **DON'T:**
- Require mouse
- Hide shortcuts
- Trap users in views
- Make navigation inconsistent

---

## Common Mistakes

### 1. Duplicating Components

**Bad:**
```go
// In commit_view.go
func renderMyCustomModal() string {
    // 80 lines of modal code
}

// In merge_view.go
func renderAnotherModal() string {
    // 80 lines of nearly identical code
}
```

**Good:**
```go
modal := components.NewConfirmModal(...)
return modal.RenderCentered(width, height)
```

### 2. Hardcoded Values

**Bad:**
```go
viewportHeight := msg.Height - 17  // Magic number!
style := lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a"))
padding := 3  // Why 3?
```

**Good:**
```go
viewportHeight := layout.CalculateViewportHeight(msg.Height)
style := lipgloss.NewStyle().Background(styles.BackgroundModal)
padding := layout.SpacingMD
```

### 3. Inconsistent Error Handling

**Bad:**
```go
if err != nil {
    return "ERROR: " + err.Error()  // Inconsistent formatting
}
```

**Good:**
```go
if err != nil {
    banner := components.NewErrorBanner(err.Error())
    return banner.Render()
}
```

### 4. Breaking Theme System

**Bad:**
```go
// Directly styling without theme
textStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("orange")).
    Background(lipgloss.Color("black"))
```

**Good:**
```go
styles := ui.GetGlobalThemeManager().GetStyles()
textStyle := styles.Primary  // Respects active theme
```

---

## Component Checklist

Before creating a new component, verify:

- [ ] Similar component doesn't exist in `/components`
- [ ] Uses theme colors, no hardcoded values
- [ ] Uses layout constants for spacing
- [ ] Supports keyboard navigation
- [ ] Shows focus states
- [ ] Has inline documentation
- [ ] Includes usage example
- [ ] Added to components README
- [ ] Tested with all themes
- [ ] Tested at different terminal sizes

---

## Questions & Updates

This design system evolves as the application grows. When in doubt:

1. Check existing components in `internal/ui/components/`
2. Reference this document
3. Look at how similar components handle it
4. Ask: "Does this maintain consistency?"

**For updates to this document:**
- Document new patterns as they emerge
- Add examples from real usage
- Keep do's and don'ts up to date
- Maintain the component checklist

---

Last Updated: 2025-11-19
