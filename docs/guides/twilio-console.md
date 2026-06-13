# Twilio Console Configuration

Step-by-step guide for configuring your Twilio phone number to work with omni-twilio.

## Prerequisites

- A [Twilio account](https://www.twilio.com/try-twilio)
- A Twilio phone number
- Your application running with a public URL (or ngrok for development)

---

## Finding Your Credentials

1. Log in to [Twilio Console](https://console.twilio.com/)
2. On the dashboard, locate:
   - **Account SID**: Starts with `AC`
   - **Auth Token**: Click the eye icon to reveal

3. Set as environment variables:

```bash
export TWILIO_ACCOUNT_SID="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export TWILIO_AUTH_TOKEN="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

---

## Phone Number Configuration

### Navigate to Your Phone Number

1. Go to **Phone Numbers** → **Manage** → **Active numbers**
2. Click on the phone number you want to configure

### Voice Configuration

Scroll to the **Voice Configuration** section:

| Field | Value | Description |
|-------|-------|-------------|
| **Configure with** | Webhooks | Use webhook URLs |
| **A call comes in** | `https://your-server.com/voice/inbound` | Incoming call handler |
| **HTTP Method** | POST | Required method |
| **Primary handler fails** | (optional) | Fallback URL |
| **Call status changes** | `https://your-server.com/voice/status` | Status updates |

Click **Save configuration**.

### Messaging Configuration

Scroll to the **Messaging Configuration** section:

| Field | Value | Description |
|-------|-------|-------------|
| **Configure with** | Webhooks | Use webhook URLs |
| **A message comes in** | `https://your-server.com/webhook/twilio/sms` | Incoming SMS/MMS |
| **HTTP Method** | POST | Required method |
| **Primary handler fails** | (optional) | Fallback URL |

Click **Save configuration**.

---

## Voice Gateway Setup

### Webhook URLs

Configure these URLs in Twilio Console:

| Webhook | Path | Purpose |
|---------|------|---------|
| Voice URL | `/voice/inbound` | Handles incoming calls |
| Status Callback | `/voice/status` | Call status updates |

### Code Example

```go
import "github.com/plexusone/omni-twilio/omnivoice/gateway"

gw, err := gateway.New(gateway.Config{
    AccountSID:  os.Getenv("TWILIO_ACCOUNT_SID"),
    AuthToken:   os.Getenv("TWILIO_AUTH_TOKEN"),
    PhoneNumber: "+15551234567",
    PublicURL:   "https://your-server.com",
    ListenAddr:  ":8080",

    // Voice pipeline
    STTProvider: "deepgram",
    TTSProvider: "elevenlabs",
    LLMProvider: "anthropic",
})

// Start handling calls
gw.Start(ctx)
```

---

## SMS/MMS Setup

### Webhook URLs

Configure this URL in Twilio Console:

| Webhook | Path | Purpose |
|---------|------|---------|
| Message URL | `/webhook/twilio/sms` | Incoming SMS/MMS |

### Code Example

```go
import "github.com/plexusone/omni-twilio/omnichat"

provider, err := omnichat.New(
    omnichat.WithAccountSID(os.Getenv("TWILIO_ACCOUNT_SID")),
    omnichat.WithAuthToken(os.Getenv("TWILIO_AUTH_TOKEN")),
    omnichat.WithPhoneNumber("+15551234567"),
)

// Connect provider
provider.Connect(ctx)

// Mount webhook handler
http.Handle("/webhook/twilio/sms", provider.WebhookHandler())

// Handle incoming messages
provider.OnMessage(func(ctx context.Context, msg provider.IncomingMessage) error {
    fmt.Printf("From: %s, Body: %s\n", msg.SenderID, msg.Content)

    // Handle MMS attachments
    for _, media := range msg.Media {
        fmt.Printf("Media: %s (%s)\n", media.URL, media.MimeType)
    }

    return nil
})
```

---

## RCS Setup (Rich Communication Services)

RCS enables rich messaging with branded senders, rich cards, and suggested actions.

### Step 1: Create a Messaging Service

1. Go to **Messaging** → **Services**
2. Click **Create Messaging Service**
3. Name it (e.g., "My App RCS")
4. Click **Create**

### Step 2: Add Phone Number to Sender Pool

1. In your Messaging Service, click **Sender Pool**
2. Click **Add Senders** → **Phone Numbers**
3. Select your phone number
4. Click **Add Phone Numbers**

### Step 3: Add RCS Sender (Optional)

1. Click **Add Senders** → **RCS Sender**
2. Complete the RCS onboarding (requires carrier approval)

### Step 4: Get Messaging Service SID

1. Go to your Messaging Service overview
2. Copy the **Messaging Service SID** (starts with `MG`)

### Step 5: Configure Environment

```bash
export TWILIO_MESSAGING_SERVICE_SID="MGxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

### Code Example

```go
import "github.com/plexusone/omni-twilio/omnichat"

provider, err := omnichat.New(
    omnichat.WithAccountSID(os.Getenv("TWILIO_ACCOUNT_SID")),
    omnichat.WithAuthToken(os.Getenv("TWILIO_AUTH_TOKEN")),
    omnichat.WithMessagingServiceSid(os.Getenv("TWILIO_MESSAGING_SERVICE_SID")),
)

// Messages sent via RCS with automatic SMS/MMS fallback
provider.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Hello via RCS!",
})

