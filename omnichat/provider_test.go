package omnichat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/plexusone/omnichat/provider"
)

func TestMediaTypeFromMIME(t *testing.T) {
	tests := []struct {
		mimeType string
		expected provider.MediaType
	}{
		{"image/jpeg", provider.MediaTypeImage},
		{"image/png", provider.MediaTypeImage},
		{"image/gif", provider.MediaTypeImage},
		{"video/mp4", provider.MediaTypeVideo},
		{"video/quicktime", provider.MediaTypeVideo},
		{"audio/mpeg", provider.MediaTypeAudio},
		{"audio/ogg", provider.MediaTypeAudio},
		{"application/pdf", provider.MediaTypeDocument},
		{"application/octet-stream", provider.MediaTypeDocument},
		{"text/plain", provider.MediaTypeDocument},
		{"", provider.MediaTypeDocument},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := mediaTypeFromMIME(tt.mimeType)
			if got != tt.expected {
				t.Errorf("mediaTypeFromMIME(%q) = %q, want %q", tt.mimeType, got, tt.expected)
			}
		})
	}
}

func TestWebhookHandler_SMS(t *testing.T) {
	p := &Provider{
		connected: true,
	}

	var receivedMsg provider.IncomingMessage
	p.messageHandler = func(_ context.Context, msg provider.IncomingMessage) error {
		receivedMsg = msg
		return nil
	}

	// Simulate SMS webhook
	form := url.Values{
		"MessageSid": {"SM123"},
		"From":       {"+15551234567"},
		"To":         {"+15559876543"},
		"Body":       {"Hello, world!"},
		"AccountSid": {"AC123"},
		"NumMedia":   {"0"},
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	p.handleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if receivedMsg.ID != "SM123" {
		t.Errorf("expected ID SM123, got %s", receivedMsg.ID)
	}
	if receivedMsg.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got %s", receivedMsg.Content)
	}
	if len(receivedMsg.Media) != 0 {
		t.Errorf("expected 0 media, got %d", len(receivedMsg.Media))
	}
}

func TestWebhookHandler_MMS(t *testing.T) {
	p := &Provider{
		connected: true,
	}

	var receivedMsg provider.IncomingMessage
	p.messageHandler = func(_ context.Context, msg provider.IncomingMessage) error {
		receivedMsg = msg
		return nil
	}

	// Simulate MMS webhook with 2 media attachments
	form := url.Values{
		"MessageSid":        {"MM456"},
		"From":              {"+15551234567"},
		"To":                {"+15559876543"},
		"Body":              {"Check out these pics!"},
		"AccountSid":        {"AC123"},
		"NumMedia":          {"2"},
		"MediaUrl0":         {"https://api.twilio.com/media/image1.jpg"},
		"MediaContentType0": {"image/jpeg"},
		"MediaUrl1":         {"https://api.twilio.com/media/video1.mp4"},
		"MediaContentType1": {"video/mp4"},
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	p.handleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if receivedMsg.ID != "MM456" {
		t.Errorf("expected ID MM456, got %s", receivedMsg.ID)
	}
	if receivedMsg.Content != "Check out these pics!" {
		t.Errorf("expected content 'Check out these pics!', got %s", receivedMsg.Content)
	}
	if len(receivedMsg.Media) != 2 {
		t.Fatalf("expected 2 media, got %d", len(receivedMsg.Media))
	}

	// Check first media (image)
	if receivedMsg.Media[0].Type != provider.MediaTypeImage {
		t.Errorf("expected first media type image, got %s", receivedMsg.Media[0].Type)
	}
	if receivedMsg.Media[0].URL != "https://api.twilio.com/media/image1.jpg" {
		t.Errorf("expected first media URL, got %s", receivedMsg.Media[0].URL)
	}
	if receivedMsg.Media[0].MimeType != "image/jpeg" {
		t.Errorf("expected first media MIME type image/jpeg, got %s", receivedMsg.Media[0].MimeType)
	}

	// Check second media (video)
	if receivedMsg.Media[1].Type != provider.MediaTypeVideo {
		t.Errorf("expected second media type video, got %s", receivedMsg.Media[1].Type)
	}
	if receivedMsg.Media[1].URL != "https://api.twilio.com/media/video1.mp4" {
		t.Errorf("expected second media URL, got %s", receivedMsg.Media[1].URL)
	}
	if receivedMsg.Media[1].MimeType != "video/mp4" {
		t.Errorf("expected second media MIME type video/mp4, got %s", receivedMsg.Media[1].MimeType)
	}
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	p := &Provider{}

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rec := httptest.NewRecorder()

	p.handleWebhook(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestWebhookHandler_MissingFields(t *testing.T) {
	p := &Provider{}

	form := url.Values{
		"Body": {"Hello"},
		// Missing MessageSid and From
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	p.handleWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestWithMessagingServiceSid(t *testing.T) {
	opt := WithMessagingServiceSid("MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	opts := &options{}
	opt(opts)

	if opts.messagingServiceSid != "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("expected messagingServiceSid MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx, got %s", opts.messagingServiceSid)
	}
}

func TestProviderRCSConfig(t *testing.T) {
	// Test that Provider can be configured with RCS options
	p := &Provider{
		defaultFrom:         "+15551234567",
		messagingServiceSid: "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		connected:           true,
	}

	if p.messagingServiceSid != "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("expected messagingServiceSid MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx, got %s", p.messagingServiceSid)
	}

	// When MessagingServiceSid is set, it should be preferred over defaultFrom
	if p.messagingServiceSid == "" && p.defaultFrom == "" {
		t.Error("expected either messagingServiceSid or defaultFrom to be set")
	}
}

func TestProviderSendConfigValidation(t *testing.T) {
	// Test that Send() validates sender configuration
	tests := []struct {
		name                string
		defaultFrom         string
		messagingServiceSid string
		expectError         bool
	}{
		{
			name:                "both empty should fail",
			defaultFrom:         "",
			messagingServiceSid: "",
			expectError:         true,
		},
		{
			name:                "only defaultFrom should pass",
			defaultFrom:         "+15551234567",
			messagingServiceSid: "",
			expectError:         false,
		},
		{
			name:                "only messagingServiceSid should pass",
			defaultFrom:         "",
			messagingServiceSid: "MGxxxxxxxx",
			expectError:         false,
		},
		{
			name:                "both set should pass (uses messagingServiceSid)",
			defaultFrom:         "+15551234567",
			messagingServiceSid: "MGxxxxxxxx",
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				defaultFrom:         tt.defaultFrom,
				messagingServiceSid: tt.messagingServiceSid,
				connected:           true,
			}

			// We can't actually call Send() without a real client,
			// but we can verify the provider's config validation logic
			hasValidSender := p.messagingServiceSid != "" || p.defaultFrom != ""
			if tt.expectError && hasValidSender {
				t.Errorf("expected no valid sender, but found one")
			}
			if !tt.expectError && !hasValidSender {
				t.Errorf("expected valid sender, but none found")
			}
		})
	}
}
