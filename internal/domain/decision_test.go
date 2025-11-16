package domain

import (
	"testing"
)

func TestNewDecision(t *testing.T) {
	tests := []struct {
		name        string
		action      ActionType
		confidence  float64
		reasoning   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid decision",
			action:     ActionCommitDirect,
			confidence: 0.9,
			reasoning:  "Small bug fix, safe to commit directly",
			wantErr:    false,
		},
		{
			name:        "empty reasoning",
			action:      ActionCommitDirect,
			confidence:  0.9,
			reasoning:   "",
			wantErr:     true,
			errContains: "reasoning cannot be empty",
		},
		{
			name:        "confidence too high",
			action:      ActionCommitDirect,
			confidence:  1.5,
			reasoning:   "Test",
			wantErr:     true,
			errContains: "confidence must be between 0.0 and 1.0",
		},
		{
			name:        "confidence too low",
			action:      ActionCommitDirect,
			confidence:  -0.1,
			reasoning:   "Test",
			wantErr:     true,
			errContains: "confidence must be between 0.0 and 1.0",
		},
		{
			name:       "confidence at boundary 0.0",
			action:     ActionReview,
			confidence: 0.0,
			reasoning:  "Very uncertain",
			wantErr:    false,
		},
		{
			name:       "confidence at boundary 1.0",
			action:     ActionCommitDirect,
			confidence: 1.0,
			reasoning:  "Very certain",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := NewDecision(tt.action, tt.confidence, tt.reasoning)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewDecision() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewDecision() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewDecision() unexpected error = %v", err)
				return
			}

			if decision.Action() != tt.action {
				t.Errorf("Action() = %v, want %v", decision.Action(), tt.action)
			}
			if decision.Confidence() != tt.confidence {
				t.Errorf("Confidence() = %v, want %v", decision.Confidence(), tt.confidence)
			}
			if decision.Reasoning() != tt.reasoning {
				t.Errorf("Reasoning() = %v, want %v", decision.Reasoning(), tt.reasoning)
			}
		})
	}
}

func TestNewAlternative(t *testing.T) {
	tests := []struct {
		name        string
		action      ActionType
		description string
		confidence  float64
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid alternative",
			action:      ActionCreateBranch,
			description: "Create feature branch",
			confidence:  0.7,
			wantErr:     false,
		},
		{
			name:        "empty description",
			action:      ActionCreateBranch,
			description: "",
			confidence:  0.7,
			wantErr:     true,
			errContains: "description cannot be empty",
		},
		{
			name:        "invalid confidence",
			action:      ActionCreateBranch,
			description: "Test",
			confidence:  1.5,
			wantErr:     true,
			errContains: "confidence must be between 0.0 and 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alt, err := NewAlternative(tt.action, tt.description, tt.confidence)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAlternative() expected error, got nil")
					return
				}
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewAlternative() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAlternative() unexpected error = %v", err)
				return
			}

			if alt.Action != tt.action {
				t.Errorf("Action = %v, want %v", alt.Action, tt.action)
			}
			if alt.Description != tt.description {
				t.Errorf("Description = %v, want %v", alt.Description, tt.description)
			}
		})
	}
}

