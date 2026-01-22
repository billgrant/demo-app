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

## 2026-01-09 — Session 6: Frontend Dashboard

### What We Built
- Single-page dashboard with four panels (Health, System, Items, Display)
- Vanilla JavaScript — no frameworks, no build step
- Static file serving from Go

### File Structure
```
static/
  index.html    # Page structure, panel containers
  style.css     # Dark theme dashboard layout (CSS Grid)
  app.js        # All JavaScript — API calls, rendering, modals
```

### JavaScript Concepts Covered

**`async/await` — Modern Asynchronous Code**
```javascript
async function fetchHealth() {
    const response = await fetch('/health');
    return await response.json();
}
```
- `async` marks function as asynchronous
- `await` pauses until Promise resolves
- Cleaner than callback chains or `.then()`

**`fetch()` — Browser's HTTP Client**
- Built-in function to make HTTP requests
- Returns a Promise
- Same as curl, just from JavaScript

**DOM Manipulation**
```javascript
const container = document.getElementById('health-content');
container.innerHTML = `<div>${data.status}</div>`;
```
- `getElementById` finds elements by their `id` attribute
- `innerHTML` sets the HTML content inside an element
- Template literals (backticks) allow variable interpolation

**Event Listeners**
```javascript
document.getElementById('add-item-btn').addEventListener('click', handleAddItem);
```
- Attach functions to run when events happen (click, load, etc.)
- Arrow functions `() => {}` as callbacks

**`setInterval()` — Repeated Execution**
```javascript
setInterval(refreshHealth, 10000);  // every 10 seconds
```
- Runs a function repeatedly at specified interval
- Used for auto-refreshing health panel

### Go Changes

**Serving Static Files**
```go
http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
```
- `http.FileServer` serves files from a directory
- `http.StripPrefix` removes URL prefix before looking up file

**Root Redirect**
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.Redirect(w, r, "/static/index.html", http.StatusFound)
        return
    }
    http.NotFound(w, r)
})
```

### Key Insight: Frontend-Backend Separation

The browser (JavaScript) and server (Go) are completely decoupled:
- JS makes HTTP requests to API endpoints
- Go returns JSON responses
- They don't know anything about each other's implementation
- Could run on different machines, be written in different languages
- This is why REST APIs are universal interfaces between services

### Files Changed
- `static/index.html` — new file, page structure
- `static/style.css` — new file, dark theme dashboard
- `static/app.js` — new file, all JavaScript logic
- `main.go` — added static file serving and root redirect

### Phase 3 Progress
- [x] Single-page dashboard (vanilla JS)
- [x] Health panel (auto-refresh)
- [x] System info panel
- [x] Items panel with CRUD
- [x] Display panel
- [x] Embed static files in binary

---

## 2026-01-12 — Session 7: Embed Static Files

### What We Built
- Embedded static files into the Go binary using `embed` package
- Binary is now fully self-contained (15MB)
- Can run from any directory without needing the static folder

### Go Concepts Covered

**`//go:embed` Directive**
```go
//go:embed static/*
var staticFiles embed.FS
```
- Compiler directive (not a regular comment)
- Tells Go to embed files matching the pattern at build time
- Files become part of the binary itself

**`embed.FS` — Embedded File System**
- Implements `fs.FS` interface
- Read-only file system backed by embedded data
- Preserves directory structure (files at `static/index.html`)

**`fs.Sub` — Sub-Filesystem**
```go
staticFS, err := fs.Sub(staticFiles, "static")
```
- Creates a new filesystem rooted at a subdirectory
- Needed because embed preserves full paths
- After `fs.Sub`, files are at `index.html` not `static/index.html`

### The Change
```go
// Before: reads from disk
http.FileServer(http.Dir("static"))

// After: reads from embedded files
staticFS, _ := fs.Sub(staticFiles, "static")
http.FileServer(http.FS(staticFS))
```

### Testing
```bash
# Build
go build -o demo-app

# Copy to /tmp (no static folder)
cp demo-app /tmp/
cd /tmp
./demo-app  # Dashboard works!
```

### Files Changed
- `main.go` — added `embed` and `io/fs` imports, `//go:embed` directive, changed file server

### Phase 3 Complete ✓
All frontend items done:
- [x] Single-page dashboard
- [x] All four panels working
- [x] Static files embedded in binary

