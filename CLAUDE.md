# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitMind (gm) is an AI-powered Git CLI manager built with Clean Architecture and Domain-Driven Design. It uses the Cerebras API (free tier optimized) to generate intelligent commit messages, suggest branching strategies, and automate merge workflows.

**Key Technologies:**
- Go 1.24+
- Bubble Tea (TUI framework)
- Cobra (CLI framework)
- Lipgloss (terminal styling)
- Cerebras AI API (llama-3.3-70b model)

## Build & Test Commands

```bash
# Build
go build -o bin/gm.exe ./cmd/gm

# Test all packages
go test -v ./...

# Test with coverage
go test -coverprofile=coverage.out ./...

# Test specific package
go test -v ./internal/domain
go test -v ./internal/usecase

# Run the application
bin/gm.exe commit
bin/gm.exe merge
bin/gm.exe config

# Test in a sandbox
mkdir test-repo && cd test-repo
git init
echo "test" > file.txt
../bin/gm.exe commit
```

## Architecture Overview

GitMind follows **Clean Architecture** with strict dependency rules flowing inward:

```
┌─────────────────────────────────────────────┐
│  cmd/gm (main.go)                           │  ← Entry point
│  - CLI commands via Cobra                   │
│  - Wires up dependencies                    │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│  internal/ui                                │  ← Presentation
│  - Bubble Tea TUI models                    │
│  - app_model.go: Root TUI coordinator       │
│  - dashboard_view.go: Main dashboard grid   │
│  - commit_view.go: Commit decision UI       │
│  - merge_view.go: Merge strategy selection  │
│  - styles.go: Lipgloss theme                │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│  internal/usecase                           │  ← Application Logic
│  - analyze_commit.go: AI commit analysis    │
│  - execute_commit.go: Execute git commits   │
│  - analyze_merge.go: AI merge analysis      │
│  - execute_merge.go: Execute git merges     │
└──────────────────┬──────────────────────────┘
                   │
      ┌────────────┴────────────┐
      ▼                         ▼
┌─────────────────┐     ┌──────────────────┐
│ internal/domain │     │ internal/adapter │  ← Interfaces & External
│  - Repository   │     │  /ai             │
│  - BranchInfo   │     │  - provider.go   │
│  - Decision     │     │  - cerebras.go   │
│  - CommitMsg    │     │  /git            │
│  - APIKey       │     │  - operations.go │
│                 │     │  - exec.go       │
│                 │     │  /config         │
│                 │     │  - config.go     │
└─────────────────┘     └──────────────────┘
    Pure business         External integrations
    entities & rules      (Git, AI, Config)
```

**Dependency Rule:** Code can only depend on layers inside it, never outside.

## UI Architecture

### Unified TUI Session

GitMind uses a **unified Bubble Tea session** managed by `AppModel` that keeps the dashboard visible throughout all workflows. This eliminates process exits and enables seamless circular navigation.

**State Machine Pattern:**
```go
type AppState int

const (
    StateDashboard AppState = iota
    StateCommitAnalyzing
    StateCommitView
    StateCommitExecuting
    StateMergeAnalyzing
    StateMergeView
    StateMergeExecuting
)
```

**Key components:**
- `AppModel` (app_model.go): Root coordinator managing all child views and state transitions
- `DashboardModel` (dashboard_view.go): 2x3 grid of cards showing repo status, actions
- `CommitViewModel` (commit_view.go): Commit decision interface (rendered as overlay)
- `MergeViewModel` (merge_view.go): Merge strategy selection (rendered as overlay)

### Navigation Flow

```
Dashboard (persistent)
    ↓ Select "Commit" card
StateCommitAnalyzing (loading overlay)
    ↓ AI analysis complete
StateCommitView (overlay)
    ↓ User selects option
StateCommitExecuting (loading overlay)
    ↓ Git commit complete
Dashboard (refresh)
```

