# NotLinkTree - AI Agent Development Guide

## Project Overview

NotLinkTree is a self-contained LinkTree clone built with Go and Next.js/Tailwind CSS. It provides a beautiful landing page for social media links, an admin interface for management, and click tracking analytics. All configuration is stored in a simple YAML file.

### Key Features
- Modern, responsive UI built with Next.js and Tailwind CSS v4
- Secure admin interface for managing links (add, edit, delete)
- YAML-based configuration (`config.yaml`)
- Basic click tracking for analytics with rate limiting
- Single Go binary with embedded UI
- Avatar upload and cropping for profile photo
- Docker support for easy deployment
- Standardized API responses with error codes
- Configuration management interface
- Password change functionality
- Localhost-only config reload endpoint

## Architecture

### Backend (Go)
- **Entry Point**: `main.go` - Server setup, route wiring, and main application logic
- **Configuration**: `config.go` - YAML config management, data structures, and utilities
- **Authentication**: `auth.go` - JWT token validation and admin authentication
- **Rate Limiting**: `rate_limiter.go` - IP-based rate limiting for click tracking
- **Middleware**: `middleware.go` - Logging and CORS middleware
- **SPA Handler**: `spa.go` - Static file serving for embedded UI
- **API Handlers**: `handlers.go` - All HTTP API endpoints and business logic
- **Response Utilities**: `responses.go` - Standardized JSON response helpers and error codes

### Frontend (Next.js/React)
- **Main Page**: `ui/app/page.js` - Public landing page with links
- **Admin Interface**: `ui/app/admin/page.js` - Link management interface
- **Admin Components**: `ui/app/admin/` - Admin-specific components (AdminHeader, AdminFooter)
- **Configuration Page**: `ui/app/admin/config/page.js` - UI configuration management
- **Avatar Upload**: `ui/app/admin/upload-avatar/page.js` - Profile picture upload and cropping
- **Link Card**: `ui/app/LinkCard.js` - Reusable link display component
- **Styling**: Tailwind CSS v4 with custom configuration and PostCSS

### Data Storage
- **Configuration**: `config.yaml` - YAML file containing all settings, links, and UI configuration
- **Avatars**: `avatar.png` - Profile picture stored in data directory
- **Click Tracking**: In-memory with periodic persistence to config file

## Project Structure

```
notlinktree/
├── main.go                 # Application entry point and routing
├── config.go              # Configuration management and data structures
├── auth.go                # JWT authentication and validation
├── rate_limiter.go        # Rate limiting for click tracking
├── middleware.go          # HTTP middleware (logging, CORS)
├── spa.go                 # Static file serving for SPA
├── handlers.go            # API endpoint handlers
├── responses.go           # Standardized API response utilities
├── config.yaml            # Main configuration file
├── avatar.png             # Profile avatar image
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── Makefile               # Build automation
├── Dockerfile             # Container image definition
├── docker-compose.yml     # Container orchestration
├── LICENSE                # MIT License
├── README.md              # Main project documentation
├── README-docker.md       # Docker deployment guide
├── README-release-verification.md # Release verification guide
├── ui/                    # Next.js frontend application
│   ├── app/               # React components and pages
│   │   ├── admin/         # Admin interface components
│   │   │   ├── AdminFooter.js
│   │   │   ├── AdminHeader.js
│   │   │   ├── config/    # Configuration management page
│   │   │   ├── upload-avatar/ # Avatar upload functionality
│   │   │   └── page.js    # Main admin page
│   │   ├── LinkCard.js    # Link display component
│   │   ├── page.js        # Public landing page
│   │   └── layout.js      # Root layout component
│   ├── build.sh           # UI build script (legacy)
│   ├── next.config.js     # Next.js configuration
│   ├── package.json       # Node.js dependencies
│   ├── tailwind.config.js # Tailwind CSS configuration
│   ├── postcss.config.mjs # PostCSS configuration
│   └── out/               # Built static files (generated)
├── embed/                 # Embedded static files (generated)
│   └── ui/                # UI files copied from ui/out/
├── dist/                  # Built binaries (generated)
│   ├── notlinktree-linux-amd64
│   └── notlinktree-darwin-arm64
└── test/                  # Test files and configurations
    ├── data/              # Test data files
    └── docker/            # Docker test configurations
```

