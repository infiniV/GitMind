package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/gitman/internal/domain"
)

const (
	defaultCerebrasBaseURL = "https://api.cerebras.ai/v1"
	defaultModel           = "llama-3.3-70b" // Good balance of quality and speed
	defaultTimeout         = 30 * time.Second
	maxRetries             = 3
)

// CerebrasProvider implements the Provider interface for Cerebras AI.
type CerebrasProvider struct {
	apiKey     *domain.APIKey
	baseURL    string
	model      string
	httpClient *http.Client
	maxRetries int
}

// NewCerebrasProvider creates a new Cerebras provider.
func NewCerebrasProvider(apiKey *domain.APIKey, config ProviderConfig) *CerebrasProvider {
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	maxRetries := maxRetries
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}

	baseURL := defaultCerebrasBaseURL
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	model := defaultModel
	if config.Model != "" {
		model = config.Model
	}

	return &CerebrasProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
	}
}

// GetName returns the provider name.
func (c *CerebrasProvider) GetName() string {
	return "cerebras"
}

// ValidateKey checks if the API key is valid.
func (c *CerebrasProvider) ValidateKey(ctx context.Context) error {
	// Simple validation by making a minimal API call
	reqBody := cerebrasRequest{
		Model: c.model,
		Messages: []message{
			{Role: "user", Content: "test"},
		},
		MaxCompletionTokens: 10,
	}

	_, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	return nil
}

// DetectTier attempts to detect the API key tier.
func (c *CerebrasProvider) DetectTier(ctx context.Context) (domain.APITier, error) {
	// For Cerebras, we can't automatically detect tier from the API
	// We'll use a heuristic: try a moderate request and see if rate limits are hit

	// For now, assume free tier and let users upgrade in config
	// A production implementation could track rate limits from response headers
	return domain.TierFree, nil
}

// Analyze analyzes git changes and returns a decision.
func (c *CerebrasProvider) Analyze(ctx context.Context, request AnalysisRequest) (*AnalysisResponse, error) {
	if request.Repository == nil {
		return nil, errors.New("repository cannot be nil")
	}

	startTime := time.Now()

	// Build the prompt
	prompt := c.buildPrompt(request)

	// Prepare the request with structured output
	reqBody := c.buildStructuredRequest(prompt)

	// Make the API call with retry logic
	var resp *cerebrasResponse
	var err error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err = c.makeRequestWithRetry(ctx, reqBody, attempt)
		if err == nil {
			break
		}

		// Check if it's a rate limit error
		if strings.Contains(err.Error(), "rate limit") && request.APIKey.IsFree() {
			return nil, &FreeTierLimitError{
				Message: "Rate limit reached. Please wait a moment or upgrade to a pro API key for higher limits.",
				RetryAfter: 60,
			}
		}

		// Check if we should retry
		if attempt < c.maxRetries && isRetryableError(err) {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second) // Exponential backoff
			continue
		}

		return nil, fmt.Errorf("AI analysis failed after %d attempts: %w", attempt+1, err)
	}

	// Parse the structured response
	decision, err := c.parseResponse(resp, request.UseConventionalCommits)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	processingTime := time.Since(startTime).Milliseconds()

	return &AnalysisResponse{
		Decision:         decision,
		TokensUsed:       resp.Usage.TotalTokens,
		Model:            resp.Model,
		ProcessingTimeMs: int(processingTime),
	}, nil
}