---

## 2026-01-12 — Session 8: Docker & Multi-Arch

### What We Built
- Updated Dockerfile for embedded static files
- Added multi-arch build support (amd64 + arm64)
- Tested containerized deployment

### Dockerfile Changes

**Copy static files for embedding:**
```dockerfile
COPY static/ ./static/
```

**Multi-arch support with buildx:**
```dockerfile
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /demo-app
```
- `TARGETOS` and `TARGETARCH` are set automatically by `docker buildx`
- Defaults to `linux/amd64` for regular `docker build`

### Docker Buildx Setup

**Install QEMU for cross-platform emulation:**
```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

**Create multi-arch builder:**
```bash
docker buildx create --name multiarch --driver docker-container --bootstrap --use
```

**Build for multiple platforms:**
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t demo-app:multiarch .
```

### Notes
- Local multi-arch builds via QEMU emulation are slow
- Production builds will use CI/CD with native arm64 runners (Phase 8)
- The Dockerfile is ready, just needs the right build environment

### Files Changed
- `Dockerfile` — added static folder copy, TARGETOS/TARGETARCH args

### Phase 4 Complete ✓

---

## 2026-01-15 — Session 9: BadgerDB Refactor & Demo Examples

### What We Built
- Replaced SQLite with BadgerDB for concurrent write support
- Created `demo-app-examples` repo with baseline Terraform demo
- Single `terraform apply` that provisions container AND populates data

### Why BadgerDB?

**The Problem:**
SQLite's in-memory mode (`:memory:`) doesn't support concurrent writes. When Terraform creates multiple resources in parallel, SQLite returns "database is locked" errors.

**Options Considered:**
1. SQLite WAL mode — doesn't work with `:memory:`
2. File-based SQLite — inconsistent experience between binary/container
3. **BadgerDB** — K/V store with native concurrent write support

**Decision:** BadgerDB gives consistent ephemeral behavior everywhere while supporting parallel Terraform operations.

### Go Concepts Covered

**K/V Database Pattern**
```go
// SQL: rows and columns
db.Query("SELECT * FROM items WHERE id = ?", id)

// K/V: keys and values
txn.Get([]byte("item:1"))  // returns JSON blob
```
- Data stored as `key → JSON blob`
- No schema, no migrations
- Simpler for our use case

**BadgerDB Transactions**
```go
// Read-only (concurrent safe)
db.View(func(txn *badger.Txn) error {
    item, _ := txn.Get(key)
    return item.Value(func(val []byte) error {
        return json.Unmarshal(val, &result)
    })
})

// Read-write
db.Update(func(txn *badger.Txn) error {
    return txn.Set(key, value)
})
```
- `View()` for reads — multiple can run concurrently
- `Update()` for writes — serialized but fast
- Callback pattern ensures cleanup

**Sequences for Auto-Increment IDs**
```go
itemSeq, _ := db.GetSequence([]byte("seq:items"), 100)
id, _ := itemSeq.Next()  // atomic, concurrent-safe
```
- Replaces SQL's `AUTOINCREMENT`
- Pre-allocates IDs in batches for performance

**Iterator with Prefix**
```go
prefix := []byte("item:")
for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
    // process each item
}
```
- Replaces `SELECT * FROM items`
- Scans all keys with matching prefix

### Baseline Demo Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Docker         │     │  HTTP           │     │  DemoApp        │
│  Provider       │────►│  Provider       │────►│  Provider       │
│                 │     │  (data source)  │     │                 │
│  Creates        │     │  Fetches from   │     │  Posts to       │
│  container      │     │  /api/system    │     │  /api/display   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**Key Design Decision:** `lifecycle { ignore_changes = [data] }` on the display resource. Terraform is the persistence layer — if app crashes, `terraform apply` restores the same data, not new data that might have changed.

### Files Changed
- `main.go` — replaced SQLite with BadgerDB (~100 lines changed)
- `go.mod` / `go.sum` — new dependency
- `PLAN.md` — Phase 6 complete, tech stack updated
- New repo: `demo-app-examples/baseline/`

### Testing Results

| Test | Before (SQLite) | After (BadgerDB) |
|------|-----------------|------------------|
| 10 parallel creates | 2 failures | All succeed |
| `-parallelism=1` needed | Yes | No |