## Development Guidelines

### Code Organization

#### Go Backend
- **Single Package**: All Go code is in the `main` package for simplicity
- **Global State**: Uses global variables for config, JWT secret, and rate limiter
- **Concurrency**: Uses `sync.RWMutex` for thread-safe config access
- **Error Handling**: Consistent error responses with appropriate HTTP status codes

#### React Frontend
- **Client-Side**: All React components use `'use client'` directive
- **State Management**: Local state with `useState` and `useEffect`
- **API Communication**: Direct fetch calls to backend endpoints
- **Styling**: Tailwind CSS v4 with responsive design patterns
- **Dependencies**: 
  - Next.js (latest)
  - React (latest)
  - Tailwind CSS v4
  - FontAwesome icons
  - Headless UI components
  - Heroicons
  - React Easy Crop for avatar cropping

### API Endpoints

#### Public Endpoints
- `GET /api/config` - Get UI configuration
- `GET /api/links` - Get all public links
- `POST /api/click/{id}` - Track link click (rate limited)
- `GET /api/avatar` - Get profile avatar
- `POST /api/refresh-config` - Reload config from disk (localhost only)

#### Admin Endpoints (Require JWT)
- `POST /api/admin/login` - Admin authentication
- `GET /api/admin/config` - Get full configuration with links
- `POST /api/admin/config` - Update UI configuration
- `POST /api/admin/links` - Add new link
- `PUT /api/admin/links/{id}` - Update existing link
- `DELETE /api/admin/links/{id}` - Delete link
- `POST /api/admin/refresh-config` - Reload config from disk (localhost only)
- `GET /api/admin/avatar` - Get admin avatar
- `POST /api/admin/avatar` - Upload avatar (requires authentication)
- `POST /api/admin/password` - Change admin password

### Configuration Schema

```yaml
admin:
  password: <bcrypt-hashed-password>
links:
  <link-id>:
    id: <unique-id>
    title: <link-title>
    url: <link-url>
    description: <optional-description>
    type: <link-type>
    clicks: <click-count>
ui:
  username: <profile-name>
  title: <page-title>
  primaryColor: <hex-color>
  secondaryColor: <hex-color>
  backgroundColor: <hex-color>
```

### Environment Variables

- `NLT_PORT` - Server port (default: 8080)
- `NLT_DATA` - Config directory (default: current directory)
- `NLT_JWT_SECRET` - JWT signing secret (required)

### Build Process

1. **UI Build**: Next.js application builds to static files in `ui/out/`
2. **Embed**: Static files copied to `embed/ui/` directory
3. **Go Build**: Go binary built with embedded files and version information
4. **Cross-Platform**: Supports Linux AMD64 and macOS ARM64
5. **Versioning**: Automatic version detection from git tags and commit hash
6. **Dependencies**: Automatic dependency installation for both Go and Node.js

#### Available Make Targets

**Build Targets:**
- `make build` - Build for all platforms (default, includes UI and Go builds)
- `make build-linux` - Build Linux AMD64 binary only
- `make build-darwin` - Build macOS ARM64 binary only
- `make build-ui` - Build UI only (copies to embed/)
- `make build-embed` - Clean embed directory and build UI
- `make build-go` - Build Go binaries for all platforms (requires UI already built)

**Development Targets:**
- `make dev-ui` - Start UI development server
- `make deps` - Install all dependencies (Go and Node.js)

**Test Targets:**
- `make test` - Quick test (Go unit tests only)
- `make test-go` - Run Go unit tests with race detection
- `make test-ui` - Run UI tests (linting and build)
- `make test-all` - Run all tests (Go and UI)
- `make test-coverage` - Run Go tests with coverage report

