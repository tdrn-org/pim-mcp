# PIM-MCP Architecture

**Date**: 2026-06-20
**Status**: Living document — updated as architecture evolves

## Vision

PIM-MCP is an MCP (Model Context Protocol) server that gives AI agents secure, read-and-write access to personal information management services — email, calendar, tasks, and contacts. It abstracts provider-specific APIs behind clean Go interfaces, with Microsoft Graph as the primary backend.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Agent (Judy)                                           │
│  MCP Tools: searchEmails, getEmail, updateEmail,        │
│             searchEvents, getEvent, createEvent,        │
│             searchTasks, getTask, createTask, updateTask│
│             searchContacts, getContact                  │
└──────────────────┬──────────────────────────────────────┘
                   │ X-API-Key header
┌──────────────────▼──────────────────────────────────────┐
│  MCP Middleware (internal/adapters/middleware/mcp/)      │
│  • Validates API key → injects Session into context     │
│  • AccessMode gate: ReadWrite → registers write tools   │
│  • Notifications handled via shadow-mode cron jobs      │
└──────────────────┬──────────────────────────────────────┘
                   │ domain.EmailProvider, etc.
┌──────────────────▼──────────────────────────────────────┐
│  PIM Adapter (internal/adapters/pim/)                   │
│  • pim.Provider interface                               │
│  • Credential management (CheckCredentials, Refresh)    │
└──────────────────┬──────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────┐
│  MS Graph Adapter (internal/adapters/pim/msgraph/)      │
│  • OAuth2 with OBO (On-Behalf-Of) flow                  │
│  • Two-token architecture (see ADR-001)                 │
│  • Graph client per MCP call (no shared state)          │
└──────────────────┬──────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────┐
│  REST API (internal/adapters/middleware/rest/)          │
│  /api/v1/session, /api/v1/login, /api/v1/ping          │
│  • Browser-based OAuth2 UI flow                         │
│  • Cookie + API-Key dual auth                           │
└──────────────────┬──────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────┐
│  Session Store (internal/session/)                      │
│  • SQLite via go-database                              │
│  • Session{ID, APIKey, Credentials, LastUpdate}         │
│  • Credentials = serialized oauth2.Token (JSON)        │
│  • db-tag-based ScanStruct for clean DB mapping         │
└─────────────────────────────────────────────────────────┘
```

## ADR-001: Two-Token Architecture

**Date**: 2026-06-18
**Status**: Accepted

### Problem

OAuth2 login produces TWO token pairs, not one:

1. **User-Token** (Access + Refresh) — for the app itself. NOT usable for Graph.
2. **Graph-Token** (OBO credential) — for Microsoft Graph API. Self-renewing, opaque.

The User-Token is the "key" to create the OBO credential via `OnBehalfOfCredentialWithSecret`. The OBO credential manages its own Graph tokens transparently.

### Decision

| What | Where | Why |
|------|-------|-----|
| User-Token (full `oauth2.Token`) | SQLite Session.credentials | Key for OBO creation; includes RefreshToken for renewal |
| Graph-Token (OBO credential) | Memory cache (1h TTL) | Not serializable, self-renewing via Azure SDK |
| Cache key | `token.AccessToken` | Links Session → OBO credential |

### Consequences

- Server restart: OBO cache is empty, but `createOBOCredential` rebuilds from stored User-Token
- User-Token expiry (~1h): `RefreshCredentials` (5-min ticker) proactively refreshes before expiry
- If all tokens expire: `domain.ErrAuthenticationRequired` → re-connect flow via UI
- `AADSTS90009` prevented: never try to refresh User-Token for Graph — use OBO credential instead

## ADR-002: Write-Tool Interface Split

**Date**: 2026-06-19
**Status**: Accepted

### Problem

Write operations need to be gated by a single configuration switch without blowing up provider interfaces.

### Decision

Split each domain provider into read + write interfaces:

```go
type EmailProvider interface { SearchEmails, GetEmail }        // all adapters
type EmailWriteProvider interface { EmailProvider; UpdateEmail } // ReadWrite only
```

One config switch: `access_mode = "read_only" | "read_write"`

Write tools are only registered when `caps.AccessMode == ReadWrite` AND the provider implements the write interface.

### Safety Philosophy

| Domain | Write-Ops | Boundaries |
|--------|-----------|------------|
| Email | UpdateEmail | Only `IsRead: true` — no body/sender manipulation |
| Calendar | CreateEvent | No attendees, no recurrence |
| Tasks | CreateTask, UpdateTask | No Delete |

## ADR-003: Re-Connect Flow

**Date**: 2026-06-20
**Status**: Accepted

### Problem

After token expiry or server restart, the agent receives `ErrAuthenticationRequired`. Previously the only fix was "Disconnect + new API key" — which broke the agent configuration.

### Decision

Two entry points for re-authentication:

1. **Cookie-based** (session page): `POST /login {reconnect: true}` — browser cookie identifies session, OAuth2 redirect, API key unchanged.
2. **API-key-based** (landing page): `POST /login {api_key, reconnect: true}` — for cold recovery.

### Consequences

- API key survives credential expiry — no agent reconfiguration needed
- UI shows "Re-connect Required" with amber lock icon when credentials invalid
- Separate from initial "Connect to Provider" (new session, new API key)

## ADR-004: UTC as Default Timezone

**Date**: 2026-06-20
**Status**: Accepted

### Problem

`marshalTZTime` fell back to `time.Location().String()` when timezone was empty, returning `"Local"` which MS Graph rejects. MS Graph requires Windows timezone names.

### Decision

Always default to `"UTC"` in `marshalTZTime` when no explicit timezone is set. All existing events use UTC. RFC3339 offsets are normalized correctly.

No config option — eliminates a failure mode.

## Key Design Patterns

### Builder Pattern for MS Graph Requests

```go
requestBuilder := eventRequestBuilder{Request: models.NewEvent()}
request := requestBuilder.
    Title(&create.Title).
    Start(&create.Start).
    End(&create.End).
    Request
