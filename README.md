# Omni-Twilio

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omni-twilio/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omni-twilio/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omni-twilio/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omni-twilio/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omni-twilio/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omni-twilio/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omni-twilio
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omni-twilio
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/omni-twilio
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomni-twilio
 [loc-svg]: https://tokei.rs/b1/github/plexusone/omni-twilio
 [repo-url]: https://github.com/plexusone/omni-twilio
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omni-twilio/blob/main/LICENSE

Go SDK for Twilio with adapters for [OmniChat](https://github.com/plexusone/omnichat) (SMS) and [OmniVoice](https://github.com/plexusone/omnivoice-core) (voice).

## Features

- 🎙️ **Voice Gateway**: Full-duplex phone calls with STT → LLM → TTS pipeline
- 📞 **CallSystem**: PSTN call handling (incoming/outgoing phone calls)
- 📡 **Transport**: Twilio Media Streams for real-time audio
- 🗣️ **TTS**: Text-to-speech via Twilio's Say verb (Alice, Polly, Google voices)
- 👂 **STT**: Speech recognition via Gather verb and real-time transcription
- 💬 **SMS/MMS/RCS**: Send/receive SMS, MMS, and RCS messages via OmniChat provider interface
- 🧠 **Memory**: Twilio Memory API provider for OmniMemory (semantic search, observations)

## Installation

```bash
go get github.com/plexusone/omni-twilio
```

## Package Structure

```
omni-twilio/
├── client/           # Exported Twilio REST API client
├── omnichat/         # SMS provider for omnichat
├── omnimemory/       # Memory provider for omnimemory (Twilio Memory API)
└── omnivoice/
    ├── gateway/      # Full-duplex voice gateway (STT→LLM→TTS)
    ├── callsystem/   # Call handling provider
    ├── transport/    # Media Streams provider
    ├── stt/          # Speech-to-text provider
    └── tts/          # Text-to-speech provider
```

## Quick Start

### SMS/MMS (OmniChat)

```go
import (
    "github.com/plexusone/omnichat/provider"
    "github.com/plexusone/omni-twilio/omnichat"
)

p, _ := omnichat.New(
    omnichat.WithAccountSID("ACxxxxxxxx"),
    omnichat.WithAuthToken("your-token"),
    omnichat.WithPhoneNumber("+15551234567"),
)

// Connect
p.Connect(ctx)

// Send SMS
p.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Hello from Twilio!",
})

// Send MMS with images
p.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Check out this photo!",
    Media: []provider.Media{
        {
            Type: provider.MediaTypeImage,
            URL:  "https://example.com/image.jpg",
        },
    },
})

// Enable RCS with automatic SMS/MMS fallback
pRCS, _ := omnichat.New(
    omnichat.WithAccountSID("ACxxxxxxxx"),
    omnichat.WithAuthToken("your-token"),
    omnichat.WithMessagingServiceSid("MGxxxxxxxx"), // RCS-enabled Messaging Service
)

// Send message (RCS with fallback to SMS/MMS)
pRCS.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Hello via RCS!",
})

// Send RCS with content template
pRCS.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Order update",
    Metadata: map[string]any{
        "content_sid":       "HXxxxxxxxx", // Pre-created content template
        "content_variables": `{"1": "John", "2": "#12345"}`,
    },
})

// Handle incoming SMS/MMS via webhook
http.Handle("/sms", p.WebhookHandler())

// In your message handler, access incoming media
p.OnMessage(func(ctx context.Context, msg provider.IncomingMessage) error {
    for _, media := range msg.Media {
        fmt.Printf("Received %s: %s\n", media.Type, media.URL)
    }
    return nil
})
```

### Memory (OmniMemory)

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omni-twilio/omnimemory" // Register Twilio provider
)

// Create client with Twilio Memory provider
client, _ := omnimemory.NewClient(core.ClientConfig{
    Providers: []core.ProviderConfig{
        {
            Name: core.ProviderNameTwilio,
            Options: map[string]any{
                "account_sid": os.Getenv("TWILIO_ACCOUNT_SID"),
                "auth_token":  os.Getenv("TWILIO_AUTH_TOKEN"),
            },
        },
    },
})