**Container Targets:**
- `make docker-test` - Run full container test suite
- `make docker-lint` - Lint Dockerfile using hadolint
- `make docker-clean` - Clean container test artifacts
- `make runtime-info` - Show detected container runtime (Docker/Podman)

**Utility Targets:**
- `make clean` - Clean all build artifacts
- `make clean-embed` - Clean only embed directory
- `make all` - Alias for `make build`

**Container Runtime Support:**
- **Auto-Detection**: Automatically detects Docker or Podman
- **Compose**: Uses `docker-compose` or `podman-compose` as appropriate
- **Cross-Platform**: Supports both container runtimes seamlessly

## CI/CD Pipeline

### GitHub Actions Workflows

The project uses GitHub Actions for automated testing, building, and releasing. The main workflow is triggered on git tag pushes (e.g., `v1.0.0`).

#### Release Workflow (`.github/workflows/release.yml`)

**Trigger:** Push to tags matching `v*` pattern

**Jobs:**

1. **Test Job** (`test`)
   - **Platform**: Ubuntu Latest
   - **Go Version**: 1.22 (with caching)
   - **Node.js Version**: 20 (with npm caching)
   - **Steps**:
     - Checkout code
     - Install Go and Node.js dependencies
     - Run Go unit tests with race detection and coverage
     - Generate HTML coverage report
     - Upload coverage artifacts
     - Run UI linting (`npm run lint`)
     - Test UI build (`npm run build`)
     - Run container tests using Makefile
     - Clean up container test artifacts

2. **Build and Release Job** (`build-release`)
   - **Dependencies**: Requires `test` job to pass
   - **Permissions**: `contents: write`, `packages: write`
   - **Steps**:
     - Checkout code and set up Go/Node.js
     - Install dependencies using `make deps`
     - Build binaries using `make build`
     - Test built binaries (help command and architecture verification)
     - Import GPG key for signing
     - Generate and sign SHA256 checksums
     - Create GitHub release with generated release notes
     - Upload all binaries and checksums as release assets

3. **Container Release Job** (`container-release`)
   - **Dependencies**: Requires `test` job to pass
   - **Platforms**: Linux AMD64, Linux ARM64
   - **Registry**: GitHub Container Registry (ghcr.io)
   - **Steps**:
     - Checkout code
     - Set up Docker Buildx for multi-platform builds
     - Login to GitHub Container Registry
     - Extract metadata and generate tags
     - Build and push multi-platform Docker image
     - Use GitHub Actions cache for build optimization

#### Dependabot Configuration (`.github/dependabot.yml`)

- **Package Ecosystem**: GitHub Actions
- **Schedule**: Weekly updates
- **Pull Request Limit**: 10 open PRs maximum
- **Scope**: Root directory

### Release Process

1. **Tag Creation**: Create and push a git tag (e.g., `git tag v1.0.0 && git push origin v1.0.0`)
2. **Automated Testing**: GitHub Actions runs comprehensive test suite
3. **Binary Building**: Cross-platform binaries built for Linux AMD64 and macOS ARM64
4. **Container Building**: Multi-platform Docker images built and pushed to GHCR
5. **Release Creation**: GitHub release created with:
   - Generated release notes
   - Binary assets for both platforms
   - SHA256 checksums (signed with GPG)
   - Container images tagged with version and `latest`

### Security Features

- **GPG Signing**: All checksums are signed with GPG for verification
- **Multi-Platform**: Binaries and containers built for multiple architectures
- **Dependency Updates**: Automated dependency updates via Dependabot
- **Test Coverage**: Comprehensive test coverage with HTML reports
- **Container Testing**: Full container test suite before release

### Testing

#### Test Types

