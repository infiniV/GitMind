package domain

import (
	"strings"
	"testing"
)

func TestNewCommitMessage(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid title",
			title:   "Add new feature",
			wantErr: false,
		},
		{
			name:    "title with whitespace",
			title:   "  Add new feature  ",
			wantErr: false,
		},
		{
			name:        "empty title",
			title:       "",
			wantErr:     true,
			errContains: "commit title cannot be empty",
		},
		{
			name:        "title too long",
			title:       strings.Repeat("a", 73),
			wantErr:     true,
			errContains: "commit title too long",
		},
		{
			name:    "title at max length",
			title:   strings.Repeat("a", 72),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewCommitMessage(tt.title)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewCommitMessage() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewCommitMessage() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewCommitMessage() unexpected error = %v", err)
				return
			}

			if msg == nil {
				t.Error("NewCommitMessage() returned nil message")
				return
			}

			// Title should be trimmed
			if strings.TrimSpace(tt.title) != msg.Title() {
				t.Errorf("Title() = %q, want %q", msg.Title(), strings.TrimSpace(tt.title))
			}
		})
	}
}

func TestNewConventionalCommit(t *testing.T) {
	tests := []struct {
		name        string
		commitType  string
		scope       string
		title       string
		wantErr     bool
		errContains string
		wantTitle   string
	}{
		{
			name:       "valid feat without scope",
			commitType: "feat",
			scope:      "",
			title:      "add user authentication",
			wantErr:    false,
			wantTitle:  "feat: add user authentication",
		},
		{
			name:       "valid fix with scope",
			commitType: "fix",
			scope:      "api",
			title:      "handle null pointer",
			wantErr:    false,
			wantTitle:  "fix(api): handle null pointer",
		},
		{
			name:        "empty type",
			commitType:  "",
			scope:       "",
			title:       "something",
			wantErr:     true,
			errContains: "commit type cannot be empty",
		},
		{
			name:        "invalid type",
			commitType:  "invalid",
			scope:       "",
			title:       "something",
			wantErr:     true,
			errContains: "invalid commit type",
		},
		{
			name:        "empty title",
			commitType:  "feat",
			scope:       "",
			title:       "",
			wantErr:     true,
			errContains: "commit title cannot be empty",
		},
		{
			name:        "title too long with type",
			commitType:  "feat",
			scope:       "verylongscope",
			title:       strings.Repeat("a", 72),
			wantErr:     true,
			errContains: "commit title too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := NewConventionalCommit(tt.commitType, tt.scope, tt.title)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewConventionalCommit() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewConventionalCommit() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewConventionalCommit() unexpected error = %v", err)
				return
			}

			if msg.Title() != tt.wantTitle {
				t.Errorf("Title() = %q, want %q", msg.Title(), tt.wantTitle)
			}

			if !msg.IsConventional() {
				t.Error("IsConventional() = false, want true")
			}

			if msg.Type() != tt.commitType {
				t.Errorf("Type() = %q, want %q", msg.Type(), tt.commitType)
			}

			if msg.Scope() != tt.scope {
				t.Errorf("Scope() = %q, want %q", msg.Scope(), tt.scope)
			}
		})
	}
}

func TestCommitMessage_Body(t *testing.T) {
	msg, _ := NewCommitMessage("Test commit")

	// Initially no body
	if msg.Body() != "" {
		t.Errorf("Body() = %q, want empty string", msg.Body())
	}

	// Set body
	body := "This is the commit body\nwith multiple lines"
	msg.SetBody(body)

	if msg.Body() != body {
		t.Errorf("Body() = %q, want %q", msg.Body(), body)
	}

	// Body with extra whitespace should be trimmed
	msg.SetBody("  \n  body with whitespace  \n  ")
	if msg.Body() != "body with whitespace" {
		t.Errorf("Body() = %q, want trimmed body", msg.Body())
	}
}

func TestCommitMessage_FullMessage(t *testing.T) {
	tests := []struct {
		name  string
		title string
		body  string
		want  string
	}{
		{
			name:  "title only",
			title: "Add feature",
			body:  "",
			want:  "Add feature",
		},
		{
			name:  "title and body",
			title: "Add feature",
			body:  "This adds a new feature",
			want:  "Add feature\n\nThis adds a new feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewCommitMessage(tt.title)
			if tt.body != "" {
				msg.SetBody(tt.body)
			}

			if got := msg.FullMessage(); got != tt.want {
				t.Errorf("FullMessage() = %q, want %q", got, tt.want)
			}

			// String() should return the same as FullMessage()
			if got := msg.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommitMessage_Validate(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		body        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid simple message",
			title:   "Add feature",
			body:    "",
			wantErr: false,
		},
		{
			name:    "valid with body",
			title:   "Add feature",
			body:    "This is a valid body\nwith multiple lines\nthat are not too long",
			wantErr: false,
		},
		{
			name:        "title ends with period",
			title:       "Add feature.",
			body:        "",
			wantErr:     true,
			errContains: "should not end with a period",
		},
		{
			name:        "body line too long",
			title:       "Add feature",
			body:        strings.Repeat("a", 73),
			wantErr:     true,
			errContains: "body line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewCommitMessage(tt.title)
			if tt.body != "" {
				msg.SetBody(tt.body)
			}

			err := msg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestCommitStrategy_String(t *testing.T) {
	tests := []struct {
		strategy CommitStrategy
		want     string
	}{
		{StrategyDirectCommit, "direct-commit"},
		{StrategyNewBranch, "new-branch"},
		{StrategySplitCommits, "split-commits"},
		{CommitStrategy(99), "CommitStrategy(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.want {
				t.Errorf("CommitStrategy.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
