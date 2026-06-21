# PIM-MCP

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-StreamableHTTP-blueviolet)](https://modelcontextprotocol.io/)

**Personal Information Management MCP Server** — give AI agents secure, read-and-write access to your email, calendar, tasks, and contacts. Built for [Hermes Agent](https://github.com/nousresearch/hermes-agent) and similar agent, currently powered by Microsoft Graph.

---

## What It Does

PIM-MCP is an [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that lets AI agents interact with your personal data through 12 clean, typed tools:

| Domain | Read Tools | Write Tools |
|--------|-----------|-------------|
| 📧 Email | `searchEmails`, `getEmail` | `updateEmail` (mark read) |
| 📅 Calendar | `searchEvents`, `getEvent` | `createEvent` |
| ✅ Tasks | `searchTasks`, `getTask` | `createTask`, `updateTask` |
| 👤 Contacts | `searchContacts`, `getContact` | — |

No raw API calls. No OAuth2 dance. Your agent just calls MCP tools — PIM-MCP handles authentication, token refresh, and provider abstraction.

## Security Considerations
The actual authentication stays with the user. The agent uses the PIM-MCP API Key to access the MCP tools and has no direct access.
Per default the access mode is set to Read-Only. With Read-Write selected controlled write operations are permitted for the agent. 

---

## Architecture

```
Agent (any MCP client)
  │ X-API-Key header
  ▼
MCP Middleware  ─── validates API key, injects session
  │
  ▼
PIM Adapter     ─── domain.EmailProvider, CalendarProvider, etc.
  │
  ▼
MS Graph        ─── OAuth2 OBO flow, two-token architecture
  │
  ▼
REST API        ─── /api/v1/session, /api/v1/login (browser UI)
  │
  ▼
Session Store   ─── SQLite (sessions, API keys, credentials)
```

**Key design principles:**
- **Provider-agnostic** — MS Graph today, IMAP/Google tomorrow. Adapters implement Go interfaces.
- **Two-token architecture** — User-Token (OAuth2) stored in SQLite; Graph-Token (OBO) cached in memory, self-renewing via Azure SDK. See [ARCHITECTURE.md](ARCHITECTURE.md).
- **API-key auth** — Agents authenticate with `X-API-Key` header. No OAuth2 redirects for machines.
- **Write safety** — Single `access_mode` switch (`read_only` / `read_write`). No delete operations. No attendee management. Minimal, auditable surface.

---

## MCP Tools Reference

All tools use StreamableHTTP transport at the `/mcp` endpoint. Authentication via `X-API-Key` header.

### Email

| Tool | Parameters | Returns |
|------|-----------|---------|
| `searchEmails` | `query?`, `unread_only?`, `folder?`, `since?`, `limit?` | `{emails: [...]}` |
| `getEmail` | `id` | Single `Email` |
| `updateEmail` | `id`, `is_read` | Updated `Email` 🔐 |

### Calendar

| Tool | Parameters | Returns |
|------|-----------|---------|
| `searchEvents` | `query?`, `from?`, `to?`, `limit?` | `{events: [...]}` |
| `getEvent` | `id` | Single `Event` |
| `createEvent` | `title`, `start`, `end`, `description?`, `location?`, `is_all_day?` | Created `Event` 🔐 |

### Tasks

| Tool | Parameters | Returns |
|------|-----------|---------|
| `searchTasks` | `query?`, `status?`, `due_after?`, `due_before?`, `limit?` | `{tasks: [...]}` |
| `getTask` | `id` | Single `Task` |
| `createTask` | `title`, `description?`, `status?`, `priority?`, `due_at?` | Created `Task` 🔐 |
| `updateTask` | `id`, `title?`, `description?`, `status?`, `priority?`, `due_at?` | Updated `Task` 🔐 |

### Contacts

| Tool | Parameters | Returns |
|------|-----------|---------|
| `searchContacts` | `query?`, `limit?` | `{contacts: [...]}` |
| `getContact` | `id` | Single `Contact` |

> 🔐 = Requires `access_mode = "read_write"` in config. Read-only mode silently omits these tools.

### Common Types

**TZTime** — Timezone-aware timestamp:
```json
{ "date_time": "2026-07-06T05:30:00Z", "timezone": "UTC" }
```

**NamedEmailAddress** — Email + display name:
```json
{ "email": "holger@carne.de", "name": "Holger de Carne" }
```

**TaskStatus**: `todo` | `in_progress` | `done`
**TaskPriority**: `low` | `medium` | `high`
**EventStatus**: `confirmed` | `tentative` | `canceled`

---

## Configuration

Create a `config.toml`:

```toml
# Server settings
[http]
address = ":9125"
base_url = "https://pim-mcp.example.com"

# Session store
[database]
url = "file:pim-mcp.db"

# Provider: "demo" for testing, "msgraph" for production
[provider]
adapter = "msgraph"
default_time_location = "Europe/Berlin"

[provider.msgraph]
client_id = "your-azure-client-id"
client_secret = "your-azure-client-secret"
tenant_id = "your-azure-tenant-id"

# Controls write tool availability
# "read_only" = search/get only. "read_write" = search/get + create/update.
access_mode = "read_write"
```

The `demo` adapter provides static test data — no Azure setup needed. Perfect for development and CI.

---

## Hermes Agent Integration

PIM-MCP is designed for [Hermes Agent](https://github.com/nousresearch/hermes-agent) integration via native MCP:

```bash
# Register the MCP server
hermes mcp add pim-mcp --url "https://pim-mcp.example.com/mcp"

# Add the API key header to ~/.hermes/config.yaml:
# mcp_servers:
#   pim-mcp:
#     url: https://pim-mcp.example.com/mcp
#     headers:
#       X-API-Key: ${MCP_PIM_MCP_API_KEY}
#     enabled: true
```

After registration, Hermes discovers all 12 tools automatically. Write tools are only available when `access_mode = "read_write"`.

**Session management:**
1. Open `https://pim-mcp.example.com/` in a browser
2. Click "Connect to Provider" → OAuth2 login with Microsoft
3. Copy the `pim_mcp_...` API key
4. Configure the key in Hermes → done

When credentials expire, use "Re-connect" on the session page — your API key stays the same.

---

## Azure App Registration

> 📝 **TODO** — Document the exact Azure Portal steps:
> - Required API permissions (Mail.ReadWrite, Tasks.ReadWrite, Calendars.ReadWrite, Contacts.Read, User.Read, offline_access)
> - Redirect URI format
> - Client secret creation
> - Supported account types (single tenant / multi-tenant)
> - Consent workflow for third-party users

---

## Development

```bash
# Build
make build

# Run with demo adapter (no Azure needed)
build/bin/pim-mcp --config config.toml  # adapter = "demo"

# Run with MS Graph
build/bin/pim-mcp --config config.toml  # adapter = "msgraph"
```

**Requirements:** Go 1.26+, Azure AD app registration (for MS Graph).

See [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions, ADRs, and data flow diagrams.

---

## License

> 📝 **TODO** — License section. Currently Apache 2.0 per source headers.

---

## Status

✅ Production — 12 MCP tools live, token refresh working, re-connect flow stable.

**Known limitations:**
- No auto-reconnect for MCP clients after server restart (Hermes gateway restart required)
