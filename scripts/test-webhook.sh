#!/bin/bash
#
# Test the log webhook functionality
#
# This script:
#   1. Builds and starts the webhook receiver (Go)
#   2. Starts demo-app with LOG_WEBHOOK_URL pointing to the receiver
#   3. Makes requests to generate log entries
#   4. Shows the webhook output
#
# Usage: ./scripts/test-webhook.sh
#
# Prerequisites:
#   - Go installed
#   - Ports 8080 and 9999 available

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# Cleanup function to kill background processes
cleanup() {
    echo ""
    echo "Cleaning up..."
    [ -n "$RECEIVER_PID" ] && kill $RECEIVER_PID 2>/dev/null || true
    [ -n "$APP_PID" ] && kill $APP_PID 2>/dev/null || true
}
trap cleanup EXIT

# Build demo-app if needed
if [ ! -f "./demo-app" ]; then
    echo "Building demo-app..."
    go build -o demo-app .
fi

echo "=== Starting webhook receiver on port 9999 ==="
go run scripts/webhook-receiver/main.go -port 9999 &
RECEIVER_PID=$!
sleep 1

echo "=== Starting demo-app with webhook enabled ==="
LOG_WEBHOOK_URL="http://localhost:9999/logs" ./demo-app 2>&1 &
APP_PID=$!
sleep 1

echo ""
echo "=== Making requests to generate logs ==="
echo ""

echo "GET /health"
curl -s http://localhost:8080/health
echo ""
sleep 0.5

echo ""
echo "POST /api/items"
curl -s -X POST http://localhost:8080/api/items \
    -H "Content-Type: application/json" \
    -d '{"name":"Test Item","description":"Created by test script"}'
echo ""
sleep 0.5

echo ""
echo "GET /api/items"
curl -s http://localhost:8080/api/items
echo ""

# Give webhooks time to complete
sleep 1

echo ""
echo "=== Test complete (check webhook output above) ==="
