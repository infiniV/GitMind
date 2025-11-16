package domain

import (
	"testing"
)

func TestNewAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		provider    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid API key",
			key:      "sk-test-123456",
			provider: "cerebras",
			wantErr:  false,
		},
		{
			name:        "empty key",
			key:         "",
			provider:    "cerebras",
			wantErr:     true,
			errContains: "API key cannot be empty",
		},
		{
			name:        "empty provider",
			key:         "sk-test-123456",
			provider:    "",
			wantErr:     true,
			errContains: "provider cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, err := NewAPIKey(tt.key, tt.provider)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAPIKey() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewAPIKey() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewAPIKey() unexpected error = %v", err)
				return
			}

			if apiKey.Key() != tt.key {
				t.Errorf("APIKey.Key() = %v, want %v", apiKey.Key(), tt.key)
			}
			if apiKey.Provider() != tt.provider {
				t.Errorf("APIKey.Provider() = %v, want %v", apiKey.Provider(), tt.provider)
			}
			if apiKey.Tier() != TierUnknown {
				t.Errorf("APIKey.Tier() = %v, want %v", apiKey.Tier(), TierUnknown)
			}
		})
	}
}

func TestAPIKey_SetTier(t *testing.T) {
	apiKey, _ := NewAPIKey("test-key", "cerebras")

	tests := []struct {
		name string
		tier APITier
	}{
		{"set to free", TierFree},
		{"set to pro", TierPro},
		{"set to unknown", TierUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey.SetTier(tt.tier)
			if apiKey.Tier() != tt.tier {
				t.Errorf("APIKey.Tier() = %v, want %v", apiKey.Tier(), tt.tier)
			}
		})
	}
}

func TestAPIKey_IsFree(t *testing.T) {
	tests := []struct {
		name string
		tier APITier
		want bool
	}{
		{"free tier", TierFree, true},
		{"pro tier", TierPro, false},
		{"unknown tier", TierUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, _ := NewAPIKey("test-key", "cerebras")
			apiKey.SetTier(tt.tier)

			if got := apiKey.IsFree(); got != tt.want {
				t.Errorf("APIKey.IsFree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKey_IsPro(t *testing.T) {
	tests := []struct {
		name string
		tier APITier
		want bool
	}{
		{"free tier", TierFree, false},
		{"pro tier", TierPro, true},
		{"unknown tier", TierUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, _ := NewAPIKey("test-key", "cerebras")
			apiKey.SetTier(tt.tier)

			if got := apiKey.IsPro(); got != tt.want {
				t.Errorf("APIKey.IsPro() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKey_MaxTokensPerRequest(t *testing.T) {
	tests := []struct {
		name string
		tier APITier
		want int
	}{
		{"free tier", TierFree, 2000},
		{"pro tier", TierPro, 8000},
		{"unknown tier", TierUnknown, 8000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, _ := NewAPIKey("test-key", "cerebras")
			apiKey.SetTier(tt.tier)

			if got := apiKey.MaxTokensPerRequest(); got != tt.want {
				t.Errorf("APIKey.MaxTokensPerRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKey_ShouldReduceContext(t *testing.T) {
	tests := []struct {
		name string
		tier APITier
		want bool
	}{
		{"free tier should reduce", TierFree, true},
		{"pro tier should not reduce", TierPro, false},
		{"unknown tier should reduce", TierUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, _ := NewAPIKey("test-key", "cerebras")
			apiKey.SetTier(tt.tier)

			if got := apiKey.ShouldReduceContext(); got != tt.want {
				t.Errorf("APIKey.ShouldReduceContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		provider string
		tier     APITier
		want     string
	}{
		{
			name:     "masks long key",
			key:      "sk-test-123456789",
			provider: "cerebras",
			tier:     TierFree,
			want:     "APIKey{provider: cerebras, tier: free, key: sk-t***}",
		},
		{
			name:     "masks short key",
			key:      "abc",
			provider: "cerebras",
			tier:     TierPro,
			want:     "APIKey{provider: cerebras, tier: pro, key: ***}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, _ := NewAPIKey(tt.key, tt.provider)
			apiKey.SetTier(tt.tier)

			if got := apiKey.String(); got != tt.want {
				t.Errorf("APIKey.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPITier_String(t *testing.T) {
	tests := []struct {
		name string
		tier APITier
		want string
	}{
		{"free tier", TierFree, "free"},
		{"pro tier", TierPro, "pro"},
		{"unknown tier", TierUnknown, "unknown"},
		{"invalid tier", APITier(99), "APITier(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("APITier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAPITier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    APITier
		wantErr bool
	}{
		{"parse free", "free", TierFree, false},
		{"parse pro", "pro", TierPro, false},
		{"parse unknown", "unknown", TierUnknown, false},
		{"parse invalid", "invalid", TierUnknown, true},
		{"parse empty", "", TierUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAPITier(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAPITier() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseAPITier() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ParseAPITier() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
