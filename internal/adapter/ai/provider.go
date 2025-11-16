package ai

import (
	"context"

	"github.com/yourusername/gitman/internal/domain"
)

// Provider defines the interface for AI providers (Cerebras, OpenAI, Claude, etc.)
type Provider interface {
	// Analyze analyzes git changes and returns a decision about how to proceed.
	Analyze(ctx context.Context, request AnalysisRequest) (*AnalysisResponse, error)

	// GenerateMergeMessage generates a merge commit message based on branch commits.
	GenerateMergeMessage(ctx context.Context, request MergeMessageRequest) (*MergeMessageResponse, error)

	// DetectTier attempts to detect the API key tier (free vs pro).
	DetectTier(ctx context.Context) (domain.APITier, error)

	// GetName returns the provider name (e.g., "cerebras", "openai").
	GetName() string

	// ValidateKey checks if the API key is valid.
	ValidateKey(ctx context.Context) error
}

// AnalysisRequest contains all information needed for the AI to analyze changes.
type AnalysisRequest struct {
	Repository             *domain.Repository // Current repository state
	BranchInfo             *domain.BranchInfo // Branch context and metadata
	Diff                   string             // Git diff content
	RecentLog              []string           // Recent commit messages for context
	UserPrompt             string             // Optional user-provided context
	APIKey                 *domain.APIKey     // API key with tier information
	UseConventionalCommits bool               // Whether to use conventional commit format
	MergeOpportunity       bool               // Whether branch is ready for merge
	MergeTargetBranch      string             // Target branch for merge (if MergeOpportunity is true)
	MergeCommitCount       int                // Number of commits to be merged
}

// AnalysisResponse contains the AI's analysis and recommendations.
type AnalysisResponse struct {
	Decision         *domain.Decision // Primary decision and reasoning
	TokensUsed       int              // Number of tokens consumed
	Model            string           // Model used for analysis
	ProcessingTimeMs int              // Processing time in milliseconds
}

// MergeMessageRequest contains information needed to generate a merge commit message.
type MergeMessageRequest struct {
	SourceBranch string   // Branch being merged from
	TargetBranch string   // Branch being merged into
	Commits      []string // Commit messages to summarize
	CommitCount  int      // Number of commits being merged
	APIKey       *domain.APIKey
}

// MergeMessageResponse contains the AI-generated merge message and strategy.
type MergeMessageResponse struct {
	MergeMessage      *domain.CommitMessage // Generated merge commit message
	SuggestedStrategy string                // Suggested merge strategy ("squash", "regular", etc.)
	Reasoning         string                // Explanation for the suggestion
	TokensUsed        int                   // Number of tokens consumed
	Model             string                // Model used
}

// ProviderConfig contains configuration for creating a provider.
type ProviderConfig struct {
	APIKey    string
	BaseURL   string // Optional custom base URL
	Model     string // Model to use (optional, provider will choose default)
	Timeout   int    // Request timeout in seconds (default: 30)
	MaxRetries int   // Maximum number of retries (default: 3)
}

// Factory creates AI providers.
type Factory struct {
	providers map[string]func(*domain.APIKey, ProviderConfig) Provider
}

// NewFactory creates a new provider factory.
func NewFactory() *Factory {
	factory := &Factory{
		providers: make(map[string]func(*domain.APIKey, ProviderConfig) Provider),
	}

	// Register default providers
	factory.Register("cerebras", func(apiKey *domain.APIKey, config ProviderConfig) Provider {
		return NewCerebrasProvider(apiKey, config)
	})

	return factory
}

// Register registers a provider constructor.
func (f *Factory) Register(name string, constructor func(*domain.APIKey, ProviderConfig) Provider) {
	f.providers[name] = constructor
}

// Create creates a provider by name.
func (f *Factory) Create(name string, apiKey *domain.APIKey, config ProviderConfig) (Provider, error) {
	constructor, ok := f.providers[name]
	if !ok {
		return nil, &ProviderNotFoundError{ProviderName: name}
	}

	return constructor(apiKey, config), nil
}

// ProviderNotFoundError is returned when a provider is not found.
type ProviderNotFoundError struct {
	ProviderName string
}

func (e *ProviderNotFoundError) Error() string {
	return "provider not found: " + e.ProviderName
}
