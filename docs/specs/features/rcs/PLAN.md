# RCS Support Implementation Plan

Add Rich Communication Services (RCS) support to omni-twilio.

## Overview

Twilio's RCS API uses the same Message resource as SMS/MMS, making integration straightforward. RCS provides rich messaging features including branded senders, rich cards, carousels, and suggested actions with automatic fallback to SMS/MMS.

## References

- [Twilio RCS Documentation](https://www.twilio.com/docs/rcs)
- [Send and receive RCS messages](https://www.twilio.com/docs/rcs/send-an-rcs-message)
- [RCS Onboarding Guide](https://www.twilio.com/docs/rcs/onboarding)
- [Content API](https://www.twilio.com/docs/content)

---

## Phase 1: Basic RCS Parameters

**Status**: Complete

Add RCS-specific parameters to existing SMS/MMS infrastructure.

### Changes

#### client/client.go

```go
// SendSMSParams are parameters for sending an SMS, MMS, or RCS message.
type SendSMSParams struct {
    To                string
    From              string   // Phone number (SMS/MMS) or omit for RCS
    Body              string
    MediaURLs         []string // MMS media attachments

    // RCS-specific fields
    MessagingServiceSid string   // Required for RCS senders
    ContentSid          string   // Template ID for rich content (starts with "HX")
    ContentVariables    string   // JSON object for template variables
}
```

Update `SendSMS` method:

```go
func (c *Client) SendSMS(ctx context.Context, params *SendSMSParams) (*Message, error) {
    createParams := &openapi.CreateMessageParams{}
    createParams.SetTo(params.To)

    // Use MessagingServiceSid for RCS, From for SMS/MMS
    if params.MessagingServiceSid != "" {
        createParams.SetMessagingServiceSid(params.MessagingServiceSid)
    } else if params.From != "" {
        createParams.SetFrom(params.From)
    }

    if params.Body != "" {
        createParams.SetBody(params.Body)
    }
    if len(params.MediaURLs) > 0 {
        createParams.SetMediaUrl(params.MediaURLs)
    }

    // RCS rich content
    if params.ContentSid != "" {
        createParams.SetContentSid(params.ContentSid)
    }
    if params.ContentVariables != "" {
        createParams.SetContentVariables(params.ContentVariables)
    }

    // ... rest of method
}
```

#### omnichat/provider.go

Add RCS configuration option:

```go
type options struct {
    accountSID          string
    authToken           string
    phoneNumber         string
    messagingServiceSid string  // NEW: For RCS senders
    logger              *slog.Logger
}

// WithMessagingServiceSid sets the Messaging Service SID for RCS.
func WithMessagingServiceSid(sid string) Option {
    return func(o *options) {
        o.messagingServiceSid = sid
    }
}
```

Update `Send` to use MessagingServiceSid when available:

```go
func (p *Provider) Send(ctx context.Context, chatID string, msg provider.OutgoingMessage) error {
    // ... existing validation ...

    params := &client.SendSMSParams{
        To:   chatID,
        Body: msg.Content,
    }

    // Prefer MessagingServiceSid for RCS, fall back to From
    if p.messagingServiceSid != "" {
        params.MessagingServiceSid = p.messagingServiceSid
    } else {
        params.From = p.defaultFrom
    }

    // Extract media URLs
    for _, m := range msg.Media {
        if m.URL != "" {
            params.MediaURLs = append(params.MediaURLs, m.URL)
        }
    }

    // RCS content template (from metadata)
    if contentSid, ok := msg.Metadata["content_sid"].(string); ok {
        params.ContentSid = contentSid
    }
    if contentVars, ok := msg.Metadata["content_variables"].(string); ok {
        params.ContentVariables = contentVars
    }

    // ... rest of method
}
```

### Tests

```go
func TestSendRCS(t *testing.T) {
    // Test with MessagingServiceSid
    params := &client.SendSMSParams{
        To:                  "+15559876543",
        MessagingServiceSid: "MGxxxxxxxx",
        Body:                "Hello via RCS!",
    }
    // Verify MessagingServiceSid is used instead of From
}

func TestSendRCSWithContentTemplate(t *testing.T) {
    params := &client.SendSMSParams{
        To:                  "+15559876543",
        MessagingServiceSid: "MGxxxxxxxx",
        ContentSid:          "HXxxxxxxxx",
        ContentVariables:    `{"1": "John"}`,
    }
    // Verify ContentSid and ContentVariables are set
}
```

### Deliverables

- [x] Add `MessagingServiceSid`, `ContentSid`, `ContentVariables` to `SendSMSParams`
- [x] Update `SendSMS` to use MessagingServiceSid when provided
- [x] Add `WithMessagingServiceSid` option to omnichat provider
- [x] Update `Send` to extract content template from metadata
- [x] Add unit tests
- [x] Update README with RCS examples

---

## Phase 2: Rich Content Types

**Status**: Pending

Add Go types for RCS rich content and helper methods to build ContentVariables.

### New Package: client/rcs

```
client/rcs/
├── doc.go           # Package documentation
├── types.go         # RCS content types
├── card.go          # Rich card builder
├── carousel.go      # Carousel builder
├── actions.go       # Suggested actions/replies
└── builder.go       # ContentVariables builder
```

### Types

```go
// types.go
package rcs

// ContentType identifies the Twilio content template type.
type ContentType string

const (
    ContentTypeText     ContentType = "twilio/text"
    ContentTypeMedia    ContentType = "twilio/media"
    ContentTypeCard     ContentType = "twilio/card"
    ContentTypeCarousel ContentType = "twilio/carousel"
    ContentTypeLocation ContentType = "twilio/location"
)

// Action represents a suggested action or reply.
type Action struct {
    Type    ActionType `json:"type"`
    Title   string     `json:"title"`
    ID      string     `json:"id,omitempty"`      // For postback
    Payload string     `json:"payload,omitempty"` // For postback
    URL     string     `json:"url,omitempty"`     // For URL action
    Phone   string     `json:"phone,omitempty"`   // For dial action
}

type ActionType string

const (
    ActionTypeReply    ActionType = "reply"
    ActionTypeURL      ActionType = "url"
    ActionTypeDial     ActionType = "dial"
    ActionTypeLocation ActionType = "location"
)

// Card represents an RCS rich card.
type Card struct {
    Title       string   `json:"title,omitempty"`
    Subtitle    string   `json:"subtitle,omitempty"`
    MediaURL    string   `json:"media_url,omitempty"`
    Actions     []Action `json:"actions,omitempty"`
}

// Carousel represents an RCS carousel of cards.
type Carousel struct {
    Cards []Card `json:"cards"`
}
```

### Builders

```go
// card.go
package rcs

// CardBuilder builds an RCS rich card.
type CardBuilder struct {
    card Card
}

func NewCard() *CardBuilder {
    return &CardBuilder{}
}

func (b *CardBuilder) Title(title string) *CardBuilder {
    b.card.Title = title
    return b
}

func (b *CardBuilder) Subtitle(subtitle string) *CardBuilder {
    b.card.Subtitle = subtitle
    return b
}

func (b *CardBuilder) Media(url string) *CardBuilder {
    b.card.MediaURL = url
    return b
}

func (b *CardBuilder) AddReply(title, id string) *CardBuilder {
    b.card.Actions = append(b.card.Actions, Action{
        Type:  ActionTypeReply,
        Title: title,
        ID:    id,
    })
    return b
}

func (b *CardBuilder) AddURL(title, url string) *CardBuilder {
    b.card.Actions = append(b.card.Actions, Action{
        Type:  ActionTypeURL,
        Title: title,
        URL:   url,
    })
    return b
}

func (b *CardBuilder) Build() Card {
    return b.card
}
```

### Usage Example

```go
import "github.com/plexusone/omni-twilio/client/rcs"

// Build a rich card
card := rcs.NewCard().
    Title("Order Confirmation").
    Subtitle("Your order #12345 has shipped").
    Media("https://example.com/package.jpg").
    AddReply("Track Order", "track_12345").
    AddURL("View Details", "https://example.com/orders/12345").
    Build()

// Convert to ContentVariables JSON
vars, _ := json.Marshal(map[string]any{
    "1": card.Title,
    "2": card.Subtitle,
    "3": card.MediaURL,
})

client.SendSMS(ctx, &client.SendSMSParams{
    To:                  "+15559876543",
    MessagingServiceSid: "MGxxxxxxxx",
    ContentSid:          "HXcard_template",
    ContentVariables:    string(vars),
})
```

### Deliverables

- [ ] Create `client/rcs/` package
- [ ] Add `types.go` with RCS content types
- [ ] Add `card.go` with CardBuilder
- [ ] Add `carousel.go` with CarouselBuilder
- [ ] Add `actions.go` with action types
- [ ] Add unit tests
- [ ] Add documentation

---

## Phase 3: OmniChat Integration

**Status**: Pending

Integrate RCS rich content with the OmniChat provider interface.

### Approach Options

**Option A: Extend provider.Media**

Add RCS-specific fields to existing Media type:

```go
// In omnichat provider/types.go
type Media struct {
    Type     MediaType
    URL      string
    Data     []byte
    MimeType string
    Filename string
    Caption  string

    // RCS-specific (optional)
    Actions []MediaAction `json:",omitempty"` // Suggested actions
}

type MediaAction struct {
    Type    string // "reply", "url", "dial"
    Title   string
    Payload string // ID for reply, URL for url, phone for dial
}
```

**Option B: Use Metadata**

Keep Media unchanged, pass RCS content via Metadata:

```go
router.Send(ctx, "twilio", chatID, provider.OutgoingMessage{
    Content: "Check out our products!",
    Metadata: map[string]any{
        "rcs_card": rcs.NewCard().
            Title("Product Name").
            AddReply("Buy Now", "buy_123").
            Build(),
    },
})
```

**Option C: New RCS-specific types** (Recommended)

Add optional RCS types that providers can support:

```go
// In omnichat provider/types.go
type OutgoingMessage struct {
    Content  string
    Media    []Media
    ReplyTo  string
    Format   MessageFormat
    Metadata map[string]any

    // Rich messaging (RCS, WhatsApp, etc.)
    RichContent *RichContent `json:",omitempty"`
}

type RichContent struct {
    Card     *RichCard     `json:",omitempty"`
    Carousel []RichCard    `json:",omitempty"`
    Actions  []RichAction  `json:",omitempty"` // Suggested replies/actions
}

type RichCard struct {
    Title    string
    Subtitle string
    MediaURL string
    Actions  []RichAction
}

type RichAction struct {
    Type    RichActionType // Reply, URL, Dial, Location
    Title   string
    Payload string
}
```

This approach:
- Keeps backward compatibility
- Can be used by multiple providers (RCS, WhatsApp, Telegram)
- Clear separation of concerns

### Webhook Handling

Update webhook handler to parse incoming RCS messages:

```go
func (p *Provider) handleWebhook(w http.ResponseWriter, r *http.Request) {
    // ... existing parsing ...

    // Check for RCS-specific fields
    if r.Form.Get("ChannelPrefix") == "rcs" {
        // Parse RCS-specific content
        // Handle suggested reply callbacks
        // Extract rich content from incoming messages
    }

    // ... rest of handler
}
```

### Deliverables

- [ ] Decide on approach (recommend Option C)
- [ ] Add RichContent types to omnichat if Option C
- [ ] Update Twilio provider to convert RichContent to RCS
- [ ] Update webhook handler for incoming RCS
- [ ] Handle suggested reply callbacks
- [ ] Add integration tests
- [ ] Update documentation

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TWILIO_MESSAGING_SERVICE_SID` | Messaging Service SID with RCS sender |

---

## Testing

### Unit Tests

```bash
go test ./client/... ./client/rcs/... ./omnichat/...
```

### Integration Tests (requires Twilio RCS sender)

```bash
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_AUTH_TOKEN="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
go test -v -tags=integration ./...
```

---

## Timeline

| Phase | Scope | Status |
|-------|-------|--------|
| Phase 1 | Basic RCS parameters | Complete |
| Phase 2 | Rich content types | Pending |
| Phase 3 | OmniChat integration | Pending |

---

## Notes

1. **RCS Sender Approval**: RCS senders require carrier approval before production use
2. **Fallback**: RCS automatically falls back to SMS/MMS if recipient doesn't support RCS
3. **Content Templates**: Rich content requires pre-created templates in Twilio Console
4. **Regional Availability**: RCS available in 22+ countries as of 2025
