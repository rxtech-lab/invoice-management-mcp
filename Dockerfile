# Multi-stage Dockerfile for invoice-management service
# Supports linux/amd64 and linux/arm64 architectures

# Build stage - compile Go binary
FROM golang:1.24-bookworm AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=docker
ARG COMMIT_HASH
ARG BUILD_TIME

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags "-X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildTime=${BUILD_TIME}" \
    -o invoice-management \
    ./cmd/server/main.go

# Final runtime stage
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS and wget for healthcheck
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -m appuser

# Set working directory
WORKDIR /app

# Create data directory for SQLite
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

# Copy the binary from build stage
COPY --from=builder /app/invoice-management ./

# Switch to non-root user
USER appuser

# Expose port (default 8080, configurable via PORT env var)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./invoice-management"]