// buildPrompt builds the analysis prompt with context reduction for free tier.
func (c *CerebrasProvider) buildPrompt(request AnalysisRequest) string {
	var sb strings.Builder

	sb.WriteString("You are an expert Git workflow assistant. Analyze the following code changes and provide recommendations.\n\n")

	// Repository context
	sb.WriteString(fmt.Sprintf("Repository: %s\n", request.Repository.Path()))

	// Branch context (enhanced)
	if request.BranchInfo != nil {
		branchDesc := request.BranchInfo.Name()
		if request.BranchInfo.Parent() != "" {
			branchDesc += fmt.Sprintf(" (parent: %s", request.BranchInfo.Parent())
			if request.BranchInfo.CommitCount() > 0 {
				branchDesc += fmt.Sprintf(", %d commits on this branch", request.BranchInfo.CommitCount())
			}
			branchDesc += ")"
		}

		if request.BranchInfo.IsProtected() {
			branchDesc += " [PROTECTED BRANCH]"
		} else {
			branchDesc += fmt.Sprintf(" [%s branch]", request.BranchInfo.Type())
		}

		sb.WriteString(fmt.Sprintf("Current branch: %s\n", branchDesc))
	} else {
		sb.WriteString(fmt.Sprintf("Current branch: %s\n", request.Repository.CurrentBranch()))
	}

	sb.WriteString(fmt.Sprintf("Changes: %s\n\n", request.Repository.ChangeSummary()))

	// Recent commits for context (with scope indicator)
	if len(request.RecentLog) > 0 {
		commitScope := "Recent commits"
		if request.BranchInfo != nil && request.BranchInfo.Parent() != "" {
			commitScope = fmt.Sprintf("Commits on this branch (since %s)", request.BranchInfo.Parent())
		}
		sb.WriteString(fmt.Sprintf("%s:\n", commitScope))

		for i, log := range request.RecentLog {
			if i >= 3 {
				break // Limit to 3 recent commits
			}
			sb.WriteString(fmt.Sprintf("- %s\n", log))
		}
		sb.WriteString("\n")
	}

	// Diff content (with reduction for free tier)
	if request.Diff != "" {
		diff := request.Diff

		// Reduce context for free tier or large changesets
		if request.APIKey.ShouldReduceContext() || request.Repository.IsLargeChangeset() {
			diff = reduceDiffContext(diff, request.APIKey.MaxTokensPerRequest())
		}

		sb.WriteString("Changes (git diff):\n")
		sb.WriteString(diff)
		sb.WriteString("\n\n")
	}

	// User context
	if request.UserPrompt != "" {
		sb.WriteString(fmt.Sprintf("User context: %s\n\n", request.UserPrompt))
	}

	// Merge opportunity detection
	if request.MergeOpportunity {
		sb.WriteString("**MERGE OPPORTUNITY DETECTED**\n")
		sb.WriteString("- Working directory is clean (no uncommitted changes)\n")
		sb.WriteString(fmt.Sprintf("- Branch has %d commits ready to merge into '%s'\n", request.MergeCommitCount, request.MergeTargetBranch))
		sb.WriteString("- Consider recommending a MERGE action instead of commit\n\n")
	}

	// Instructions (enhanced with branch-aware guidance)
	sb.WriteString("Based on these changes, provide:\n")
	sb.WriteString("1. A clear, concise commit message")
	if request.UseConventionalCommits {
		sb.WriteString(" following conventional commits format (type(scope): description)")
	}
	sb.WriteString("\n")
	sb.WriteString("2. Your recommendation:\n")
	if request.MergeOpportunity {
		sb.WriteString("   - MERGE OPPORTUNITY: Branch is clean with multiple commits. Consider recommending 'merge' action.\n")
		sb.WriteString("   - User can merge to parent branch or continue working.\n")
	} else if request.BranchInfo != nil && request.BranchInfo.IsProtected() {
		sb.WriteString("   - IMPORTANT: User is on a PROTECTED branch. Strongly recommend creating a feature branch.\n")
	} else if request.BranchInfo != nil && request.BranchInfo.Type() != "protected" {
		sb.WriteString("   - User is on a feature branch. Consider if changes fit the branch purpose.\n")
		sb.WriteString("   - If changes match the branch theme, recommend committing directly.\n")
		sb.WriteString("   - If changes are unrelated or substantial, consider creating a new branch.\n")
	}
	sb.WriteString("3. Brief reasoning for your recommendation\n")
	sb.WriteString("4. Alternative approaches if applicable\n")

	return sb.String()
}

// buildStructuredRequest builds a Cerebras API request with JSON schema for structured output.
func (c *CerebrasProvider) buildStructuredRequest(prompt string) cerebrasRequest {
	falseBool := false

	schema := analysisSchema{
		Type: "object",
		Properties: map[string]property{
			"commit_message": {
				Type:        "string",
				Description: "Clear, concise commit message describing the changes",
			},
			"action": {
				Type:        "string",
				Enum:        []string{"commit-direct", "create-branch", "review", "merge"},
				Description: "Recommended action to take",
			},
			"confidence": {
				Type:        "number",
				Description: "Confidence level between 0.0 and 1.0",
			},
			"reasoning": {
				Type:        "string",
				Description: "Brief explanation for the recommendation",
			},
			"branch_name": {
				Type:        "string",
				Description: "Suggested branch name if action is create-branch",
			},
			"alternatives": {
				Type: "array",
				Items: &property{
					Type: "object",
					Properties: map[string]property{
						"action": {Type: "string"},
						"description": {Type: "string"},
						"confidence": {Type: "number"},
					},
					Required:             []string{"action", "description", "confidence"},
					AdditionalProperties: &falseBool,
				},
			},
		},
		Required:             []string{"commit_message", "action", "confidence", "reasoning"},
		AdditionalProperties: &falseBool,
	}

	return cerebrasRequest{
		Model: c.model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchema{
				Name:   "commit_analysis",
				Strict: true,
				Schema: schema,
			},
		},
		MaxCompletionTokens: 1000,
		Temperature:         ptrFloat(0.7),
	}
}