### Phase 6 Complete ✓

---

## 2026-01-16 — Session 10: Phase 7 Planning

### What We Planned

Phase 7 scope defined — "Observability & Polish":

| Item | Description |
|------|-------------|
| Prometheus `/metrics` | App metrics + Go runtime + process metrics |
| Log webhook shipping | Optional `LOG_WEBHOOK_URL` for pushing logs to any HTTP endpoint |
| Request header display | Show incoming headers in `/api/system` |
| Environment variable filtering | Regex-based via `ENV_FILTER` |
| Configuration documentation | Document all env vars |

### Metrics Discussion

**App-specific metrics planned:**
- `demoapp_http_requests_total` (counter, labels: method, path, status)
- `demoapp_http_request_duration_seconds` (histogram)
- `demoapp_items_total` (gauge)
- `demoapp_display_updates_total` (counter)
- `demoapp_info` (gauge with version/commit labels)

**System metrics (free from library):**
- Go runtime: goroutines, memory, GC
- Process: CPU, file descriptors

**Why Prometheus format over OpenTelemetry:**
- Simpler to implement
- Wide compatibility — Splunk, Datadog, Grafana all support it
- OTel is more complex; Bill's Grafana interview experience confirmed setup pain

### Log Webhook Design

**Problem:** Stdout logs work great with container runtimes (Docker captures, K8s ships via agents), but what about demos where you're running the binary directly and want logs in Splunk/Loki?

**Solution:** Optional webhook — if `LOG_WEBHOOK_URL` is set, POST each log entry as JSON.

```
Log event → Write to stdout (always) → If webhook configured, also POST to URL
```

**Keeps app vendor-neutral:**
- No Splunk SDK, no Loki SDK
- Just HTTP POST with JSON
- Receiving end handles format transformation

### Deferred Items

Moved to Future Considerations:

| Item | Reason |
|------|--------|
| External IP detection | Needs more thought — client IP (from headers) may be more useful than app's public IP |
| Highlights endpoint | New idea about auto-parsing display data; not ready to design |
| UI polish | Wait until feature set is finalized |

### Files Changed
- `PLAN.md` — Phase 7 scope updated, deferred items documented
- `DEVLOG.md` — this session

### Next Session
Implement Phase 7: Prometheus metrics endpoint first, then log webhook, then header display.

---

## 2026-01-19 — Session 11: Prometheus Metrics Implementation

### What We Built
- Prometheus `/metrics` endpoint using `prometheus/client_golang`
- Custom app metrics for HTTP requests, items, and display updates
- Path normalization to prevent cardinality explosion

### Metrics Implemented

| Metric | Type | Labels | Purpose |
|--------|------|--------|---------|
| `demoapp_http_requests_total` | Counter | method, path, status | Track request counts |
| `demoapp_http_request_duration_seconds` | Histogram | method, path | Track response times |
| `demoapp_items_total` | Gauge | — | Current item count |
| `demoapp_display_updates_total` | Counter | — | Track display POSTs |
| `demoapp_info` | Gauge | version | Build info (always 1) |

**Plus free from the library:**
- `go_goroutines` — active goroutines
- `go_memstats_*` — memory allocation stats
- `go_gc_*` — garbage collection timing
- `process_*` — CPU, file descriptors

### Go/Prometheus Concepts Covered

**Metric Types**
- **Counter** — only increases (requests, errors, updates)
- **Gauge** — can go up or down (current items, connections)
- **Histogram** — tracks distribution of values (response times)

**Labels — Dimensions for Slicing Data**
```go
httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "demoapp_http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},  // labels
)
```
- Labels create sub-metrics: `demoapp_http_requests_total{method="GET",path="/health",status="200"}`
- Each unique label combination is a separate time series
- Too many labels = "cardinality explosion" (memory/storage issues)

**Path Normalization — Avoiding High Cardinality**
```go
func normalizePath(path string) string {
    if strings.HasPrefix(path, "/api/items/") {
        return "/api/items/:id"  // /api/items/123 -> /api/items/:id
    }
    return path
}
```
- Without this, `/api/items/1`, `/api/items/2`, etc. would each create a new metric series
- Normalizing to `/api/items/:id` keeps cardinality bounded

