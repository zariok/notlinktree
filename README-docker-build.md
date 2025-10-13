# Docker Deployment

This document provides instructions for deploying NotLinkTree using Docker.

## Using Docker Compose (recommended)

1. Clone the repository
2. Run with docker-compose:
```bash
docker-compose up -d
```

The `docker-compose.yml` is pre-configured to use the `NLT_JWT_SECRET` environment variable. You can set it in a `.env` file in the same directory as your `docker-compose.yml`:

**.env**
```
NLT_JWT_SECRET=your_super_secret_and_long_jwt_secret_here
```

Or you can pass it directly on the command line:
```bash
NLT_JWT_SECRET="your_secret" docker-compose up -d
```

The config file and avatars will be stored in a Docker volume. To use a local directory instead:

1. (Optional) Create a local config directory if you want files on the host:
```bash
mkdir -p config
```

2. Modify docker-compose.yml to use the local directory:
```yaml
    volumes:
      - ./config:/config
```

3. Run docker-compose:
```bash
docker-compose up -d
```

Notes:
- You do NOT need to pre-seed `config/config.yaml` or `config/avatar.png`. If they are missing, the container will copy defaults into `/config` on first start.

## Using Docker directly

Build the image:
```bash
docker build -t notlinktree .
```

If you've changed any UI (Next.js) code, build the UI first so static files are embedded:
```bash
make build-ui   # or: make build-embed
docker build -t notlinktree .
```

Run with a Docker volume and passing the secret:
```bash
docker run -d \
  -p 8080:8080 \
  -e NLT_JWT_SECRET="your_secret" \
  -v notlinktree_data:/config \
  --name notlinktree \
  notlinktree
```

Or run with a local directory:
```bash
docker run -d \
  -p 8080:8080 \
  -e NLT_JWT_SECRET="your_secret" \
  -v $(pwd)/config:/config \
  --name notlinktree \
  notlinktree
```

Podman equivalents:
```bash
podman build -t notlinktree .
podman run -d \
  -p 8080:8080 \
  -e NLT_JWT_SECRET="your_secret" \
  -v $(pwd)/config:/config \
  --name notlinktree \
  notlinktree
```

## Configuration

- All configuration files are stored in `/config` inside the container
- This includes:
  - `config.yaml` - Main configuration file
  - Any avatar or image files referenced in the config
- The directory is exposed as a volume for persistence
- If no config exists, a default `config.yaml` and `avatar.png` will be created on first run by copying defaults from inside the image (`/usr/share/notlinktree/*.default`).
- You can modify files in the mounted directory and they will be available to the container

## Environment Variables

- `NLT_JWT_SECRET` - **Required.** JWT secret for admin authentication
- `NLT_PORT` - HTTP port (default: 8080)
- `NLT_DATA` - Config directory (default: /config in container)

## See Also

For general setup and configuration, see [README.md](README.md).