func TestDecision_ConfidenceLevels(t *testing.T) {
	tests := []struct {
		name            string
		confidence      float64
		wantHigh        bool
		wantMedium      bool
		wantLow         bool
		wantLevel       string
		wantRequiresReview bool
	}{
		{
			name:            "high confidence",
			confidence:      0.9,
			wantHigh:        true,
			wantMedium:      false,
			wantLow:         false,
			wantLevel:       "high",
			wantRequiresReview: false,
		},
		{
			name:            "medium confidence",
			confidence:      0.7,
			wantHigh:        false,
			wantMedium:      true,
			wantLow:         false,
			wantLevel:       "medium",
			wantRequiresReview: false,
		},
		{
			name:            "low confidence",
			confidence:      0.4,
			wantHigh:        false,
			wantMedium:      false,
			wantLow:         true,
			wantLevel:       "low",
			wantRequiresReview: true,
		},
		{
			name:            "boundary high-medium",
			confidence:      0.8,
			wantHigh:        true,
			wantMedium:      false,
			wantLow:         false,
			wantLevel:       "high",
			wantRequiresReview: false,
		},
		{
			name:            "boundary medium-low",
			confidence:      0.5,
			wantHigh:        false,
			wantMedium:      true,
			wantLow:         false,
			wantLevel:       "medium",
			wantRequiresReview: true, // confidence < 0.7 requires review
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, _ := NewDecision(ActionCommitDirect, tt.confidence, "test")

			if got := decision.IsHighConfidence(); got != tt.wantHigh {
				t.Errorf("IsHighConfidence() = %v, want %v", got, tt.wantHigh)
			}
			if got := decision.IsMediumConfidence(); got != tt.wantMedium {
				t.Errorf("IsMediumConfidence() = %v, want %v", got, tt.wantMedium)
			}
			if got := decision.IsLowConfidence(); got != tt.wantLow {
				t.Errorf("IsLowConfidence() = %v, want %v", got, tt.wantLow)
			}
			if got := decision.ConfidenceLevel(); got != tt.wantLevel {
				t.Errorf("ConfidenceLevel() = %v, want %v", got, tt.wantLevel)
			}
			if got := decision.RequiresReview(); got != tt.wantRequiresReview {
				t.Errorf("RequiresReview() = %v, want %v", got, tt.wantRequiresReview)
			}
		})
	}
}

func TestDecision_SuggestedMessage(t *testing.T) {
	decision, _ := NewDecision(ActionCommitDirect, 0.9, "test")

	// Initially no message
	if decision.SuggestedMessage() != nil {
		t.Error("SuggestedMessage() expected nil, got message")
	}

	// Set message
	msg, _ := NewCommitMessage("Add feature")
	decision.SetSuggestedMessage(msg)

	if decision.SuggestedMessage() == nil {
		t.Error("SuggestedMessage() expected message, got nil")
	}
	if decision.SuggestedMessage().Title() != "Add feature" {
		t.Errorf("SuggestedMessage().Title() = %v, want %v", decision.SuggestedMessage().Title(), "Add feature")
	}
}

func TestDecision_BranchName(t *testing.T) {
	decision, _ := NewDecision(ActionCreateBranch, 0.8, "test")

	// Initially no branch name
	if decision.BranchName() != "" {
		t.Errorf("BranchName() = %v, want empty string", decision.BranchName())
	}

	// Set branch name
	decision.SetBranchName("feature/new-auth")

	if decision.BranchName() != "feature/new-auth" {
		t.Errorf("BranchName() = %v, want %v", decision.BranchName(), "feature/new-auth")
	}
}

func TestDecision_Alternatives(t *testing.T) {
	decision, _ := NewDecision(ActionCommitDirect, 0.9, "test")

	// Initially no alternatives
	if len(decision.Alternatives()) != 0 {
		t.Errorf("Alternatives() length = %v, want 0", len(decision.Alternatives()))
	}

	// Add alternatives
	alt1, _ := NewAlternative(ActionCreateBranch, "Create branch instead", 0.7)
	alt2, _ := NewAlternative(ActionReview, "Manual review recommended", 0.5)

	decision.AddAlternative(*alt1)
	decision.AddAlternative(*alt2)

	if len(decision.Alternatives()) != 2 {
		t.Errorf("Alternatives() length = %v, want 2", len(decision.Alternatives()))
	}
}

