## 1. Proxy opens SQLite read-only

- [x] 1.1 Add `OpenReadOnly()` function to `internal/db/schema/schema.go` (opens with `mode=ro`, no WAL, no migrate)
- [x] 1.2 Update `main.go` to use `OpenReadOnly()` instead of `Open()` for the proxy
- [x] 1.3 Update `TestMaliciousDBIntegration` proxy test to verify proxy works with read-only opened DB
- [x] 1.4 Run existing tests — confirm no regressions

## 2. Embed packages.txt at build time

- [x] 2.1 Expand `internal/engine/packages.txt` to ~2000 popular PyPI/NPM packages
- [x] 2.2 Use `//go:embed packages.txt` in engine package instead of `os.ReadFile` in `main.go`
- [x] 2.3 Update typo check tests to use embedded data
- [x] 2.4 Remove `loadPopularPackages()` from `main.go`
- [x] 2.5 Run existing tests — confirm no regressions

## 3. Populate last_updated timestamp

- [x] 3.1 Pass `time.Now().Format(time.RFC3339)` to `SetUpdateState` instead of empty string
- [x] 3.2 Update `TestUpdateState` to verify `last_updated` is populated
- [x] 3.3 Run existing tests — confirm no regressions