**Unit Tests**: Go test files for individual components (`*_test.go`)
- `main_test.go` - Main application tests
- `config_test.go` - Configuration management tests
- `auth_test.go` - Authentication tests
- `handlers_test.go` - API handler tests
- `middleware_test.go` - Middleware tests
- `rate_limiter_test.go` - Rate limiting tests
- `spa_test.go` - Static file serving tests

**Integration Tests**: Docker-based testing with test containers
- Full container lifecycle testing
- Multi-container orchestration testing
- Volume mounting and persistence testing
- Network connectivity testing

**UI Tests**: Frontend testing and validation
- Linting with ESLint
- Build verification
- Component functionality testing

**Manual Testing**: Admin interface and public page functionality
- User interface validation
- Cross-browser compatibility
- Responsive design testing

#### Test Execution

**Local Testing:**
```bash
# Quick Go tests only
make test

# All Go tests with race detection
make test-go

# Go tests with coverage report
make test-coverage

# UI tests (linting and build)
make test-ui

# All tests (Go + UI)
make test-all

# Container tests
make docker-test
```

**GitHub Actions Testing:**
- **Automatic**: Runs on every tag push
- **Go Tests**: Race detection, coverage reporting, HTML coverage generation
- **UI Tests**: Linting and build verification
- **Container Tests**: Full Docker/Podman test suite
- **Coverage**: HTML coverage reports uploaded as artifacts
- **Parallel Execution**: Tests run in parallel for faster execution

#### Test Coverage

- **Go Coverage**: Generated with `go test -coverprofile=coverage.out`
- **Coverage Report**: HTML format generated with `go tool cover -html=coverage.out`
- **Coverage Mode**: Count mode for accurate branch coverage
- **Artifact Upload**: Coverage reports uploaded to GitHub Actions artifacts
- **Threshold**: No specific coverage threshold enforced (monitoring only)

#### Test Data

- **Location**: `test/data/` directory
- **Files**:
  - `avatar_test.png` - Test avatar image
  - `config_invalid.yaml` - Invalid configuration for error testing
  - `index.html` - Test HTML file
- **Docker Test Data**: `test/docker/test_config/` for container testing

#### Container Testing

- **Runtime Detection**: Automatically uses Docker or Podman
- **Test Containers**: Uses docker-compose for multi-container testing
- **Cleanup**: Automatic cleanup of test artifacts
- **Linting**: Dockerfile linting with hadolint
- **Multi-Platform**: Tests both Linux AMD64 and ARM64 architectures

## Common Development Tasks

### Adding New Link Types

1. Update `LINK_TYPES` array in `ui/app/admin/page.js`
2. Add URL examples to `LINK_TYPE_URL_EXAMPLES`
3. Add title examples to `LINK_TYPE_TITLE_EXAMPLES`

### Modifying API Endpoints

1. Add route in `main.go` mux configuration
2. Implement handler function in `handlers.go`
3. Update frontend to call new endpoint
4. Add appropriate error handling and validation

### Changing UI Styling

1. Modify Tailwind classes in React components
2. Update `tailwind.config.js` for custom configurations
3. Test responsive design across different screen sizes

### Database/Storage Changes

1. Update `Config` struct in `config.go`
2. Modify `loadConfig` and `saveConfig` functions
3. Update frontend to handle new data structure
4. Consider migration strategy for existing configs

## Security Considerations

- **JWT Tokens**: 24-hour expiration for admin sessions
- **Password Hashing**: bcrypt with default cost
- **Rate Limiting**: 10 clicks per hour per IP address
- **Input Validation**: URL validation and sanitization
- **File Upload**: 1MB limit for avatar uploads
- **CORS**: Simple allow-all policy (consider restricting in production)

## Performance Considerations

- **Click Tracking**: In-memory with periodic persistence
- **Static Assets**: 1-year cache for embedded files
- **Config Access**: Read-write mutex for thread safety
- **Background Tasks**: Goroutine for click count flushing

## Deployment