**Key behaviors:**
- Dashboard always stays in memory (never exits)
- Loading states show animated overlay on dashboard
- Commit/merge views render as full-screen overlays
- `q`/`esc` quit from dashboard when no submenu active
- `esc` in views shows confirmation dialog before returning to dashboard

### Dashboard Design

The dashboard uses a **6-card grid layout** (3 columns × 2 rows):

**Top row:**
1. **REPOSITORY** - Current branch and status
2. **COMMIT** - AI-powered commit workflow
3. **MERGE** - AI-powered merge workflow

**Bottom row:**
4. **RECENT COMMITS** - Last 3 commits with hashes
5. **BRANCHES** - Branch list and count
6. **QUICK ACTIONS** - Help and refresh

**Card structure:**
- Fixed dimensions: 36 width (content) + 2 padding + 2 border = 40 total width
- Title at top (bold, primary color)
- Content at bottom (muted color, bottom-aligned)
- Selected card: orange border (`#C15F3C`)
- Unselected card: muted border

**Card rendering (dashboard_view.go:renderCard):**
```go
// Uses lipgloss.Place() to enforce exact dimensions
titleBlock := lipgloss.Place(36, 1, lipgloss.Left, lipgloss.Top, titleLine)
contentBlock := lipgloss.Place(36, 4, lipgloss.Left, lipgloss.Bottom, contentStr)
interior := titleBlock + spacer + contentBlock
```

### UI Design Principles

1. **No emojis** - Professional text-only interface
2. **Consistent card heights** - All cards same size using `lipgloss.Place()`
3. **Color-coded status** - Success (green), warning (orange), error (red), info (blue)
4. **Minimal text** - Concise labels, no redundant "Press Enter" prompts
5. **Bottom-aligned content** - Content anchored to bottom of card for clean hierarchy
6. **Instant transitions** - No animations, immediate state changes
7. **Claude Code color scheme** - Primary orange `#C15F3C`, muted gray `#787878`

### Loading States

During async operations (AI analysis, git execution), a loading overlay renders on top of dashboard:

```go
func (m AppModel) renderLoadingOverlay() string {
    dots := strings.Repeat(".", m.loadingDots)
    loadingText := loadingStyle.Render(m.loadingMessage + dots)
    return commitBoxStyle.Render(loadingText)
}
```

**Loading messages:**
- "Analyzing changes with AI..." (commit analysis)
- "Analyzing merge with AI..." (merge analysis)
- "Executing commit..." (git commit)
- "Executing merge..." (git merge)

Dots animate every 500ms for visual feedback.

## Core Domain Concepts

### 1. BranchInfo Intelligence

The `BranchInfo` domain entity encapsulates Git branch metadata:

- **Branch Type Detection**: Automatically detects `feature/*`, `hotfix/*`, `bugfix/*`, `refactor/*`, `release/*`, protected branches
- **Parent Tracking**: Stores parent branch in `git config branch.<name>.parent`
- **Scoped Commit History**: Counts commits unique to this branch (relative to parent)
- **Merge Target Suggestion**: Smart fallback logic (parent → main → master → develop)

**Key methods:**
```go
branchInfo.Parent()              // Get configured parent
branchInfo.CommitCount()         // Commits on this branch only
branchInfo.Type()                // BranchType enum
branchInfo.SuggestedMergeTarget() // Fallback merge target
```

### 2. Decision System

The AI returns a `Decision` which contains:
- **Action**: `ActionCommitDirect`, `ActionCreateBranch`, `ActionMerge`, `ActionReview`
- **SuggestedMessage**: AI-generated commit message
- **Reasoning**: Why this action was chosen
- **Alternatives**: Other viable options with confidence scores
- **BranchName**: Suggested name if creating branch

### 3. Merge Detection

When `gm commit` is run with:
- Clean working directory (no uncommitted changes)
- 3+ commits on current branch
- Branch has a configured parent

→ AI is informed of "merge opportunity" and can suggest `ActionMerge`

## AI Provider Architecture

### Interface-Based Design

All AI interactions go through the `Provider` interface (`internal/adapter/ai/provider.go`):

