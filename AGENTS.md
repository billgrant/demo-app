# AI Coding Guidelines — demo-app

> Instructions for Claude and other AI assistants when contributing to this project.

## Project Context

This is a **learning project** where I (Bill) am:
1. Learning Go — I have a Python background, this is my first real Go project
2. Building something useful — a universal demo app for my Solutions Engineering work
3. Documenting publicly — blog posts at billgrant.io about the development journey

## Collaboration Style

**I drive creative and strategic direction; AI helps refine and execute.**

- Explain the "why" behind suggestions, not just the "what"
- Use analogies to Python or networking concepts when introducing Go patterns
- Call out Go idioms that differ from Python (error handling, interfaces, etc.)
- If I'm about to make a mistake, stop me and explain — don't just fix it silently
- When there are multiple valid approaches, present options with tradeoffs

## Learning-First Development

**This is critical: I want to understand the code, not just have working code.**

Think of it like my underlay/overlay analogy — I want to understand the underlay (how Go works, why patterns exist) so I can effectively work with the overlay (this specific application).

### Before Writing Code
Provide a high-level explanation of what we're about to build:
- What problem does this code solve?
- What's the overall structure/approach?
- Are there Go-specific patterns or idioms involved?
- How does this fit into the larger architecture?

### During Implementation
- I may stop and ask "explain this function" or "why are we doing it this way"
- When I ask, pause and explain before continuing
- Don't just write code and move on — make sure I'm following along
- If a concept is new to me (channels, goroutines, interfaces), take time to explain it

### After Writing Code
- Walk through what we built if it was complex
- Point out anything I should pay special attention to
- Note any "gotchas" a Python developer might trip over

### Example Interaction
```
Bill: "Let's implement the health endpoint"

Claude: "Before we write this, here's what we're doing:
- We'll create an HTTP handler function that responds to GET /health
- In Go, handlers have a specific signature: func(w http.ResponseWriter, r *http.Request)
- Unlike Flask where you return a response, in Go you write directly to the ResponseWriter
- We'll return JSON with status and timestamp

Here's the code:
[code]

The http.HandleFunc registers our handler with the default mux (router).
That w.Header().Set() call — unlike Python dicts, you use .Set() method..."

Bill: "Wait, explain the ResponseWriter thing more"

Claude: [pauses and explains before continuing]
```

---

## Code Standards

### Go Conventions
- Follow standard Go project layout
- Use `gofmt` / `goimports` formatting
- Prefer stdlib where reasonable; justify external dependencies
- Handle errors explicitly — no silent failures
- Write idiomatic Go, but comment when doing something non-obvious for a Go beginner

### Documentation
- Update README.md when adding features
- **Update README.md status and quick start when completing each phase**
- Keep PLAN.md current with architectural decisions
- Add session notes to DEVLOG.md for blog material
- Add comments explaining "why" for complex logic
- Include examples in API documentation

### Git
- Small, focused commits
- Conventional commit style: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Reference PLAN.md milestones in commits when relevant

## Testing Approach

- Write tests for API endpoints
- Test happy paths first, edge cases as needed
- Integration tests over excessive unit tests for this project size
- Manual testing is fine during early development

## What to Avoid

- Over-engineering — this is a demo app, not production infrastructure
- Adding libraries/abstractions prematurely — start with stdlib, add dependencies only when there's a clear need
- External service dependencies — should run anywhere with zero setup
- Vendor lock-in — no cloud-specific SDKs baked into core functionality
- Premature optimization — get it working, then make it fast/clean
- **Writing code without explanation** — see "Learning-First Development" above

---

## Claude Code / CLI Sessions

Bill typically opens Claude CLI from `~/code` so Claude can access multiple repos simultaneously.

### Key Paths
| Path | Contents |
|------|----------|
| `~/code/demo-app` | This project |
| `~/code/billgrant.github.io` | Blog (Jekyll site) |
| `~/code/music-graph` | Previous project (Flask, for reference) |

### Development Environment

**OS:** PopOS 24.04

**Tooling:**
| Tool | Version/Details |
|------|-----------------|
| Go | 1.25.5 (official tarball, `/usr/local/go`) |
| Docker | Docker Engine (official repo) |
| VS Code | With Go and Claude extensions |
| gh CLI | GitHub CLI for repo operations |
| git | System package |
| make | System package |
| curl | System package |

**GitHub Auth:** SSH and HTTPS configured

### Go Process Naming Quirk
When running `go run main.go`, Go compiles to a temporary binary named after the source file (`main`), NOT the project name. The process will be something like `/tmp/go-build.../main`.

**To kill the dev server:**
```bash
pkill main          # kills by process name
# NOT pkill demo-app  # won't work during development
```

The `demo-app` name only exists when explicitly built with `go build -o demo-app`.

### Starting a Session
1. Read `PLAN.md` first — understand current phase and milestones
2. Read this file (`AGENTS.md`) for collaboration guidelines
3. Check recent commits to understand where we left off
4. Ask if context seems missing before making changes

### Session Context Bill May Provide
- Current phase/milestone we're working on
- Specific files to focus on
- Error messages or issues to debug
- "Let's write a blog post about X"

---

## Blog Documentation

Each phase gets a blog post at https://billgrant.io

### Blog Details
- **Repo:** `~/code/billgrant.github.io` (Jekyll site)
- **Posts directory:** `_posts/`
- **Filename format:** `YYYY-MM-DD-title.md`
- **Tag:** `#demo-app`

### Writing Style & Workflow

**The `*** ***` markers serve two purposes:**
1. Bill's voice — intro sets context, outro reflects on learnings
2. **AI transparency disclaimer** — signals to readers that the main content below was written by Claude, not Bill. This is intentional transparency about AI use.

**Workflow:**
1. Claude drafts the full post using DEVLOG.md session notes
2. Bill rewrites the intro/outro sections in his own voice
3. Bill reviews main content, adds screenshots or additional context
4. Claude spell-checks and grammar-checks Bill's edits
5. Bill commits and pushes to the blog repo

**Content guidelines:**
- Documents successes AND failures — the learning journey matters
- Include code snippets, decisions made, and lessons learned
- Main content is Claude's synthesis of the sessions, not purely Bill's writing

### Jekyll/Template Code in Posts
When showing Go templates or any curly-brace syntax in blog posts, wrap with raw tags to prevent Jekyll from processing:

```
{% raw %}
{{ .SomeVariable }} or {{ if .Condition }}{{ end }}
{% endraw %}
```

This applies to:
- Go `html/template` or `text/template` examples
- Any `{{ }}` syntax that Jekyll might try to interpret

### Blog Post Structure (suggested)
```markdown
---
layout: post
title: "Demo App: [Phase/Topic]"
date: YYYY-MM-DD
tags: demo-app go learning-in-public
---

*** Bill's intro — sets context, what we set out to do ***

[Main content — what happened, code examples, decisions]
```

Note: Intro only, no outro. The `*** ***` markers signal AI transparency — content below is Claude's synthesis of the sessions.

---

## References

- [PLAN.md](PLAN.md) — Architecture, milestones, decisions
- [Music Graph CLAUDE.md](https://github.com/billgrant/music-graph) — Previous project's AI guidelines (for reference)
- [Blog](https://billgrant.io) — Published development journey

---

*Last updated: 2026-01-09*
