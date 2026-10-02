# ==========================================
# Stage 1: Build stage
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install git and ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Build the binary statically
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/bin/server \
    ./cmd/server

# ==========================================
# Stage 2: Production minimal runtime
# ==========================================
FROM alpine:3.21

# Install security certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create a non-privileged system group and user
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server /app/server
COPY --from=builder /build/.env.example /app/.env.example

# Set correct ownership
RUN chown -R appuser:appgroup /app

# Run as non-root user
USER appuser

EXPOSE 8080

CMD ["/app/server"]
