package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// PRStatus represents the state of a pull request.
type PRStatus int

const (
	// PRStatusOpen indicates an open pull request.
	PRStatusOpen PRStatus = iota
	// PRStatusClosed indicates a closed pull request (not merged).
	PRStatusClosed
	// PRStatusMerged indicates a merged pull request.
	PRStatusMerged
	// PRStatusDraft indicates a draft pull request.
	PRStatusDraft
)

// String returns the string representation of the PR status.
func (s PRStatus) String() string {
	switch s {
	case PRStatusOpen:
		return "open"
	case PRStatusClosed:
		return "closed"
	case PRStatusMerged:
		return "merged"
	case PRStatusDraft:
		return "draft"
	default:
		return fmt.Sprintf("PRStatus(%d)", s)
	}
}

// PRAction represents actions that can be performed on a pull request.
type PRAction int

const (
	// PRActionNone represents no action (view-only).
	PRActionNone PRAction = iota
	// PRActionCreate creates a new pull request.
	PRActionCreate
	// PRActionUpdate updates an existing pull request.
	PRActionUpdate
	// PRActionClose closes a pull request without merging.
	PRActionClose
	// PRActionMerge merges a pull request.
	PRActionMerge
	// PRActionConvertToDraft converts a ready PR to draft.
	PRActionConvertToDraft
	// PRActionMarkReady marks a draft PR as ready for review.
	PRActionMarkReady
)

// String returns the string representation of the PR action.
func (a PRAction) String() string {
	switch a {
	case PRActionNone:
		return "none"
	case PRActionCreate:
		return "create"
	case PRActionUpdate:
		return "update"
	case PRActionClose:
		return "close"
	case PRActionMerge:
		return "merge"
	case PRActionConvertToDraft:
		return "convert-to-draft"
	case PRActionMarkReady:
		return "mark-ready"
	default:
		return fmt.Sprintf("PRAction(%d)", a)
	}
}

// PRInfo represents information about a pull request.
type PRInfo struct {
	id             string
	number         int
	title          string
	body           string
	state          PRStatus
	author         string
	baseRef        string // Base branch (e.g., "main")
	headRef        string // Head branch (e.g., "feature/foo")
	isDraft        bool
	labels         []string
	createdAt      time.Time
	updatedAt      time.Time
	htmlURL        string
	mergeableState string // "mergeable", "conflicting", "unknown"
}

