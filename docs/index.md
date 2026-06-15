# Omni-Twilio

Go SDK for Twilio with adapters for [OmniChat](https://github.com/plexusone/omnichat) (SMS) and [OmniVoice](https://github.com/plexusone/omnivoice-core) (voice).

## Features

- **Voice Gateway**: Full-duplex phone calls with dual pipeline modes:
  - **Text Mode**: STT → LLM → TTS (500-1000ms latency, custom voices)
  - **Realtime Mode**: Native voice-to-voice (100-200ms latency)
- **Client**: Exported Twilio REST API client for calls and SMS
- **Transport**: Twilio Media Streams for real-time audio
- **TTS**: Text-to-speech via Twilio's Say verb (Alice, Polly, Google voices)
- **STT**: Speech recognition via Gather verb and real-time transcription
- **SMS/MMS/RCS**: Send/receive messages via OmniChat provider interface

## Package Structure

```
omni-twilio/
├── client/           # Exported Twilio REST API client
├── omnichat/         # SMS provider for omnichat
└── omnivoice/
    ├── gateway/      # Full-duplex voice gateway (STT→LLM→TTS)
    ├── callsystem/   # Call handling provider
    ├── transport/    # Media Streams provider
    ├── stt/          # Speech-to-text provider
    └── tts/          # Text-to-speech provider
```

## Quick Start

### SMS (OmniChat)

```go
import "github.com/plexusone/omni-twilio/omnichat"

provider, _ := omnichat.New(
    omnichat.WithAccountSID("ACxxxxxxxx"),
    omnichat.WithAuthToken("your-token"),
    omnichat.WithPhoneNumber("+15551234567"),
)

// Connect and send SMS
provider.Connect(ctx)
provider.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Hello from Twilio!",
})

// Handle incoming SMS via webhook
http.Handle("/sms", provider.WebhookHandler())
```

### Voice (OmniVoice)

```go
import "github.com/plexusone/omni-twilio/omnivoice/callsystem"

provider, _ := callsystem.New(
    callsystem.WithAccountSID("ACxxxxxxxx"),
    callsystem.WithAuthToken("your-token"),
    callsystem.WithPhoneNumber("+15551234567"),
)

// Make an outbound call
call, _ := provider.MakeCall(ctx, "+15559876543", callbackURL)
```

## Installation

```bash
go get github.com/plexusone/omni-twilio
```

## Links

- [GitHub Repository](https://github.com/plexusone/omni-twilio)
- [Go Package Documentation](https://pkg.go.dev/github.com/plexusone/omni-twilio)
- [Changelog](https://github.com/plexusone/omni-twilio/blob/main/CHANGELOG.md)
