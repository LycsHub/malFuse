# 📖 MalFuse — Malicious Package Firewall

[中文](README_CN.md) | English

`MalFuse` is a local HTTP proxy built in **Go** that intercepts `pip install` / `npm install` traffic and blocks malicious packages **before they land on disk** via inline HTTP 403. The name comes from **Mal**icious + **Fuse** (circuit breaker).

P0 + P1 complete. P2 in progress (whitelist done). 252,637 malicious package records across 11 ecosystems.

---

# Usage

## Quick Start

```bash
# 1. Generate malicious package database
./malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 2. Configure package manager
./malfuse link

# 3. Start proxy
./malfuse start            # daemon mode (background)
# or
./malfuse -config config.json  # foreground

# 4. Install dependencies normally (traffic routed through proxy)
pip install requests
npm install lodash
```

---

## CLI Commands

### `malfuse` — Proxy + Management

```bash
# Running
malfuse                         # foreground (default config.json)
malfuse -c /path/to/config.json # custom config
malfuse start                   # daemon (background)
malfuse stop                    # stop daemon
malfuse status                  # check daemon status

# Package manager configuration
malfuse link                    # configure all installed managers
malfuse link --target pip       # pip only
malfuse link --target npm       # npm only
malfuse link --target pnpm      # pnpm only
malfuse link --target yarn      # yarn v1 only
malfuse unlink                  # restore all
malfuse unlink --target pip     # restore pip only

# Whitelist management
malfuse allow add requests --ecosystem pypi               # allow all versions
malfuse allow add lodash --ecosystem npm --version 4.17.21 # allow specific version
malfuse allow remove requests --ecosystem pypi              # remove from whitelist
malfuse allow list                                          # list all
malfuse allow list --ecosystem npm                          # filter by ecosystem
```

### `malfuse-db` — Database Management

```bash
# Direct SQLite write (online use)
malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# Generate SQL incremental file (offline/air-gapped use)
malfuse-db --mode sql --output updates-20260527.sql --repo ossf-malicious-packages

# Specify config (reads repo_proxy for GitHub acceleration)
malfuse-db --config config.json --mode direct
```

| Flag | Default | Description |
|------|---------|-------------|
| `--mode` | `direct` | `direct` write to SQLite / `sql` generate incremental SQL file |
| `--db` | `malfuse.db` | SQLite database path |
| `--repo` | `ossf-malicious-packages` | git repo cache directory |
| `--output` | `updates-YYYYMMDDHHmm.sql` | SQL output path (sql mode only) |
| `--config` | `config.json` | config file path (reads `repo_proxy` for proxy) |

**Update mechanism:** First run does a full scan (clone + parse all OSV JSON). Subsequent runs use `git fetch` + `git diff --name-status` for incremental updates, completing in seconds.

**Output:** `direct` mode writes to `malfuse.db` (WAL mode, readable by proxy concurrently). `sql` mode generates `INSERT OR REPLACE` + `DELETE` statements for importing into air-gapped databases.

---

## Package Manager Configuration

### One-click (`malfuse link`)

| Tool | Command | What it does |
|------|---------|--------------|
| pip | `malfuse link --target pip` | `pip config set global.index-url http://127.0.0.1:8080/pypi/simple/` |
| npm | `malfuse link --target npm` | `npm config set registry http://127.0.0.1:8080/npm/` |
| pnpm | `malfuse link --target pnpm` | `pnpm config set registry http://127.0.0.1:8080/npm/` |
| yarn v1 | `malfuse link --target yarn` | `yarn config set registry http://127.0.0.1:8080/npm/` |

`malfuse link` backs up original values to `~/.malfuse_backup.json`. `malfuse unlink` restores from backup.

### Manual Configuration

These tools don't support CLI config commands—manual setup required:

| Tool | Documentation |
|------|---------------|
| yarn v2+ (Berry) | [docs/yarn-v2-config.md](docs/yarn-v2-config.md) |
| uv | [docs/uv-config.md](docs/uv-config.md) |
| poetry | [docs/poetry-config.md](docs/poetry-config.md) |
| conda | [docs/conda-config.md](docs/conda-config.md) |

---

## Blocking Granularity

Different package managers have different proxy behaviors, resulting in different version-matching precision:

| Ecosystem | Granularity | Mechanism |
|-----------|-------------|-----------|
| **PyPI (pip)** | Version-precise | Simple API only blocks `version=NULL` entries. Proxy rewrites HTML download links so pip's tarball requests go back through the proxy, enabling exact version matching |
| **npm** | Package-level (all-or-nothing) | npm's `pacote` download library bypasses proxy caching, preventing precise version matching on download. Only `version=NULL` entries block at Simple API stage with 403 |

### npm Version-Precise Workaround

For npm version-level control, use whitelist with all-or-nothing blocking:

