---
title: "Build, Distribution & Dev Workflow"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, build, development, workflow]
---

# Build, Distribution & Dev Workflow

## Project Setup

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Compilation, module management |
| PostgreSQL | 15+ | Database |
| golang-migrate | latest | Database migrations |
| Make | any | Build automation |

### Initial Setup

```bash
# Clone the repo
git clone https://github.com/CodeAkram/niotebook.git
cd niotebook

# Install Go dependencies
go mod download

# Start PostgreSQL (macOS with Homebrew)
brew services start postgresql@15

# Create the database
createdb niotebook_dev

# Run migrations
migrate -path migrations -database "postgres://localhost/niotebook_dev?sslmode=disable" up

# Copy example config
cp .env.example .env
# Edit .env with your JWT secret and database URL
```

### Environment Variables

| Variable | Required | Example | Description |
|----------|----------|---------|-------------|
| `NIOTEBOOK_DB_URL` | Yes | `postgres://localhost/niotebook_dev?sslmode=disable` | PostgreSQL connection string |
| `NIOTEBOOK_JWT_SECRET` | Yes | `your-256-bit-secret-here` | JWT signing key |
| `NIOTEBOOK_PORT` | No | `8080` | Server listen port (default: 8080) |
| `NIOTEBOOK_HOST` | No | `0.0.0.0` | Server listen host (default: localhost) |
| `NIOTEBOOK_ACCESS_TOKEN_TTL` | No | `24h` | Access token lifetime |
| `NIOTEBOOK_REFRESH_TOKEN_TTL` | No | `168h` | Refresh token lifetime |
| `NIOTEBOOK_LOG_LEVEL` | No | `info` | Log level: debug, info, warn, error |

The `.env` file is loaded by the server binary at startup. It is gitignored.

### .env.example

```env
NIOTEBOOK_DB_URL=postgres://localhost/niotebook_dev?sslmode=disable
NIOTEBOOK_JWT_SECRET=change-me-to-a-secure-random-string-at-least-32-bytes
NIOTEBOOK_PORT=8080
NIOTEBOOK_HOST=localhost
NIOTEBOOK_LOG_LEVEL=debug
```

## Makefile

```makefile
.PHONY: build server tui test lint migrate-up migrate-down clean dev

# Build both binaries
build: server tui

# Build server binary
server:
	go build -o bin/niotebook-server ./cmd/server

# Build TUI binary
tui:
	go build -o bin/niotebook-tui ./cmd/tui

# Run server in development mode (with auto-reload via air if installed)
dev:
	go run ./cmd/server

# Run TUI pointing at local server
dev-tui:
	go run ./cmd/tui --server http://localhost:8080

# Run all tests
test:
	go test ./... -v -race

# Run tests with coverage
test-cover:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Run database migrations up
migrate-up:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" up

# Rollback last database migration
migrate-down:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" down 1

# Create a new migration file pair
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Cross-compile for release
release:
	GOOS=linux GOARCH=amd64 go build -o bin/niotebook-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -o bin/niotebook-server-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o bin/niotebook-tui-darwin-amd64 ./cmd/tui
	GOOS=darwin GOARCH=arm64 go build -o bin/niotebook-tui-darwin-arm64 ./cmd/tui
	GOOS=linux GOARCH=amd64 go build -o bin/niotebook-tui-linux-amd64 ./cmd/tui
	GOOS=linux GOARCH=arm64 go build -o bin/niotebook-tui-linux-arm64 ./cmd/tui
```

## Development Workflow

### Daily Development

```
Terminal 1: make dev              # runs the server on :8080
Terminal 2: make dev-tui          # runs the TUI pointing at localhost:8080
Terminal 3: (editor/IDE)
```

Changes to Go files require a restart. For faster iteration:
- Install [air](https://github.com/cosmtrek/air) for live reload of the server: `air` instead of `make dev`
- TUI requires manual restart (Bubble Tea's event loop doesn't support hot reload)

### Database Changes

```bash
# Create a new migration
make migrate-create name=add_followers_table

# Apply migration
make migrate-up

# Rollback if something went wrong
make migrate-down
```

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/server/service/... -v

# With coverage report
make test-cover
# Open coverage.html in browser
```

## CLI Flags

### Server (`niotebook-server`)

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | Listen port |
| `--host` | `localhost` | Listen host |
| `--migrate` | `false` | Run pending migrations on startup |
| `--env` | `.env` | Path to .env file |

### TUI (`niotebook-tui`)

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | `https://api.niotebook.com` | Server URL |
| `--config` | `~/.config/niotebook/config.yaml` | Config file path |
| `--version` | — | Print version and exit |

Flag parsing via Go's standard `flag` package (no Cobra for MVP — single command, no subcommands needed).

## Distribution

### go install

```bash
go install github.com/CodeAkram/niotebook/cmd/tui@latest
```

Only the TUI is distributed via `go install`. The server is deployed by the operator, not installed by end users.

### GitHub Releases

Automated via GitHub Actions. On git tag push (`v0.1.0`):

1. CI runs tests
2. Cross-compiles TUI for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
3. Creates GitHub Release with binaries attached
4. Users download and run: `chmod +x niotebook-tui-darwin-arm64 && ./niotebook-tui-darwin-arm64`

### Homebrew (Post-MVP)

```bash
brew install niotebook
```

Requires maintaining a Homebrew formula. Deferred to after initial launch.

## Git Conventions

### Branch Strategy

- `main` — stable, always deployable
- `dev` — integration branch for features
- `feature/*` — feature branches off `dev`
- `fix/*` — bug fix branches off `dev`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add timeline view with post card rendering
fix: handle expired JWT in TUI client
docs: add ADR for cursor-based pagination
test: add integration tests for auth endpoints
refactor: extract post validation into service layer
chore: update Go dependencies
```

## .gitignore

```gitignore
# Binaries
bin/
*.exe

# Environment
.env
!.env.example

# Coverage
coverage.out
coverage.html

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Go
vendor/

# Niotebook config (local dev)
auth.json
```

## Production Deployment

### On the VPS

```bash
# Upload binary
scp bin/niotebook-server-linux-amd64 user@server:/opt/niotebook/niotebook-server

# Create systemd service: /etc/systemd/system/niotebook.service
[Unit]
Description=Niotebook API Server
After=network.target postgresql.service

[Service]
Type=simple
User=niotebook
WorkingDirectory=/opt/niotebook
ExecStart=/opt/niotebook/niotebook-server
EnvironmentFile=/opt/niotebook/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target

# Enable and start
systemctl enable niotebook
systemctl start niotebook
```

### Caddy Reverse Proxy

```
api.niotebook.com {
    reverse_proxy localhost:8080
}
```

Caddy automatically provisions and renews TLS certificates via Let's Encrypt.

### Database Backups

```bash
# Cron job: daily at 2am
0 2 * * * pg_dump niotebook | gzip > /backups/niotebook-$(date +\%Y\%m\%d).sql.gz

# Retain 30 days of backups
0 3 * * * find /backups -name "niotebook-*.sql.gz" -mtime +30 -delete
```