// NewPRInfo creates a new PRInfo instance.
func NewPRInfo(number int, title, author, baseRef, headRef string) (*PRInfo, error) {
	if number <= 0 {
		return nil, errors.New("PR number must be positive")
	}
	if title == "" {
		return nil, errors.New("PR title cannot be empty")
	}
	if author == "" {
		return nil, errors.New("PR author cannot be empty")
	}
	if baseRef == "" {
		return nil, errors.New("base branch cannot be empty")
	}
	if headRef == "" {
		return nil, errors.New("head branch cannot be empty")
	}

	return &PRInfo{
		number:    number,
		title:     title,
		author:    author,
		baseRef:   baseRef,
		headRef:   headRef,
		state:     PRStatusOpen,
		labels:    make([]string, 0),
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

// ID returns the PR ID.
func (p *PRInfo) ID() string {
	return p.id
}

// SetID sets the PR ID.
func (p *PRInfo) SetID(id string) {
	p.id = id
}

// Number returns the PR number.
func (p *PRInfo) Number() int {
	return p.number
}

// Title returns the PR title.
func (p *PRInfo) Title() string {
	return p.title
}

// SetTitle sets the PR title.
func (p *PRInfo) SetTitle(title string) {
	p.title = title
}

// Body returns the PR body/description.
func (p *PRInfo) Body() string {
	return p.body
}

// SetBody sets the PR body/description.
func (p *PRInfo) SetBody(body string) {
	p.body = body
}

// State returns the PR state.
func (p *PRInfo) State() PRStatus {
	return p.state
}

// SetState sets the PR state.
func (p *PRInfo) SetState(state PRStatus) {
	p.state = state
}

// Author returns the PR author.
func (p *PRInfo) Author() string {
	return p.author
}

// BaseRef returns the base branch.
func (p *PRInfo) BaseRef() string {
	return p.baseRef
}

// HeadRef returns the head branch.
func (p *PRInfo) HeadRef() string {
	return p.headRef
}

// IsDraft returns true if this is a draft PR.
func (p *PRInfo) IsDraft() bool {
	return p.isDraft
}

// SetIsDraft sets the draft status.
func (p *PRInfo) SetIsDraft(draft bool) {
	p.isDraft = draft
}

// Labels returns the PR labels.
func (p *PRInfo) Labels() []string {
	return p.labels
}

// SetLabels sets the PR labels.
func (p *PRInfo) SetLabels(labels []string) {
	p.labels = labels
}

// AddLabel adds a label to the PR.
func (p *PRInfo) AddLabel(label string) {
	p.labels = append(p.labels, label)
}

// CreatedAt returns the creation timestamp.
func (p *PRInfo) CreatedAt() time.Time {
	return p.createdAt
}

// SetCreatedAt sets the creation timestamp.
func (p *PRInfo) SetCreatedAt(t time.Time) {
	p.createdAt = t
}

// UpdatedAt returns the last update timestamp.
func (p *PRInfo) UpdatedAt() time.Time {
	return p.updatedAt
}

// SetUpdatedAt sets the last update timestamp.
func (p *PRInfo) SetUpdatedAt(t time.Time) {
	p.updatedAt = t
}

// HTMLURL returns the web URL of the PR.
func (p *PRInfo) HTMLURL() string {
	return p.htmlURL
}

// SetHTMLURL sets the web URL.
func (p *PRInfo) SetHTMLURL(url string) {
	p.htmlURL = url
}

// MergeableState returns the mergeable state.
func (p *PRInfo) MergeableState() string {
	return p.mergeableState
}

// SetMergeableState sets the mergeable state.
func (p *PRInfo) SetMergeableState(state string) {
	p.mergeableState = state
}

// IsOpen returns true if the PR is open.
func (p *PRInfo) IsOpen() bool {
	return p.state == PRStatusOpen || p.state == PRStatusDraft
}

// IsMerged returns true if the PR is merged.
func (p *PRInfo) IsMerged() bool {
	return p.state == PRStatusMerged
}

// IsClosed returns true if the PR is closed (but not merged).
func (p *PRInfo) IsClosed() bool {
	return p.state == PRStatusClosed
}

// HasConflicts returns true if the PR has merge conflicts.
func (p *PRInfo) HasConflicts() bool {
	return p.mergeableState == "conflicting"
}

// String returns a string representation of the PR.
func (p *PRInfo) String() string {
	return fmt.Sprintf("PR #%d: %s (%s → %s) [%s]",
		p.number, p.title, p.headRef, p.baseRef, p.state)
}

// PROptions represents options for creating or updating a pull request.
type PROptions struct {
	title       string
	body        string
	baseBranch  string
	headBranch  string
	isDraft     bool
	labels      []string
	useTemplate bool
	assignees   []string
	reviewers   []string
}

// NewPROptions creates a new PROptions instance.
func NewPROptions(title, baseBranch, headBranch string) (*PROptions, error) {
	if title == "" {
		return nil, errors.New("PR title cannot be empty")
	}
	if baseBranch == "" {
		return nil, errors.New("base branch cannot be empty")
	}
	if headBranch == "" {
		return nil, errors.New("head branch cannot be empty")
	}
	if baseBranch == headBranch {
		return nil, errors.New("base and head branches cannot be the same")
	}

	return &PROptions{
		title:      title,
		baseBranch: baseBranch,
		headBranch: headBranch,
		labels:     make([]string, 0),
		assignees:  make([]string, 0),
		reviewers:  make([]string, 0),
	}, nil
}

// Title returns the PR title.
func (o *PROptions) Title() string {
	return o.title
}

// SetTitle sets the PR title.
func (o *PROptions) SetTitle(title string) {
	o.title = title
}

// Body returns the PR body.
func (o *PROptions) Body() string {
	return o.body
}

// SetBody sets the PR body.
func (o *PROptions) SetBody(body string) {
	o.body = body
}

// BaseBranch returns the base branch.
func (o *PROptions) BaseBranch() string {
	return o.baseBranch
}

// SetBaseBranch sets the base branch.
func (o *PROptions) SetBaseBranch(branch string) {
	o.baseBranch = branch
}

// HeadBranch returns the head branch.
func (o *PROptions) HeadBranch() string {
	return o.headBranch
}

// SetHeadBranch sets the head branch.
func (o *PROptions) SetHeadBranch(branch string) {
	o.headBranch = branch
}

// IsDraft returns true if this should be a draft PR.
func (o *PROptions) IsDraft() bool {
	return o.isDraft
}

// SetIsDraft sets the draft status.
func (o *PROptions) SetIsDraft(draft bool) {
	o.isDraft = draft
}

// Labels returns the labels.
func (o *PROptions) Labels() []string {
	return o.labels
}

// SetLabels sets the labels.
func (o *PROptions) SetLabels(labels []string) {
	o.labels = labels
}

// AddLabel adds a label.
func (o *PROptions) AddLabel(label string) {
	o.labels = append(o.labels, label)
}

// UseTemplate returns true if PR template should be loaded.
func (o *PROptions) UseTemplate() bool {
	return o.useTemplate
}

// SetUseTemplate sets whether to use PR template.
func (o *PROptions) SetUseTemplate(use bool) {
	o.useTemplate = use
}

// Assignees returns the assignees.
func (o *PROptions) Assignees() []string {
	return o.assignees
}

// SetAssignees sets the assignees.
func (o *PROptions) SetAssignees(assignees []string) {
	o.assignees = assignees
}

// AddAssignee adds an assignee.
func (o *PROptions) AddAssignee(assignee string) {
	o.assignees = append(o.assignees, assignee)
}

// Reviewers returns the reviewers.
func (o *PROptions) Reviewers() []string {
	return o.reviewers
}

// SetReviewers sets the reviewers.
func (o *PROptions) SetReviewers(reviewers []string) {
	o.reviewers = reviewers
}

// AddReviewer adds a reviewer.
func (o *PROptions) AddReviewer(reviewer string) {
	o.reviewers = append(o.reviewers, reviewer)
}

// Validate validates the PR options.
func (o *PROptions) Validate() error {
	if err := ValidatePRTitle(o.title); err != nil {
		return err
	}
	if o.baseBranch == "" {
		return errors.New("base branch cannot be empty")
	}
	if o.headBranch == "" {
		return errors.New("head branch cannot be empty")
	}
	if o.baseBranch == o.headBranch {
		return errors.New("base and head branches cannot be the same")
	}
	return nil
}

// ValidatePRTitle validates a PR title.
func ValidatePRTitle(title string) error {
	if title == "" {
		return errors.New("PR title cannot be empty")
	}
	if len(title) < 5 {
		return errors.New("PR title must be at least 5 characters")
	}
	if len(title) > 256 {
		return errors.New("PR title must not exceed 256 characters")
	}
	// Trim and check for meaningful content
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return errors.New("PR title cannot be only whitespace")
	}
	return nil
}

// ValidatePRBody validates a PR body/description.
func ValidatePRBody(body string) error {
	// Body is optional, but if provided should have some content
	if body != "" {
		trimmed := strings.TrimSpace(body)
		if trimmed == "" {
			return errors.New("PR body cannot be only whitespace")
		}
		if len(body) > 65536 {
			return errors.New("PR body must not exceed 65536 characters")
		}
	}
	return nil
}

// ValidateBaseBranch validates a base branch name.
func ValidateBaseBranch(branch string) error {
	if branch == "" {
		return errors.New("base branch cannot be empty")
	}
	if strings.Contains(branch, " ") {
		return errors.New("base branch name cannot contain spaces")
	}
	return nil
}
