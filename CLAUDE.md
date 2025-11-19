# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GitMind** is an AI-powered Git CLI manager built with Go and Bubble Tea TUI framework. It uses AI (Cerebras free tier) to intelligently analyze commits and merges, suggest workflows, and automate Git operations.

- **Binary:** `gitmind` (formerly `gm`)
- **Language:** Go 1.24.2
- **Architecture:** Clean Architecture + Domain-Driven Design
- **UI Framework:** Bubble Tea (charmbracelet)
- **Module:** `github.com/yourusername/gitman`

## Build & Run Commands

```bash
# Build
make build                    # Outputs: bin/gitmind.exe
go build -o bin/gitmind.exe ./cmd/gm

# Test
go test ./...                 # All tests
go test -v ./...              # Verbose
go test ./internal/domain/... # Specific package

# Run
gitmind                       # Launch dashboard
gitmind commit                # Commit workflow
gitmind merge                 # Merge workflow
gitmind config                # Settings
gitmind onboard               # Setup wizard
gitmind --version
```

## Clean Architecture Layers

### Domain (`internal/domain/`) - Zero Dependencies
Core business entities with no external dependencies:

- **`apikey.go`** - API key management with tier detection (Free/Pro)
- **`decision.go`** - AI decision framework with confidence levels (High ≥0.8, Medium 0.5-0.8, Low <0.5)
- **`commit.go`** - Commit message domain with conventional commits validation
- **`config.go`** - Complete configuration structure (197 lines)
- **`branch.go`** - Branch information and metadata
- **`repository.go`** - Repository state representation
- **`theme.go`** - Theme system (ThemeColors, ThemeBackgrounds)

**Action Types:** CommitDirect, CreateBranch, SplitCommits, Review, Merge

### Use Case (`internal/usecase/`) - Application Logic
Workflows orchestrating domain and adapters:

- **`analyze_commit.go`** - Orchestrates AI commit analysis (90s timeout)
- **`execute_commit.go`** - Executes commit actions (direct commit or branch creation)
- **`analyze_merge.go`** - AI merge strategy suggestions with conflict detection
- **`execute_merge.go`** - Executes merge (squash, regular, fast-forward)

**Context Timeouts:** 90s for AI analysis, 120s for git operations

### Adapter (`internal/adapter/`) - External Integrations

#### Git Adapter (`adapter/git/`)
- **`operations.go`** - Git Operations interface (27 operations)
- **`exec.go`** - Implementation using os/exec + git.exe
- Tracks parent branches via git config
- Detects merge conflicts before attempting merge
- Calculates ahead/behind commit counts

#### AI Adapter (`adapter/ai/`)
- **`provider.go`** - Provider interface with Factory pattern
- **`cerebras.go`** - Cerebras AI (currently only fully implemented provider)
- Free tier optimization: reduces context, handles rate limits with FreeTierLimitError
- Token counting and usage tracking

#### GitHub Adapter (`adapter/github/`)
- **`operations.go`** - Repository creation via `gh` CLI
- CreateRepoOptions: visibility, license, .gitignore, issues, wiki, projects

#### Config Adapter (`adapter/config/`)
- **`config.go`** - JSON persistence at `~/.gitman.json`
- Automatic migration from legacy key=value format

### UI (`internal/ui/`) - Bubble Tea Components

#### Main Views
- **`app_model.go`** - Unified TUI entry (600 lines)
- **`dashboard_view.go`** - 6-card grid (status, commits, branches, quick actions)
- **`commit_view.go`** - Commit analysis and execution workflow
- **`merge_view.go`** - Merge analysis and execution workflow
- **`settings_view.go`** - 5 nested tabs (Git, GitHub, Commits, Naming, AI)
- **`onboarding_*.go`** - 8-step wizard (8 files)

#### Component Library (`ui/components/`)
**CRITICAL:** Always use these instead of creating duplicates:

- **`modal.go`** - 5 types (Info, Error, Warning, Confirm, Form) with keyboard nav
- **`card.go`** - 6 types (Default, Primary, Success, Warning, Error, Info)
- **`error.go`** - ErrorBanner, ValidationError with severity levels
- **`logo.go`** - 5 variants (Dashboard, Commit, Merge, Settings, Onboard)
- **`footer.go`** - Pre-built footers with keyboard shortcuts
- **`text_utils.go`** - 20+ utilities (wrap, truncate, padding, pluralize)

