# demo-app

A universal demo application for infrastructure, security, and platform demonstrations.

## Why This Exists

When demoing infrastructure tools (Terraform, Vault, CI/CD pipelines, network appliances), you need something to deploy. Most demo apps are either too simple ("Hello World") or too complex (full production apps). This app sits in the sweet spot:

- **Real enough** — REST API, database, frontend, structured logging
- **Simple enough** — single binary, SQLite, one container
- **Universal** — doesn't assume what you're demoing; accepts injected data
- **Observable** — structured logs, system info, network details for any monitoring stack

## Status

🚧 **Planning Phase** — See [PLAN.md](PLAN.md) for architecture and roadmap.

## Links

- [Project Plan](PLAN.md) — Architecture, milestones, decisions
- [AI Coding Guidelines](AGENTS.md) — Instructions for AI-assisted development
- [Blog](https://billgrant.io) — Development journey posts

## License

MIT
