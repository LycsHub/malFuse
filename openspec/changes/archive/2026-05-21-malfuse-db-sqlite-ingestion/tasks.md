## 1. Project Setup

- [x] 1.1 Add `modernc.org/sqlite` dependency to go.mod (pure Go SQLite)
- [x] 1.2 Create `cmd/malfuse-db/` directory
- [x] 1.3 Create `internal/db/schema/`, `internal/db/ingest/`, `internal/db/output/` directories

## 2. Database Schema

- [x] 2.1 Define `malicious_packages` table DDL (name, version, ecosystem, published, source)
- [x] 2.2 Define `update_state` table DDL (ecosystem, last_commit, last_updated)
- [x] 2.3 Create composite index on (name, ecosystem, version)
- [x] 2.4 Implement `Open()` function with WAL mode enable
- [x] 2.5 Implement `InsertOrReplace()` for upserting malicious package records
- [x] 2.6 Implement `Delete()` for removing records by name/ecosystem/version
- [x] 2.7 Implement `Lookup()` for querying by name+ecosystem+version
- [x] 2.8 Implement `GetUpdateState()` and `SetUpdateState()` for commit tracking

## 3. OSV Ingestion

- [x] 3.1 Implement OSV JSON struct types (matching OSV schema 1.5.0)
- [x] 3.2 Implement `ParseFile()` to extract package name, ecosystem, versions, published from a single JSON file
- [x] 3.3 Implement ecosystem normalization ("PyPI" → "pypi", "npm" → "npm")
- [x] 3.4 Implement `ListFiles()` to enumerate all OSV JSON files under a directory
- [x] 3.5 Skip files under `osv/withdrawn/` and `osv/unmergable/`
- [x] 3.6 Handle parse errors gracefully (log and skip, don't abort)

## 4. Git Operations

- [x] 4.1 Implement `Clone()` with `--depth 1` for initial clone
- [x] 4.2 Implement `Fetch()` to pull latest changes
- [x] 4.3 Implement `Diff()` to get changed files via `git diff --name-status <from>..<to>`
- [x] 4.4 Implement `HeadHash()` to get current HEAD commit
- [x] 4.5 Handle git command errors gracefully

## 5. Incremental Update Logic

- [x] 5.1 Implement full scan mode (first run, no prior commit)
- [x] 5.2 Implement incremental mode (parse only added/modified/deleted files from git diff)
- [x] 5.3 Update `update_state` after successful update
- [x] 5.4 Handle empty diff (no changes since last update)

## 6. Database Updater (Direct Mode)

- [x] 6.1 Implement direct mode: `InsertOrReplace` for added/modified packages
- [x] 6.2 Implement direct mode: `Delete` for removed packages
- [x] 6.3 Wrap batch updates in a SQL transaction

## 7. SQL Generator (Offline Mode)

- [x] 7.1 Implement SQL file output: generate INSERT OR REPLACE statements
- [x] 7.2 Implement SQL file output: generate DELETE statements for removed packages
- [x] 7.3 Write output to user-specified file path

## 8. malfuse-db CLI

- [x] 8.1 Implement `malfuse-db update` command with flags: `--mode` (direct|sql), `--db` (sqlite path), `--output` (sql file path), `--repo` (git clone path)
- [x] 8.2 Implement `cmd/malfuse-db/main.go` entry point
- [x] 8.3 Wire up the full pipeline: git clone/fetch → diff → parse → write

## 9. Engine Integration

- [x] 9.1 Implement `MaliciousDBCheck(db *sql.DB)` checkFunc in engine package
- [x] 9.2 Query: `SELECT COUNT(*) FROM malicious_packages WHERE name=? AND ecosystem=? AND (version=? OR version IS NULL)`
- [x] 9.3 Return BLOCK with reason "malicious-db" on match, PASS on no match
- [x] 9.4 Graceful fallback: if db is nil, log warning, return PASS
- [x] 9.5 Remove `BlacklistCheck()` and `BlacklistEntry` from engine package

## 10. Config & Proxy Changes

- [x] 10.1 Remove `BlacklistConfig` and `BlacklistEntry` from `internal/config/config.go`
- [x] 10.2 Add `DBPath` field to config for SQLite database path
- [x] 10.3 Update `config.json` to remove blacklist section, add db_path
- [x] 10.4 Update `main.go` to open SQLite in read-only mode and pass to engine
- [x] 10.5 Update engine construction in main.go to use `MaliciousDBCheck` instead of `BlacklistCheck`

## 11. Integration Tests

- [x] 11.1 Test `malfuse-db update` produces valid SQLite database
- [x] 11.2 Test `malfuse-db update --mode sql` produces valid SQL file
- [x] 11.3 Test incremental update only processes changed files
- [x] 11.4 Test malFuse proxy blocks package found in SQLite
- [x] 11.5 Test malFuse proxy passes when db file is missing
- [x] 11.6 Test `CGO_ENABLED=0 go build` for both binaries
- [x] 11.7 Run all existing tests — confirm no regressions