**`init()` Function — Auto-Run Before Main**
```go
func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    // ...
}
```
- `init()` runs automatically before `main()`
- Common pattern for registering metrics, database drivers, etc.
- Multiple `init()` functions allowed in a package

**Why /metrics Isn't Logged**
```go
// No loggingMiddleware wrapper
http.Handle("/metrics", promhttp.Handler())
```
- Prometheus scrapes every 15-60 seconds
- Would flood logs with noise
- Would create metrics about the metrics endpoint (recursive)

### Testing Results

```bash
# After creating 2 items and 1 display update:
demoapp_items_total 2
demoapp_display_updates_total 1
demoapp_http_requests_total{method="POST",path="/api/items",status="201"} 2
demoapp_http_requests_total{method="POST",path="/api/display",status="201"} 1

# After deleting 1 item:
demoapp_items_total 1  # gauge decreased
```

### Files Changed
- `main.go` — added imports, metrics definitions, `init()`, instrumented middleware and handlers
- `go.mod` / `go.sum` — added `prometheus/client_golang` dependency

### Phase 7 Progress
- [x] Prometheus `/metrics` endpoint
- [ ] **Code refactoring** (added this session)
- [ ] Log webhook shipping
- [ ] Request header display
- [ ] Environment variable filtering
- [ ] Configuration documentation

### End-of-Session Discussion: Future Plans

**1. Architecture Diagram for Demos**
- Visual documentation showing how the app works
- Format: Mermaid (renders on GitHub, version controlled)
- Content: architecture overview, request flow, demo scenarios
- Timing: Wait until feature set is finalized (added to Phase 9)

**2. Refactoring main.go**
- File has grown to ~700 lines — mixing handlers, middleware, database, metrics
- Plan: Split into multiple files within same package:
  - `main.go` — startup, routing, configuration
  - `handlers.go` — HTTP handlers
  - `store.go` — BadgerDB operations
  - `middleware.go` — logging with metrics
  - `metrics.go` — Prometheus definitions
- **Key requirement:** Provide detailed explanations during refactor. Bill noted he's having trouble following code as it moves around. This aligns with AGENTS.md "Learning-First Development" — explain what's moving and why.
- Timing: Do before remaining Phase 7 items (cleaner foundation)

### Next Session
Code refactoring with detailed explanations, then continue Phase 7 items.

---

## 2026-01-21 — Session 12: Code Refactoring

### What We Built
- Split `main.go` (~730 lines) into 5 focused files
- Zero behavior changes — same functionality, better organization
- All tests pass

### File Structure After Refactoring

| File | Lines | Responsibility |
|------|-------|----------------|
| `handlers.go` | 455 | HTTP endpoint logic (health, items, display, system) |
| `main.go` | 149 | Startup, configuration, routing |
| `middleware.go` | 90 | Request logging, metrics recording |
| `metrics.go` | 74 | Prometheus metric definitions, `init()` |
| `store.go` | 67 | Data model, database setup |
| **Total** | 835 | (was ~730, grew due to added comments) |

### Go Concepts Covered

**Package Scope — Multiple Files, One Namespace**

All `.go` files in the same directory with `package main` share variables and functions automatically:

```go
// store.go
var db *badger.DB  // package-level

// handlers.go
func listItems(w http.ResponseWriter, r *http.Request) {
    db.View(...)  // uses db from store.go directly — no import needed
}
```

This is why you don't need imports between files in the same package. Think of it like all files being concatenated into one big file at compile time.

**Terraform Parallel**

Bill noticed this is exactly how Terraform works — all `.tf` files in a directory share the same namespace. Variables, resources, and data sources from any file are accessible in all other files. Same design pattern, because Terraform is written in Go.

**Variable Scope Levels**

```go
var X = 10  // package-level — all files can access

func main() {
    X := 5  // function-level — shadows package-level X (new variable!)

    if true {
        X := 3  // block-level — shadows function-level X
    }
}
```

Key insight: `X := 5` inside a function creates a NEW local variable, even if a package-level `X` exists. The package-level `X` is "shadowed" (hidden), not modified.

**Shadowing Is Just the Name**

"Shadowing" creates a completely new, independent variable that happens to have the same name. No copying, no connection to the outer variable:

```go
var count = 10  // package-level

func example() {
    count := 5   // NEW variable, doesn't touch outer count
    fmt.Println(count)  // 5
}

func main() {
    example()
    fmt.Println(count)  // 10 — unchanged
}
```

