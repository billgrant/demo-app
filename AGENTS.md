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

## Code Standards

### Go Conventions
- Follow standard Go project layout
- Use `gofmt` / `goimports` formatting
- Prefer stdlib where reasonable; justify external dependencies
- Handle errors explicitly — no silent failures
- Write idiomatic Go, but comment when doing something non-obvious for a Go beginner

### Documentation
- Update README.md when adding features
- Keep PLAN.md current with architectural decisions
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
- External service dependencies — should run anywhere with zero setup
- Vendor lock-in — no cloud-specific SDKs baked into core functionality
- Premature optimization — get it working, then make it fast/clean

## Blog Integration

I'm writing about this journey at billgrant.io. When we make significant decisions or overcome interesting challenges:
- Note it as potential blog content
- Help me articulate the learning if I ask
- The posts should reflect my actual thinking, enhanced by AI — not AI-generated content

## Context Files

When starting a coding session, I may provide:
- `PLAN.md` — current architecture and milestones
- Specific files we're working on
- Error messages or issues to debug

Read these before suggesting changes. Ask if context seems missing.

---

*Last updated: 2026-01-02*
