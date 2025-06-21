# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy the entire project
COPY . .

# Create bin directory
RUN mkdir -p /app/bin

# Build the binary directly
WORKDIR /app/cmd/bridge
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/syncora .

# Runtime stage
FROM alpine:3.18

# Install ca-certificates for SSL
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Create bin directory in runtime stage
RUN mkdir -p /app/bin

# Copy binary from builder
COPY --from=builder /app/bin/syncora /app/bin/syncora

# Make sure binary is executable
RUN chmod +x /app/bin/syncora

# Set entrypoint
ENTRYPOINT ["/app/bin/syncora"]

# Default command
CMD ["help"]