#### Layout System (`ui/layout/constants.go`)
**CRITICAL:** Never use magic numbers. Always use layout constants:

```go
// Spacing scale
layout.SpacingXS  // 1
layout.SpacingSM  // 2
layout.SpacingMD  // 3 (default)
layout.SpacingLG  // 4
layout.SpacingXL  // 6

// Modal sizes
layout.ModalWidthSM   // 50
layout.ModalWidthMD   // 60
layout.ModalWidthLG   // 70

// Split ratios
layout.SplitRatio35_65
layout.SplitRatio50_50

// Helper functions
layout.CalculateContentHeight(windowHeight)
layout.CalculateSplitWidths(totalWidth, ratio)
```

#### Theme System
**CRITICAL:** Never hardcode colors. Always use the theme system:

```go
styles := ui.GetGlobalThemeManager().GetStyles()

// Colors
styles.ColorPrimary
styles.ColorSuccess
styles.ColorError
styles.ColorWarning
styles.ColorMuted
styles.ColorText
styles.ColorBorder

// Status styles
styles.StatusOk
styles.StatusError
styles.StatusWarning
styles.StatusInfo

// Get theme backgrounds
theme := ui.GetGlobalThemeManager().GetCurrentTheme()
theme.Backgrounds.Modal
theme.Backgrounds.Confirmation
```

**Available Themes (8):** claude-warm, ocean-blue, forest-green, monochrome, magma, viridis, plasma, twilight

## Configuration System

**Location:** `~/.gitman.json`

**Structure (domain.Config):**
- **Git:** main_branch, protected_branches, auto_push, auto_pull
- **GitHub:** enabled, default_visibility, default_license, default_gitignore, enable_issues/wiki/projects
- **Commits:** convention (conventional/custom/none), types, require_scope, require_breaking
- **Naming:** enforce, pattern, allowed_prefixes
- **AI:** provider, api_key, api_tier (free/pro), default_model, fallback_model, max_diff_size, include_context
- **UI:** theme

**Migration:** Automatic from legacy format with backup at `~/.gitman.json.backup`

## Critical Design Patterns

### AI Integration Flow
```
User Action → Use Case → Git Adapter (diff) → AI Provider → Decision → UI → Execute
```

### Decision Confidence System
- **High (≥0.8):** AI very confident, execute directly
- **Medium (0.5-0.8):** AI suggests alternatives
- **Low (<0.5):** Requires manual review

### Dependency Inversion
- Domain has zero dependencies
- Use Cases depend only on domain interfaces
- Adapters implement domain ports
- UI depends on use cases and domain

### Provider Pattern
AI providers are pluggable via Factory pattern:
```go
provider := ai.NewProvider(config.AI.Provider, config.AI.APIKey, config.AI.APITier)
```

## Development Guidelines

### Adding New UI Components
1. Create in `internal/ui/components/`
2. Use layout constants from `layout/constants.go`
3. Integrate with global theme system
4. Add to `components/README.md`
5. Follow existing patterns (New*, Render, Update methods)

### Adding New Git Operations
1. Add method to `GitOperations` interface in `adapter/git/operations.go`
2. Implement in `adapter/git/exec.go`
3. Use context for cancellation
4. Return domain objects, not raw strings

### Adding New Themes
1. Define in `internal/ui/themes.go`
2. Include ThemeColors and ThemeBackgrounds
3. Add to theme manager's available themes
4. Test with all UI components

### Adding New AI Providers
1. Implement `AIProvider` interface in `adapter/ai/`
2. Add to Factory in `provider.go`
3. Support tier detection and token tracking
4. Handle rate limits gracefully

## Testing

**Test Files (13):**
- Domain: `apikey_test.go`, `commit_test.go`, `decision_test.go`, `repository_test.go`
- Adapter: `exec_test.go` (git operations)
- UI: `form_components_test.go`, `onboarding_*_test.go`, `checkbox_integration_test.go`

**Patterns:**
- Table-driven tests for domain logic
- Context-based timeout testing
- Integration tests for UI components

## Key Constraints

**DO NOT:**
- Hardcode hex colors (`#1a1a1a`) - use theme system
- Use magic numbers - use layout constants
- Duplicate components - use component library
- Create modals from scratch - use `components.NewModal*()`
- Assume git operations succeed - check errors