```go
type Provider interface {
    Analyze(ctx, AnalysisRequest) (*AnalysisResponse, error)
    GenerateMergeMessage(ctx, MergeMessageRequest) (*MergeMessageResponse, error)
    DetectTier(ctx) (domain.APITier, error)
    GetName() string
    ValidateKey(ctx) error
}
```

This allows swapping AI providers without changing business logic.

### Cerebras Implementation

Located in `internal/adapter/ai/cerebras.go`:

**Key features:**
- **Structured JSON output** via `response_format` parameter (enforces schema)
- **Token reduction** for free tier (reduces diff context when needed)
- **Retry logic** with exponential backoff
- **Free tier rate limit handling** (429 errors with retry-after)

**Prompt Engineering:**
- `buildPrompt()`: Constructs analysis prompt with branch context
- `buildMergePrompt()`: Constructs merge message prompt with commit history
- Both use branch intelligence (parent, commit count, type) to guide AI decisions

**Structured Output Schema:**
```json
{
  "commit_message": "string",
  "action": "commit-direct | create-branch | review | merge",
  "reasoning": "string",
  "alternatives": [...],
  "suggested_branch_name": "string (optional)"
}
```

### Adding a New AI Provider

1. Create `internal/adapter/ai/<provider>.go`
2. Implement `Provider` interface
3. Register in `Factory` (`provider.go:NewFactory()`)
4. Update config to support new provider name

## Git Operations Abstraction

All Git commands go through `git.Operations` interface (`internal/adapter/git/operations.go`).

**Critical Operations:**
- `GetBranchInfo()`: Returns `domain.BranchInfo` with parent, type, commit count
- `GetBranchCommits()`: Gets commits unique to branch (uses `git log branch1 ^branch2`)
- `GetParentBranch()` / `SetParentBranch()`: Uses `git config branch.<name>.parent`
- `CanMerge()`: Detects merge conflicts before executing
- `Merge()`: Executes merge with strategy (squash, regular, fast-forward, rebase)

**Implementation:**
- `internal/adapter/git/exec.go`: Uses `os/exec` to shell out to git commands
- Allows for future implementations (go-git, libgit2)

## Use Case Workflows

### Commit Workflow (`gm commit` or dashboard)

**Entry points:**
- CLI: `gm commit` → launches dashboard
- Dashboard: Select COMMIT card → triggers analysis

**Workflow:**

1. **User activates commit** (from dashboard or CLI)
   - Dashboard sets `action = ActionCommit`
   - AppModel transitions to `StateCommitAnalyzing`
   - Loading overlay displays: "Analyzing changes with AI..."

2. **AnalyzeCommitUseCase** (`internal/usecase/analyze_commit.go`):
   - Validate git repo
   - Get repository status (changed files, branch)
   - Get BranchInfo (parent, type, commits)
   - Get diff (staged + unstaged)
   - **Special:** Read untracked file contents directly (without staging) to preserve clean state
   - Get scoped commit history (commits on this branch only, if parent exists)
   - **Check merge opportunity**: If clean + 3+ commits + has parent → set `MergeOpportunity = true`
   - Call AI with full context
   - Return `AnalysisResponse` with `Decision`

3. **AppModel receives analysis result** (`commitAnalysisMsg`)
   - On error: Show error, return to dashboard
   - On success: Transition to `StateCommitView`
   - Create `CommitViewModel` with analysis results

4. **CommitViewModel renders as overlay** (`internal/ui/commit_view.go`):
   - Display AI suggestion and alternatives
   - Show confidence badges for each option
   - User navigates with arrow keys, Enter confirms
   - Esc shows confirmation dialog
   - Sets `hasDecision = true` when user selects

5. **AppModel detects decision**
   - Transitions to `StateCommitExecuting`
   - Loading overlay: "Executing commit..."
   - Calls `executeCommit()` with selected option

