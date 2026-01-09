# demo-app

A universal demo application for infrastructure, security, and platform demonstrations.

## Why This Exists

When demoing infrastructure tools (Terraform, Vault, CI/CD pipelines, network appliances), you need something to deploy. Most demo apps are either too simple ("Hello World") or too complex (full production apps). This app sits in the sweet spot:

- **Real enough** — REST API, database, frontend, structured logging
- **Simple enough** — single binary, SQLite, one container
- **Universal** — doesn't assume what you're demoing; accepts injected data
- **Observable** — structured logs, system info, network details for any monitoring stack

## Status

🚧 **Phase 2 In Progress** — Core API endpoints built. See [PLAN.md](PLAN.md) for roadmap.

### What's Working
- HTTP server with structured JSON logging
- SQLite database (in-memory or file-based)
- Docker container with hardened images
- Full CRUD API for items
- Display panel for injected demo data
- System info endpoint

### Quick Start

```bash
# Run locally
go run main.go

# Or with Docker (requires docker login dhi.io)
docker login dhi.io
docker build -t demo-app .
docker run --rm -p 8080:8080 demo-app
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Items (CRUD)
```bash
# List all items
curl http://localhost:8080/api/items

# Create item
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"My Item","description":"Optional description"}'

# Get single item
curl http://localhost:8080/api/items/1

# Update item
curl -X PUT http://localhost:8080/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name","description":"New description"}'

# Delete item
curl -X DELETE http://localhost:8080/api/items/1
```

### Display Panel
Store arbitrary JSON for display in demos (in-memory, not persisted):
```bash
# Get current display data
curl http://localhost:8080/api/display

# Set display data (any valid JSON)
curl -X POST http://localhost:8080/api/display \
  -H "Content-Type: application/json" \
  -d '{"terraform_output":{"region":"us-east-1"},"status":"deployed"}'
```

### System Info
Returns hostname, IP addresses, and selected environment variables:
```bash
curl http://localhost:8080/api/system
```
Useful for demos showing load balancing, container orchestration, or multi-node deployments.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Listen port |
| `DB_PATH` | `:memory:` | SQLite path (`:memory:` or file path) |

## Links

- [Project Plan](PLAN.md) — Architecture, milestones, decisions
- [Development Log](DEVLOG.md) — Session notes and learnings
- [AI Coding Guidelines](AGENTS.md) — Instructions for AI-assisted development
- [Blog](https://billgrant.io) — Development journey posts

## License

MIT
