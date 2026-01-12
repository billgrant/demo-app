# syntax=docker/dockerfile:1

## -----------------------------------------------------
## Build stage: compile the Go binary
FROM dhi.io/golang:1.25-alpine3.22-dev AS build-stage

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY static/ ./static/

# Build for the target architecture (set by docker buildx)
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /demo-app

## -----------------------------------------------------
## Runtime stage: minimal hardened image
FROM dhi.io/static:20250911-alpine3.22 AS runtime-stage

WORKDIR /

COPY --from=build-stage /demo-app /demo-app

# Document the port (doesn't publish, just metadata)
EXPOSE 8080

# Health check using the /health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/demo-app", "healthcheck"]

ENTRYPOINT ["/demo-app"]
