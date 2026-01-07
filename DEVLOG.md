# Demo App — Development Log

> Session notes for blog posts and future reference.

---

## 2026-01-06 — Session 1: Foundation + Logging

### What We Built
- Initialized Go module (`github.com/billgrant/demo-app`)
- Basic HTTP server with `/health` endpoint returning JSON
- Structured JSON logging with request middleware

### Decisions Made

| Decision | Choice | Reasoning |
|----------|--------|-----------|
| Router | stdlib `net/http` | Learn fundamentals before adding frameworks |
| Logging | `log/slog` | Stdlib, structured JSON, no external deps |
| SQLite driver | `modernc.org/sqlite` | Pure Go, easier cross-compilation for multi-arch Docker |
| Project structure | Flat (just `main.go`) | Start simple, refactor into packages when there's a reason |

### Go Concepts Covered

**Handler signature: `func(w http.ResponseWriter, r *http.Request)`**
- `w` is where you write your response — not a return value
- `r` is the incoming request (pointer to avoid copying)
- Key mental shift from Flask: you don't return responses, you write to `w`

**Parameter syntax: `name type`**
- `r *http.Request` means "r is a pointer to an http.Request"
- Go requires explicit types; Python infers them
- The `*` means pointer — a memory address, not a copy

**Pointers**
- `*Type` = "pointer to Type" (in type declarations)
- `&variable` = "address of variable" (to get a pointer)
- Without pointer: function gets a copy, changes stay local
- With pointer: function gets reference, changes affect original
- Python does this automatically for objects; Go makes it explicit

**Middleware pattern**
- Function that takes a handler, returns a new handler
- Wraps behavior before/after the actual handler runs
- Like Python decorators but explicit

**Struct embedding**
- Put a type inside a struct without a field name
- The struct "inherits" all methods of the embedded type
- Used for `responseRecorder` to wrap `http.ResponseWriter`

### "Aha" Moments

1. **ResponseWriter IS the return** — Flask's `return jsonify(...)` model hides that you're writing to a network connection. Go hands you the pipe directly.

2. **`next` is just a variable** — In the middleware, `next` isn't a keyword; it's the parameter name holding `healthHandler`. Could be called anything.

3. **`r *http.Request` syntax** — Reading it wrong initially. `r` is the name, `*http.Request` is the type. Go is `name type`, not `type name`.

4. **Why we can't read status codes** — `ResponseWriter` only has write methods. To log the status code, we wrap it with `responseRecorder` that intercepts `WriteHeader()` and saves the value.

### Code Highlights

```go
// Middleware wraps a handler to add behavior
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        recorder := &responseRecorder{ResponseWriter: w, statusCode: 200}
        next(recorder, r)  // call the actual handler
        slog.Info("request", "method", r.Method, "status", recorder.statusCode, ...)
    }
}
```

### Files Changed
- `go.mod` — module definition
- `main.go` — server, health handler, logging middleware

### Next Up
- SQLite integration
- Dockerfile
- Then Phase 2: CRUD endpoints

---

## 2026-01-07 — Session 2: SQLite Integration

### What We Built
- SQLite database integration using `modernc.org/sqlite`
- `initDB()` function that opens database and creates tables
- Support for `DB_PATH` environment variable (`:memory:` default, or file path)
- `items` table ready for Phase 2 CRUD

### Go Concepts Covered

**`database/sql` — The Database Abstraction Layer**
- Stdlib provides common interface for all databases
- Drivers implement the interface for specific databases
- You code to `database/sql`, swap drivers without changing queries

**Underscore Import: `_ "modernc.org/sqlite"`**
- The `_` means "import for side effects only"
- Driver's `init()` function registers itself with `database/sql`
- You never call the driver directly — `sql.Open("sqlite", path)` looks it up
- Looks weird but is standard Go pattern

**`:=` vs `=` — Declaration vs Assignment**
- `:=` declares a new variable and infers type
- `=` assigns to an existing variable
- Can't use `:=` twice for same variable in same scope

**Error Handling Pattern**
- Functions that can fail return `(result, error)`
- Check `if err != nil` immediately after every call
- No exceptions in Go — errors are values you handle explicitly
- You'll write `if err != nil` hundreds of times

**`defer` — Cleanup Scheduling**
- `defer db.Close()` schedules `Close()` to run when function exits
- Like Python's `with` statement or `finally` block
- Runs even if function exits due to error (after defer is registered)
- Belongs to the function it's written in, not the function that's called

**Scope in `if` statements**
```go
if err := db.Ping(); err != nil {
    // err only exists in this block
}
```

### "Aha" Moments

1. **Driver registration magic** — The underscore import felt weird until understanding that drivers self-register via `init()`. You import them, they register, `database/sql` finds them by name.

2. **`defer` scope** — It runs when the *surrounding* function exits, not when the called function exits. `defer db.Close()` in `main()` runs when `main()` returns.

3. **Errors are just values** — No try/catch, no exceptions. Every function that can fail returns an error, and you check it. Verbose but explicit.

### Container Consideration
For containerized deployments, the database file needs to live on a mounted volume to persist. The app creates/uses whatever path `DB_PATH` points to — just mount a volume there.

### Files Changed
- `go.mod` — added `modernc.org/sqlite` dependency
- `go.sum` — new file with dependency checksums
- `main.go` — added `initDB()` function, database initialization in `main()`

### Next Up
- Dockerfile (last Phase 1 item)
- Then Phase 2: CRUD endpoints

### Dockerfile Planning Notes
Want to explore **Docker Hardened Images** (https://www.docker.com/products/hardened-images/) instead of standard base images. Rationale:

1. **Shift-left security** — Start secure, don't fix later
2. **Clean baseline for security demos** — If demo-app tests a code scanner, a CVE-free base image means any vulnerabilities found are intentional (added for demo purposes)
3. **Learning opportunity** — Have used Chainguard/distroless before, want to understand Docker's implementation

This may add complexity to the Dockerfile phase but fits the project's purpose.

---
