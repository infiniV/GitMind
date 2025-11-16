# Testing GitMind

This guide explains how to test GitMind in different scenarios.

## Prerequisites

1. Build the binary: `go build -o gm.exe ./cmd/gm`
2. Have a Cerebras API key ready (get one free at https://cloud.cerebras.ai/)

## Test Scenario 1: First Time Setup

**What to test**: Configuration wizard

```bash
.\gm.exe config
```

**Expected behavior**:
- Shows configuration wizard with clear prompts
- Accepts your API key
- Saves configuration to `~/.gitman.json`
- Shows success message with path to config file

**How to verify**:
- Check that `~/.gitman.json` was created
- File should contain your settings

## Test Scenario 2: Simple Commit

**What to test**: Basic commit workflow with small changes

**Steps**:
1. Create a test directory or use an existing git repo
   ```bash
   mkdir test-repo
   cd test-repo
   git init
   git config user.name "Test User"
   git config user.email "test@example.com"
   ```

2. Make a small change
   ```bash
   echo "console.log('hello world');" > hello.js
   ```

3. Run GitMind
   ```bash
   ..\gm.exe commit
   ```

**Expected behavior**:
- Shows "Analyzing your changes..." message
- Displays TUI with:
  - Repository info (path, branch, changes summary)
  - AI-suggested commit message
  - Action options with confidence levels
  - Navigation instructions
- Arrow keys change selection
- Enter confirms
- Esc/Q cancels

**What to check**:
- ✓ Does the commit message make sense?
- ✓ Is the confidence level reasonable (>50%)?
- ✓ Can you navigate with arrow keys?
- ✓ Does pressing Enter execute the commit?
- ✓ Does `git log` show the new commit?

## Test Scenario 3: Multiple Files

**What to test**: AI analysis with multiple file changes

**Steps**:
1. In your test repo, create multiple files:
   ```bash
   echo "export function add(a, b) { return a + b; }" > math.js
   echo "export function subtract(a, b) { return a - b; }" > math2.js
   echo "import { add } from './math';" > index.js
   ```

2. Run GitMind
   ```bash
   ..\gm.exe commit
   ```

**Expected behavior**:
- AI should analyze all changes
- Commit message should reference multiple files or the common theme
- May suggest creating a branch if changes are significant

**What to check**:
- ✓ Does the message capture all changes?
- ✓ Is the reasoning sound?

## Test Scenario 4: With User Context

**What to test**: Using the `-m` flag to provide context

```bash
echo "// Fixed bug #42" > bugfix.js
..\gm.exe commit -m "This fixes the null pointer exception in the auth module"
```

**Expected behavior**:
- AI incorporates your context into the commit message
- Reasoning might reference your context

**What to check**:
- ✓ Does the commit message include your context?
- ✓ Is the message coherent?

## Test Scenario 5: Conventional Commits

**What to test**: Conventional commit format

```bash
echo "export const API_URL = 'https://api.example.com';" > config.js
..\gm.exe commit --conventional
```

**Expected behavior**:
- Commit message should follow format: `type(scope): description`
- Example: `feat: add API configuration`
- Example: `fix(auth): handle null pointer`

**What to check**:
- ✓ Message starts with a type (feat, fix, chore, etc.)?
- ✓ Follows conventional format?

## Test Scenario 6: No Changes

**What to test**: Error handling when no changes exist

```bash
# After committing everything
..\gm.exe commit
```

**Expected behavior**:
- Clear error message: "no changes to commit"
- No crash or confusing output

## Test Scenario 7: Not a Git Repo

**What to test**: Error handling in non-git directory

```bash
cd ..
mkdir not-a-repo
cd not-a-repo
..\gm.exe commit
```

**Expected behavior**:
- Clear error message: "not a git repository"
- Helpful suggestion to run `git init`

## Test Scenario 8: Free Tier Rate Limit

**What to test**: Handling rate limits gracefully

**Note**: This is hard to test without actually hitting rate limits

**If you hit a rate limit, check**:
- ✓ Friendly error message (not raw API error)
- ✓ Tells you to wait or upgrade
- ✓ Mentions how long to wait

## Test Scenario 9: Invalid API Key

**What to test**: Handling invalid credentials

**Steps**:
1. Edit `~/.gitman.json` and change API key to something invalid
2. Try to commit

**Expected behavior**:
- Clear error message about API key validation
- Suggestion to run `gm config`

## Test Scenario 10: Cancel Operation

**What to test**: User cancellation

```bash
echo "test" > test.txt
..\gm.exe commit
# Press Esc or Q in the TUI
```

**Expected behavior**:
- Shows "Operation cancelled" message
- No commit is created
- No errors

## Performance Tests

### Token Usage
Check the tokens used display at bottom of TUI:
- Small changes (1-2 files): Should use < 1000 tokens
- Medium changes (5-10 files): Should use 1000-3000 tokens
- Large changes (20+ files): Should reduce context and use < 3000 tokens

### Speed
- Analysis should complete in < 10 seconds on free tier
- TUI should respond instantly to keypresses

## What to Report

When testing, please note:

1. **What worked well**:
   - Which scenarios felt smooth?
   - Was the AI analysis helpful?
   - Did the TUI feel intuitive?

2. **What didn't work**:
   - Any crashes or errors?
   - Confusing messages?
   - Unexpected behavior?

3. **What could be better**:
   - Suggestions for improvements
   - Features you wish existed
   - UX issues

## Known Limitations (MVP)

- No push/pull integration (yet)
- No diff preview in TUI (coming)
- Simple config format (will upgrade to JSON)
- Windows-only testing (cross-platform coming)

---

**Thank you for testing GitMind! Your feedback is invaluable. 🙏**
