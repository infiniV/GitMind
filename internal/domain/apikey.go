package domain

import (
	"errors"
	"fmt"
)

// APITier represents the tier level of an API key, determining rate limits and features.
type APITier int

const (
	// TierUnknown indicates the tier hasn't been determined yet.
	TierUnknown APITier = iota
	// TierFree indicates a free tier API key with reduced rate limits.
	TierFree
	// TierPro indicates a paid tier API key with higher rate limits.
	TierPro
)

// String returns the string representation of the API tier.
func (t APITier) String() string {
	switch t {
	case TierFree:
		return "free"
	case TierPro:
		return "pro"
	case TierUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("APITier(%d)", t)
	}
}

// ParseAPITier parses a string into an APITier.
func ParseAPITier(s string) (APITier, error) {
	switch s {
	case "free":
		return TierFree, nil
	case "pro":
		return TierPro, nil
	case "unknown":
		return TierUnknown, nil
	default:
		return TierUnknown, fmt.Errorf("invalid API tier: %s", s)
	}
}

// APIKey represents an AI provider API key with tier information.
type APIKey struct {
	key      string
	provider string
	tier     APITier
}

// NewAPIKey creates a new APIKey with unknown tier.
func NewAPIKey(key, provider string) (*APIKey, error) {
	if key == "" {
		return nil, errors.New("API key cannot be empty")
	}
	if provider == "" {
		return nil, errors.New("provider cannot be empty")
	}
	return &APIKey{
		key:      key,
		provider: provider,
		tier:     TierUnknown,
	}, nil
}

// Key returns the API key value.
func (a *APIKey) Key() string {
	return a.key
}

// Provider returns the provider name.
func (a *APIKey) Provider() string {
	return a.provider
}

// Tier returns the current tier.
func (a *APIKey) Tier() APITier {
	return a.tier
}

// SetTier updates the tier information.
func (a *APIKey) SetTier(tier APITier) {
	a.tier = tier
}

// IsFree returns true if the API key is on the free tier.
func (a *APIKey) IsFree() bool {
	return a.tier == TierFree
}

// IsPro returns true if the API key is on the pro tier.
func (a *APIKey) IsPro() bool {
	return a.tier == TierPro
}

// MaxTokensPerRequest returns the recommended maximum tokens per request based on tier.
func (a *APIKey) MaxTokensPerRequest() int {
	if a.IsFree() {
		return 2000 // Conservative for free tier
	}
	return 8000 // Pro tier can handle more
}

// ShouldReduceContext returns true if context should be reduced for this API key.
func (a *APIKey) ShouldReduceContext() bool {
	return a.tier == TierFree || a.tier == TierUnknown
}

// String returns a safe string representation (doesn't expose the key).
func (a *APIKey) String() string {
	maskedKey := "***"
	if len(a.key) > 4 {
		maskedKey = a.key[:4] + "***"
	}
	return fmt.Sprintf("APIKey{provider: %s, tier: %s, key: %s}", a.provider, a.tier, maskedKey)
}
