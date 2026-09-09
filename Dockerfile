# syntax=docker/dockerfile:1

# --- Build stage -------------------------------------------------------
# Pinned to the exact Go version this project requires (see go.mod: go 1.26.0).
FROM golang:1.26 AS builder

WORKDIR /src

# Copy module files first so dependency resolution is cached
# independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy application source.
COPY . .

# Build the main CacheDB server as a fully static Linux binary.
# CacheDB has no CGO/runtime dependencies.
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/cachedb ./app

# Build the dedicated healthcheck binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/healthcheck ./app/healthcheck

# Create the persistent-data directory that will be copied into
# the runtime image with ownership suitable for the non-root user.
RUN mkdir -p /out/data

# --- Runtime stage -----------------------------------------------------
# Distroless static image:
# - no shell
# - no package manager
# - no unnecessary OS utilities
# - runs as non-root via the :nonroot tag
FROM gcr.io/distroless/static-debian12:nonroot

# Copy the application binaries.
COPY --from=builder /out/cachedb /cachedb
COPY --from=builder /out/healthcheck /healthcheck

# IMPORTANT:
# The distroless :nonroot image runs as UID 65532.
#
# The /data directory therefore needs to be writable by UID 65532.
# When Docker creates a fresh named volume at /data, Docker populates
# the empty volume from this directory (unless volume copy-up is
# explicitly disabled), preserving its ownership.
COPY --from=builder --chown=65532:65532 /out/data /data

# Persistent CacheDB data directory.
VOLUME ["/data"]

EXPOSE 6379

# Perform a real RESP PING against CacheDB.
# This is a binary healthcheck because the Distroless image has no shell
# or redis-cli available.
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck", "127.0.0.1:6379"]

ENTRYPOINT ["/cachedb"]

CMD ["--port", "6379", "--dir", "/data", "--appendonly", "yes"]