// Add a memory (observation)
memory, _ := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "store-id",   // Twilio Store ID
        SubjectID: "profile-id", // Twilio Profile ID
    },
    Type:    core.MemoryTypeObservation,
    Content: "User prefers dark mode interfaces",
})

// Semantic search
results, _ := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "store-id",
        SubjectID: "profile-id",
    },
    Query: "interface preferences",
    Limit: 10,
})

// Recall relevant memories
recalled, _ := client.Recall(ctx, &core.RecallRequest{
    Context: core.Context{
        TenantID:  "store-id",
        SubjectID: "profile-id",
    },
    Query:      "What does the user prefer?",
    MaxResults: 5,
})
```

**Concept Mapping:**

| OmniMemory | Twilio Memory API |
|------------|-------------------|
| TenantID | Store ID |
| SubjectID | Profile ID |
| Memory | Observation |
| Search/Recall | Recall API |

### Voice Calls (OmniVoice)

```go
import (
    "github.com/plexusone/omni-twilio/omnivoice/callsystem"
    "github.com/plexusone/omni-twilio/omnivoice/transport"
)

// Create call system
cs, _ := callsystem.New(
    callsystem.WithPhoneNumber("+15551234567"),
    callsystem.WithWebhookURL("wss://your-server.com/media-stream"),
)

// Handle incoming calls
cs.OnIncomingCall(func(call callsystem.Call) error {
    fmt.Printf("Incoming call from %s\n", call.From())
    return call.Answer(context.Background())
})

// Make outbound call
call, _ := cs.MakeCall(ctx, "+15559876543")
fmt.Printf("Call initiated: %s\n", call.ID())

// Set up webhooks
http.HandleFunc("/incoming", handleIncoming(cs))
http.HandleFunc("/media-stream", handleMediaStream(cs.Transport()))
```

### Voice Gateway (Full-Duplex)

The gateway package provides a complete voice agent solution with bidirectional audio. Two pipeline modes are available:

| Mode | Latency | Description |
|------|---------|-------------|
| `text` | 500-1000ms | Traditional STT → LLM → TTS pipeline |
| `realtime` | 100-200ms | Native voice-to-voice via OpenAI Realtime or Gemini Live |

#### Text Mode (STT → LLM → TTS)

```go
import "github.com/plexusone/omni-twilio/omnivoice/gateway"

gw, _ := gateway.New(gateway.Config{
    AccountSID:      os.Getenv("TWILIO_ACCOUNT_SID"),
    AuthToken:       os.Getenv("TWILIO_AUTH_TOKEN"),
    PhoneNumber:     "+15551234567",
    PublicURL:       "https://your-server.com",
    ListenAddr:      ":8080",

    // Voice pipeline providers
    STTProvider:     "deepgram",
    TTSProvider:     "elevenlabs",
    LLMProvider:     "anthropic",
    LLMModel:        "claude-sonnet-4-20250514",
    LLMSystemPrompt: "You are a helpful voice assistant.",
})

gw.Start(ctx)
```

#### Realtime Mode (Voice-to-Voice)

```go
import (
    "github.com/plexusone/omni-twilio/omnivoice/gateway"
    coregateway "github.com/plexusone/omnivoice-core/gateway"
    openaiRealtime "github.com/plexusone/omni-openai/omnivoice/realtime"
)

gw, _ := gateway.New(gateway.Config{
    AccountSID:  os.Getenv("TWILIO_ACCOUNT_SID"),
    AuthToken:   os.Getenv("TWILIO_AUTH_TOKEN"),
    PhoneNumber: "+15551234567",
    PublicURL:   "https://your-server.com",
    ListenAddr:  ":8080",

    // Realtime mode (~100ms latency)
    Mode:             coregateway.PipelineModeRealtime,
    RealtimeProvider: openaiRealtime.NewFactory(),
    RealtimeConfig: &coregateway.RealtimeConfig{
        Provider:     "openai",
        APIKey:       os.Getenv("OPENAI_API_KEY"),
        Model:        "gpt-4o-realtime-preview-2024-12-17",
        Voice:        "alloy",
        Instructions: "You are a helpful voice assistant.",
    },
})

