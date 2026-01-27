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
| Database | BadgerDB | Embedded K/V store, concurrent writes, in-memory or file |
| Frontend | Embedded SPA | Baked into binary, zero file path management |
| Router | stdlib `net/http` | Learn fundamentals first, add frameworks only if needed |
| Logging | `log/slog` (JSON) | Stdlib structured logging, no external deps |

### Development Requirements

| Tool | Purpose | Installation |
|------|---------|--------------|
| Go 1.23+ | Build demo-app and provider | [golang.org](https://golang.org/dl/) |
| Terraform 1.0+ | Test provider locally | `apt` via [HashiCorp repo](https://www.hashicorp.com/official-packaging-guide) |
| Docker | Container builds | Standard installation |
| GoReleaser | Provider releases | `go install github.com/goreleaser/goreleaser@latest` |

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
│                    BadgerDB                             │
│         (Embedded K/V store, in-memory or file)         │
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
| `DB_PATH` | `:memory:` | BadgerDB path, `:memory:` for ephemeral |
| `SHOW_ENV` | `false` | Expose env vars in system info |
| `ENV_FILTER` | `""` | Regex filter for displayed env vars |
| `LOG_WEBHOOK_URL` | `""` | URL to POST log entries (Phase 7) |
| `LOG_WEBHOOK_TOKEN` | `""` | Auth token for log webhook (Phase 7) |

---

## Milestones

### Phase 1: Foundation ✓
- [x] Go project structure
- [x] Basic HTTP server with health endpoint
- [x] ~~SQLite~~ BadgerDB integration (changed in Phase 6)
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

### Phase 5: Terraform Provider ✓
- [x] Create `terraform-provider-demoapp` repository
- [x] Implement provider using `terraform-plugin-framework`
- [x] `demoapp_item` resource (CRUD maps to REST API)
- [x] `demoapp_display` resource (POST arbitrary JSON)
- [x] Provider documentation (README, docs/ for registry)
- [x] GoReleaser config for automated releases

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

### Phase 6: Demo-for-the-Demo ✓
- [x] **Concurrency fix** — Replaced SQLite with BadgerDB (K/V store with native concurrent write support)
- [x] Create `demo-app-examples/` repo
- [x] Baseline demo: Docker provider + demoapp provider + http provider
- [x] Single `terraform apply` provisions container AND populates data
- [x] Documentation showing the full flow

**Purpose:** Show how to use demo-app in real demos. The "demo of the demo app."

**Database Change:** Originally planned SQLite WAL mode fix, but switched to BadgerDB for better concurrency and consistent behavior across in-memory/file modes.

**Future Improvement:** Once container images are published to ghcr.io (Phase 8), refactor baseline demo to pull from registry instead of requiring local build. Goal: `git clone` → `terraform init` → `terraform apply` with no prerequisites.

**Demo Architecture:**
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Docker         │     │  HTTP           │     │  DemoApp        │
│  Provider       │────►│  Provider       │────►│  Provider       │
│                 │     │  (data source)  │     │                 │
│  Creates        │     │  Fetches from   │     │  Posts to       │
│  container      │     │  /api/system    │     │  /api/display   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**Display Panel Strategy:**
- **Now:** HTTP provider fetches `/api/system`, posts to display (shows data flow, slight duplication with System Info panel is intentional — raw JSON vs formatted)
- **Future:** When `/metrics` endpoint exists (Phase 7), fetch that instead — dynamic metrics snapshot at apply time, no dedicated panel needed

### Phase 7: Observability & Polish ✓
- [x] Prometheus `/metrics` endpoint (using `prometheus/client_golang`)
  - App metrics: `demoapp_http_requests_total`, `demoapp_http_request_duration_seconds`, `demoapp_items_total`, `demoapp_display_updates_total`, `demoapp_info`
  - Go runtime metrics: goroutines, memory, GC (included by default)
  - Process metrics: CPU, file descriptors (included by default)
- [x] **Code refactoring** — split `main.go` into multiple files before adding more features
  - `main.go` — startup, routing, configuration
  - `handlers.go` — HTTP handlers (items, display, system, health)
  - `store.go` — BadgerDB operations
  - `middleware.go` — logging middleware with metrics instrumentation
  - `metrics.go` — Prometheus metric definitions and registration
- [x] Log webhook shipping — optional `LOG_WEBHOOK_URL` + `LOG_WEBHOOK_TOKEN` for pushing logs to any HTTP endpoint (Splunk HEC, Loki, etc.)
- [x] Request header display — show incoming headers in `/api/system` response
- [x] Environment variable filtering — case-insensitive regex via `ENV_FILTER` env var
- [x] Configuration documentation — `docs/CONFIGURATION.md` with full env var details and examples

**Design Decisions:**
- **Prometheus format** chosen over OpenTelemetry for simplicity and wide compatibility. Most observability platforms (Splunk, Datadog, Grafana, etc.) can ingest Prometheus format natively or via collectors.
- **Log webhook** keeps the app vendor-neutral — just HTTP POST with JSON. No Splunk SDK, no Loki SDK. The receiving end handles any format transformation needed.
- **Refactoring approach:** Split by responsibility into separate files within the same package (not separate packages under `internal/`). This keeps imports simple while improving organization. **Important:** Provide detailed explanations during refactor — explain what's moving, why it belongs together, and how Go's package system works with multiple files.

### Phase 8: CI/CD (demo-app) ✓ `v0.8.0`

**Step 1: Tests**
- [x] Unit tests for handlers (`handlers_test.go`)
  - Health endpoint: returns 200, has status field
  - Items CRUD: create, list, get, update, delete
  - Items errors: 404 for non-existent, 400 for invalid ID/JSON
  - Display: GET empty, POST JSON, GET returns it
  - System: returns hostname, ips fields
- [x] Test coverage target: core API paths (~48% statement coverage)

**Step 2: CI Pipeline**
- [x] GitHub Actions workflow for CI (build, test, lint on push/PR)
- [x] Go vet / staticcheck for code quality
- [x] Docker build verification (container starts, /health responds)
- [x] `paths-ignore` for markdown/docs-only changes
- [x] DHI registry authentication via GitHub secrets

**Step 3: Release Automation**
- [x] Release workflow triggered by git tags
- [x] Multi-arch binary builds (linux/mac/windows × amd64/arm64)
- [x] Multi-arch Docker image builds
- [x] Push to GitHub Container Registry (ghcr.io)
- [x] Create GitHub Release with binaries attached

**Why before Distribution:** CI/CD produces the artifacts (binaries, containers). Distribution documents how to consume them. Helm charts reference container images that CI/CD publishes.

### Phase 8b: CI/CD (terraform-provider-demoapp) ✓ `v0.1.0`

Separate track for the Terraform provider repo.

- [x] GPG signing setup for provider binaries (personal email, not corporate)
- [x] GitHub Actions release workflow (triggered by git tags)
- [x] GoReleaser configuration for provider builds (fixed `formats` deprecation)
- [x] Publish to registry.terraform.io (personal account)
- [x] Update demo-app-examples to use published provider (remove dev overrides)
- [x] Switch demo-app-examples to ghcr.io container image (no local build needed)

### Phase 9: Distribution & Documentation
- [ ] GitHub releases with binaries (automated by Phase 8)
- [ ] Terraform module example
- [ ] Kubernetes manifest example
- [ ] Helm chart
- [ ] **Architecture diagram** — visual documentation for demos
  - Mermaid format (renders on GitHub, version controlled)
  - Architecture overview: API → handlers → BadgerDB
  - Request flow: what happens when requests hit endpoints
  - Demo scenarios: Terraform → app → display panel flow
  - Wait until feature set is finalized before creating

### Phase 10: Demo Library

Expanded demos showcasing demo-app with various technologies. Lives in `demo-app-examples/` repo.

**Multi-Tier Architecture:**
- [ ] Two demo-app containers: "backend" (`:8081`) and "frontend" (`:8080`)
- [ ] Custom `DEMO_*` env vars on each container with `ENV_FILTER="^DEMO_"`
- [ ] HTTP provider fetches `/api/system` from backend
- [ ] DemoApp provider posts backend's system info to frontend's `/api/display`
- [ ] Shows: multi-container orchestration, service-to-service data flow, ENV_FILTER in action
- [ ] "Demo for the demo" — demo-app instances demonstrating each other

**Secrets Management:**
- [ ] Vault demo (works with CE and Enterprise)
- [ ] Inject secrets into demo-app, display in panel
- [ ] Show dynamic secrets workflow

**Observability:**
- [ ] Grafana/Loki stack
- [ ] Ship structured logs to dashboards
- [ ] Visualize demo-app traffic patterns

**CI/CD:**
- [ ] GitHub Actions pipeline demo
- [ ] Pipeline deploys demo-app, posts build status to display panel
- [ ] Show full GitOps workflow

**AI/Agentic:**
- [ ] MCP server for demo-app (connects to "MCP Server" in Future Considerations)
- [ ] Claude Desktop as demo UI ("customer" perspective)
- [ ] Show AI agent managing application state

**Purpose:** Build a library of examples that:
1. Show what's possible with demo-app
2. Provide templates others can adapt
3. Help us learn what's missing in demo-app itself

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

**Status:** Deferred from Phase 7. New idea: could the display panel auto-parse incoming JSON and extract key values to highlight? Needs more thought before implementing a separate endpoint.

### External IP Detection

Show the app's public IP address in the system info panel. Useful for demos in cloud environments (AWS, GCP, Azure) to prove deployment location.

**Considerations:**
- Requires outbound call to external service (ifconfig.me, icanhazip.com, etc.)
- Less useful for local Docker containers
- Should be opt-in via environment variable
- Need to handle timeout/failure gracefully

**Status:** Deferred from Phase 7. Needs more thought on the exact use case — showing client IP (from request headers) may be more valuable than showing the app's own public IP.

### UI Polish

Make the frontend more visually appealing — colors, typography, layout, animations.

**Status:** Deferred until feature set is finalized. No point polishing UI that might change.

### Other Future Ideas
- WebSocket endpoint for real-time demo scenarios
- ~~Prometheus metrics endpoint (`/metrics`)~~ — **Promoted to Phase 7**
- Configurable response delays (for latency/timeout demos)

---

## Open Questions

- [x] Frontend framework choice — **Vanilla JS** (decided: learn fundamentals, no build step, same philosophy as backend)
- [x] Router choice — stdlib `net/http` (decided: learn fundamentals first)
- [x] ~~SQLite driver~~ — Switched to **BadgerDB** in Phase 6 for concurrent write support
- [x] ~~External IP detection~~ — Deferred to Future Considerations; showing client IP via headers may be more useful

---

## Non-Goals (for now)

- Multiple database backends
- Vendor-specific integrations baked into core (Vault SDK, Terraform SDK, etc.)

The app should stay generic. Demo-specific behavior comes from *how you use it*, not built-in features.

---

## Versioning

Follows semantic versioning (`MAJOR.MINOR.PATCH`) with phase-based milestones:

| Segment | Meaning | Example |
|---------|---------|---------|
| `MINOR` | Maps to project phase | Phase 8 = `v0.8.0`, Phase 9 = `v0.9.0` |
| `PATCH` | Bug fixes within a phase | `v0.8.1` for a fix after Phase 8 release |
| `MAJOR` | Reserved for production-ready | `v1.0.0` when the project is "complete" |

**Tagging a release:**
```bash
git tag v0.9.0
git push --tags
```

Pushing a `v*` tag triggers the release workflow, which builds multi-arch binaries, pushes a Docker image to `ghcr.io`, and creates a GitHub Release.

---

## References

- [Music Graph](https://github.com/billgrant/music-graph) — Previous project, Flask-based
- [Blog](https://billgrant.io) — Development journey documentation
