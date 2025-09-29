# Build stage
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o staticsocket .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/staticsocket .

# Change ownership to non-root user
RUN chown appuser:appgroup /app/staticsocket

# Switch to non-root user
USER appuser

# Set default command
ENTRYPOINT ["./staticsocket"]
CMD ["--help"]

# Add labels
LABEL maintainer="yuval" \
      description="Static socket analysis tool for Go codebases" \
      version="latest"