### Local Development
```bash
# Install dependencies
go mod download
cd ui && npm install && cd ..

# Build and run
make build
./dist/notlinktree-linux-amd64
```

### Docker Deployment

#### Local Development
```bash
# Using docker-compose (recommended)
NLT_JWT_SECRET="your-secret" docker-compose up -d

# Using Docker directly
docker build -t notlinktree .
docker run -d -p 8080:8080 -e NLT_JWT_SECRET="secret" notlinktree

# Container testing
make docker-test  # Run full test suite in containers
make docker-lint  # Lint Dockerfile
make docker-clean # Clean test artifacts
```

#### GitHub Container Registry (GHCR)

**Pre-built Images:**
- **Registry**: `ghcr.io/[username]/notlinktree`
- **Tags**: 
  - `latest` - Latest stable release
  - `v1.0.0` - Specific version tags
  - `main-abc1234` - Branch-based tags with commit hash
- **Platforms**: Linux AMD64, Linux ARM64
- **Auto-updates**: Images automatically built and pushed on every release

**Using Pre-built Images:**
```bash
# Pull latest image
docker pull ghcr.io/[username]/notlinktree:latest

# Run with environment variables
docker run -d -p 8080:8080 \
  -e NLT_JWT_SECRET="your-secret" \
  -e NLT_DATA="/data" \
  -v $(pwd)/data:/data \
  ghcr.io/[username]/notlinktree:latest

# Run specific version
docker run -d -p 8080:8080 \
  -e NLT_JWT_SECRET="your-secret" \
  ghcr.io/[username]/notlinktree:v1.0.0
```

**Multi-Platform Support:**
- **Architectures**: AMD64, ARM64
- **Build Process**: Uses Docker Buildx for multi-platform builds
- **Cache Optimization**: GitHub Actions cache for faster builds
- **Automatic Detection**: Docker automatically selects correct architecture

### Container Runtime Support
- **Docker**: Primary container runtime
- **Podman**: Alternative container runtime (auto-detected)
- **Compose**: Uses `docker-compose` or `podman-compose` as appropriate
- **Multi-Platform**: Supports both AMD64 and ARM64 architectures
- **Registry**: GitHub Container Registry (GHCR) for pre-built images

## Troubleshooting

### Common Issues

1. **JWT Secret Not Set**: Ensure `NLT_JWT_SECRET` environment variable is set
2. **Config File Not Found**: Check `NLT_DATA` directory and file permissions
3. **UI Not Loading**: Verify `embed/ui/` directory has built files
4. **Admin Login Fails**: Check password in `config.yaml` or use generated password
5. **Click Tracking Not Working**: Verify rate limiter configuration

### Debug Steps

1. Check server logs for error messages
2. Verify environment variables are set correctly
3. Test API endpoints directly with curl
4. Check file permissions on config directory
5. Verify JWT token validity in browser dev tools

## Contributing

### Code Style
- **Go**: Follow standard Go formatting with `gofmt`
- **JavaScript**: Use Prettier for consistent formatting
- **Comments**: Document public functions and complex logic
- **Error Handling**: Always handle errors appropriately

### Testing
- Write tests for new functionality
- Test both success and error cases
- Verify admin and public interfaces work correctly
- Test with different configuration scenarios

### Pull Requests
- Include description of changes
- Test on both Linux and macOS if possible
- Verify Docker build still works
- Update documentation if needed

## Future Enhancements

### Potential Features
- Database backend (PostgreSQL, SQLite)
- User authentication system
- Advanced analytics and reporting
- Custom themes and layouts
- API rate limiting improvements
- Multi-language support
- Link preview generation
- Social media integration

### Technical Improvements
- Structured logging
- Configuration validation
- Health check endpoints
- Metrics and monitoring
- Automated testing pipeline
- Performance optimization
- Security hardening

---

This guide should help AI agents understand the codebase structure, development patterns, and common tasks when working with NotLinkTree. Always refer to the actual code for the most up-to-date implementation details.