6. **ExecuteCommitUseCase** (`internal/usecase/execute_commit.go`):
   - Execute selected action:
     - `ActionCommitDirect`: Stage all + commit
     - `ActionCreateBranch`: Create branch, set parent, checkout, stage, commit
     - `ActionReview`: No-op, user reviews manually
     - `ActionMerge`: Redirect to merge workflow

7. **AppModel receives execution result** (`commitExecutionMsg`)
   - On error: Show error message
   - On success: Show success message
   - Return to `StateDashboard`
   - Dashboard refreshes to show new commit

### Merge Workflow (`gm merge` or dashboard)

**Entry points:**
- CLI: `gm merge` → launches dashboard
- Dashboard: Select MERGE card → triggers analysis
- CommitView: Select merge action → triggers analysis

**Workflow:**

1. **User activates merge**
   - Dashboard sets `action = ActionMerge`
   - AppModel transitions to `StateMergeAnalyzing`
   - Loading overlay: "Analyzing merge with AI..."

2. **AnalyzeMergeUseCase** (`internal/usecase/analyze_merge.go`):
   - Get source branch (current or specified via `-s`)
   - Determine target branch with fallback logic:
     - Use `-t` flag if provided
     - Try configured parent (if exists)
     - Try common branches: `main` → `master` → `develop` → `development`
     - Use `SuggestedMergeTarget()` fallback
   - Get commits to merge (`git log target..source`)
   - Check for conflicts (`CanMerge()`)
   - Call AI `GenerateMergeMessage()` with commit list
   - AI suggests strategy:
     - **Squash**: 5+ commits (clean history)
     - **Regular**: 1-4 commits (preserve history)
     - **Fast-forward**: Linear history, no divergence

3. **AppModel receives analysis result** (`mergeAnalysisMsg`)
   - On error: Show error, return to dashboard
   - On success: Transition to `StateMergeView`
   - Create `MergeViewModel` with analysis results

4. **MergeViewModel renders as overlay** (`internal/ui/merge_view.go`):
   - Display merge info (source → target)
   - Show commits being merged (up to 5)
   - Display AI-generated merge message
   - Show recommended strategy
   - User selects strategy with arrow keys
   - Esc shows confirmation dialog
   - Show conflict warnings if detected
   - Sets `hasDecision = true` when user confirms

5. **AppModel detects decision**
   - Transitions to `StateMergeExecuting`
   - Loading overlay: "Executing merge..."
   - Calls `executeMerge()` with selected strategy

6. **ExecuteMergeUseCase** (`internal/usecase/execute_merge.go`):
   - Checkout target branch
   - Execute merge with selected strategy
   - Handle conflicts (abort merge on failure)
   - Return merge commit hash

7. **AppModel receives execution result** (`mergeExecutionMsg`)
   - On error: Show error message
   - On success: Show success message with commit hash
   - Return to `StateDashboard`
   - Dashboard refreshes to show merge result

## Configuration

**Location:** `~/.gitman.json` (cross-platform via `os.UserHomeDir()`)

**Structure:**
```json
{
  "ai_provider": "cerebras",
  "api_key": "csk-...",
  "api_tier": "free",
  "default_model": "llama-3.3-70b",
  "use_conventional_commits": true,
  "protected_branches": ["main", "master", "develop"]
}
```

**Tier-specific behavior:**
- Free tier: Reduces diff context for large changes, graceful 429 handling
- Pro tier: Full context, higher rate limits

## Testing Patterns

**Domain Tests:**
- Pure unit tests (no dependencies)
- Table-driven tests for branch type detection
- Example: `internal/domain/decision_test.go`

**Use Case Tests:**
- Use mock implementations of `git.Operations` and `ai.Provider`
- Test business logic in isolation
- Example: `internal/usecase/analyze_commit_test.go` (if exists)

**Integration Tests:**
- Located in `test/` directory (if they exist)
- Test full workflows with real git repos in temp directories

**To run specific tests:**
```bash
go test -v ./internal/domain -run TestDetectBranchType
go test -v ./internal/usecase -run TestAnalyzeCommit
```

