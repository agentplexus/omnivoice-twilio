package omnimemory

import (
	"os"
	"testing"

	"github.com/plexusone/omnimemory/core"
	"github.com/plexusone/omnimemory/core/providertest"
)

func TestConformance(t *testing.T) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	storeID := os.Getenv("TWILIO_MEMORY_STORE_ID")
	profileID := os.Getenv("TWILIO_MEMORY_PROFILE_ID")

	if accountSid == "" || authToken == "" {
		t.Skip("TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN must be set for conformance tests")
	}
	if storeID == "" || profileID == "" {
		t.Skip("TWILIO_MEMORY_STORE_ID and TWILIO_MEMORY_PROFILE_ID must be set for conformance tests")
	}

	p, err := NewProvider(core.ProviderConfig{
		Options: map[string]any{
			"account_sid": accountSid,
			"auth_token":  authToken,
		},
	}, nil)
	if err != nil {
		t.Fatalf("NewProvider() error: %v", err)
	}
	defer func() { _ = p.Close() }()

	providertest.RunAll(t, providertest.Config{
		Provider:  p,
		TenantID:  storeID,
		SubjectID: profileID,
	})
}