The term "shadow" comes from the visual metaphor: the inner variable "casts a shadow" over the outer one, hiding it from view within that scope.

**`:=` vs `=` — The Critical Distinction**

```go
var db *badger.DB  // package-level

func main() {
    db, err := initStore()  // WRONG: creates local db, package-level stays nil!

    var err error
    db, err = initStore()   // RIGHT: assigns to existing package-level db
}
```

- `:=` — declares AND initializes a NEW variable (short declaration)
- `=` — assigns to an EXISTING variable

This is why our code uses `var err error` then `db, err = initStore()` — to assign to the package-level `db`, not create a local one.

**Package-Level Variables Don't Need Passing**

Unlike Python where you might need `global` to modify a module-level variable:

```go
var db *badger.DB  // package-level

func handler(w http.ResponseWriter, r *http.Request) {
    // Can read AND write db directly — no passing needed
    db.View(...)
}
```

Functions can freely access package-level variables without them being passed as parameters.

### What Moved Where

**metrics.go** — "What are we measuring?"
- Metric variable definitions (`httpRequestsTotal`, `httpRequestDuration`, etc.)
- `init()` function that registers metrics

**middleware.go** — "What happens to every request?"
- `responseRecorder` struct (captures HTTP status)
- `loggingMiddleware` function (logs + records metrics)
- `normalizePath` function (prevents cardinality explosion)

**store.go** — "How do we store data?"
- `itemKeyPrefix` constant (`"item:"`)
- Package-level variables: `db`, `itemSeq`, `displayData`
- `Item` struct
- `initStore` function

**handlers.go** — "What does each endpoint do?"
- `healthHandler`
- `itemsHandler`, `listItems`, `createItem`, `getItem`, `updateItem`, `deleteItem`
- `displayHandler`, `getDisplay`, `setDisplay`
- `systemHandler`, `getIPAddresses`, `getFilteredEnvVars`

**main.go** — "How does the app start?"
- `//go:embed` directive and `staticFiles`
- `runHealthcheck` function (Docker healthcheck)
- `main()` with config, initialization, and routing

### Testing Results

```bash
# Build succeeds
go build -o demo-app .

# All endpoints work
curl http://localhost:8080/health
# {"status":"ok","timestamp":"2026-01-21T15:07:30Z"}

curl -X POST http://localhost:8080/api/items -d '{"name":"Test"}'
# {"id":0,"name":"Test",...}

curl http://localhost:8080/metrics | grep demoapp_items_total
# demoapp_items_total 1
```

### Files Changed
- `metrics.go` — new file
- `middleware.go` — new file
- `store.go` — new file
- `handlers.go` — new file
- `main.go` — trimmed from ~730 to 149 lines

### Phase 7 Progress
- [x] Prometheus `/metrics` endpoint
- [x] **Code refactoring**
- [ ] Log webhook shipping
- [ ] Request header display
- [ ] Environment variable filtering
- [ ] Configuration documentation

---

## 2026-01-21 — Session 12 (continued): Request Headers & Client Info

### What We Built
- Added request headers to `/api/system` response (API only)
- Added `client_ip` and `user_agent` to system info (API and dashboard)

### Design Decisions

**Headers: API-only, not on dashboard**

Initially added headers to the dashboard, but removed them because:
- Browsers send many verbose headers (Accept, Accept-Encoding, Accept-Language, User-Agent, etc.)
- Takes up too much screen real estate
- Not particularly useful to show an audience during a demo

Headers remain available via `curl /api/system | jq .headers` for debugging proxy chains, auth issues, etc.

**Client IP and User Agent: Dashboard-friendly**

Instead of raw headers, added these two fields which are more demo-useful:
- `client_ip` — shows who's hitting the app (from `r.RemoteAddr`)
- `user_agent` — shows what client is making the request (from `r.UserAgent()`)

**Use case for demos:**

When demoing Terraform provisioning an EC2:
1. System Info shows the EC2's local IP (proving where app is deployed)
2. Client IP shows the demo engineer's IP (proving traffic flows from their machine)

Two different IPs on screen = simple visual proof the infrastructure works.

### Go Concepts Covered

**`r.RemoteAddr`** — The client's IP:port that made the request. For proxied requests, this is the proxy's IP (real client IP would be in `X-Forwarded-For` header).

