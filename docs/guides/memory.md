# Twilio Memory API

The `omnimemory` package provides an [OmniMemory](https://github.com/plexusone/omnimemory) provider backed by the Twilio Memory API.

## Overview

Twilio Memory API provides semantic memory storage with automatic embedding generation and recall capabilities. This provider maps OmniMemory concepts to Twilio's Memory API:

| OmniMemory | Twilio Memory API |
|------------|-------------------|
| TenantID | Store ID |
| SubjectID | Profile ID |
| Memory | Observation |
| Search/Recall | Recall API |

## Installation

```bash
go get github.com/plexusone/omni-twilio
```

## Configuration

### Environment Variables

```bash
# Required
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-auth-token"

# Optional: for conformance tests
export TWILIO_MEMORY_STORE_ID="store-id"
export TWILIO_MEMORY_PROFILE_ID="profile-id"
```

### Provider Options

The provider accepts the following options:

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `account_sid` | `TWILIO_ACCOUNT_SID` | Twilio Account SID (required) |
| `auth_token` | `TWILIO_AUTH_TOKEN` | Twilio Auth Token (required) |

## Usage

### Basic Setup

```go
import (
    "github.com/plexusone/omnimemory"
    "github.com/plexusone/omnimemory/core"
    _ "github.com/plexusone/omni-twilio/omnimemory" // Register provider
)

// Create client with Twilio provider
client, err := omnimemory.NewClient(core.ClientConfig{
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
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Adding Memories

```go
// Add an observation
memory, err := client.Add(ctx, &core.AddRequest{
    Context: core.Context{
        TenantID:  "your-store-id",   // Twilio Store ID
        SubjectID: "your-profile-id", // Twilio Profile ID
    },
    Type:    core.MemoryTypeObservation,
    Content: "User mentioned they prefer dark mode interfaces",
})
```

### Searching Memories

```go
// Semantic search
results, err := client.Search(ctx, &core.SearchRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    Query:     "interface preferences",
    Limit:     10,
    Threshold: 0.7, // Similarity threshold
})

for _, result := range results.Results {
    fmt.Printf("Score: %.2f, Content: %s\n", result.Score, result.Memory.Content)
}
```

### Recalling Memories

```go
// Recall relevant memories with optional summary
recalled, err := client.Recall(ctx, &core.RecallRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    Query:      "What are the user's preferences?",
    MaxResults: 5,
})

fmt.Printf("Summary: %s\n", recalled.Summary)
for _, mem := range recalled.Memories {
    fmt.Printf("- %s\n", mem.Content)
}
```

### Listing Memories

```go
// List all memories for a profile
list, err := client.List(ctx, &core.ListRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    Limit: 100,
})

fmt.Printf("Total: %d, HasMore: %v\n", list.TotalCount, list.HasMore)
```

### Updating Memories

```go
// Update memory content
updated, err := client.Update(ctx, &core.UpdateRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    ID:      memory.ID,
    Content: "User strongly prefers dark mode interfaces",
})
```

### Deleting Memories

```go
// Delete a memory
err := client.Delete(ctx, &core.DeleteRequest{
    Context: core.Context{
        TenantID:  "your-store-id",
        SubjectID: "your-profile-id",
    },
    ID: memory.ID,
})
```

## Direct Provider Usage

You can also use the provider directly without the client:

```go
import "github.com/plexusone/omni-twilio/omnimemory"

provider, err := omnimemory.NewProvider(core.ProviderConfig{
    Options: map[string]any{
        "account_sid": "ACxxxxxxxx",
        "auth_token":  "your-token",
    },
}, nil)
if err != nil {
    log.Fatal(err)
}
defer provider.Close()

// Use provider directly
memory, err := provider.Add(ctx, &core.AddRequest{...})
```

## Twilio Console Setup

1. **Create a Memory Store**:
   - Go to Twilio Console → Memory → Stores
   - Create a new store and note the Store ID

2. **Create a Profile**:
   - Within your store, create a profile
   - Note the Profile ID

3. **Set Credentials**:
   - Use your Account SID and Auth Token from Twilio Console

## Testing

Run conformance tests with credentials:

```bash
export TWILIO_ACCOUNT_SID="ACxxxxxxxx"
export TWILIO_AUTH_TOKEN="your-token"
export TWILIO_MEMORY_STORE_ID="your-store-id"
export TWILIO_MEMORY_PROFILE_ID="your-profile-id"

go test -v ./omnimemory/...
```

## Memory Types

The provider supports all OmniMemory types:

| Type | Description |
|------|-------------|
| `observation` | Observed behaviors or interactions |
| `fact` | Verified pieces of information |
| `preference` | User preferences |
| `summary` | Summarized information |
| `trait` | Personality traits |
| `relationship` | Relationships between entities |

## Error Handling

```go
memory, err := client.Add(ctx, req)
if err != nil {
    var validationErr *core.ValidationError
    if errors.As(err, &validationErr) {
        log.Printf("Validation error on field %s: %s", validationErr.Field, validationErr.Message)
        return
    }

    var providerErr *core.ProviderError
    if errors.As(err, &providerErr) {
        log.Printf("Provider %s error in %s: %v", providerErr.Provider, providerErr.Operation, providerErr.Err)
        return
    }

    log.Printf("Unknown error: %v", err)
}
```

## Related

- [OmniMemory Documentation](https://github.com/plexusone/omnimemory)
- [Twilio Memory API Documentation](https://www.twilio.com/docs/memory)