// makeRequest makes an API request to Cerebras.
func (c *CerebrasProvider) makeRequest(ctx context.Context, reqBody cerebrasRequest) (*cerebrasResponse, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey.Key())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	var cerebrasResp cerebrasResponse
	if err := json.Unmarshal(body, &cerebrasResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &cerebrasResp, nil
}

// makeRequestWithRetry makes a request with retry logic.
func (c *CerebrasProvider) makeRequestWithRetry(ctx context.Context, reqBody cerebrasRequest, attempt int) (*cerebrasResponse, error) {
	return c.makeRequest(ctx, reqBody)
}

// parseResponse parses the Cerebras response into a Decision.
func (c *CerebrasProvider) parseResponse(resp *cerebrasResponse, useConventional bool) (*domain.Decision, error) {
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return nil, errors.New("empty response content")
	}

	// Parse JSON response
	var analysis struct {
		CommitMessage string  `json:"commit_message"`
		Action        string  `json:"action"`
		Confidence    float64 `json:"confidence"`
		Reasoning     string  `json:"reasoning"`
		BranchName    string  `json:"branch_name,omitempty"`
		Alternatives  []struct {
			Action      string  `json:"action"`
			Description string  `json:"description"`
			Confidence  float64 `json:"confidence"`
		} `json:"alternatives,omitempty"`
	}

	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse structured output: %w", err)
	}

	// Map action string to ActionType
	actionType := mapActionType(analysis.Action)

	// Create decision
	decision, err := domain.NewDecision(actionType, analysis.Confidence, analysis.Reasoning)
	if err != nil {
		return nil, err
	}

	// Create commit message
	commitMsg, err := domain.NewCommitMessage(analysis.CommitMessage)
	if err != nil {
		return nil, fmt.Errorf("invalid commit message from AI: %w", err)
	}
	decision.SetSuggestedMessage(commitMsg)

	// Set branch name if applicable
	if analysis.BranchName != "" {
		decision.SetBranchName(analysis.BranchName)
	}

	// Add alternatives
	for _, alt := range analysis.Alternatives {
		alternative, err := domain.NewAlternative(
			mapActionType(alt.Action),
			alt.Description,
			alt.Confidence,
		)
		if err == nil {
			decision.AddAlternative(*alternative)
		}
	}

	return decision, nil
}

// GenerateMergeMessage generates a merge commit message and suggests a merge strategy.
func (c *CerebrasProvider) GenerateMergeMessage(ctx context.Context, request MergeMessageRequest) (*MergeMessageResponse, error) {
	// Build prompt for merge message generation
	prompt := c.buildMergePrompt(request)

	// Build structured request for merge message
	structuredReq := c.buildMergeStructuredRequest(prompt)

	// Call API
	resp, err := c.makeRequestWithRetry(ctx, structuredReq, 0)
	if err != nil {
		return nil, err
	}

	// Parse response
	mergeResponse, err := c.parseMergeResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse merge response: %w", err)
	}

	mergeResponse.TokensUsed = resp.Usage.TotalTokens
	mergeResponse.Model = resp.Model

	return mergeResponse, nil
}

// buildMergePrompt builds the prompt for merge message generation.
func (c *CerebrasProvider) buildMergePrompt(request MergeMessageRequest) string {
	var sb strings.Builder

	sb.WriteString("You are an expert Git workflow assistant. Generate a merge commit message for the following branch merge.\n\n")

	// Merge context
	sb.WriteString(fmt.Sprintf("Merging: %s → %s\n", request.SourceBranch, request.TargetBranch))
	sb.WriteString(fmt.Sprintf("Commits being merged: %d\n\n", request.CommitCount))

	// List commits
	sb.WriteString("Commits to merge:\n")
	maxCommits := len(request.Commits)
	if maxCommits > 10 {
		maxCommits = 10 // Limit to avoid token overflow
	}
	for i := 0; i < maxCommits; i++ {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, request.Commits[i]))
	}
	if len(request.Commits) > maxCommits {
		sb.WriteString(fmt.Sprintf("... and %d more commits\n", len(request.Commits)-maxCommits))
	}
	sb.WriteString("\n")

	// Instructions
	sb.WriteString("Provide:\n")
	sb.WriteString("1. A concise merge commit message that summarizes the changes\n")
	sb.WriteString("2. Recommended merge strategy:\n")
	sb.WriteString("   - 'squash' if many commits (5+) or commits contain WIP/fixup messages\n")
	sb.WriteString("   - 'regular' if few meaningful commits (1-4) that should be preserved\n")
	sb.WriteString("   - 'fast-forward' if linear history is possible\n")
	sb.WriteString("3. Brief reasoning for your recommendation\n")

	return sb.String()
}

