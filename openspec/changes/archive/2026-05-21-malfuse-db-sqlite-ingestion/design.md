## Context

P0 has a static JSON blacklist and a live OSV API check. The `ossf/malicious-packages` GitHub repository provides 10,000+ confirmed malicious package reports in OSV format, updated continuously. This change adds a SQLite-backed ingestion pipeline that replaces the static blacklist, while retaining the live OSV API as a complementary layer.

**Constraints:**
- `CGO_ENABLED=0` — use `modernc.org/sqlite` (pure Go SQLite)
- Git must be available at runtime for clone/pull
- Network access required for initial clone and periodic pull
- Offline mode must work without network (read existing SQLite, apply SQL files)

## Goals / Non-Goals

**Goals:**
- Independent `malfuse-db` CLI for data ingestion
- Git-based incremental update using `git diff --name-status`
- Parse OSV JSON files, extract package name/version/ecosystem/published time
- Store in SQLite with minimal schema + time info
- Two output modes: direct SQLite write, SQL file generation
- `malfuse` proxy queries SQLite per-request for blacklist lookup
- Unit test coverage

**Non-Goals:**
- Real-time push-based updates (webhook, polling API)
- Web UI or dashboard
- Multi-ecosystem stats/reporting
- bbolt migration
- CI/CD integration for automated ingest

## Decisions

**1. `modernc.org/sqlite` over `mattn/go-sqlite3`**
- **Chosen:** `modernc.org/sqlite` — pure Go SQLite, no CGo.
- **Alternative:** `mattn/go-sqlite3` — requires CGo, breaks `CGO_ENABLED=0`.
- **Rationale:** Must remain cross-compilable without a C toolchain.

**2. Per-request SQLite query over in-memory load**
- **Chosen:** `SELECT COUNT(*) FROM malicious_packages WHERE name=? AND ecosystem=? AND (version=? OR version IS NULL)` on each request.
- **Alternative:** Load entire table to memory map at startup.
- **Rationale:** The dataset is small enough that indexed single-row lookups are sub-millisecond. Per-request queries enable instant visibility when `malfuse-db` updates the database while `malfuse` is running (WAL mode).

**3. Git diff incremental over filesystem mtime**
- **Chosen:** `git fetch && git diff --name-status HEAD..origin/main` to identify changed files.
- **Alternative:** Re-scan all JSON files each time.
- **Rationale:** 10,000+ entries, full rescan is slow. Git diff is O(changed files), not O(total files). Precise: catches adds, modifications, and deletions.

**4. Separate binary over integrated CLI**
- **Chosen:** Independent `malfuse-db` binary, distinct from `malfuse` proxy.
- **Alternative:** Single binary with `malfuse ingest` subcommand.
- **Rationale:** `malfuse-db` has heavy dependencies (git, network, SQLite write access) that the proxy should not carry. Separation of concerns: ingest is an admin/maintenance task, proxy is a runtime service.

**5. SQLite WAL mode**
- **Chosen:** Enable WAL journal mode for the database.
- **Rationale:** Allows `malfuse-db` to write while `malfuse` reads without locking conflicts. WAL is the default for `modernc.org/sqlite`.

## Risks / Trade-offs

- **[Risk] Git clone of large repo on first run** — The ossf/malicious-packages repo has 11,400+ commits and many files. Initial clone may take seconds to minutes. Mitigation: Show progress, use shallow clone (`--depth 1`) on initial run to reduce transfer size, then `git fetch --unshallow` incrementally.
- **[Risk] SQLite write contention if proxy updates** — If `malfuse` were to write to SQLite while `malfuse-db` generates, there could be contention. Mitigation: `malfuse` opens SQLite read-only. Only `malfuse-db` writes.
- **[Trade-off] Two binaries to distribute** — Users need both `malfuse` and `malfuse-db`. Mitigation: Document clearly. `malfuse` fails gracefully if SQLite is missing (log warning, skip db check).
