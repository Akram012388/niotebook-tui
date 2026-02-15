#!/usr/bin/env bash
set -euo pipefail

# Niotebook MVP Sprint — Bootstrap Dependencies
# Idempotent: safe to re-run. Installs system-level tools only.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[bootstrap]${NC} $*"; }
warn() { echo -e "${YELLOW}[bootstrap]${NC} $*"; }
err()  { echo -e "${RED}[bootstrap]${NC} $*" >&2; }
die()  { err "$@"; exit 1; }

# ── Go ──────────────────────────────────────────────────────────────
check_go() {
  log "Checking Go..."
  command -v go >/dev/null 2>&1 || die "Go is not installed. Install Go 1.22+ first."
  local ver
  ver=$(go version | grep -oE '[0-9]+\.[0-9]+' | head -1)
  local major minor
  major=$(echo "$ver" | cut -d. -f1)
  minor=$(echo "$ver" | cut -d. -f2)
  if (( major < 1 || (major == 1 && minor < 22) )); then
    die "Go $ver found, need >= 1.22"
  fi
  log "Go $(go version | grep -oE 'go[0-9]+\.[^ ]+') ✓"
}

# ── Homebrew ────────────────────────────────────────────────────────
check_brew() {
  command -v brew >/dev/null 2>&1 || die "Homebrew is not installed. Install from https://brew.sh"
  log "Homebrew ✓"
}

# ── PostgreSQL ──────────────────────────────────────────────────────
install_postgres() {
  log "Checking PostgreSQL..."
  if command -v psql >/dev/null 2>&1; then
    log "PostgreSQL already installed ✓"
  else
    log "Installing PostgreSQL 15 via Homebrew..."
    brew install postgresql@15
    # Add to PATH for this session
    export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
  fi

  # Start service if not running
  if ! pg_isready -q 2>/dev/null; then
    log "Starting PostgreSQL service..."
    brew services start postgresql@15 2>/dev/null || true
    # Wait for PostgreSQL to be ready
    local retries=10
    while ! pg_isready -q 2>/dev/null && (( retries-- > 0 )); do
      sleep 1
    done
    pg_isready -q 2>/dev/null || die "PostgreSQL failed to start"
  fi
  log "PostgreSQL running ✓"

  # Create databases if they don't exist
  local user
  user=$(whoami)
  if ! psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_dev; then
    log "Creating niotebook_dev database..."
    createdb niotebook_dev 2>/dev/null || true
  fi
  if ! psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_test; then
    log "Creating niotebook_test database..."
    createdb niotebook_test 2>/dev/null || true
  fi
  log "Databases niotebook_dev + niotebook_test ✓"
}

# ── golang-migrate CLI ──────────────────────────────────────────────
install_migrate() {
  log "Checking golang-migrate..."
  if command -v migrate >/dev/null 2>&1; then
    log "golang-migrate already installed ✓"
  else
    log "Installing golang-migrate via Homebrew..."
    brew install golang-migrate
  fi
  log "golang-migrate ✓"
}

# ── golangci-lint ───────────────────────────────────────────────────
install_golangci_lint() {
  log "Checking golangci-lint..."
  if command -v golangci-lint >/dev/null 2>&1; then
    log "golangci-lint already installed ✓"
  else
    log "Installing golangci-lint via Homebrew..."
    brew install golangci-lint
  fi
  log "golangci-lint ✓"
}

# ── jq (needed by sprint.sh for progress tracking) ─────────────────
install_jq() {
  log "Checking jq..."
  if command -v jq >/dev/null 2>&1; then
    log "jq already installed ✓"
  else
    log "Installing jq via Homebrew..."
    brew install jq
  fi
  log "jq ✓"
}

# ── Claude CLI ──────────────────────────────────────────────────────
check_claude() {
  log "Checking Claude CLI..."
  command -v claude >/dev/null 2>&1 || die "Claude CLI not found. Install from https://claude.ai/claude-code"
  log "Claude CLI $(claude --version 2>/dev/null || echo 'unknown') ✓"
}

# ── .env file ───────────────────────────────────────────────────────
setup_env() {
  local env_file="$PROJECT_DIR/.env"
  if [[ -f "$env_file" ]]; then
    log ".env already exists, skipping ✓"
    return
  fi

  log "Generating .env..."
  local jwt_secret
  jwt_secret=$(openssl rand -hex 32)
  local user
  user=$(whoami)

  cat > "$env_file" <<EOF
NIOTEBOOK_DB_URL=postgres://${user}@localhost/niotebook_dev?sslmode=disable
NIOTEBOOK_JWT_SECRET=${jwt_secret}
NIOTEBOOK_PORT=8080
NIOTEBOOK_HOST=localhost
NIOTEBOOK_LOG_LEVEL=debug
EOF

  log ".env created ✓"
}

# ── Main ────────────────────────────────────────────────────────────
main() {
  echo ""
  echo "=== Niotebook MVP Sprint — Bootstrap ==="
  echo ""

  check_go
  check_brew
  install_postgres
  install_migrate
  install_golangci_lint
  install_jq
  check_claude
  setup_env

  echo ""
  log "All dependencies ready. Run ./scripts/sprint.sh to start the sprint."
  echo ""
}

main "$@"
