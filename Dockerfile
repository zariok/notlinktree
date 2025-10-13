# Build stage
FROM golang:1.25.2-alpine AS builder

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

# Add labels for proper image identification
LABEL org.opencontainers.image.title="NotLinkTree"
LABEL org.opencontainers.image.description="A self-contained LinkTree clone built with Go and Next.js"
LABEL org.opencontainers.image.vendor="NotLinkTree"
LABEL org.opencontainers.image.source="https://github.com/zariok/notlinktree"
LABEL org.opencontainers.image.licenses="MIT"

# Add non-root user for security
RUN addgroup -S notlinktree && \
    adduser -S notlinktree -G notlinktree

# Create and set permissions for config directory
RUN mkdir /config && \
    chown notlinktree:notlinktree /config

# Copy the binary from builder
COPY --from=builder /build/notlinktree /usr/local/bin/
# Copy default files to a non-mounted location so they aren't masked by /config volume
RUN mkdir -p /usr/share/notlinktree
COPY config.yaml /usr/share/notlinktree/config.yaml.default
COPY avatar.png /usr/share/notlinktree/avatar.png.default

# Use non-root user
USER notlinktree

# Document the config volume
VOLUME /config

# Document the port
EXPOSE 8080

# Environment variables
ENV NLT_DATA=/config

# Create entrypoint script
USER root
COPY <<'EOF' /usr/local/bin/docker-entrypoint.sh
#!/bin/sh

# Check if NLT_JWT_SECRET is set
if [ -z "$NLT_JWT_SECRET" ]; then
    echo "Error: NLT_JWT_SECRET environment variable is not set." >&2
    exit 1
fi

# Ensure config directory exists
mkdir -p /config

# If no config exists in mounted dir, copy defaults from image
if [ ! -f /config/config.yaml ]; then
    cp /usr/share/notlinktree/config.yaml.default /config/config.yaml
fi

# If no avatar exists, copy the default from image
if [ ! -f /config/avatar.png ]; then
    cp /usr/share/notlinktree/avatar.png.default /config/avatar.png
fi

# Pass all arguments to notlinktree
exec notlinktree "$@"
EOF

RUN chmod +x /usr/local/bin/docker-entrypoint.sh
USER notlinktree

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"] 