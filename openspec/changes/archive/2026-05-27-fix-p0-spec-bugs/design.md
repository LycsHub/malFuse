## Context

P0 is "complete" but 4 code behaviors diverge from their approved specs. These are implementation bugs, not design changes. No architectural decisions needed — this is a straight-line code fix to match the spec.

## Goals / Non-Goals

**Goals:**
- Proxy opens SQLite in read-only mode (`mode=ro`)
- `packages.txt` embedded at build time via `//go:embed`
- `packages.txt` expanded to ~2000 entries
- `last_updated` populated with real timestamp after updates

**Non-Goals:**
- Changing any spec or requirement
- Adding new features
- Refactoring unrelated code

## Decisions

**1. Separate `OpenReadOnly()` instead of modifying `Open()`**
- `OpenReadOnly()` skips WAL pragma and migration DDL — the DB is assumed to already exist.
- Alternative: Add a `readOnly bool` parameter to `Open()`. Rejected — more invasive, changes existing test signatures.
- Rationale: Clear separation of concerns. `Open()` is for `malfuse-db` (writes). `OpenReadOnly()` is for `malfuse` (reads).

**2. `//go:embed` replacing runtime `os.ReadFile`**
- `packages.txt` moved to `internal/engine/` (already there), accessed via `//go:embed packages.txt`.
- Alternative: Embed inside `typo.go` directly. Chose separate file for maintainability.
- Rationale: Binary is self-contained regardless of working directory.

**3. Timestamp format for `last_updated`**
- Use RFC3339 (`time.RFC3339`). Update tests accordingly.

## Risks / Trade-offs

- **[Risk] `OpenReadOnly()` fails silently if DB does not exist** — proxy starts without DB check. Mitigation: Already handled by `MaliciousDBCheck(nil)` graceful fallback.
- **[Risk] Larger binary due to embedded packages.txt** — ~30KB of text data. Negligible.
