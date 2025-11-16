package domain

import (
	"errors"
	"fmt"
	"strings"
)

// CommitMessage represents a structured commit message.
type CommitMessage struct {
	title       string
	body        string
	conventional bool
	commitType  string // feat, fix, chore, etc.
	scope       string // optional scope in conventional commits
}

// NewCommitMessage creates a new commit message.
func NewCommitMessage(title string) (*CommitMessage, error) {
	if title == "" {
		return nil, errors.New("commit title cannot be empty")
	}

	// Trim whitespace
	title = strings.TrimSpace(title)

	// If title is too long, truncate it intelligently
	if len(title) > 72 {
		// Try to truncate at a word boundary
		truncated := title[:69] + "..."
		// Find last space before position 69
		if lastSpace := strings.LastIndex(title[:69], " "); lastSpace > 50 {
			truncated = title[:lastSpace] + "..."
		}
		title = truncated
	}

	return &CommitMessage{
		title: title,
	}, nil
}

// NewConventionalCommit creates a new conventional commit message.
func NewConventionalCommit(commitType, scope, title string) (*CommitMessage, error) {
	if commitType == "" {
		return nil, errors.New("commit type cannot be empty")
	}

	validTypes := map[string]bool{
		"feat": true, "fix": true, "docs": true, "style": true,
		"refactor": true, "perf": true, "test": true, "chore": true,
		"build": true, "ci": true, "revert": true,
	}

	if !validTypes[commitType] {
		return nil, fmt.Errorf("invalid commit type: %s", commitType)
	}

	if title == "" {
		return nil, errors.New("commit title cannot be empty")
	}

	// Build the full title with conventional format
	fullTitle := commitType
	if scope != "" {
		fullTitle += fmt.Sprintf("(%s)", scope)
	}
	fullTitle += ": " + title

	if len(fullTitle) > 72 {
		return nil, fmt.Errorf("commit title too long (%d chars), should be <= 72", len(fullTitle))
	}

	return &CommitMessage{
		title:        fullTitle,
		conventional: true,
		commitType:   commitType,
		scope:        scope,
	}, nil
}

// Title returns the commit title.
func (cm *CommitMessage) Title() string {
	return cm.title
}

// Body returns the commit body.
func (cm *CommitMessage) Body() string {
	return cm.body
}

// SetBody sets the commit body.
func (cm *CommitMessage) SetBody(body string) {
	cm.body = strings.TrimSpace(body)
}

// IsConventional returns true if this is a conventional commit.
func (cm *CommitMessage) IsConventional() bool {
	return cm.conventional
}

// Type returns the commit type (for conventional commits).
func (cm *CommitMessage) Type() string {
	return cm.commitType
}

// Scope returns the commit scope (for conventional commits).
func (cm *CommitMessage) Scope() string {
	return cm.scope
}

// FullMessage returns the complete commit message (title + body).
func (cm *CommitMessage) FullMessage() string {
	if cm.body == "" {
		return cm.title
	}
	return cm.title + "\n\n" + cm.body
}

// String returns the string representation of the commit message.
func (cm *CommitMessage) String() string {
	return cm.FullMessage()
}

// Validate checks if the commit message follows best practices.
func (cm *CommitMessage) Validate() error {
	if cm.title == "" {
		return errors.New("commit title is empty")
	}

	if len(cm.title) > 72 {
		return fmt.Errorf("commit title too long (%d chars), should be <= 72", len(cm.title))
	}

	// Title should not end with a period
	if strings.HasSuffix(cm.title, ".") {
		return errors.New("commit title should not end with a period")
	}

	// Body lines should wrap at 72 characters (warning, not error)
	if cm.body != "" {
		lines := strings.Split(cm.body, "\n")
		for i, line := range lines {
			if len(line) > 72 {
				return fmt.Errorf("commit body line %d too long (%d chars), should wrap at 72", i+1, len(line))
			}
		}
	}

	return nil
}

// CommitStrategy represents how the commit should be made.
type CommitStrategy int

const (
	// StrategyDirectCommit commits directly to the current branch.
	StrategyDirectCommit CommitStrategy = iota
	// StrategyNewBranch creates a new branch and commits there.
	StrategyNewBranch
	// StrategySplitCommits splits changes into multiple commits.
	StrategySplitCommits
)

// String returns the string representation of the commit strategy.
func (cs CommitStrategy) String() string {
	switch cs {
	case StrategyDirectCommit:
		return "direct-commit"
	case StrategyNewBranch:
		return "new-branch"
	case StrategySplitCommits:
		return "split-commits"
	default:
		return fmt.Sprintf("CommitStrategy(%d)", cs)
	}
}
