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
