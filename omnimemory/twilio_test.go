package omnimemory

import (
	"testing"

	"github.com/plexusone/omnimemory/core"
)

func TestNewProvider_MissingAccountSid(t *testing.T) {
	// Clear env vars for test
	t.Setenv("TWILIO_ACCOUNT_SID", "")
	t.Setenv("TWILIO_AUTH_TOKEN", "")

	_, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"auth_token": "test-token",
		},
	}, nil)

	if err == nil {
		t.Error("expected error for missing account_sid")
	}

	validationErr, ok := err.(*core.ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "account_sid" {
		t.Errorf("expected field 'account_sid', got %q", validationErr.Field)
	}
}

func TestNewProvider_MissingAuthToken(t *testing.T) {
	// Clear env vars for test
	t.Setenv("TWILIO_ACCOUNT_SID", "")
	t.Setenv("TWILIO_AUTH_TOKEN", "")

	_, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"account_sid": "ACtest123",
		},
	}, nil)

	if err == nil {
		t.Error("expected error for missing auth_token")
	}

	validationErr, ok := err.(*core.ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "auth_token" {
		t.Errorf("expected field 'auth_token', got %q", validationErr.Field)
	}
}

//nolint:gosec // Test credentials are intentionally hardcoded
func TestNewProvider_FromOptions(t *testing.T) {
	// Clear env vars for test
	t.Setenv("TWILIO_ACCOUNT_SID", "")
	t.Setenv("TWILIO_AUTH_TOKEN", "")

	provider, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"account_sid": "ACtest123456789",
			"auth_token":  "test-auth-token",
		},
	}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.Name() != "twilio" {
		t.Errorf("expected name 'twilio', got %q", provider.Name())
	}
}

func TestNewProvider_FromEnv(t *testing.T) {
	t.Setenv("TWILIO_ACCOUNT_SID", "ACenv123456789")
	t.Setenv("TWILIO_AUTH_TOKEN", "env-auth-token")

	provider, err := NewProvider(core.ProviderConfig{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.Name() != "twilio" {
		t.Errorf("expected name 'twilio', got %q", provider.Name())
	}
}

func TestProvider_Close(t *testing.T) {
	t.Setenv("TWILIO_ACCOUNT_SID", "ACtest123456789")
	t.Setenv("TWILIO_AUTH_TOKEN", "test-auth-token")

	provider, err := NewProvider(core.ProviderConfig{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := provider.Close(); err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}
}

func TestGetOption(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]any
		key      string
		defVal   string
		expected string
	}{
		{
			name:     "nil options",
			options:  nil,
			key:      "foo",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "key exists",
			options:  map[string]any{"foo": "bar"},
			key:      "foo",
			defVal:   "default",
			expected: "bar",
		},
		{
			name:     "key missing",
			options:  map[string]any{"other": "value"},
			key:      "foo",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "wrong type",
			options:  map[string]any{"foo": 123},
			key:      "foo",
			defVal:   "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOption(tt.options, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("getOption() = %q, want %q", result, tt.expected)
			}
		})
	}
}