func TestDecision_ShouldShowAlternatives(t *testing.T) {
	tests := []struct {
		name         string
		confidence   float64
		alternatives []Alternative
		want         bool
	}{
		{
			name:         "high confidence no alternatives",
			confidence:   0.9,
			alternatives: []Alternative{},
			want:         false,
		},
		{
			name:       "medium confidence no alternatives",
			confidence: 0.7,
			alternatives: []Alternative{},
			want:       true,
		},
		{
			name:       "high confidence with strong alternative",
			confidence: 0.9,
			alternatives: []Alternative{
				{Action: ActionCreateBranch, Confidence: 0.7, Description: "Create branch"},
			},
			want: true,
		},
		{
			name:       "high confidence with weak alternative",
			confidence: 0.9,
			alternatives: []Alternative{
				{Action: ActionReview, Confidence: 0.3, Description: "Review"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, _ := NewDecision(ActionCommitDirect, tt.confidence, "test")
			for _, alt := range tt.alternatives {
				decision.AddAlternative(alt)
			}

			if got := decision.ShouldShowAlternatives(); got != tt.want {
				t.Errorf("ShouldShowAlternatives() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecision_RequiresReview(t *testing.T) {
	t.Run("low confidence requires review", func(t *testing.T) {
		decision, _ := NewDecision(ActionCommitDirect, 0.6, "test")
		if !decision.RequiresReview() {
			t.Error("RequiresReview() = false, want true for low confidence")
		}
	})

	t.Run("explicit review flag", func(t *testing.T) {
		decision, _ := NewDecision(ActionCommitDirect, 0.9, "test")
		decision.SetRequiresReview(true)
		if !decision.RequiresReview() {
			t.Error("RequiresReview() = false, want true when explicitly set")
		}
	})

	t.Run("high confidence no review", func(t *testing.T) {
		decision, _ := NewDecision(ActionCommitDirect, 0.9, "test")
		if decision.RequiresReview() {
			t.Error("RequiresReview() = true, want false for high confidence")
		}
	})
}

func TestDecision_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Decision
		wantErr     bool
		errContains string
	}{
		{
			name: "valid direct commit",
			setup: func() *Decision {
				d, _ := NewDecision(ActionCommitDirect, 0.9, "test reasoning")
				msg, _ := NewCommitMessage("Add feature")
				d.SetSuggestedMessage(msg)
				return d
			},
			wantErr: false,
		},
		{
			name: "valid create branch",
			setup: func() *Decision {
				d, _ := NewDecision(ActionCreateBranch, 0.8, "test reasoning")
				d.SetBranchName("feature/new-feature")
				return d
			},
			wantErr: false,
		},
		{
			name: "direct commit without message",
			setup: func() *Decision {
				d, _ := NewDecision(ActionCommitDirect, 0.9, "test reasoning")
				return d
			},
			wantErr:     true,
			errContains: "commit message required",
		},
		{
			name: "create branch without name",
			setup: func() *Decision {
				d, _ := NewDecision(ActionCreateBranch, 0.8, "test reasoning")
				return d
			},
			wantErr:     true,
			errContains: "branch name required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := tt.setup()
			err := decision.Validate()

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

func TestActionType_String(t *testing.T) {
	tests := []struct {
		action ActionType
		want   string
	}{
		{ActionCommitDirect, "commit-direct"},
		{ActionCreateBranch, "create-branch"},
		{ActionSplitCommits, "split-commits"},
		{ActionReview, "review"},
		{ActionType(99), "ActionType(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.action.String(); got != tt.want {
				t.Errorf("ActionType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecision_String(t *testing.T) {
	decision, _ := NewDecision(ActionCommitDirect, 0.9, "test reasoning")

	str := decision.String()
	if !contains(str, "commit-direct") {
		t.Errorf("String() = %v, want to contain 'commit-direct'", str)
	}
	if !contains(str, "0.90") {
		t.Errorf("String() = %v, want to contain '0.90'", str)
	}
	if !contains(str, "test reasoning") {
		t.Errorf("String() = %v, want to contain 'test reasoning'", str)
	}
}
