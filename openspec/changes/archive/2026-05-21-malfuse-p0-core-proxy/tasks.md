## 1. Project Setup

- [x] 1.1 Update go.mod with module path and Go version
- [x] 1.2 Create `internal/config/`, `internal/proxy/`, `internal/engine/`, `internal/osv/` directory structure
- [x] 1.3 Create `config.yaml` with default routing table, blacklist, and check toggles

## 2. Configuration

- [x] 2.1 Define `Config` struct and sub-structs (`Route`, `ChecksConfig`, `BlacklistConfig`, etc.)
- [x] 2.2 Implement JSON config loading with default path `./config.json`
- [x] 2.3 Implement `-config` flag parsing in main
- [x] 2.4 Validate blacklist entries on load (reject empty names, fatal on invalid)

## 3. Engine Pipeline

- [x] 3.1 Define `engine.Request` and `engine.Result` types
- [x] 3.2 Define `Engine` struct holding config and check functions
- [x] 3.3 Implement `Engine.Check(ctx, Request) Result` with sequential execution and short-circuit
- [x] 3.4 Implement per-check enable/disable gating
- [x] 3.5 Handle context cancellation (return BLOCK on timeout/cancel)

## 4. Blacklist Check

- [x] 4.1 Implement name-only matching (entry with no version)
- [x] 4.2 Implement name+version matching (entry with version constraint)
- [x] 4.3 Add blacklist lookup to engine pipeline

## 5. Typo-squatting Check

- [x] 5.1 Create embedded top-2000 packages data file under `internal/engine/`
- [x] 5.2 Implement Levenshtein distance function in Go
- [x] 5.3 Implement typo check: iterate top-2000 list, compare distance <= threshold
- [x] 5.4 Skip packages with name shorter than 3 characters
- [x] 5.5 Skip exact matches (distance 0 means it IS the popular package)
- [x] 5.6 Add typo check to engine pipeline

## 6. OSV Check

- [x] 6.1 Implement OSV API client (`internal/osv/`) for `POST /v1/query`
- [x] 6.2 Implement in-memory TTL cache with `sync.RWMutex`
- [x] 6.3 Implement cache key as `name:ecosystem` for per-ecosystem isolation
- [x] 6.4 Implement fail-open on API error: timeout (2s), non-200, connection error → log warning, return PASS
- [x] 6.5 Add OSV check to engine pipeline

## 7. Cooldown Check

- [x] 7.1 Implement upstream metadata fetch (PyPI `/pypi/<name>/json`, NPM `/<name>`)
- [x] 7.2 Parse `upload_time` / `time` field from registry JSON response
- [x] 7.3 Implement fail-closed: metadata parse error or fetch timeout (2s) → BLOCK
- [x] 7.4 Compare publish time against configurable cooldown duration
- [x] 7.5 Add cooldown check to engine pipeline (after blacklist, before typo)

## 8. Proxy Layer

- [x] 8.1 Define `proxy.Handler` struct with engine, route map, and upstream transport
- [x] 8.2 Implement route matching: strip prefix, extract upstream URL and ecosystem
- [x] 8.3 Implement package name/version extraction from URL path (PyPI and NPM patterns)
- [x] 8.4 Implement `ServeHTTP`: match route → run engine → 403 or forward
- [x] 8.5 Implement upstream forwarding: recreate request URL to upstream over HTTPS
- [x] 8.6 Implement upstream response streaming (headers + body passthrough)
- [x] 8.7 Handle unmatched routes (502)
- [x] 8.8 Handle upstream unreachable (502)

## 9. Main Entry Point & Graceful Shutdown

- [x] 9.1 Wire up main.go: load config, create engine, create proxy handler, start server
- [x] 9.2 Implement SIGINT/SIGTERM signal handling with graceful shutdown (5s drain)
- [x] 9.3 Log server start message with listening address
- [x] 9.4 Log each blocked request with package name, version, and check reason

## 10. Integration Verification

- [x] 10.1 Verify proxy starts and listens on configured port
- [x] 10.2 Verify blacklisted package returns 403 with reason in body
- [x] 10.3 Verify allowed package is proxied successfully (real PyPI test)
- [x] 10.4 Verify unmatched route returns 502
- [x] 10.5 Verify `CGO_ENABLED=0 go build` produces a static binary