gw.Start(ctx)
```

#### Registry Integration

Use omnivoice-core's provider registry for automatic discovery:

```go
import (
    omnivoice "github.com/plexusone/omnivoice-core"
    "github.com/plexusone/omnivoice-core/registry"
    _ "github.com/plexusone/omni-twilio/omnivoice/gateway" // Auto-register
)

// Get gateway via registry
gw, err := omnivoice.GetGatewayProvider("twilio",
    registry.WithAPIKey("TWILIO_AUTH_TOKEN"),
    registry.WithExtension("accountSID", "ACxxxxxxxx"),
    registry.WithExtension("phoneNumber", "+15551234567"),
)
```

**Architecture:**

```
┌──────────┐        ┌─────────────────┐        ┌───────────────────┐
│  Caller  │◄──────►│     Twilio      │◄──────►│   Voice Gateway   │
│  (PSTN)  │  PSTN  │  Media Streams  │   WS   │  (Text/Realtime)  │
└──────────┘        └─────────────────┘        └───────────────────┘
```

**Configure Twilio webhooks:**

- Voice URL: `https://your-server.com/voice/inbound` (POST)
- Status Callback: `https://your-server.com/voice/status` (POST)

**Monitor sessions:**

```go
session, _ := gw.GetSession(callSID)
for event := range session.Events() {
    switch event.Type {
    case gateway.EventUserTranscript:
        log.Printf("User: %s", event.Data)
    case gateway.EventAgentTranscript:
        log.Printf("Agent: %s", event.Data)
    case gateway.EventSessionEnded:
        log.Println("Call ended")
    }
}
```

### Direct Client Usage

```go
import "github.com/plexusone/omni-twilio/client"

c, _ := client.New(&client.Config{
    AccountSID: "ACxxxxxxxx",
    AuthToken:  "your-token",
})

// Send SMS
msg, _ := c.SendSMS(ctx, &client.SendSMSParams{
    To:   "+15559876543",
    From: "+15551234567",
    Body: "Hello!",
})

// Send MMS with media
mms, _ := c.SendSMS(ctx, &client.SendSMSParams{
    To:        "+15559876543",
    From:      "+15551234567",
    Body:      "Check this out!",
    MediaURLs: []string{"https://example.com/image.jpg"},
})

// Send RCS (with automatic SMS/MMS fallback)
rcs, _ := c.SendSMS(ctx, &client.SendSMSParams{
    To:                  "+15559876543",
    MessagingServiceSid: "MGxxxxxxxx", // RCS-enabled Messaging Service
    Body:                "Hello via RCS!",
})

// Send RCS with content template (rich cards, carousels)
rcsRich, _ := c.SendSMS(ctx, &client.SendSMSParams{
    To:                  "+15559876543",
    MessagingServiceSid: "MGxxxxxxxx",
    ContentSid:          "HXxxxxxxxx", // Pre-created template
    ContentVariables:    `{"1": "John", "2": "Welcome!"}`,
})

// Make call
call, _ := c.MakeCall(ctx, &client.MakeCallParams{
    To:    "+15559876543",
    From:  "+15551234567",
    Twiml: "<Response><Say>Hello!</Say></Response>",
})
```

### TTS (Text-to-Speech)

```go
import "github.com/plexusone/omni-twilio/omnivoice/tts"

provider, _ := tts.New(
    tts.WithVoice("Polly.Joanna"),
    tts.WithLanguage("en-US"),
)

// Generate TwiML
result, _ := provider.Synthesize(ctx, "Hello!", tts.SynthesisConfig{
    VoiceID: "Polly.Matthew",
})
// result.Audio contains TwiML
```

### STT (Speech-to-Text)

```go
import "github.com/plexusone/omni-twilio/omnivoice/stt"

provider, _ := stt.New(
    stt.WithLanguage("en-US"),
    stt.WithSpeechModel("phone_call"),
)

// Generate TwiML for speech recognition
twiml := provider.GenerateGatherTwiML(stt.GatherConfig{
    Input:         "speech",
    Language:      "en-US",
    SpeechTimeout: "auto",
    Action:        "/handle-speech",
    Prompt:        "Please say your account number",
})
```

