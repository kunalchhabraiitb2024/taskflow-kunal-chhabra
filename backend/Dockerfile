# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install build tools (needed for CGO-less builds with some packages)
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Download dependencies first (cached layer unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /server ./cmd/server

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates: needed for HTTPS outbound calls if any
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary
COPY --from=builder /server .

# Copy migration files (server runs them on startup)
COPY migrations/ ./migrations/

# Copy seed file
COPY seed/ ./seed/

EXPOSE 8080

CMD ["./server"]
