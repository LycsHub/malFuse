# malFuse P0 — Core Proxy Architecture

**Date:** 2026-05-21

## 1. Overview

malFuse P0 is a local HTTP proxy that intercepts `pip install` / `npm install` traffic, inspects packages against a multi-layered check pipeline, and either forwards to the real registry or returns HTTP 403.

## 2. Runtime Model

- Single binary, foreground process
- `malfuse -config config.yaml` or `malfuse` (defaults to `./config.yaml`)
- Graceful shutdown on SIGINT/SIGTERM (5s drain)
- No daemon mode, no backgrounding

## 3. Package Layout

```
malFuse/
├── main.go              # Flags, config load, proxy start
├── go.mod
├── config.yaml          # Default config
├── internal/
│   ├── proxy/           # HTTP handler + upstream forwarding
│   ├── engine/          # Sequential check pipeline
│   ├── config/          # Config struct + YAML unmarshaling
│   └── osv/             # OSV API client + in-memory TTL cache
```

## 4. Configuration (`config.yaml`)

```yaml
port: 8080
host: "127.0.0.1"

routing:
  - prefix: /pypi/
    upstream: https://pypi.org
    ecosystem: pypi
  - prefix: /npm/
    upstream: https://registry.npmjs.org
    ecosystem: npm

blacklist:
  entries:
    - name: malicious-pkg
    - name: bad-lib
      version: "2.0.0"

cooldown:
  enabled: false
  duration: 48h

typo:
  enabled: true
  threshold: 2

osv:
  enabled: true
  ttl: 1h
  base_url: https://api.osv.dev
```

## 5. Request Flow

1. Client sends `GET /pypi/simple/requests/` to `127.0.0.1:8080`
2. Proxy matches `/pypi/` prefix, extracts upstream URL + ecosystem
3. Proxy parses package name/version from remaining path
4. Proxy calls `engine.Check(ctx, Request{Name, Version, Ecosystem, RawPath})`
5. If BLOCK → `403 Forbidden` with reason in body
6. If PASS → strip prefix, forward to upstream over HTTPS, stream response back

## 6. Check Pipeline (sequential, short-circuit)

| # | Check | Data | Fail Mode | P0 Default |
|---|-------|------|-----------|------------|
| 1 | **Blacklist** | config.yaml entries (name + optional version) | — | enabled |
| 2 | **Cooldown** | fetch upstream metadata JSON, extract `upload_time` | fail-closed | disabled (opt-in) |
| 3 | **Typo-squatting** | embedded top-2000 packages, Levenshtein distance | — | enabled |
| 4 | **OSV** | live API query with in-memory TTL cache | fail-open | enabled |

## 7. Component Interfaces

```go
// engine.Request — input to all checks
type Request struct {
    Name      string
    Version   string
    Ecosystem string
    RawPath   string
}

// engine.Result — output from engine
type Result struct {
    Block  bool
    Reason string
}

// engine.Engine — sole contract between proxy and engine
func (e *Engine) Check(ctx context.Context, req Request) Result

// proxy.Handler — HTTP handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

## 8. Error Handling

- **Unmatched route prefix:** 502, log
- **Upstream unreachable:** 502, log
- **Cooldown metadata parse failure or timeout:** fail-closed → block
- **OSV API unreachable or timeout (2s):** fail-open → pass, log warning
- **Typo-squatting on names < 3 chars:** skip check
- **Invalid blacklist pattern on startup:** fatal, refuse to start
- **Upstream 4xx/5xx:** passthrough to client unmodified

## 9. Out of Scope (P1+)

- `malfuse link` / `malfuse unlink` CLI automation
- bbolt-backed local cache (P0 uses in-memory TTL)
- Streaming tarball content scan (`io.TeeReader`)
- Rust Cargo / Go module support
- Whitelist (`malfuse allow`)
- Daemon mode
