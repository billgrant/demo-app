# demo-app

A universal demo application for infrastructure, security, and platform demonstrations.

## Why This Exists

When demoing infrastructure tools (Terraform, Vault, CI/CD pipelines, network appliances), you need something to deploy. Most demo apps are either too simple ("Hello World") or too complex (full production apps). This app sits in the sweet spot:

- **Real enough** — REST API, database, frontend, structured logging
- **Simple enough** — single binary, SQLite, one container
- **Universal** — doesn't assume what you're demoing; accepts injected data
- **Observable** — structured logs, system info, network details for any monitoring stack

## Status

✅ **Phase 1 Complete** — Foundation built. See [PLAN.md](PLAN.md) for roadmap.

### What's Working
- HTTP server with `/health` endpoint
- Structured JSON logging
- SQLite database (in-memory or file-based)
- Docker container with hardened images

### Quick Start

```bash
# With Docker (requires docker login dhi.io)
docker login dhi.io
docker build -t demo-app .
docker run --rm -p 8080:8080 demo-app

# Or run locally
go run main.go
```

```bash
# Test it
curl http://localhost:8080/health
```

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
