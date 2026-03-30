#!/usr/bin/env bash
# init-ruflo.sh — Initialize RuFlo for prl-devops-service (Go backend)
# Run once from the project root after cloning or on a new machine.
set -euo pipefail
SRC_FOLDER="/src"

# ─── Colours ────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}[ruflo]${NC} $*"; }
success() { echo -e "${GREEN}[✓]${NC} $*"; }
warn()    { echo -e "${YELLOW}[⚠]${NC} $*"; }
error()   { echo -e "${RED}[✗]${NC} $*"; exit 1; }

# ─── Preflight ───────────────────────────────────────────────────────────────
info "Checking prerequisites..."

command -v node  >/dev/null 2>&1 || error "Node.js not found. Install Node >= 20."
command -v npm   >/dev/null 2>&1 || error "npm not found."
command -v go    >/dev/null 2>&1 || error "Go not found. Install Go 1.24+."
command -v claude >/dev/null 2>&1 || error "Claude Code CLI not found. Run: npm install -g @anthropic-ai/claude-code"
command -v git   >/dev/null 2>&1 || error "Git not found."

NODE_VERSION=$(node -e "process.exit(parseInt(process.versions.node) < 20 ? 1 : 0)" 2>&1) \
  || error "Node.js >= 20 required. Found: $(node --version)"

[[ -f "src/go.mod" ]] || error "go.mod not found. Run this script from the project root."
[[ -d "src" ]]    || error "src/ directory not found. Are you in the right project?"

success "Prerequisites OK"

info "Cleaning up any ruflo files misplaced inside src/..."
rm -rf src/.claude src/.claude-flow src/CLAUDE.md src/.mcp.json 2>/dev/null || true
success "src/ is clean"

# ─── Install ruflo locally ────────────────────────────────────────────────────
info "Installing ruflo locally..."

# Create a minimal package.json if none exists (Go projects won't have one)
if [[ ! -f "package.json" ]]; then
  info "No package.json found — creating a minimal one for ruflo..."
  cat > package.json <<'EOF'
{
  "name": "prl-devops-service-ruflo",
  "version": "1.0.0",
  "private": true,
  "description": "RuFlo tooling for prl-devops-service",
  "scripts": {
    "ruflo": "ruflo"
  }
}
EOF
  success "package.json created"
fi

npm install --save-dev ruflo@latest
success "ruflo installed to node_modules"

# ─── Register MCP servers ─────────────────────────────────────────────────────
info "Registering MCP servers with Claude Code..."

# Remove stale entries first (ignore errors if they don't exist)
claude mcp remove ruflo      2>/dev/null || true
claude mcp remove claude-flow 2>/dev/null || true
claude mcp remove ruv-swarm  2>/dev/null || true

# Register using local binary — avoids PATH issues with npx inside Claude Code
claude mcp add ruflo -- node node_modules/ruflo/bin/ruflo.js mcp start
success "ruflo MCP server registered"

# Verify
info "Verifying MCP registration..."
claude mcp list | grep -q "ruflo" || error "ruflo not found in claude mcp list after registration"
success "MCP registration verified"

# ─── Initialise ruflo project config ─────────────────────────────────────────
info "Initialising ruflo project config..."

# Only init if config doesn't already exist
if [[ ! -f ".claude-flow/config.yaml" ]]; then
  npx @claude-flow/cli@latest init --force 2>/dev/null || \
  node node_modules/ruflo/bin/ruflo.js init --force
  success "ruflo project config created"
else
  success "ruflo config already exists — skipping init"
fi

# ─── Initialise memory database ──────────────────────────────────────────────
info "Initialising memory database..."

# Wipe stale state if memory DB is missing or empty
if [[ ! -f ".swarm/memory.db" ]]; then
  rm -rf .claude-flow/data .swarm 2>/dev/null || true
  node node_modules/ruflo/bin/ruflo.js memory init
  success "Memory database initialised"
else
  success "Memory database already exists — skipping"
fi

# Seed project context into memory so agents understand the codebase
info "Seeding project memory..."

