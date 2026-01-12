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
| Router | stdlib `net/http` | Learn fundamentals first, add frameworks only if needed |
| Logging | `log/slog` (JSON) | Stdlib structured logging, no external deps |

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

### Phase 1: Foundation ✓
- [x] Go project structure
- [x] Basic HTTP server with health endpoint
- [x] SQLite integration
- [x] Structured logging
- [x] Dockerfile (Docker Hardened Images)

### Phase 2: Core API ✓
- [x] Generic CRUD endpoints (`/api/items`)
- [x] Display panel endpoints (`/api/display`)
- [x] System info endpoint (`/api/system`)

### Phase 3: Frontend ✓
- [x] Single-page dashboard (vanilla JS, no framework)
- [x] Health panel (status + timestamp, auto-refresh)
- [x] System info panel (hostname, IPs, env vars)
- [x] Items panel (list, create, edit, delete)
- [x] Display panel (pretty-printed JSON, update form)
- [x] Embed static files in binary (`embed` package)

**Frontend Architecture Decision:**
- **Vanilla JS** — Same philosophy as stdlib for backend; learn fundamentals first
- **Single dashboard** — All panels visible at once, no navigation/routing
- **No build step** — Just HTML, CSS, JS files; no npm, no bundler
- **Embedded in binary** — Use Go's `embed` package for single-file deployment

**Dashboard Layout:**
```
┌─────────────────┬─────────────────┐
│  Health         │  System Info    │
│  status: ok     │  hostname: ...  │
│  timestamp: ... │  ips: [...]     │
├─────────────────┴─────────────────┤
│  Items                            │
│  [+ New Item]                     │
│  - Item 1        [Edit] [Delete]  │
│  - Item 2        [Edit] [Delete]  │
├───────────────────────────────────┤
│  Display Panel                    │
│  { "terraform": "output", ... }   │
│  [Update Display Data]            │
└───────────────────────────────────┘
```

### Phase 4: Docker & Multi-Arch ✓
- [x] Update Dockerfile for embedded static files
- [x] Multi-arch builds (amd64 + arm64 for M1 Macs)
- [x] Test containerized deployment

**Note:** Full multi-arch builds will run in CI/CD (Phase 8) using GitHub Actions with native runners. Local buildx with QEMU emulation works but is slow.

### Phase 5: Terraform Provider
- [ ] Create `terraform-provider-demoapp` repository
- [ ] Implement provider using `terraform-plugin-framework`
- [ ] `demoapp_item` resource (CRUD maps to REST API)
- [ ] `demoapp_display` resource (POST arbitrary JSON)
- [ ] `demoapp_highlight` resource (if highlights endpoint exists)
- [ ] Provider documentation

**Provider Concept:**
```hcl
provider "demoapp" {
  endpoint = "http://localhost:8080"
}

resource "demoapp_item" "example" {
  name        = "Provisioned by Terraform"
  description = "Created at ${timestamp()}"
}

resource "demoapp_display" "status" {
  data = jsonencode({
    provisioned_by = "terraform"
    region         = var.region
  })
}
```

### Phase 6: Demo-for-the-Demo
- [ ] Reference Terraform configuration
- [ ] Provisions something simple + demo-app
- [ ] Uses `terraform-provider-demoapp` to populate data
- [ ] Documentation showing the full flow
- [ ] Potentially separate repo: `demo-app-examples/`

**Purpose:** Show how to use demo-app in real demos. The "demo of the demo app."

### Phase 7: Polish
- [ ] External IP detection
- [ ] Request header display
- [ ] Environment variable filtering
- [ ] Highlights endpoint (complement to display)
- [ ] Configuration documentation

### Phase 8: CI/CD
- [ ] GitHub Actions workflow for CI (build, test, lint on push/PR)
- [ ] Go vet / staticcheck for code quality
- [ ] Docker build verification (container starts, /health responds)
- [ ] Release workflow triggered by git tags
- [ ] Multi-arch binary builds (linux/mac/windows × amd64/arm64)
- [ ] Multi-arch Docker image builds
- [ ] Push to GitHub Container Registry (ghcr.io)
- [ ] Create GitHub Release with binaries attached

**Why before Distribution:** CI/CD produces the artifacts (binaries, containers). Distribution documents how to consume them. Helm charts reference container images that CI/CD publishes.

### Phase 9: Distribution
- [ ] GitHub releases with binaries (automated by Phase 8)
- [ ] Terraform module example
- [ ] Kubernetes manifest example
- [ ] Helm chart

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

### Highlights/Key-Values Endpoint

Potential complement to `/api/display` for demos where you want to show:
- **Display panel:** Full raw output (Vault response, Terraform output, etc.)
- **Highlights:** Specific extracted values to call attention to

Example flow:
```bash
# Post full Vault response to display
vault kv get -format=json secret/db | curl -X POST -d @- http://app/api/display

# Post just the secret value to highlights
curl -X POST http://app/api/highlights \
  -d '{"key": "db_password", "value": "s3cr3t", "source": "vault"}'
```

Design considerations:
- In-memory (transient like display) vs persistent (like items)?
- Flexible key-value pairs, user defines both key and value
- Could replace or extend `/api/items` with a `metadata` field
- Or be a separate endpoint entirely (`/api/highlights`, `/api/values`)

**Status:** Idea captured, not yet designed. Revisit after Phase 3 frontend is working.

### Other Future Ideas
- WebSocket endpoint for real-time demo scenarios
- Prometheus metrics endpoint (`/metrics`)
- Configurable response delays (for latency/timeout demos)

---

## Open Questions

- [x] Frontend framework choice — **Vanilla JS** (decided: learn fundamentals, no build step, same philosophy as backend)
- [x] Router choice — stdlib `net/http` (decided: learn fundamentals first)
- [x] SQLite driver — `modernc.org/sqlite` (decided: pure Go for easier cross-compilation)
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