**ALWAYS:**
- Use `ui.GetGlobalThemeManager().GetStyles()` for colors
- Use layout constants for spacing and sizes
- Follow Clean Architecture separation
- Check context cancellation in long operations
- Validate configuration before use

## Current Implementation Status

**Implemented:**
- AI commit message generation with smart workflow decisions
- Merge strategy suggestions (squash, regular, fast-forward)
- 8-step onboarding wizard
- Tabbed dashboard with 6 cards
- Settings editor with 5 nested tabs
- GitHub repository creation
- 8 professional themes
- Complete component library
- Protected branch management
- Conventional commits support

**NOT Implemented (see docs/ROADMAP.md):**
- Git push (commented but not implemented)
- Pull request creation/management
- Branch deletion
- Diff viewing in UI (stats only)
- Stash operations
- Multiple AI providers (only Cerebras fully implemented)

## Important Files

**Entry Point:**
- `cmd/gm/main.go` - CLI setup, dashboard launcher (600 lines)

**Critical Domain:**
- `internal/domain/decision.go` - AI decision framework (237 lines)
- `internal/domain/config.go` - Complete configuration (197 lines)

**Critical Adapters:**
- `internal/adapter/git/operations.go` - Git interface (27 operations)
- `internal/adapter/ai/provider.go` - AI provider interface + factory

**Critical UI:**
- `internal/ui/app_model.go` - Unified TUI entry
- `internal/ui/components/*.go` - Component library (8 files)

**Documentation:**
- `docs/UI_DESIGN_SYSTEM.md` - Complete design system (636 lines)
- `docs/UI_CONSISTENCY_IMPLEMENTATION.md` - Implementation guide (745 lines)
- `internal/ui/components/README.md` - Component API (619 lines)

## Keyboard Navigation

**Dashboard:**
- `1`/`2` - Switch main tabs
- `Tab` - Navigate cards
- `Enter` - Select
- `R` - Refresh
- `Q` - Quit

**Settings:**
- `←`/`→` - Switch tabs
- `↑`/`↓` - Navigate fields
- `Enter` - Toggle/edit
- `Ctrl+S` - Save
- `Esc` - Back

**Modals:**
- `Tab`/`Shift+Tab` - Navigate inputs/buttons
- `←`/`→` - Switch buttons
- `Enter` - Confirm
- `Esc` - Cancel

## Common Development Tasks

### Creating a Modal
```go
import "github.com/yourusername/gitman/internal/ui/components"

// Simple error
modal := components.NewErrorModal("Something went wrong")

// Confirmation
modal := components.NewConfirmModal(
    "Delete Branch?",
    "This cannot be undone.",
    onYes,
    onNo,
)

// Render
return modal.RenderCentered(windowWidth, windowHeight)
```

### Using Layout Constants
```go
import "github.com/yourusername/gitman/internal/ui/layout"

// Instead of hardcoded 3
padding := layout.SpacingMD

// Instead of msg.Height - 15
viewportHeight := layout.CalculateViewportHeight(msg.Height)

// Instead of manual calculation
leftWidth, rightWidth := layout.CalculateSplitWidths(totalWidth, layout.SplitRatio35_65)
```

### Accessing Theme Colors
```go
styles := ui.GetGlobalThemeManager().GetStyles()

// Text styling
textStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

// Status indicators
errorStyle := styles.StatusError
successStyle := styles.StatusOk

// Backgrounds
theme := ui.GetGlobalThemeManager().GetCurrentTheme()
background := lipgloss.Color(theme.Backgrounds.Modal)
```

## Windows Support

- Binary builds as `gitmind.exe`
- Uses Windows-compatible paths
- Integrates with `git.exe` and `gh.exe` via os/exec
- Fully tested on Windows

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/spf13/cobra` - CLI framework

## Additional Documentation

- **Design System:** `docs/UI_DESIGN_SYSTEM.md` - Comprehensive UI guidelines
- **Implementation Guide:** `docs/UI_CONSISTENCY_IMPLEMENTATION.md` - Migration guide for consistency
- **Component API:** `internal/ui/components/README.md` - Complete component reference
- **Roadmap:** `docs/ROADMAP.md` - 30 prioritized features with rationale