**`r.UserAgent()`** — Convenience method that returns the `User-Agent` header value. Equivalent to `r.Header.Get("User-Agent")`.

### Files Changed
- `handlers.go` — added `getRequestHeaders()`, `client_ip`, `user_agent` to systemHandler
- `static/app.js` — added Client IP and User Agent rows to System Info panel

### Phase 7 Progress (updated)
- [x] Prometheus `/metrics` endpoint
- [x] **Code refactoring**
- [x] Request header display (API-only + client_ip/user_agent on dashboard)
- [ ] Log webhook shipping
- [ ] Environment variable filtering
- [ ] Configuration documentation

### Next Session
Continue Phase 7: log webhook shipping, env filtering, config docs.

---

## 2026-01-21 — Session 13: Log Webhook Shipping

### What We Built
- Log webhook feature: if `LOG_WEBHOOK_URL` is set, logs POST to that URL
- Optional `LOG_WEBHOOK_TOKEN` for Authorization header
- Custom `slog.Handler` implementation that wraps JSONHandler
- Test scripts for manual verification

### Go Concepts Covered (Deep Dive)

This session included extended discussion of core Go concepts. Documenting here for future reference.

**`context.Context` — The Request Envelope**

Context carries request-scoped data across API boundaries:
- **Cancellation** — signals "stop working on this"
- **Deadlines/Timeouts** — automatic cancellation after X time
- **Values** — request-scoped data (use sparingly)

Networking analogy: Like packet metadata (TTL, source info) that travels with the payload. Functions can check if the request is still alive.

```go
func Handle(ctx context.Context, record slog.Record) error {
    // ctx tells us: is this request still alive? any deadline?
}
```

**Interfaces — The Contract**

An interface defines what something can DO (methods). A struct defines what something HAS (fields).

```go
// Interface: "anything with these methods qualifies"
type Driveable interface {
    Drive()
    Stop()
}

// Struct: "this is the data"
type Car struct {
    Brand string
    Speed int
}

// Methods make Car satisfy Driveable
func (c *Car) Drive() { ... }
func (c *Car) Stop() { ... }
```

Key insight: In Go, you don't explicitly declare "Car implements Driveable". If Car has the methods, it automatically satisfies the interface. Compile-time duck typing.

**`type` Keyword — Creating Types**

`struct` and `interface` are not types themselves — they're ways to CREATE types:

```go
type Item struct { ... }      // creates type "Item" (kind: struct)
type Handler interface { ... } // creates type "Handler" (kind: interface)
type UserID int               // creates type "UserID" (based on int)
```

**Method Receivers and Pointers**

Methods are defined OUTSIDE the struct (unlike Python classes):

```go
type Car struct { Brand string }

// Method belongs to *Car (pointer receiver)
func (c *Car) Drive() { ... }
```

The `(c *Car)` is the receiver — like `self` in Python. Because it's `*Car` (pointer), the method belongs to the pointer type. That's why we use `&Car{}`:

```go
car := &Car{Brand: "Toyota"}  // *Car satisfies Driveable
car := Car{Brand: "Toyota"}   // Car does NOT satisfy Driveable (methods are on *Car)
```

**slog.Handler Interface**

The stdlib `slog` package uses handlers to control where logs go:

```go
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Record) error
    WithAttrs(attrs []Attr) Handler
    WithGroup(name string) Handler
}
```

Our `webhookHandler` implements this interface, wrapping the JSONHandler and adding webhook functionality.

### Design Decisions

**Async webhook calls:**
Webhooks POST in a goroutine (`go w.postToWebhook(entry)`) so slow/failed webhooks don't block HTTP responses.

**Fire and forget:**
Failed webhook calls log to stderr but don't affect the application. Logs always go to stdout; webhook is best-effort.

**Vendor neutral:**
No Splunk SDK, no Loki SDK. Just HTTP POST with JSON. The receiving system handles format transformation.

### Files Changed
- `webhook.go` — new file, custom slog.Handler implementation
- `main.go` — reads LOG_WEBHOOK_URL/TOKEN, configures handler
- `scripts/test-webhook.sh` — automated test script
- `scripts/webhook-receiver/main.go` — simple Go server for manual testing

### Configuration