// buildMergeStructuredRequest builds a structured request for merge message generation.
func (c *CerebrasProvider) buildMergeStructuredRequest(prompt string) cerebrasRequest {
	falseBool := false

	schema := analysisSchema{
		Type: "object",
		Properties: map[string]property{
			"merge_message": {
				Type:        "string",
				Description: "Concise merge commit message summarizing the changes",
			},
			"strategy": {
				Type:        "string",
				Enum:        []string{"squash", "regular", "fast-forward"},
				Description: "Recommended merge strategy",
			},
			"reasoning": {
				Type:        "string",
				Description: "Brief explanation for the recommendation",
			},
		},
		Required:             []string{"merge_message", "strategy", "reasoning"},
		AdditionalProperties: &falseBool,
	}

	// Use lower temperature for more consistent merge messages
	temp := 0.3

	return cerebrasRequest{
		Model: c.model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchema{
				Name:   "merge_message_generation",
				Strict: true,
				Schema: schema,
			},
		},
		MaxCompletionTokens: 500, // Merge messages should be concise
		Temperature:         &temp,
	}
}

// parseMergeResponse parses the API response into a MergeMessageResponse.
func (c *CerebrasProvider) parseMergeResponse(resp *cerebrasResponse) (*MergeMessageResponse, error) {
	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from AI")
	}

	content := resp.Choices[0].Message.Content

	// Parse JSON response
	var mergeAnalysis struct {
		MergeMessage string `json:"merge_message"`
		Strategy     string `json:"strategy"`
		Reasoning    string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(content), &mergeAnalysis); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Create commit message
	commitMsg, err := domain.NewCommitMessage(mergeAnalysis.MergeMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit message: %w", err)
	}

	return &MergeMessageResponse{
		MergeMessage:      commitMsg,
		SuggestedStrategy: mergeAnalysis.Strategy,
		Reasoning:         mergeAnalysis.Reasoning,
	}, nil
}

// Helper functions

func mapActionType(action string) domain.ActionType {
	switch action {
	case "commit-direct":
		return domain.ActionCommitDirect
	case "create-branch":
		return domain.ActionCreateBranch
	case "review":
		return domain.ActionReview
	case "merge":
		return domain.ActionMerge
	default:
		return domain.ActionReview // Safe default
	}
}

func reduceDiffContext(diff string, maxTokens int) string {
	// Rough estimate: 4 characters per token
	maxChars := maxTokens * 4

	if len(diff) <= maxChars {
		return diff
	}

	// Truncate but try to keep complete hunks
	lines := strings.Split(diff, "\n")
	var sb strings.Builder
	currentSize := 0

	for _, line := range lines {
		if currentSize+len(line) > maxChars {
			sb.WriteString("\n... (diff truncated for token limit) ...")
			break
		}
		sb.WriteString(line)
		sb.WriteString("\n")
		currentSize += len(line) + 1
	}

	return sb.String()
}

func isRetryableError(err error) bool {
	// Check for network errors, timeouts, and 5xx status codes
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503")
}

func parseErrorResponse(statusCode int, body []byte) error {
	// Try to parse error details
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		if statusCode == 429 {
			return &FreeTierLimitError{
				Message:    errResp.Error.Message,
				RetryAfter: 60,
			}
		}
		return fmt.Errorf("API error (%d): %s", statusCode, errResp.Error.Message)
	}

	// If we can't parse the error, return the raw body for debugging
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500] + "..."
	}
	return fmt.Errorf("API error: status code %d, body: %s", statusCode, bodyStr)
}

func ptrFloat(f float64) *float64 {
	return &f
}

// Type definitions for Cerebras API

type cerebrasRequest struct {
	Model                string          `json:"model"`
	Messages             []message       `json:"messages"`
	ResponseFormat       *responseFormat `json:"response_format,omitempty"`
	MaxCompletionTokens  int             `json:"max_completion_tokens,omitempty"`
	Temperature          *float64        `json:"temperature,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *jsonSchema `json:"json_schema,omitempty"`
}

type jsonSchema struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema analysisSchema `json:"schema"`
}

type analysisSchema struct {
	Type                 string              `json:"type"`
	Properties           map[string]property `json:"properties"`
	Required             []string            `json:"required"`
	AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
}

type property struct {
	Type                 string              `json:"type"`
	Description          string              `json:"description,omitempty"`
	Enum                 []string            `json:"enum,omitempty"`
	Minimum              *float64            `json:"minimum,omitempty"`
	Maximum              *float64            `json:"maximum,omitempty"`
	Properties           map[string]property `json:"properties,omitempty"`
	Required             []string            `json:"required,omitempty"`
	Items                *property           `json:"items,omitempty"`
	AdditionalProperties *bool               `json:"additionalProperties,omitempty"`
}

type cerebrasResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Index   int     `json:"index"`
	Message message `json:"message"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// FreeTierLimitError represents a rate limit error for free tier users.
type FreeTierLimitError struct {
	Message    string
	RetryAfter int // Seconds to wait before retrying
}

func (e *FreeTierLimitError) Error() string {
	return e.Message
}