## Important Constraints & Design Decisions

### Untracked Files Strategy

**Problem:** AI needs to analyze untracked files, but staging them would pollute git state if user chooses "create branch" action.

**Solution:** `analyze_commit.go:buildUntrackedFilesDiff()` reads file contents directly from filesystem and constructs synthetic diff WITHOUT staging. This preserves clean state for branching.

### Parent Branch Persistence

**Problem:** Git doesn't natively track parent branches.

**Solution:** Store in `git config branch.<name>.parent`. This persists across sessions and is repository-local.

### Free Tier Optimization

**Context Reduction:** When changeset is large or free tier detected:
- `reduceDiffContext()` truncates diff to fit token limits
- Preserves important parts (file names, function signatures)
- Maintains commit message quality

**Rate Limit Handling:**
- Cerebras returns 429 with `retry-after` header
- Convert to `FreeTierLimitError` with friendly message
- UI shows "wait N seconds" instead of raw API error

### Merge Target Fallback Logic

Tries in order:
1. Specified `-t` flag
2. Configured parent (from git config)
3. Common branches that exist: `main`, `master`, `develop`, `development`
4. `SuggestedMergeTarget()` from BranchInfo
5. Error with helpful message listing available branches

## Common Modifications

### Adding a New Action Type

1. Add to `internal/domain/decision.go`:
   ```go
   const ActionNewType ActionType = iota + 5
   ```
2. Update `String()` method
3. Update `mapActionType()` in `cerebras.go`
4. Update JSON schema enum in `buildStructuredRequest()`
5. Add case in `commit_view.go` label functions
6. Handle in `execute_commit.go` switch statement

### Adding a New Command

1. Create command function in `cmd/gm/main.go`:
   ```go
   func newCmd() *cobra.Command {
       cmd := &cobra.Command{
           Use: "new",
           RunE: func(cmd *cobra.Command, args []string) error {
               return runNew()
           },
       }
       return cmd
   }
   ```
2. Register in `main()`: `rootCmd.AddCommand(newCmd())`
3. Implement `runNew()` function
4. Create use cases in `internal/usecase/` if needed
5. Create TUI view in `internal/ui/` if needed

### Modifying AI Prompt

Edit `internal/adapter/ai/cerebras.go`:
- `buildPrompt()`: For commit analysis
- `buildMergePrompt()`: For merge messages
- `buildStructuredRequest()`: For JSON schema changes

**Important:** If changing schema, update `parseResponse()` and `parseMergeResponse()` accordingly.

## Windows-Specific Considerations

- Paths use `filepath.Join()` for cross-platform compatibility
- Git commands executed via `exec.Command("git", ...)` work on Windows (Git Bash not required)
- Binary output: `bin/gm.exe` (Windows) vs `bin/gm` (Unix)
- Config stored in user home directory via `os.UserHomeDir()` (works on Windows)

## Key Files Reference

| File | Purpose |
|------|---------|
| `cmd/gm/main.go` | CLI entry point, command registration, launches unified dashboard |
| `internal/ui/app_model.go` | Root TUI coordinator, state machine, async operation handling |
| `internal/ui/dashboard_view.go` | 6-card dashboard grid, navigation, card rendering |
| `internal/ui/commit_view.go` | Commit decision overlay, option selection |
| `internal/ui/merge_view.go` | Merge strategy selection overlay |
| `internal/ui/styles.go` | Lipgloss styles, color scheme, badges, separators |
| `internal/domain/branch.go` | BranchInfo, BranchType, parent tracking |
| `internal/domain/decision.go` | ActionType, Decision, Alternative |
| `internal/usecase/analyze_commit.go` | Commit analysis workflow, merge detection |
| `internal/usecase/analyze_merge.go` | Merge analysis workflow, branch validation |
| `internal/adapter/ai/cerebras.go` | AI provider implementation, prompt engineering |
| `internal/adapter/git/exec.go` | Git command execution, branch operations |