### Transport (Media Streams)

```go
import "github.com/plexusone/omni-twilio/omnivoice/transport"

tr, _ := transport.New()

// Listen for Media Stream connections
connCh, _ := tr.Listen(ctx, "/media-stream")

for conn := range connCh {
    go func(c transport.Connection) {
        audio := make([]byte, 1024)
        for {
            n, _ := c.AudioOut().Read(audio)
            // Process audio...
            c.AudioIn().Write(responseAudio)
        }
    }(conn)
}
```

## Configuration

### Environment Variables

```bash
# Core credentials (required for all features)
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-auth-token"

# SMS/MMS/RCS
export TWILIO_MESSAGING_SERVICE_SID="MGxxxxxxxx"  # Optional: for RCS

# Memory API (for omnimemory provider)
export TWILIO_MEMORY_STORE_ID="store-id"      # Optional: default store
export TWILIO_MEMORY_PROFILE_ID="profile-id"  # Optional: default profile
```

### RCS Setup

RCS (Rich Communication Services) provides rich messaging features with automatic fallback to SMS/MMS:

1. Create a Messaging Service in the Twilio Console
2. Add an RCS sender to your Messaging Service (requires carrier approval)
3. Use `WithMessagingServiceSid()` or set `TWILIO_MESSAGING_SERVICE_SID`

RCS features:

- Branded sender identity
- Rich cards and carousels (via Content API templates)
- Suggested replies and actions
- Read receipts and typing indicators
- Automatic fallback to SMS/MMS when RCS unavailable

### Available Voices

**Twilio Basic**: `alice`, `man`, `woman`

**Amazon Polly**: `Polly.Joanna`, `Polly.Matthew`, `Polly.Amy`, `Polly.Brian`, etc.

**Google TTS**: `Google.en-US-Standard-A` through `D`, `Google.en-US-Wavenet-A` through `D`

## Testing

```bash
# Unit tests
go test -v ./...

# Integration tests (requires credentials)
export TWILIO_ACCOUNT_SID="ACxxxx"
export TWILIO_AUTH_TOKEN="xxxx"
export TWILIO_PHONE_NUMBER="+15551234567"
go test -v -tags=integration ./...
```

## Migration

### From twilio-go (v0.5.0)

This package was renamed from `twilio-go` to `omni-twilio` in v0.5.0 to align with the omni-* naming convention.

```bash
# Update imports
find . -name "*.go" -exec sed -i '' \
  's|github.com/plexusone/twilio-go|github.com/plexusone/omni-twilio|g' {} +

# Update go.mod
go get github.com/plexusone/omni-twilio@v0.5.0
go mod tidy
```

### From omnivoice-twilio (v0.4.0)

| Before | After |
|--------|-------|
| `github.com/plexusone/omnivoice-twilio/callsystem` | `github.com/plexusone/omni-twilio/omnivoice/callsystem` |
| `github.com/plexusone/omnivoice-twilio/transport` | `github.com/plexusone/omni-twilio/omnivoice/transport` |
| `github.com/plexusone/omnivoice-twilio/tts` | `github.com/plexusone/omni-twilio/omnivoice/tts` |
| `github.com/plexusone/omnivoice-twilio/stt` | `github.com/plexusone/omni-twilio/omnivoice/stt` |

Added in v0.4.0:

- `github.com/plexusone/omni-twilio/client` - Exported Twilio client
- `github.com/plexusone/omni-twilio/omnichat` - SMS provider for OmniChat

## Related Packages

- [omnivoice-core](https://github.com/plexusone/omnivoice-core) - Voice interfaces
- [omnichat](https://github.com/plexusone/omnichat) - Chat interfaces
- [omnimemory](https://github.com/plexusone/omnimemory) - Memory interfaces
- [elevenlabs-go](https://github.com/plexusone/elevenlabs-go) - ElevenLabs SDK

## License

MIT