// Send with content template (rich cards)
provider.Send(ctx, "+15559876543", provider.OutgoingMessage{
    Content: "Order update",
    Metadata: map[string]any{
        "content_sid":       "HXxxxxxxxx",
        "content_variables": `{"1": "John", "2": "#12345"}`,
    },
})
```

---

## Local Development with ngrok

### Install and Configure ngrok

1. Sign up at [ngrok.com](https://ngrok.com)
2. Install ngrok:
   ```bash
   # macOS
   brew install ngrok

   # Linux
   snap install ngrok
   ```
3. Authenticate:
   ```bash
   ngrok config add-authtoken YOUR_AUTH_TOKEN
   ```

### Start ngrok Tunnel

```bash
# Start your server on port 8080
go run main.go

# In another terminal, start ngrok
ngrok http 8080
```

ngrok will display a URL like:
```
Forwarding    https://abc123.ngrok.io -> http://localhost:8080
```

### Update Twilio Webhooks

1. Copy the ngrok HTTPS URL
2. Go to **Phone Numbers** → **Active numbers** → your number
3. Update webhook URLs:
   - Voice: `https://abc123.ngrok.io/voice/inbound`
   - SMS: `https://abc123.ngrok.io/webhook/twilio/sms`
4. Click **Save configuration**

!!! warning "URL Changes on Restart"
    Free ngrok URLs change every time you restart ngrok. Update Twilio webhooks accordingly, or use a paid ngrok plan for persistent URLs.

---

## Testing Your Setup

### Test Voice

1. Call your Twilio phone number
2. Check your server logs for incoming call events
3. Verify the call connects and TwiML is served

### Test SMS

1. Send an SMS to your Twilio phone number
2. Check your server logs for incoming message
3. Verify the webhook handler receives the message

### Test via Twilio Console

1. Go to **Phone Numbers** → **Active numbers** → your number
2. Under Voice, click **Make a test call**
3. Under Messaging, use the **Send a test message** feature

---

## Debugging

### Twilio Debugger

1. Go to **Monitor** → **Logs** → **Errors**
2. Click on errors to see request/response details

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| 11200 HTTP retrieval failure | Webhook URL not accessible | Check URL is public, HTTPS valid |
| 11205 Connection failure | Server not running | Start your server |
| 12100 Document parse failure | Invalid TwiML response | Check response format |
| No webhook received | Wrong URL configured | Verify URL in Console |

### Enable Request Logging

```go
import "log/slog"

provider, err := omnichat.New(
    omnichat.WithAccountSID("..."),
    omnichat.WithAuthToken("..."),
    omnichat.WithLogger(slog.Default()), // Enable logging
)
```

---

## Security

### Validate Webhook Signatures

omni-twilio validates webhook signatures automatically when the Auth Token is configured. This ensures requests come from Twilio.

### Use HTTPS

Twilio requires HTTPS for all webhook URLs. ngrok provides this automatically for local development.

### Protect Auth Token

Never commit your Auth Token to version control:

```bash
# .gitignore
.env
.envrc
```

---

## Environment Variables Reference

| Variable | Description | Required |
|----------|-------------|----------|
| `TWILIO_ACCOUNT_SID` | Account SID (starts with AC) | Yes |
| `TWILIO_AUTH_TOKEN` | Auth Token | Yes |
| `TWILIO_PHONE_NUMBER` | Phone number (E.164 format) | For outbound |
| `TWILIO_MESSAGING_SERVICE_SID` | Messaging Service SID for RCS | For RCS |

---

## Next Steps

- [Voice Gateway](../getting-started/overview.md) - Full-duplex voice calls
- [OmniChat Provider](../getting-started/overview.md#sms-omnichat) - SMS/MMS integration
- [RCS Implementation](../specs/features/rcs/PLAN.md) - Rich messaging features
