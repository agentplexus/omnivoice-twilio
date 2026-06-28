// Package omnimemory provides a Twilio Memory provider for github.com/plexusone/omnimemory.
//
// This is a "thick" provider that wraps the Twilio Memory API using the official
// Twilio Go SDK. It enables omnimemory applications to use Twilio Memory as a
// backend for storing and retrieving conversational memories.
//
// # Installation
//
// Import this package alongside omnimemory:
//
//	import (
//	    "github.com/plexusone/omnimemory"
//	    _ "github.com/plexusone/omni-twilio/omnimemory" // Register Twilio provider
//	)
//
// # Configuration
//
// The provider can be configured via options or environment variables:
//
//	client, err := omnimemory.NewClient(omnimemory.ClientConfig{
//	    Providers: []omnimemory.ProviderConfig{
//	        {
//	            Name: omnimemory.ProviderNameTwilio,
//	            Options: map[string]any{
//	                "account_sid": "ACxxx",
//	                "auth_token":  "xxx",
//	            },
//	        },
//	    },
//	})
//
// Or via environment variables:
//
//	export TWILIO_ACCOUNT_SID=ACxxx
//	export TWILIO_AUTH_TOKEN=xxx
//
// # Concept Mapping
//
// Omnimemory concepts map to Twilio Memory API as follows:
//
//   - TenantID → Twilio Store ID (you must create stores beforehand)
//   - SubjectID → Twilio Profile ID (created via Twilio API or auto-resolved)
//   - Memory → Twilio Observation
//   - Memory.AgentID → Observation.Source
//   - Search/Recall → Twilio Recall API with semantic search
//
// # Prerequisites
//
// Before using this provider, you need to:
//
//  1. Create a Twilio account and obtain API credentials
//  2. Create a Memory Store via the Twilio Console or API
//  3. Create or obtain Profile IDs for your users
//
// # Features
//
//   - Add: Creates observations in Twilio Memory
//   - Get: Retrieves a specific observation by ID
//   - List: Lists observations for a profile
//   - Search: Semantic search using Twilio Recall API
//   - Recall: Retrieves relevant memories with optional summaries
//   - Delete: Removes observations
//
// # Limitations
//
//   - Update is implemented as delete + create (Twilio doesn't support direct updates)
//   - Add doesn't return the observation ID (Twilio limitation)
//   - Memory types and scopes are not natively supported by Twilio (stored in metadata)
//   - Embeddings are handled by Twilio internally, not via the Embedder interface
package omnimemory