seed_memory() {
  local key="$1"
  local value="$2"
  local namespace="$3"
  local type="${4:-procedural}"
  local id="entry_$(date +%s%3N)_$(echo $key | tr '/' '_')"

  sqlite3 .swarm/memory.db <<EOF
INSERT INTO memory_entries (id, key, namespace, content, type, status, created_at, updated_at)
VALUES (
  '$id',
  '$key',
  '$namespace',
  '${value//\'/\'\'}',
  '$type',
  'active',
  strftime('%s', 'now') * 1000,
  strftime('%s', 'now') * 1000
)
ON CONFLICT(namespace, key) DO UPDATE SET
  content    = excluded.content,
  updated_at = strftime('%s', 'now') * 1000;
EOF
  success "Seeded $namespace/$key"
}

seed_memory "project/stack" \
  "Go 1.24 service, all source under src/, DDD bounded contexts, JSON flat-file DB with dataMutex, event-driven via WebSocket PDFM events" \
  "prl-devops-service" "procedural"

seed_memory "project/build" \
  "make build | go build ./src/... | make test | cd src && go test ./... | make lint (golangci-lint via Docker)" \
  "prl-devops-service" "procedural"

seed_memory "project/patterns" \
  "New endpoint: controllers/ -> restapi/ -> models/ -> data/ -> mappers/. Atomic DB writes only. Never lock dataMutex manually outside data/. Use ctx.LogInfof for logging." \
  "prl-devops-service" "procedural"

seed_memory "project/packages" \
  "orchestrator/ (host mgmt VM sync WS), controllers/ (HTTP), data/ (JSON DB), models/ (API), data/models/ (DTO), mappers/ (API<->DTO), serviceprovider/ (DI), restapi/ (router), security/ (auth), jobs/ (background), handlers/ (WS events)" \
  "prl-devops-service" "procedural"

success "Project memory seeded"

# Verify
info "Verifying seeded entries..."
sqlite3 .swarm/memory.db \
  "SELECT namespace || '/' || key FROM memory_entries WHERE namespace='prl-devops-service';"

success "Project memory seeded"

# ─── Start daemon ─────────────────────────────────────────────────────────────
info "Starting ruflo daemon..."
node node_modules/ruflo/bin/ruflo.js daemon stop 2>/dev/null || true
sleep 1
node node_modules/ruflo/bin/ruflo.js daemon start
success "Daemon started"

# ─── .gitignore ──────────────────────────────────────────────────────────────
info "Updating .gitignore..."

GITIGNORE_ENTRIES=(
  ""
  "# RuFlo runtime"
  ".swarm/"
  ".claude-flow/data/"
  ".claude-flow/logs/"
  ".claude-flow/daemon.pid"
  ".claude-flow/sessions/"
  ".claude-flow/metrics/"
  ".claude-flow/security/" 
  ".claude-flow/daemon-state.json"
  ".claude-flow/swarm/swarm-state.json"
  "*.db"
  "node_modules/"
  "package-lock.json"
)

for entry in "${GITIGNORE_ENTRIES[@]}"; do
  if [[ -z "$entry" ]] || ! grep -qxF "$entry" .gitignore 2>/dev/null; then
    echo "$entry" >> .gitignore
  fi
done
success ".gitignore updated"

# ─── Health check ─────────────────────────────────────────────────────────────
info "Running health check..."
echo ""
node node_modules/ruflo/bin/ruflo.js doctor
echo ""

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  RuFlo initialised for prl-devops-service${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${CYAN}Next steps:${NC}"
echo -e "  1. Restart Claude Code completely (fully quit and reopen)"
echo -e "  2. Verify MCP is connected: run ${YELLOW}/mcp${NC} inside Claude Code"
echo -e "  3. Test the swarm with a real task, e.g.:"
echo -e "     ${YELLOW}Following the swarm rules in CLAUDE.md, add a new endpoint to the users controller${NC}"
echo ""
echo -e "  ${CYAN}Files to commit:${NC}"
echo -e "  ${GREEN}✓${NC} CLAUDE.md"
echo -e "  ${GREEN}✓${NC} .mcp.json"
echo -e "  ${GREEN}✓${NC} .claude-flow/config.yaml"
echo -e "  ${GREEN}✓${NC} .claude/settings.json"
echo -e "  ${GREEN}✓${NC} .claude/commands/"
echo -e "  ${GREEN}✓${NC} package.json"
echo -e "  ${GREEN}✓${NC} .gitignore"
echo -e "  ${RED}✗${NC} node_modules/  .swarm/  .claude-flow/data/  *.db"
echo ""