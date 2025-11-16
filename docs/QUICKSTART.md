# GitMind Quick Start Guide

Welcome to GitMind! This guide will help you get started in minutes.

## Prerequisites

1. **Git** - Make sure git is installed and accessible from command line
2. **Cerebras API Key** - Get a free API key:
   - Go to https://cloud.cerebras.ai/
   - Sign up for a free account
   - Navigate to API Keys section
   - Create a new API key
   - Copy the key (starts with something like `csk-...`)

## Installation

### Option 1: Use the Built Binary (Quickest)

The binary `gm.exe` is already built. You can:

1. Move `gm.exe` to a directory in your PATH, OR
2. Run it directly from this directory: `.\gm.exe`

### Option 2: Build from Source

```bash
go build -o gm.exe ./cmd/gm
```

## First Time Setup

Run the configuration wizard:

```bash
.\gm.exe config
```

You'll be prompted for:

1. **AI Provider**: Press Enter (defaults to "cerebras")
2. **API Key**: Paste your Cerebras API key
3. **API Tier**:
   - Press `1` for Free tier (recommended to start)
   - Press `2` if you have a Pro account
4. **Conventional Commits**: Type `y` if you want to use conventional commits format, or `n`
5. **Default Model**: Press `1` for llama-3.3-70b (recommended)

Your configuration will be saved to `~/.gitman.json`

## Using GitMind

### Basic Workflow

1. **Make changes to your code**
   ```bash
   # Edit some files...
   echo "console.log('hello')" > test.js
   ```

2. **Run GitMind commit**
   ```bash
   .\gm.exe commit
   ```

3. **Review AI suggestions**
   - GitMind analyzes your changes
   - Shows a suggested commit message
   - Recommends whether to commit directly or create a branch
   - Displays confidence level and reasoning

4. **Select an action**
   - Use arrow keys (↑/↓) to navigate options
   - Press Enter to confirm
   - Press Esc or Q to cancel

5. **Done!** GitMind executes the commit for you

### Advanced Usage

#### Add context to help the AI

```bash
.\gm.exe commit -m "This fixes the authentication bug reported in issue #42"
```

#### Force conventional commits format

```bash
.\gm.exe commit --conventional
```

or shorthand:

```bash
.\gm.exe commit -c
```

### What You'll See

The TUI (Text User Interface) shows:

```
╔══════════════════════════════════════════════════════════════════════╗
║                     GitMind - AI Commit Assistant                    ║
╚══════════════════════════════════════════════════════════════════════╝

Repository: D:\projects\gitman
Branch: main
Changes: 3 file(s) changed, +120 -15

Suggested Commit Message:
┌────────────────────────────────────────────────────────────────────┐
│ Add AI-powered commit analysis with interactive TUI
└────────────────────────────────────────────────────────────────────┘

Select an action:

▶ 1. ✓ Commit directly (confidence: 85%)
     The changes are well-contained improvements to existing features

  2. Create a new branch instead
     Consider isolating this feature for review

────────────────────────────────────────────────────────────────────
↑/↓: Navigate  Enter: Confirm  Esc/Q: Cancel
AI Model: llama-3.3-70b | Tokens: 1234
```

## Free Tier Limits

Cerebras free tier provides:
- **60,000 tokens per minute**
- **30 requests per minute**
- **14,400 requests per day**

This is plenty for regular use! If you hit rate limits:
- Wait 60 seconds before trying again
- Consider upgrading to Pro for higher limits
- GitMind automatically reduces context for free tier users

## Troubleshooting

### "not a git repository"
Make sure you're in a directory initialized with git:
```bash
git init
```

### "no changes to commit"
Make sure you have uncommitted changes:
```bash
git status
```

### "API key not configured"
Run the config wizard:
```bash
.\gm.exe config
```

### "API key validation failed"
- Check your API key is correct
- Ensure you have internet connection
- Verify the API key is active at https://cloud.cerebras.ai/

## Tips for Best Results

1. **Descriptive changes**: The AI works best with clear, focused changes
2. **Add context**: Use `-m "context"` to give the AI more information
3. **Review suggestions**: The AI is smart but always review before confirming
4. **Conventional commits**: Enable this for consistent commit message format
5. **Break up large changes**: If you have 50+ files changed, consider splitting commits

## Next Steps

- Try it on a real project!
- Experiment with different commit scenarios
- Use the `-m` flag to guide the AI
- Share feedback on what works and what doesn't

## Commands Reference

```bash
# Show version
.\gm.exe --version

# Show help
.\gm.exe --help
.\gm.exe commit --help

# Configure GitMind
.\gm.exe config

# Make a commit with AI assistance
.\gm.exe commit

# Add context for the AI
.\gm.exe commit -m "Your context here"

# Use conventional commits
.\gm.exe commit --conventional
```

## Example Workflow

Here's a complete example:

```bash
# 1. Navigate to your project
cd D:\projects\myproject

# 2. Make some changes
echo "function hello() { return 'world'; }" > hello.js

# 3. Run GitMind
.\gm.exe commit

# 4. Review the AI suggestions in the TUI

# 5. Press Enter to confirm

# 6. Done! Your commit is created
```

---

**Enjoy using GitMind! 🚀**

Have questions or found a bug? Open an issue on GitHub.