| Env Var | Description | Default |
|---------|-------------|---------|
| `LOG_WEBHOOK_URL` | URL to POST log entries to | (disabled) |
| `LOG_WEBHOOK_TOKEN` | Value for Authorization header | (none) |

### Testing

```bash
# Automated test
./scripts/test-webhook.sh

# Manual test (two terminals)
# Terminal 1:
go run scripts/webhook-receiver/main.go

# Terminal 2:
LOG_WEBHOOK_URL="http://localhost:9999/logs" ./demo-app
```

### Phase 7 Progress
- [x] Prometheus `/metrics` endpoint
- [x] Code refactoring
- [x] Request header display
- [x] **Log webhook shipping**
- [ ] Environment variable filtering
- [ ] Configuration documentation

### Next Session
Continue Phase 7: env filtering, config docs.

---

## 2026-01-22 — Session 14: ENV_FILTER & Configuration Docs

### What We Built
- `ENV_FILTER` environment variable for regex-based filtering of displayed env vars
- `docs/CONFIGURATION.md` — comprehensive configuration documentation
- CSS fix for long environment variable names in dashboard
- Updated README.md with current status and features

### Go Concepts Covered

**`os.Getenv()` vs `os.Hostname()` — Two Different Sources**

```go
os.Getenv("HOSTNAME")  // Reads from environment variables (may not exist)
os.Hostname()          // Syscall to kernel (always works)
```

- Environment variables are inherited from the parent process (your shell)
- `os.Hostname()` asks the kernel directly — same as running `hostname` in bash
- Docker/K8s often set `HOSTNAME` env var automatically; desktop Linux typically doesn't

Bash equivalent to see all env vars:
```bash
env        # or printenv — same data os.Environ() returns
```

**`regexp` Package — Pattern Matching**

```go
// Compile once, use many times
re, err := regexp.Compile("(?i)" + pattern)  // (?i) = case-insensitive

// Test if string matches
if re.MatchString(key) { ... }
```

- Similar to Python's `re.compile()` and `pattern.match()`
- `(?i)` prefix makes the pattern case-insensitive
- Invalid patterns return an error on `Compile()` — handle gracefully

**`strings.SplitN()` — Controlled Splitting**

```go
// os.Environ() returns "KEY=value" strings
parts := strings.SplitN(envVar, "=", 2)
key, value := parts[0], parts[1]
```

- `SplitN(s, sep, n)` splits into at most `n` parts
- Important because values can contain `=` characters
- Without the `2`, `DB_URL=postgres://user:pass@host` would split wrong

### Design Decisions

**Regex replaces allowlist when set:**
- Default (no `ENV_FILTER`): safe allowlist of common vars
- With `ENV_FILTER`: user takes full control, regex matches against ALL env vars
- Security note documented — user responsible for not exposing secrets

**Case-insensitive matching:**
- `ENV_FILTER="^demo"` matches `DEMO_VERSION`
- More user-friendly, less error-prone

**Invalid regex handling:**
- Logs error, returns empty map
- Better to show nothing than crash or expose unintended vars

### CSS Fix

**Problem:** Environment variable names longer than 100px (like `DEMO_LONG_VARIABLE_NAME`) overlapped with values in the dashboard.

**Fix:** Changed `.info-label` from `width: 100px` to `min-width: 100px` + `margin-right: 1rem`. Labels now grow to fit content.

### Documentation Structure

Created `docs/` folder for detailed documentation:
- `docs/CONFIGURATION.md` — full env var docs with examples
- README.md has quick reference table + link to full docs

This keeps the README scannable while providing depth for users who need it.

### Files Changed
- `handlers.go` — added `regexp` import, rewrote `getFilteredEnvVars()` with regex support
- `static/style.css` — fixed `.info-label` width for long env var names
- `docs/CONFIGURATION.md` — new file, comprehensive configuration docs
- `README.md` — updated status to Phase 7, fixed SQLite→BadgerDB refs, added config link
- `PLAN.md` — marked Phase 7 complete

### Phase 7 Complete ✓

All items done:
- [x] Prometheus `/metrics` endpoint
- [x] Code refactoring
- [x] Request header display
- [x] Log webhook shipping
- [x] Environment variable filtering
- [x] Configuration documentation

### Next Up
Phase 8: CI/CD — tests, GitHub Actions, releases.

---
