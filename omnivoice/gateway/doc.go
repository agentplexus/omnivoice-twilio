// Package gateway provides an HTTP/WebSocket gateway for handling Twilio voice calls
// with full-duplex bidirectional audio using Twilio Media Streams.
//
// # Architecture
//
// The gateway handles the complete voice call flow:
//
//	┌──────────┐        ┌─────────────────┐        ┌───────────────────┐
//	│  Caller  │◄──────►│     Twilio      │◄──────►│   Voice Gateway   │
//	│  (PSTN)  │  PSTN  │  Media Streams  │   WS   │   (STT→LLM→TTS)   │
//	└──────────┘        └─────────────────┘        └───────────────────┘
//
// # Quick Start
//
// Create and start a voice gateway:
//
//	gw, err := gateway.New(gateway.Config{
//	    AccountSID:  os.Getenv("TWILIO_ACCOUNT_SID"),
//	    AuthToken:   os.Getenv("TWILIO_AUTH_TOKEN"),
//	    PhoneNumber: "+1234567890",
//	    PublicURL:   "https://your-server.com",
//	    ListenAddr:  ":8080",
//
//	    // Voice pipeline providers
//	    STTProvider: "deepgram",
//	    TTSProvider: "elevenlabs",
//	    LLMProvider: "anthropic",
//	    LLMModel:    "claude-sonnet-4-20250514",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Handle incoming calls
//	gw.OnCall(func(call *gateway.CallInfo) error {
//	    log.Printf("Incoming call from %s", call.From)
//	    return nil // Accept the call
//	})
//
//	// Start the gateway
//	ctx := context.Background()
//	if err := gw.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// # Twilio Configuration
//
// Configure your Twilio phone number webhooks:
//
//   - Voice webhook URL: https://your-server.com/voice/inbound (POST)
//   - Status callback URL: https://your-server.com/voice/status (POST)
//
// # Voice Pipeline
//
// The gateway processes audio through a three-stage pipeline:
//
//  1. STT (Speech-to-Text): Converts user speech to text
//  2. LLM (Language Model): Generates a response
//  3. TTS (Text-to-Speech): Converts response to audio
//
// Each stage can be configured with different providers:
//
//   - STT: deepgram, whisper, google
//   - LLM: anthropic, openai
//   - TTS: elevenlabs, openai, google
//
// # Session Events
//
// Monitor call progress through session events:
//
//	session, _ := gw.GetSession(callSID)
//	for event := range session.Events() {
//	    switch event.Type {
//	    case gateway.EventUserTranscript:
//	        log.Printf("User said: %s", event.Data)
//	    case gateway.EventAgentTranscript:
//	        log.Printf("Agent said: %s", event.Data)
//	    case gateway.EventSessionEnded:
//	        log.Printf("Call ended")
//	    }
//	}
//
// # Outbound Calls
//
// Initiate outbound calls:
//
//	session, err := gw.MakeCall(ctx, "+1987654321")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Monitor the call
//	for event := range session.Events() {
//	    // Handle events
//	}
//
// # Interruption Handling
//
// The gateway supports three interruption modes:
//
//   - "immediate": Stop agent speech immediately when user starts speaking
//   - "after_sentence": Finish current sentence before stopping
//   - "disabled": Ignore interruptions
//
// Configure via InterruptionMode in Config.
//
// # Environment Variables
//
// The gateway reads credentials from environment variables if not provided in config:
//
//   - TWILIO_ACCOUNT_SID: Twilio Account SID
//   - TWILIO_AUTH_TOKEN: Twilio Auth Token
//   - DEEPGRAM_API_KEY: Deepgram API key (for STT)
//   - OPENAI_API_KEY: OpenAI API key (for STT/LLM/TTS)
//   - ANTHROPIC_API_KEY: Anthropic API key (for LLM)
//   - ELEVENLABS_API_KEY: ElevenLabs API key (for TTS)
//   - GOOGLE_API_KEY: Google API key (for STT/TTS)
package gateway