```

Fluent API avoids verbose Set/Get chains. Each method handles nil gracefully.

### Provider per MCP Call

No global access token. Each MCP call:
1. Extracts `X-API-Key` → looks up session → injects into context
2. `credentialFromContext(ctx)` → retrieves OBO credential from cache
3. `graphClient(ctx)` → fresh Graph client for this call

### Access Mode Gate

```go
func addEmailTools(server, caps, provider) {
    addSearchEmailsTool(server, provider)
    addGetEmailTool(server, provider)
    if caps.AccessMode == ReadWrite {
        if wp, ok := provider.(EmailWriteProvider); ok {
            addUpdateEmailTool(server, wp)
        }
    }
}
```

ReadOnly adapter? Type assertion fails → write tools silently not registered.

## Data Flow: MCP Call

```
Agent → POST /mcp {X-API-Key}
  → getServerFromRequest: LookupSessionByAPIKey
  → Session in context
  → ReceivingMiddleware: logs call
  → Tool handler: SessionFromContext(ctx)
      → credentialFromContext: unmarshal oauth2.Token
      → credentialCache.Get(accessToken): createOBOCredential
      → graphClient: OBO → Graph API call
      → Response → structuredContent
```

## Cron Job: Token Refresh

Every 5 minutes (`serverJobTickerSchedule`):

```
runJobs():
  for each session with credentials:
    RefreshCredentials(credentials, 5min)
      → token.Valid()? skip
      → time.Until(expiry) < 10min? proactive refresh
      → persist refreshed token → session.Update()
```

## Error Sentinel: ErrAuthenticationRequired

Defined in `internal/domain/domain.go`. Returned by `credentialFromContext` when:
- No session or credentials empty
- Cannot unmarshal token (old format)
- OBO credential not in cache (expired or never created)

Agent (Judy) detects this and alerts user: "Re-connect required."

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go |
| Database | SQLite via go-database (tdrn-org) |
| HTTP Server | go-httpserver (tdrn-org) |
| MCP Protocol | modelcontextprotocol/go-sdk |
| OAuth2 | golang.org/x/oauth2 + Azure SDK |
| MS Graph | microsoftgraph/msgraph-sdk-go |
| Frontend | SvelteKit 5 (embedded via //go:embed) |
| Caching | go-cache/memory (tdrn-org) |
| Auth | Custom API-Key + OAuth2 dual auth |

## Project Evolution

| Date | Milestone |
|------|-----------|
| 2026-05-14 | Project initiated (client-credentials OAuth2, stdio MCP) |
| 2026-06-15 | Adapter architecture: PIM providers + middleware split |
| 2026-06-16 | API-Key auth + session SQLite store |
| 2026-06-17 | MS Graph cross-folder search + uniqueBody fix |
| 2026-06-18 | OBO two-token architecture + entrypoint separation |
| 2026-06-19 | Domain write interfaces + task Create/Update + deploy |
| 2026-06-20 | Email/Calendar write tools + token persistence + re-connect + folder field |
