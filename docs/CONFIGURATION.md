# Configuration

demo-app is configured entirely through environment variables. No config files needed.

## Quick Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `:memory:` | Database path (`:memory:` or file path) |
| `ENV_FILTER` | (allowlist) | Regex pattern for displayed env vars |
| `LOG_WEBHOOK_URL` | (disabled) | URL to POST log entries |
| `LOG_WEBHOOK_TOKEN` | (none) | Authorization header for log webhook |

## Server

### `PORT`

The port the HTTP server listens on.

```bash
PORT=3000 ./demo-app
```

**Default:** `8080`

## Database

### `DB_PATH`

Controls where BadgerDB stores data.

| Value | Behavior |
|-------|----------|
| `:memory:` | In-memory, data lost on restart (default) |
| `/path/to/dir` | Persistent storage in directory |

```bash
# Ephemeral (default) - good for demos
DB_PATH=":memory:" ./demo-app

# Persistent - data survives restarts
DB_PATH="/data/demo-app" ./demo-app
```

**Default:** `:memory:`

**Note:** When using persistent storage, BadgerDB creates multiple files in the specified directory. For containers, mount a volume to this path.

## Environment Display

### `ENV_FILTER`

Controls which environment variables appear in the System Info panel on the dashboard and in `/api/system` responses.

**When not set (default):** Uses a safe allowlist of common variables:
- `PORT`, `DB_PATH`
- `HOSTNAME`, `CONTAINER_ID` (Docker)
- `POD_NAME`, `POD_NAMESPACE`, `NODE_NAME` (Kubernetes)

**When set:** Uses the value as a case-insensitive regex pattern to match variable names against ALL environment variables.

```bash
# Show only vars starting with DEMO_
ENV_FILTER="^DEMO_" ./demo-app

# Show Kubernetes-related vars
ENV_FILTER="^(POD_|NODE_|KUBERNETES_)" ./demo-app

# Show vars containing "CONFIG" anywhere in name
ENV_FILTER="CONFIG" ./demo-app

# Show everything (use with caution!)
ENV_FILTER=".*" ./demo-app
```

**Security note:** When `ENV_FILTER` is set, you control what's exposed. Be careful not to expose sensitive variables like `AWS_SECRET_ACCESS_KEY`, database passwords, or API tokens.

**Invalid patterns:** If the regex is invalid, the app logs an error and returns an empty environment list (safe fallback).

## Log Shipping

Optional feature to POST log entries to an HTTP endpoint. Useful for shipping logs to Splunk HEC, Grafana Loki, or any webhook-compatible logging system.

### `LOG_WEBHOOK_URL`

URL to POST log entries to. Each log entry is sent as a JSON object.

```bash
LOG_WEBHOOK_URL="https://splunk.example.com:8088/services/collector" ./demo-app
```

**Default:** (disabled — logs only go to stdout)

### `LOG_WEBHOOK_TOKEN`

Optional authorization token. When set, adds an `Authorization` header to webhook requests.

```bash
LOG_WEBHOOK_URL="https://splunk.example.com:8088/services/collector" \
LOG_WEBHOOK_TOKEN="Splunk abc123" \
./demo-app
```

**Default:** (no Authorization header)

**Behavior notes:**
- Logs always go to stdout regardless of webhook configuration
- Webhook calls are asynchronous (don't block HTTP responses)
- Failed webhook calls are logged to stderr but don't affect the app
- No retry logic — webhook is best-effort

## Examples

### Local Development

```bash
# Defaults work fine
./demo-app
```

### Docker with Persistent Storage

```bash
docker run -p 8080:8080 \
  -v demo-data:/data \
  -e DB_PATH=/data \
  demo-app
```

### Kubernetes Demo

```bash
# Show K8s environment info on dashboard
ENV_FILTER="^(POD_|NODE_|KUBERNETES_)" ./demo-app
```

### Custom Demo Variables

```bash
# Set your own demo variables and filter to show them
DEMO_CUSTOMER="Acme Corp" \
DEMO_ENVIRONMENT="staging" \
DEMO_VERSION="2.1.0" \
ENV_FILTER="^DEMO_" \
./demo-app
```

### Shipping Logs to Splunk HEC

```bash
LOG_WEBHOOK_URL="https://splunk.example.com:8088/services/collector/event" \
LOG_WEBHOOK_TOKEN="Splunk your-hec-token" \
./demo-app
```
