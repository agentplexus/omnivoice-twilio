package client

import "testing"

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"10", 10},
		{"", 0},
		{"invalid", 0},
		{"-1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseInt(tt.input)
			if got != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSendSMSParams(t *testing.T) {
	// Test that SendSMSParams can hold both SMS and MMS data
	smsParams := SendSMSParams{
		To:   "+15551234567",
		From: "+15559876543",
		Body: "Hello, SMS!",
	}

	if smsParams.To != "+15551234567" {
		t.Errorf("expected To +15551234567, got %s", smsParams.To)
	}
	if len(smsParams.MediaURLs) != 0 {
		t.Errorf("expected 0 MediaURLs for SMS, got %d", len(smsParams.MediaURLs))
	}

	mmsParams := SendSMSParams{
		To:   "+15551234567",
		From: "+15559876543",
		Body: "Check out this image!",
		MediaURLs: []string{
			"https://example.com/image1.jpg",
			"https://example.com/image2.png",
		},
	}

	if len(mmsParams.MediaURLs) != 2 {
		t.Errorf("expected 2 MediaURLs for MMS, got %d", len(mmsParams.MediaURLs))
	}
}

func TestMessage(t *testing.T) {
	// Test that Message can hold MMS data
	msg := Message{
		SID:       "MM123",
		To:        "+15551234567",
		From:      "+15559876543",
		Body:      "Hello!",
		NumMedia:  2,
		MediaURLs: []string{"https://example.com/1.jpg", "https://example.com/2.jpg"},
	}

	if msg.NumMedia != 2 {
		t.Errorf("expected NumMedia 2, got %d", msg.NumMedia)
	}
	if len(msg.MediaURLs) != 2 {
		t.Errorf("expected 2 MediaURLs, got %d", len(msg.MediaURLs))
	}
}

func TestSendSMSParamsRCS(t *testing.T) {
	// Test RCS with MessagingServiceSid
	rcsParams := SendSMSParams{
		To:                  "+15551234567",
		MessagingServiceSid: "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Body:                "Hello via RCS!",
	}

	if rcsParams.MessagingServiceSid != "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("expected MessagingServiceSid MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx, got %s", rcsParams.MessagingServiceSid)
	}
	if rcsParams.From != "" {
		t.Errorf("expected empty From for RCS, got %s", rcsParams.From)
	}
}

func TestSendSMSParamsRCSWithContentTemplate(t *testing.T) {
	// Test RCS with content template
	rcsParams := SendSMSParams{
		To:                  "+15551234567",
		MessagingServiceSid: "MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		ContentSid:          "HXxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		ContentVariables:    `{"1": "John", "2": "Welcome!"}`,
	}

	if rcsParams.ContentSid != "HXxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("expected ContentSid HXxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx, got %s", rcsParams.ContentSid)
	}
	if rcsParams.ContentVariables != `{"1": "John", "2": "Welcome!"}` {
		t.Errorf("expected ContentVariables JSON, got %s", rcsParams.ContentVariables)
	}
}

func TestMessageType(t *testing.T) {
	tests := []struct {
		name     string
		params   *SendSMSParams
		expected string
	}{
		{
			name: "SMS",
			params: &SendSMSParams{
				To:   "+15551234567",
				From: "+15559876543",
				Body: "Hello SMS",
			},
			expected: "SMS",
		},
		{
			name: "MMS",
			params: &SendSMSParams{
				To:        "+15551234567",
				From:      "+15559876543",
				Body:      "Hello MMS",
				MediaURLs: []string{"https://example.com/image.jpg"},
			},
			expected: "MMS",
		},
		{
			name: "RCS with MessagingServiceSid",
			params: &SendSMSParams{
				To:                  "+15551234567",
				MessagingServiceSid: "MGxxxxxxxx",
				Body:                "Hello RCS",
			},
			expected: "RCS",
		},
		{
			name: "RCS with ContentSid",
			params: &SendSMSParams{
				To:         "+15551234567",
				From:       "+15559876543",
				ContentSid: "HXxxxxxxxx",
			},
			expected: "RCS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messageType(tt.params)
			if got != tt.expected {
				t.Errorf("messageType() = %s, want %s", got, tt.expected)
			}
		})
	}
}
