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

### Commit Workflow (`gm commit`)

1. **AnalyzeCommitUseCase** (`internal/usecase/analyze_commit.go`):
   - Validate git repo
   - Get repository status (changed files, branch)
   - Get BranchInfo (parent, type, commits)
   - Get diff (staged + unstaged)
   - **Special:** Read untracked file contents directly (without staging) to preserve clean state
   - Get scoped commit history (commits on this branch only, if parent exists)
   - **Check merge opportunity**: If clean + 3+ commits + has parent → set `MergeOpportunity = true`
   - Call AI with full context
   - Return `AnalysisResponse` with `Decision`

2. **CommitViewModel** (`internal/ui/commit_view.go`):
   - Display TUI with AI suggestion
   - User selects action (arrow keys, Enter confirms)
   - If `ActionMerge` selected → launch merge workflow

3. **ExecuteCommitUseCase** (`internal/usecase/execute_commit.go`):
   - Execute selected action:
     - `ActionCommitDirect`: Stage all + commit
     - `ActionCreateBranch`: Create branch, set parent, checkout, stage, commit
     - `ActionReview`: No-op, user reviews manually
     - `ActionMerge`: Redirect to merge workflow

### Merge Workflow (`gm merge` or from commit)

1. **AnalyzeMergeUseCase** (`internal/usecase/analyze_merge.go`):
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

2. **MergeViewModel** (`internal/ui/merge_view.go`):
   - Display merge info (source → target)
   - Show commits being merged
   - Display AI-generated merge message
   - User selects strategy
   - Show conflict warnings if detected

3. **ExecuteMergeUseCase** (`internal/usecase/execute_merge.go`):
   - Checkout target branch
   - Execute merge with selected strategy
   - Handle conflicts (abort merge on failure)
   - Return merge commit hash

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
| `cmd/gm/main.go` | CLI entry point, command registration, workflow orchestration |
| `internal/domain/branch.go` | BranchInfo, BranchType, parent tracking |
| `internal/domain/decision.go` | ActionType, Decision, Alternative |
| `internal/usecase/analyze_commit.go` | Commit analysis workflow, merge detection |
| `internal/usecase/analyze_merge.go` | Merge analysis workflow, branch validation |
| `internal/adapter/ai/cerebras.go` | AI provider implementation, prompt engineering |
| `internal/adapter/git/exec.go` | Git command execution, branch operations |
| `internal/ui/commit_view.go` | Commit decision TUI |
| `internal/ui/merge_view.go` | Merge strategy selection TUI |

## Development Workflow

When making changes:

1. **Understand the layer:** Changes should respect architecture boundaries
2. **Domain changes:** Update domain entities first, then propagate upward
3. **New features:** Start with use case, add to UI, wire in main.go
4. **AI changes:** Modify prompts/schema in cerebras.go, test with real API
5. **Always rebuild:** `go build -o bin/gm.exe ./cmd/gm` after changes
6. **Test in real repo:** Use GitMind on itself for dogfooding
