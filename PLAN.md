# Demo App — Project Plan

> A universal demo application for infrastructure, security, and platform demonstrations.

## Problem Statement

As a Solutions Engineer, I frequently need to deploy an application to demonstrate infrastructure tooling. The app itself isn't the point — it's a vehicle to show that Terraform provisioned something, Vault injected secrets, the CI/CD pipeline works, or traffic is flowing through a network appliance.

Existing options are either:
- **Too simple** — nginx default page, "Hello World" doesn't impress anyone
- **Too complex** — real apps require databases, dependencies, configuration
- **Too specific** — built for one vendor's demo, not reusable

## Solution

A self-contained application that:
1. Looks and feels like a real multi-tier app
2. Deploys as a single artifact (binary + embedded assets)
3. Accepts arbitrary data injection for demo-specific content
4. Exposes system/network info proving deployment worked
5. Produces structured logs for observability demos

---

## Architecture

### Tech Stack

| Component | Choice | Rationale |
|-----------|--------|----------|
| Language | Go | Single binary, fast startup, HashiCorp ecosystem alignment |
| Database | SQLite | Embedded, no external dependencies, file or in-memory |
| Frontend | Embedded SPA | Baked into binary, zero file path management |
| Router | TBD (stdlib or Gin) | Keep it simple |
| Logging | Structured JSON | Ships to any observability stack |

### Core Components

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend                             │
│         (Embedded static files, simple SPA)             │
├─────────────────────────────────────────────────────────┤
│                    REST API                             │
│  /api/items     - CRUD operations (generic data)        │
│  /api/display   - Injected demo content                 │
│  /api/system    - System/network info                   │
│  /api/health    - Health check endpoint                 │
├─────────────────────────────────────────────────────────┤
│                    SQLite                               │
│         (Embedded, file or in-memory)                   │
└─────────────────────────────────────────────────────────┘
```

---

## Features

### 1. Display Panel (Injected Content)

Accepts arbitrary JSON via `POST /api/display`, stores it, frontend renders it.

**Use cases:**
- Terraform outputs: `terraform output -json | curl -X POST -d @- http://app/api/display`
- Vault audit events: script polls audit log, posts to display
- CI/CD status: pipeline posts build info
- Custom demo data: whatever the demo needs

The app doesn't know or care what the data means — it just displays it.

### 2. System Info Panel

Always-visible panel showing:
- Hostname / Container ID
- Internal & external IP addresses
- Network interfaces
- Environment variables (filtered, configurable)
- Deployment timestamp
- App version / git commit
- Request headers (optional, shows what's hitting the app)

**Use cases:**
- Prove Terraform deployed to the right region/zone
- Show container orchestration info
- Validate network routing through appliances

### 3. Generic CRUD API

Simple items/inventory/tasks endpoint:
- `GET /api/items` — list all
- `POST /api/items` — create
- `GET /api/items/:id` — read one
- `PUT /api/items/:id` — update
- `DELETE /api/items/:id` — delete

**Use cases:**
- Prove the app is functional (not just a static page)
- Generate realistic log traffic
- Demo database connectivity

### 4. Structured Logging

Every request logged as JSON:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "method": "POST",
  "path": "/api/items",
  "status": 201,
  "latency_ms": 12,
  "client_ip": "10.0.1.50",
  "user_agent": "curl/7.68.0"
}
```

**Use cases:**
- Ship to Splunk, Datadog, ELK, Loki — any observability stack
- Demo SIEM ingestion
- Show traffic patterns through network appliances

### 5. Configuration via Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Listen port |
| `LOG_LEVEL` | `info` | debug, info, warn, error |
| `LOG_FORMAT` | `json` | json or text |
| `DB_PATH` | `:memory:` | SQLite path, `:memory:` for ephemeral |
| `SHOW_ENV` | `false` | Expose env vars in system info |
| `ENV_FILTER` | `""` | Regex filter for displayed env vars |

---

## Milestones

### Phase 1: Foundation
- [ ] Go project structure
- [ ] Basic HTTP server with health endpoint
- [ ] SQLite integration
- [ ] Structured logging
- [ ] Dockerfile

### Phase 2: Core API
- [ ] Generic CRUD endpoints (`/api/items`)
- [ ] Display panel endpoints (`/api/display`)
- [ ] System info endpoint (`/api/system`)

### Phase 3: Frontend
- [ ] Simple SPA (React, Vue, or vanilla JS — TBD)
- [ ] Embed static files in binary
- [ ] Display panel rendering
- [ ] System info panel
- [ ] Items list/form

### Phase 4: Polish
- [ ] Network interface detection
- [ ] External IP detection
- [ ] Request header display
- [ ] Environment variable filtering
- [ ] Configuration documentation

### Phase 5: Distribution
- [ ] Multi-arch Docker builds
- [ ] GitHub releases with binaries
- [ ] Terraform module example
- [ ] Kubernetes manifest example

---

## Future Considerations

Features that aren't part of the initial build but fit the "universal demo app" vision. These come *after* the app is in a "ready to demo" state (Phase 5 complete).

### MCP Server for AI Agent Demos

Build an MCP (Model Context Protocol) server that connects AI assistants (Claude, etc.) directly to demo-app. This enables demos of AI agents interacting with real applications — far more compelling than the typical weather API example.

**Potential MCP tools:**
- `list_items` / `create_item` / `delete_item` — CRUD operations
- `get_system_info` — retrieve deployment info
- `post_to_display` — inject content into the display panel
- `get_recent_logs` — fetch recent request logs

**Use cases:**
- "Watch Claude manage inventory in a real app"
- "Show an AI agent reading system state and making decisions"
- Demo agentic workflows with persistent state
- MCP server development patterns and best practices

**Design principle:** The MCP server is a *separate* component that talks to demo-app's API. The app itself doesn't know or care that an AI is calling it — same universal design as everything else.

### Authentication as a Service Provider

The app could optionally act as a SAML SP or OIDC client, enabling demos of identity providers:
- Okta, Auth0, Azure AD/Entra ID
- HashiCorp Vault OIDC auth
- Any SAML 2.0 or OIDC-compliant IdP

**Design principle:** Auth remains optional. The app works without it, but can be configured to require authentication when demoing identity products. Disabled by default, enabled via environment variables pointing to IdP metadata/endpoints.

### Other Future Ideas
- WebSocket endpoint for real-time demo scenarios
- Prometheus metrics endpoint (`/metrics`)
- Configurable response delays (for latency/timeout demos)

---

## Open Questions

- [ ] Frontend framework choice — React (know it), Vue (simpler), vanilla JS (no build step)?
- [ ] Router choice — stdlib `net/http` or Gin/Echo?
- [ ] SQLite driver — `modernc.org/sqlite` (pure Go) or `mattn/go-sqlite3` (CGO)?
- [ ] How to handle external IP detection reliably across cloud providers?

---

## Non-Goals (for now)

- Multiple database backends
- Vendor-specific integrations baked into core (Vault SDK, Terraform SDK, etc.)

The app should stay generic. Demo-specific behavior comes from *how you use it*, not built-in features.

---

## References

- [Music Graph](https://github.com/billgrant/music-graph) — Previous project, Flask-based
- [Blog](https://billgrant.io) — Development journey documentation
