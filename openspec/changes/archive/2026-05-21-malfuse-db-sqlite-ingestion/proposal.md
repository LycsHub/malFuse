## Why

P0 uses a static `config.json` blacklist with hand-maintained entries. The OpenSSF maintains a comprehensive database of 10,000+ confirmed malicious packages at `ossf/malicious-packages` in OSV format, updated daily. We need to ingest this dataset into a local SQLite database and wire it into the malFuse proxy, making the blacklist dynamic, authoritative, and auto-updating.

## What Changes

- New `malfuse-db` CLI binary for data ingestion and database management
- `internal/db/` packages: SQLite schema, OSV JSON parser, git-based incremental updater
- **BREAKING**: `config.json` static blacklist (`blacklist.entries`) is removed — replaced by SQLite lookup
- `internal/engine/` gains a new `malicious-db` check that queries SQLite on each request
- OSV API check (`internal/osv/`) is retained as a complementary layer
- `malfuse` proxy opens SQLite read-only at startup, queries it per-request
- Supports two update modes: direct write to SQLite, and SQL incremental file generation for offline use

## Capabilities

### New Capabilities

- `db-schema`: SQLite schema with `malicious_packages` (name, version, ecosystem, published, source) and `update_state` (ecosystem, last_commit, last_updated) tables
- `osv-ingestion`: Clone/fetch/pull ossf/malicious-packages repo, parse OSV JSON files, extract package names/versions/ecosystems/publish times
- `git-incremental-update`: Use `git diff --name-status HEAD..origin/main` to identify changed files, only parse deltas
- `sql-generator`: Generate SQL INSERT/DELETE/UPDATE statements for offline deployment to air-gapped environments
- `db-updater`: Direct-mode: write parsed records into SQLite with upsert semantics (INSERT OR REPLACE)
- `malicious-db-check`: Query SQLite per-request to check if a package is known malicious, complementing existing OSV API check

### Modified Capabilities

- `blacklist-check`: **BREAKING** — static `config.json` blacklist is removed. Replaced by SQLite-backed lookup that supports name-only and name+version matching with ecosystem awareness.

## Impact

- New binary: `malfuse-db` (cmd/malfuse-db/main.go)
- New packages: `internal/db/schema/`, `internal/db/ingest/`, `internal/db/output/`
- Modified: `internal/engine/blacklist.go` → replaced with SQLite query
- Modified: `internal/config/config.go` → remove `BlacklistConfig` struct
- Modified: `internal/config/config.json` → remove `blacklist` section
- Modified: `main.go` → open SQLite, pass db handle to engine
- New dependency: `modernc.org/sqlite` (pure Go, `CGO_ENABLED=0` compatible)
- New dependency: git client available at runtime (`git` command)
