# NotLinkTree

A self-contained LinkTree clone built with Go and a Next.js/Tailwind CSS frontend. 

## Features

- Secure admin interface for managing links (add, edit, delete)
- YAML-based configuration (`config.yaml`)
- Basic click tracking for analytics w/rate limiting
- Single Go binary with embedded UI


## Prerequisites

- Go 1.22 or later
- Node.js 18 or later
- npm (or yarn)

## Setup & Build

### 1. Install dependencies
```bash
go mod download
cd ui && npm install && cd ..
```

### 2. Build everything (UI, embed, Go binary)
Use the makefile to:
- Build the UI
- Embed the static files
- Build the Go binary for (`build-linux`) Linux x86_64 and macOS ARM64 (`build-darwin`)
- Binaries will be output to the `dist/` directory.

```bash
make build
```

## Running

### Environment Variables

- `NLT_PORT`: Set the port for the server (default: `8080`).
- `NLT_DATA`: Set the directory where `config.yaml` is stored (default: current directory `.`).
- `NLT_JWT_SECRET`: **Required.** A long, random string used to sign and verify admin JWTs.  Keeping this the same between restarts will keep admins logged in.  Changing it will invalidate all logins.

**Example:**
```bash
export NLT_JWT_SECRET=$(openssl rand -base64 32)
./notlinktree
```

Start the server:
```bash
./notlinktree
```
The app will be available at [http://localhost:8080](http://localhost:8080) by default.
- The main landing page is at `/`
- The admin interface is at `/admin`

### Admin password

The admin password is printed to the console during first run.  If you ever forget it, you can run the following:

```bash
# Set new password (automatically reloads running instance)
./notlinktree -setadminpw newpassword
```

**Note:** The `-setadminpw` command will automatically attempt to reload the configuration in any running instance of the application. If no instance is running or the reload fails, you'll see a warning message.

## Docker Deployment

For Docker deployment instructions, see [README-docker.md](README-docker.md).

## Verifying Releases

To verify the authenticity and integrity of official releases, see [README-release-verification.md](README-release-verification.md).

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
