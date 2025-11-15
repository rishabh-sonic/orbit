# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install build dependencies (git for VCS info, ca-certs for HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache dependencies separately from source code
COPY go.mod go.sum ./
RUN go mod download

# Copy full source and build both binaries
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/api  ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/ws   ./cmd/ws

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.20

# Needed for HTTPS outbound calls (Resend email, OAuth userinfo, etc.)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /bin/api ./api
COPY --from=builder /bin/ws  ./ws

# Both binaries listen on different ports; the compose file sets CMD per service.
# Default: run the API server.
CMD ["./api"]
