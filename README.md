# Niotebook

A standalone TUI social media platform built in Go. Monorepo producing two binaries: `niotebook-server` (REST API) and `niotebook-tui` (terminal client).

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- golangci-lint (optional, for linting)

### Setup

```bash
# Clone and install dependencies
git clone https://github.com/Akram012388/niotebook-tui.git
cd niotebook-tui
go mod download

# Create database and run migrations
createdb niotebook_dev
cp .env.example .env  # Edit with your database URL and JWT secret
make migrate-up

# Run server
make dev

# Run TUI (in another terminal)
make dev-tui
```

### Build

```bash
make build          # Build both binaries to bin/
make test           # Run all tests with race detector
make lint           # Run golangci-lint
make test-cover     # Tests with coverage report
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `NIOTEBOOK_DB_URL` | Yes | PostgreSQL connection string |
| `NIOTEBOOK_JWT_SECRET` | Yes | JWT signing key (min 32 bytes) |
| `NIOTEBOOK_PORT` | No | Server port (default: 8080) |
| `NIOTEBOOK_HOST` | No | Server host (default: localhost) |
| `NIOTEBOOK_CORS_ORIGIN` | No | Allowed CORS origin |
| `NIOTEBOOK_LOG_LEVEL` | No | Log level: info, debug |

## Documentation

Full documentation is in `docs/vault/`. Start at `docs/vault/00-home/index.md`.

## Architecture

- **Server:** Three-layer architecture (handler -> service -> store) with JWT auth
- **TUI:** Bubble Tea Elm architecture (Model-Update-View) with async HTTP via tea.Cmd
- **Database:** PostgreSQL with golang-migrate sequential migrations

## License

MIT
