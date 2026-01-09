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

---

## 2026-01-07 — Session 3: Dockerfile with Docker Hardened Images

### What We Built
- Multi-stage Dockerfile using Docker Hardened Images (DHI)
- Built-in healthcheck subcommand for Docker HEALTHCHECK
- Phase 1 complete!

### Docker Hardened Images (DHI)

**Why DHI?**
1. **Shift-left security** — Start secure, don't fix later
2. **Clean baseline for security demos** — CVE-free base means intentional vulnerabilities stand out
3. **Learning opportunity** — Understand Docker's hardened image implementation

**Registry & Access:**
- Registry: `dhi.io`
- Auth: `docker login dhi.io` (uses Docker Hub credentials)
- Free tier available, no subscription required

**Images Used:**
| Stage | Image | Purpose |
|-------|-------|---------|
| Build | `dhi.io/golang:1.25-alpine3.22-dev` | Full Go SDK for compilation |
| Runtime | `dhi.io/static:20250911-alpine3.22` | Minimal static base (~CVE-free) |

### Go Concepts Covered

**`os.Args` — Command Line Arguments**
```go
if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
    runHealthcheck()
    return
}
```
- `os.Args[0]` is the program name
- `os.Args[1:]` are the arguments
- Used to add subcommand support without a CLI framework

**Self-contained healthcheck:**
- Static images have no curl/wget
- Solution: binary checks itself via HTTP
- `./demo-app healthcheck` makes request to `localhost:8080/health`
- Returns exit code 0 (healthy) or 1 (unhealthy)

### Docker Concepts Covered

**Multi-stage builds:**
- Build stage has full SDK (large)
- Runtime stage has only the binary (tiny)
- `COPY --from=build-stage` transfers artifacts between stages

**HEALTHCHECK directive:**
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/demo-app", "healthcheck"]
```
- Docker runs this periodically
- Shows in `docker ps` as `(healthy)` or `(unhealthy)`

### Known Quirk: DHI Authentication

**Issue:** Intermittent "pull access denied" errors during build, even when logged in:
```
ERROR: failed to build: pull access denied, repository does not exist
or may require authorization: server message: insufficient_scope
```

**Workaround:** Run `docker login dhi.io` immediately before building. This seems to refresh the auth token. May be related to DHI being a newer service.

**Commands:**
```bash
docker login dhi.io
docker build -t demo-app .
docker run --rm -p 8080:8080 demo-app
```

### Files Changed
- `Dockerfile` — new file, multi-stage DHI build
- `main.go` — added `runHealthcheck()` function and subcommand check

### Phase 1 Complete!
All foundation items done:
- [x] Go project structure
- [x] HTTP server with /health
- [x] Structured logging
- [x] SQLite integration
- [x] Dockerfile

### Next Up
- Phase 2: CRUD endpoints (`/api/items`)
- Display panel endpoints (`/api/display`)
- System info endpoint (`/api/system`)

---

## 2026-01-08 — Session 4: CRUD Endpoints (/api/items)

### What We Built
- Full CRUD API for `/api/items` endpoint
- Manual URL routing (no framework)
- Package-level database variable for handler access

### Go Concepts Covered

**Struct Tags for JSON**
```go
type Item struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
}
```
- Backtick strings are struct tags — metadata for serialization
- `omitempty` skips the field if empty

**`:=` vs `=` with Package-Level Variables**
```go
var db *sql.DB  // package-level

func main() {
    db, err := initDB(...)  // WRONG: creates new local db, shadows package-level

    var err error
    db, err = initDB(...)   // RIGHT: assigns to package-level db
}
```
- `:=` always creates a new variable in current scope
- To assign to existing variable, use `=`

**Manual Routing (stdlib limitation)**
```go
path := strings.TrimPrefix(r.URL.Path, "/api/items")
path = strings.TrimPrefix(path, "/")
// /api/items/123 -> "123"
// /api/items -> ""
```
- stdlib `net/http` doesn't support path parameters like `:id`
- Parse URL manually, route with switch on method
- This is what router libraries (Gin, Chi) do for you

**Switch Statements**
```go
switch r.Method {
case http.MethodGet:
    listItems(w, r)
case http.MethodPost:
    createItem(w, r)
default:
    http.Error(w, "not allowed", 405)
}
```
- Cleaner than if/else chains
- No fallthrough by default (unlike C)

**Query vs QueryRow vs Exec**
| Method | Use When |
|--------|----------|
| `db.Query()` | Multiple rows — returns `*Rows` to iterate |
| `db.QueryRow()` | Single row — returns `*Row` to scan once |
| `db.Exec()` | No rows returned (INSERT, UPDATE, DELETE) |

**Scan with Pointers**
```go
rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
```
- `Scan` needs pointers so it can write into your variables
- `&` gives the address (where to write)

### "Aha" Moments

1. **`:=` shadows package variables** — Using `:=` in a function creates a local variable even if a package-level one exists. Must use `=` to assign to existing variables.

2. **Pointers for sharing data** — If a function receives a value (not pointer), it gets a copy with its own address. Changes don't affect the original. Pointers let functions modify the original.

3. **`*` means different things** — In types: "pointer to" (`*Item`). In expressions: "value at" (dereference). Context determines meaning.

### API Endpoints Built

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/items` | List all items |
| POST | `/api/items` | Create item |
| GET | `/api/items/:id` | Get single item |
| PUT | `/api/items/:id` | Update item |
| DELETE | `/api/items/:id` | Delete item |

