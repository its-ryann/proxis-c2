# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /proxis-server ./cmd/server/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S proxis && \
    adduser -u 1001 -S proxis -G proxis

WORKDIR /app

# Copy binary from builder
COPY --from=builder /proxis-server /app/proxis-server

# Copy configuration
COPY configs/server.yaml /app/configs/server.yaml

# Create directories for certificates and logs
RUN mkdir -p /etc/proxis/certs /var/log/proxis

# Set ownership
RUN chown -R proxis:proxis /app /etc/proxis /var/log/proxis

# Switch to non-root user
USER proxis

# Expose port
EXPOSE 443

# Run the server
CMD ["/app/proxis-server"]