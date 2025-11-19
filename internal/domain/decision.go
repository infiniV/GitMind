package domain

import (
	"errors"
	"fmt"
)

// ActionType represents the type of action the AI recommends.
type ActionType int

const (
	// ActionCommitDirect recommends committing directly to the current branch.
	ActionCommitDirect ActionType = iota
	// ActionCreateBranch recommends creating a new branch first.
	ActionCreateBranch
	// ActionSplitCommits recommends splitting changes into multiple commits.
	ActionSplitCommits
	// ActionReview recommends manual review before proceeding.
	ActionReview
	// ActionMerge recommends merging the current branch to parent/target.
	ActionMerge
	// ActionCreatePR recommends creating a pull request instead of direct merge.
	ActionCreatePR
)

// String returns the string representation of the action type.
func (at ActionType) String() string {
	switch at {
	case ActionCommitDirect:
		return "commit-direct"
	case ActionCreateBranch:
		return "create-branch"
	case ActionSplitCommits:
		return "split-commits"
	case ActionReview:
		return "review"
	case ActionMerge:
		return "merge"
	case ActionCreatePR:
		return "create-pr"
	default:
		return fmt.Sprintf("ActionType(%d)", at)
	}
}

// Alternative represents an alternative action the user could take.
type Alternative struct {
	Action      ActionType
	BranchName  string  // Suggested branch name (if ActionCreateBranch)
	Confidence  float64 // Confidence in this alternative (0.0 to 1.0)
	Description string  // Human-readable description
}

// NewAlternative creates a new Alternative.
func NewAlternative(action ActionType, description string, confidence float64) (*Alternative, error) {
	if description == "" {
		return nil, errors.New("alternative description cannot be empty")
	}
	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}

	return &Alternative{
		Action:      action,
		Description: description,
		Confidence:  confidence,
	}, nil
}

// Decision represents the AI's decision about how to handle a commit.
type Decision struct {
	action         ActionType
	confidence     float64
	reasoning      string
	suggestedMsg   *CommitMessage
	branchName     string
	alternatives   []Alternative
	requiresReview bool
	mergeStrategy  string      // Suggested merge strategy (for ActionMerge)
	targetBranch   string      // Target branch for merge (for ActionMerge)
	suggestedPR    *PROptions  // Suggested PR options (for ActionCreatePR)
}

// NewDecision creates a new Decision.
func NewDecision(action ActionType, confidence float64, reasoning string) (*Decision, error) {
	if reasoning == "" {
		return nil, errors.New("decision reasoning cannot be empty")
	}
	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}

	return &Decision{
		action:       action,
		confidence:   confidence,
		reasoning:    reasoning,
		alternatives: make([]Alternative, 0),
	}, nil
}

// Action returns the recommended action.
func (d *Decision) Action() ActionType {
	return d.action
}

// Confidence returns the confidence level (0.0 to 1.0).
func (d *Decision) Confidence() float64 {
	return d.confidence
}

// Reasoning returns the explanation for this decision.
func (d *Decision) Reasoning() string {
	return d.reasoning
}

// SuggestedMessage returns the suggested commit message.
func (d *Decision) SuggestedMessage() *CommitMessage {
	return d.suggestedMsg
}

// SetSuggestedMessage sets the suggested commit message.
func (d *Decision) SetSuggestedMessage(msg *CommitMessage) {
	d.suggestedMsg = msg
}

// BranchName returns the suggested branch name (if ActionCreateBranch).
func (d *Decision) BranchName() string {
	return d.branchName
}

// SetBranchName sets the suggested branch name.
func (d *Decision) SetBranchName(name string) {
	d.branchName = name
}

// Alternatives returns the list of alternative actions.
func (d *Decision) Alternatives() []Alternative {
	return d.alternatives
}

// AddAlternative adds an alternative action to the decision.
func (d *Decision) AddAlternative(alt Alternative) {
	d.alternatives = append(d.alternatives, alt)
}

// RequiresReview returns true if manual review is recommended.
func (d *Decision) RequiresReview() bool {
	return d.requiresReview || d.confidence < 0.7
}

// SetRequiresReview sets whether manual review is required.
func (d *Decision) SetRequiresReview(required bool) {
	d.requiresReview = required
}

// MergeStrategy returns the suggested merge strategy.
func (d *Decision) MergeStrategy() string {
	return d.mergeStrategy
}

// SetMergeStrategy sets the suggested merge strategy.
func (d *Decision) SetMergeStrategy(strategy string) {
	d.mergeStrategy = strategy
}

// TargetBranch returns the target branch for merge.
func (d *Decision) TargetBranch() string {
	return d.targetBranch
}

// SetTargetBranch sets the target branch for merge.
func (d *Decision) SetTargetBranch(branch string) {
	d.targetBranch = branch
}

// SuggestedPR returns the suggested PR options.
func (d *Decision) SuggestedPR() *PROptions {
	return d.suggestedPR
}

// SetSuggestedPR sets the suggested PR options.
func (d *Decision) SetSuggestedPR(pr *PROptions) {
	d.suggestedPR = pr
}

// IsHighConfidence returns true if confidence is >= 0.8.
func (d *Decision) IsHighConfidence() bool {
	return d.confidence >= 0.8
}

// IsMediumConfidence returns true if confidence is between 0.5 and 0.8.
func (d *Decision) IsMediumConfidence() bool {
	return d.confidence >= 0.5 && d.confidence < 0.8
}

// IsLowConfidence returns true if confidence is < 0.5.
func (d *Decision) IsLowConfidence() bool {
	return d.confidence < 0.5
}

// ConfidenceLevel returns a human-readable confidence level.
func (d *Decision) ConfidenceLevel() string {
	if d.IsHighConfidence() {
		return "high"
	}
	if d.IsMediumConfidence() {
		return "medium"
	}
	return "low"
}

// ShouldShowAlternatives returns true if alternatives should be presented to the user.
// This happens when confidence is not high or when there are viable alternatives.
func (d *Decision) ShouldShowAlternatives() bool {
	if !d.IsHighConfidence() {
		return true
	}
	// Even with high confidence, show alternatives if we have some with decent confidence
	for _, alt := range d.alternatives {
		if alt.Confidence >= 0.6 {
			return true
		}
	}
	return false
}

// String returns a string representation of the decision.
func (d *Decision) String() string {
	return fmt.Sprintf("Decision{action: %s, confidence: %.2f, reasoning: %s}",
		d.action, d.confidence, d.reasoning)
}

// Validate checks if the decision is valid and complete.
func (d *Decision) Validate() error {
	if d.reasoning == "" {
		return errors.New("decision must have reasoning")
	}
	if d.confidence < 0.0 || d.confidence > 1.0 {
		return fmt.Errorf("invalid confidence value: %f", d.confidence)
	}
	if d.action == ActionCreateBranch && d.branchName == "" {
		return errors.New("branch name required for create-branch action")
	}
	if d.action == ActionCommitDirect && d.suggestedMsg == nil {
		return errors.New("commit message required for commit-direct action")
	}
	if d.action == ActionMerge && d.targetBranch == "" {
		return errors.New("target branch required for merge action")
	}
	if d.action == ActionCreatePR && d.suggestedPR == nil {
		return errors.New("PR options required for create-pr action")
	}
	return nil
}
