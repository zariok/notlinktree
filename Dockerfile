# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build
# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .
# Build with version info from build args
RUN CGO_ENABLED=0 go build -o notlinktree .

# Final stage
FROM alpine:latest

# Add non-root user for security
RUN addgroup -S notlinktree && \
    adduser -S notlinktree -G notlinktree

# Create and set permissions for config directory
RUN mkdir /config && \
    chown notlinktree:notlinktree /config

# Copy the binary from builder
COPY --from=builder /build/notlinktree /usr/local/bin/
# Copy a default config if none exists
COPY config.yaml /config/config.yaml.default

# Use non-root user
USER notlinktree

# Document the config volume
VOLUME /config

# Document the port
EXPOSE 8080

# Environment variables
ARG NLT_JWT_SECRET
ENV NLT_JWT_SECRET=${NLT_JWT_SECRET}
ENV NLT_DATA=/config

# Entrypoint script to handle config file
COPY <<EOF /usr/local/bin/docker-entrypoint.sh
#!/bin/sh

# Check if NLT_JWT_SECRET is set
if [ -z "$NLT_JWT_SECRET" ]; then
    echo "Error: NLT_JWT_SECRET environment variable is not set." >&2
    exit 1
fi

# If no config exists, copy the default
if [ ! -f /config/config.yaml ]; then
    cp /config/config.yaml.default /config/config.yaml
fi
exec notlinktree
EOF

# Make entrypoint executable
USER root
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
USER notlinktree

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"] 