```bash
# 1. Mark package as all-versions blocked (version=NULL)
sqlite3 malfuse.db "INSERT INTO malicious_packages VALUES ('bad-lib', NULL, 'npm', '', '');"

# 2. Whitelist the safe version
./malfuse allow add bad-lib --ecosystem npm --version 2.0.5
```

Now `npm install bad-lib@2.0.5` passes, all other versions are blocked.

### Blocking Behavior Matrix

| Scenario | pip | npm |
|----------|-----|-----|
| DB `version=NULL` | All versions blocked | All versions blocked |
| DB `version="1.0"`, install 1.0 | Blocked (exact match on download) | Blocked (unless whitelisted) |
| DB `version="1.0"`, install 2.0 | Pass | Blocked (unless whitelisted) |

---

## Detection Pipeline

Each install request passes through 6 checks (including whitelist). The first match stops the pipeline:

| # | Check | Data Source | Result | Default |
|---|-------|-------------|--------|---------|
| 0 | **Whitelist** | SQLite `whitelist` table | Match → immediate PASS, skip all remaining checks | On |
| 1 | **Malicious DB** | SQLite `malicious_packages` (252,637 records) | Match → 403 Forbidden | On |
| 2 | **Cooldown** | `malicious_packages.published` (OSV report timestamp) | Report age < 48h → 403 | Off |
| 3 | **Typo-Squatting** | Embedded 2,790 popular packages + Levenshtein distance | Name similarity → 403 | On |
| 4 | **OSV API** | `api.osv.dev/v1/query` + in-memory TTL cache | Vuln found + `block_on_vuln=true` → 403 | On (log only) |
| 5 | **Stream Script Scan** | TeeReader streaming (tar/gzip extraction) | Malicious script → connection reset | Off |

**Failure policies:**
- Whitelist / typo — no failure possible (pure memory / SQLite)
- Malicious DB — skip on DB missing or corrupt (log WARN)
- Cooldown — skip on missing DB or `published` field; DB-only query, no extra network calls
- OSV API — network unreachable → pass (fail-open); `block_on_vuln=false` → log only
- Script scan — parse/archive error → pass (fail-open)