## Common UI Modifications

### Modifying Dashboard Cards

**Important:** All dashboard cards must return exactly **4 lines** of content to maintain consistent heights.

**Card render functions pattern:**
```go
func (m DashboardModel) renderMyCard() string {
    var lines []string

    // Early returns must have 4 lines
    if loading {
        lines = append(lines, "Loading...")
        lines = append(lines, "")
        lines = append(lines, "")
        lines = append(lines, "")
        return strings.Join(lines, "\n")
    }

    // Build content (max 4 lines)
    lines = append(lines, "Line 1")
    lines = append(lines, "Line 2")

    // Pad to exactly 4 lines
    for len(lines) < 4 {
        lines = append(lines, "")
    }

    return strings.Join(lines, "\n")
}
```

**The renderCard function uses lipgloss.Place() for consistency:**
- Title block: `Place(36, 1, Left, Top, title)` - 1 line at top
- Content block: `Place(36, 4, Left, Bottom, content)` - 4 lines at bottom
- Spacer: 2 blank lines between title and content
- Total interior: 7 lines (1 + 2 + 4)

**Common mistakes:**
- ❌ Returning variable number of lines based on content
- ❌ Using `Height()` style on card border (doesn't include padding)
- ❌ Splitting/joining strings in renderCard (breaks styled text)
- ✅ Always return exactly 4 lines from card render functions
- ✅ Use `lipgloss.Place()` with Bottom alignment for content
- ✅ Keep render functions simple: just build string arrays

### Adding a New Dashboard Card

1. Update grid size in `renderTopRow()` or `renderBottomRow()`
2. Create render function following 4-line pattern above
3. Add card to `renderCard()` call with appropriate index
4. Update `handleCardActivation()` for new card index
5. Add submenu if needed in `handleSubmenuSelection()`

### Modifying AppModel State Machine

To add a new workflow state:

1. Add state constant to `AppState` enum in app_model.go
2. Add message type for async results (e.g., `type newWorkflowMsg struct {...}`)
3. Add state transition in `Update()` method
4. Add case in `View()` to render overlay for new state
5. Add escape handling in `esc` key handler

**Example:**
```go
// 1. Add state
const (
    // ... existing states
    StateNewWorkflow AppState = iota + 7
)

// 2. Add message type
type newWorkflowMsg struct {
    result *SomeResult
    err    error
}

// 3. Handle in Update()
case newWorkflowMsg:
    if msg.err != nil {
        PrintError(msg.err)
        m.state = StateDashboard
        return m, m.dashboard.Init()
    }
    // Process result...
```

### Troubleshooting Card Height Issues

If cards have inconsistent heights:

1. **Check render functions return exactly 4 lines:**
   ```bash
   # Add debug logging
   lines := buildCardContent()
   fmt.Fprintf(os.Stderr, "Card lines: %d\n", len(strings.Split(strings.Join(lines, "\n"), "\n")))
   ```

2. **Verify Place() dimensions in renderCard:**
   - Content block must be `Place(36, 4, ...)`
   - Title block must be `Place(36, 1, ...)`

3. **Check for styled text breaking line counts:**
   - Styles don't add lines, but splitting on `\n` after styling can break counts
   - Build content as string arrays, join once, then style

4. **Verify padding logic:**
   - Spacer should be 2 newlines (`strings.Repeat("\n", 2)`)
   - Total: 1 (title) + 2 (spacer) + 4 (content) = 7 lines interior

## Development Workflow

When making changes:

1. **Understand the layer:** Changes should respect architecture boundaries
2. **Domain changes:** Update domain entities first, then propagate upward
3. **New features:** Start with use case, add to UI, wire in main.go
4. **AI changes:** Modify prompts/schema in cerebras.go, test with real API
5. **UI changes:** Follow 4-line card pattern, test with different content states
6. **Always rebuild:** `go build -o bin/gm.exe ./cmd/gm` after changes
7. **Test in real repo:** Use GitMind on itself for dogfooding
