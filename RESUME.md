# Demo App — Re-Entry Document

> Last active: 2026-01-27. Updated: 2026-02-21.
> Use this file to get up to speed before starting a session.
> Archive into DEVLOG.md after first session back.

---

## Where We Left Off

Phases 1–8b are complete. The full pipeline is operational:

- **demo-app** `v0.8.1` — CI/CD via GitHub Actions, multi-arch binary releases, Docker image on `ghcr.io/billgrant/demo-app`
- **terraform-provider-demoapp** `v0.1.0` — published on registry.terraform.io, GPG signed
- **demo-app-examples/baseline** — works with `terraform init` + `terraform apply`, no local prerequisites

The app itself has: CRUD API, display panel (arbitrary JSON injection), system info panel, Prometheus `/metrics`, structured JSON logging with optional webhook shipping, ENV_FILTER for regex-based env var filtering, and an embedded frontend dashboard.

---

## What's Next

### Phase 9 — Distribution & Documentation
Scope is unchanged and still on target:
- [ ] Terraform module example
- [ ] Kubernetes manifest example
- [ ] Helm chart
- [ ] Mermaid architecture diagram (wait until feature set is final)

### Phase 10 — Demo Library (revised direction)
The examples repo was always meant to be a library. The order of stocking the shelves has changed.

**Round 1 — HashiCorp/IBM stack first**

Bill is a Staff Solutions Engineer at HashiCorp (4 years, Strategics + Global Finserv focus). With IBM's acquisition of HashiCorp, there's increasing R&D scope and more to learn. The goal is to build demos that solve real problems Bill's customers actually have — using tools he knows deeply. These demos become reusable customer-facing artifacts.

The key design principle carries forward: **demo-app stays vendor-neutral and dumb**. External tooling (Vault Agent, Ansible, Terraform, etc.) does the heavy lifting. The demo IS the integration layer, not the app. This is a more realistic and compelling customer story — "here's how you add Vault to an existing app without touching its code."

**Round 2 — Broaden the library**

Once the HashiCorp/IBM examples are solid, expand into observability, CI/CD, multi-tier, AI/agentic demos, etc.

> Note: The rounds above are directional, not prescriptive. Specific demos will be planned at the time based on what's most relevant. Don't over-plan before knowing which customer problems to solve.

**MCP Server** — still planned, still high value. An AI agent (Claude) managing demo-app state is a compelling demo for any audience. Likely fits in Round 2 or as a standalone phase.

---

## Suggested First Session Goals

1. Read PLAN.md and this file
2. Decide Phase 9 starting point
3. Pick one concrete thing to ship

---

## Quick Reference

```bash
# Run dev server
cd ~/code/demo-app
go run .

# Kill dev server (Go compiles to 'main', not 'demo-app')
pkill main

# Run tests
go test ./...

# Tag a release (triggers CI/CD)
git tag v0.9.0
git push --tags
```

| Repo | Purpose |
|------|---------|
| `~/code/demo-app` | Main app |
| `~/code/terraform-provider-demoapp` | Terraform provider |
| `~/code/demo-app-examples` | Demo configs |
| `~/code/billgrant.github.io` | Blog (Jekyll) |

---

*Archive into DEVLOG.md after first session back.*