**Script scan (#5) attack vectors covered:**

| JS Ecosystem | Python Ecosystem |
|--------------|-----------------|
| `package.json` `scripts.preinstall/postinstall/install` | `setup.py` full content |
| `package.json` scripts-referenced `.js` files | `__init__.py` full content |
| Standalone `.js` files | `.pth` file `import` lines |
| — | `pyproject.toml` `build-system.build-backend` |

Each vector analyzed by three detectors: **Shannon entropy** (threshold 4.5), **code obfuscation** (base64/hex/eval chains), **network detection** (URLs/IPs).

---

## Health Check

```bash
$ curl http://127.0.0.1:8080/health
{"db":true,"status":"ok","uptime":"2h34m5s"}
```

| Field | Description |
|-------|-------------|
| `status` | `ok` (healthy) or `degraded` (DB unavailable) |
| `db` | SQLite connection status |
| `uptime` | Proxy process uptime |

---

## Configuration (config.json)

```json
{
  "port": "8080",
  "host": "127.0.0.1",
  "db_path": "malfuse.db",
  "pid_file": "malfuse.pid",
  "repo_proxy": "ghfast.top",
  "logging": {
    "level": "info",
    "format": "text",
    "output": "stdout"
  },
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.tuna.tsinghua.edu.cn", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmmirror.com", "ecosystem": "npm"}
  ],
  "cooldown": {
    "enabled": false,
    "duration": "48h"
  },
  "typo": {
    "enabled": true,
    "threshold": 2
  },
  "osv": {
    "enabled": true,
    "block_on_vuln": false,
    "ttl": "1h",
    "base_url": "https://api.osv.dev"
  },
  "script_scan": {
    "enabled": false,
    "max_file_size": 5242880,
    "max_total_size": 52428800,
    "entropy": { "enabled": true, "threshold": 4.5 },
    "obfuscation": { "enabled": true, "base64_min_length": 100, "hex_min_length": 20 },
    "network": { "enabled": true, "allow_private_ips": false }
  }
}
```

**Configuration reference:**

| Section | Field | Description |
|---------|-------|-------------|
| Base | `port` / `host` | Proxy listen address |
| Base | `db_path` | SQLite path; auto-skip DB check if file absent |
| Base | `pid_file` | Daemon mode PID file path (default `malfuse.pid`) |
| Base | `repo_proxy` | GitHub acceleration proxy domain (e.g. `ghfast.top`); omit for no proxy |
| `logging` | `level` | `debug` / `info` / `warn` / `error` |
| `logging` | `format` | `text` or `json` (JSON for log collection systems) |
| `logging` | `output` | `stdout` or file path (file mode also writes to stdout) |
| `routing` | `prefix` | URL prefix matching request path |
| `routing` | `upstream` | Real registry URL (proxy internally uses HTTPS) |
| `routing` | `ecosystem` | Ecosystem identifier (`pypi` / `npm`, used for DB + OSV queries) |
| `cooldown` | `enabled` | Default off, must explicitly enable |
| `cooldown` | `duration` | Block if OSV report published less than this duration ago |
| `typo` | `threshold` | Block if Levenshtein distance ≤ this value |
| `osv` | `block_on_vuln` | Whether to block on vulnerability found (default `false`, log only) |
| `osv` | `ttl` | Query result cache duration |
| `osv` | `base_url` | OSV API endpoint |
| `script_scan` | `enabled` | Default off, must explicitly enable |
| `script_scan` | `max_file_size` | Max single file to analyze (bytes), skip if larger |
| `script_scan` | `max_total_size` | Max total stream size, stop scanning if exceeded |
| `script_scan.entropy` | `threshold` | Shannon entropy threshold (~4.5 = upper bound of English text) |
| `script_scan.obfuscation` | `base64_min_length` | Min base64 string length to trigger detection |
| `script_scan.obfuscation` | `hex_min_length` | Min consecutive `\xNN` count to trigger detection |
| `script_scan.network` | `allow_private_ips` | Whether to allow private IPs (`10.x`, `192.168.x`, etc.) |

---

## Build

```bash
CGO_ENABLED=0 go build -o malfuse .
CGO_ENABLED=0 go build -o malfuse-db ./cmd/malfuse-db/
```

Pure Go, zero CGo dependencies. Single build runs on Linux / macOS (Intel + Apple Silicon) / Windows.

---

# Development

## Directory Structure

```
malFuse/
├── main.go                    # malfuse proxy entry (cobra CLI)
├── config.json                # Configuration
├── cmd/
│   └── malfuse-db/            # Database management CLI
├── internal/
│   ├── config/                # JSON config loading + validation
│   ├── proxy/                 # HTTP proxy (routing, forwarding, health)
│   ├── engine/                # Detection pipeline (whitelist, mal-db, cooldown, typo, OSV) + StreamChecker
│   ├── scanner/               # Streaming script scan (entropy/obfuscation/network + JS/Python analysis)
│   ├── osv/                   # OSV API client + in-memory TTL cache
│   ├── logger/                # logrus structured logging wrapper
│   ├── daemon/                # Background process management (PID, signals)
│   ├── linker/                # Package manager config (pip/npm/pnpm/yarn)
│   └── db/
│       ├── schema/            # SQLite DDL + CRUD (WAL mode, DBExec interface)
│       ├── ingest/            # OSV JSON 1.5.0 parsing + Git operations
│       └── output/            # Direct DB write / SQL incremental file generation
├── .github/workflows/         # CI/CD pipelines
├── docs/                      # Manual configuration guides
├── malfuse.db                 # SQLite database (generated by malfuse-db)
└── ossf-malicious-packages/   # Git repo cache (gitignored)
```

## Running Tests

```bash
# All unit + integration tests
go test ./internal/...

# Specific package
go test -v ./internal/scanner/
go test -v ./internal/engine/
```

130+ tests covering all packages.

## Roadmap

### ✅ P0 — Core Skeleton (Complete)

- [x] HTTP reverse proxy + routing
- [x] Malicious package SQLite database (252,637 records, 11 ecosystems)
- [x] `malfuse-db` CLI (git incremental fetch + SQL offline mode)
- [x] Detection pipeline (malicious-db / cooldown / typo / OSV)
- [x] Streaming script scanner (entropy / obfuscation / network + JS/Python analysis)

### ✅ P1 — Automation & Operations (Complete)

- [x] `malfuse link` / `malfuse unlink` (pip / npm / pnpm / yarn)
- [x] logrus structured logging (level control, JSON format, file output)
- [x] `/health` health check endpoint
- [x] Daemon mode (`malfuse start/stop/status`)
- [x] End-to-end integration test suite

### 🟢 P2 — Deep Scanning & Ecosystem Expansion

- [x] `malfuse allow` whitelist management
- [x] CI/CD Pipeline (test + DB auto-update + release on tag)
- [ ] More ecosystem routes (RubyGems, NuGet, Crates.io, Go modules)
- [ ] Docker image distribution
- [ ] Install script AST analysis

## Tech Stack

| Component | Library |
|-----------|---------|
| CLI framework | `github.com/spf13/cobra` |
| Structured logging | `github.com/sirupsen/logrus` |
| SQLite | `modernc.org/sqlite` (pure Go, zero CGo) |
| HTTP proxy | `net/http/httputil.ReverseProxy` (stdlib) |
| Malicious package format | [OSV Schema 1.5.0](https://ossf.github.io/osv-schema/) |
| Typo detection | Custom Levenshtein distance implementation |
| Entropy detection | Custom Shannon Entropy implementation |
| Obfuscation detection | regexp (stdlib) |