### Quick Reference: curl Commands
```bash
# List
curl http://localhost:8080/api/items

# Create
curl -X POST http://localhost:8080/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Item Name","description":"Optional"}'

# Get
curl http://localhost:8080/api/items/1

# Update
curl -X PUT http://localhost:8080/api/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"New Name","description":"New desc"}'

# Delete
curl -X DELETE http://localhost:8080/api/items/1
```

### Files Changed
- `main.go` — added Item struct, package-level db, CRUD handlers

### Next Up
- `/api/display` endpoint (injected demo content)
- `/api/system` endpoint (hostname, IPs, env vars)

---

## 2026-01-09 — Session 5: Display & System Endpoints

### What We Built
- `/api/display` — In-memory storage for arbitrary JSON (demo data injection)
- `/api/system` — Returns hostname, IPs, and filtered environment variables

### Go Concepts Covered

**`json.RawMessage` — Storing Arbitrary JSON**
```go
var displayData json.RawMessage
```
- Holds raw JSON bytes without parsing into a specific struct
- Accepts any valid JSON — objects, arrays, primitives
- Perfect when you don't know the structure ahead of time
- Like storing a JSON string in Python, but type-safe

**`map[string]interface{}` — Dynamic Maps**
```go
response := map[string]interface{}{
    "hostname": hostname,
    "ips":      ips,        // []string
    "environment": envVars, // map[string]string
}
```
- Map with string keys and any value type
- Like Python dict — values can be different types
- Modern Go prefers `any` over `interface{}` (same thing, cleaner syntax)

**`net.Interfaces()` — Network Information**
```go
interfaces, err := net.Interfaces()
for _, iface := range interfaces {
    if iface.Flags&net.FlagLoopback != 0 {
        continue // skip loopback
    }
    addrs, _ := iface.Addrs()
    // ...
}
```
- Returns all network interfaces on the system
- `&` is bitwise AND — checking if flag bit is set
- Each interface has addresses (IPs) attached

**Type Assertion — Extracting Concrete Types**
```go
if ipnet, ok := addr.(*net.IPNet); ok {
    // ipnet is now usable as *net.IPNet
}
```
- `addr` is an interface (can hold any type)
- `addr.(*net.IPNet)` tries to extract it as that specific type
- Returns value and boolean (success/failure)
- Like Python's `isinstance()` but also does the conversion

**`make()` — Initializing Maps**
```go
result := make(map[string]string)
```
- Creates an initialized, empty map
- Required before writing to a map
- `var m map[string]string` creates a nil map — can't write to it
- Like needing `m = {}` before `m["key"] = "value"` in Python

### Design Decisions

**Display: In-memory vs Database**
- Chose in-memory (`json.RawMessage` package variable)
- Data is transient demo content, doesn't need persistence
- Simplifies code — no schema needed for arbitrary JSON

**System: Environment Allowlist**
- Only expose specific env vars, not all
- `os.Environ()` would leak secrets
- Allowlist includes: PORT, DB_PATH, HOSTNAME, POD_NAME, etc.

### Files Changed
- `main.go` — added `net` import, displayHandler, systemHandler, helper functions
- `AGENTS.md` — documented Go process naming quirk (`pkill main` not `pkill demo-app`)
- `README.md` — updated status, added API endpoint documentation

### Phase 2 Progress
- [x] CRUD endpoints (`/api/items`)
- [x] Display panel (`/api/display`)
- [x] System info (`/api/system`)
- [ ] Frontend (Phase 2 remaining)

---
