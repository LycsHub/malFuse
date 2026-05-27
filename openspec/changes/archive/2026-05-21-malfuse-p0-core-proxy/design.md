## Context

The malFuse project is at zero-code state (stub `main.go`). P0 builds the core proxy skeleton: intercept pip/npm install traffic, inspect packages, forward or block. The README describes a 3-phase roadmap; this design covers Phase 0 only.

**Constraints:**
- `CGO_ENABLED=0` — fully static, cross-compilable binaries
- Zero external dependencies (Go stdlib only) for P0
- Must support PyPI (PEP 503 Simple API) and NPM registry protocols
- No persistent storage in P0 (in-memory only)

## Goals / Non-Goals

**Goals:**
- Working HTTP proxy that blocks malicious packages based on configurable checks
- Path-based routing to support multiple registries (PyPI, npm, future ones)
- Sequential check pipeline with per-check enable/disable
- OSV vulnerability API integration with in-memory TTL cache
- Graceful shutdown and clear error handling

**Non-Goals:**
- CLI automation (`malfuse link`/`unlink`) — users configure pip/npm manually
- bbolt persistent storage (P1)
- Streaming tarball content scan (P2)
- Daemon/background mode
- Cargo/Go module support (P2)
- Whitelist (`malfuse allow`) command (P2)

## Decisions

**1. Path-based routing over transparent forward proxy**
- **Chosen:** Map URL prefixes (`/pypi/`, `/npm/`) to upstream registries. Users point `pip --index-url http://127.0.0.1:8080/pypi/` at the proxy.
- **Alternative:** Transparent HTTP forward proxy using CONNECT. Rejected because it requires client proxy settings (HTTP_PROXY), which pip/npm handle inconsistently, and adds TLS MITM complexity.
- **Rationale:** Path-based routing avoids TLS interception entirely. Local traffic is plain HTTP. Proxy issues HTTPS upstream.

**2. Sequential checks over concurrent goroutine engine**
- **Chosen:** Linear `shouldBlock()` that runs checks one-by-one, short-circuiting on first BLOCK.
- **Alternative:** Fan-in goroutine model with channel aggregation. Rejected for P0 as premature optimization — the checks are mostly fast (local map lookups) or network-bound (OSV, cooldown), and the sequential ordering already prioritizes cheapest-first.
- **Rationale:** Simpler to debug, test, and trace. The engine interface (`Check(ctx, Request) → Result`) is unchanged regardless of internal concurrency, so this can be refactored later.

**3. In-memory TTL cache over bbolt for OSV**
- **Chosen:** `map[string]cachedResult` with per-entry expiration, guarded by `sync.RWMutex`.
- **Alternative:** bbolt from day one. Rejected for P0 because it adds a persistent storage dependency without proven need, and the OSV cache is ephemeral (restart = fresh cache is fine).
- **Rationale:** Keeps P0 zero-dependency. bbolt can be added in P1 when local blacklist/whitelist persistence becomes necessary.

**4. Fail-open for remote checks, fail-closed for local checks**
- **Remote (OSV):** If the OSV API is unreachable or times out, the check passes (allows). Rationale: a transient network issue shouldn't break all installs; OSV is advisory, not the primary defense.
- **Local (cooldown):** If metadata can't be parsed or the fetch times out, the check blocks. Rationale: a package whose age can't be verified is inherently suspicious; better to block than to guess.
- **Local (blacklist, typo):** No network dependency — no failure mode to handle.

## Risks / Trade-offs

- **[Risk] OSV fail-open could be exploited** — an attacker on the same network could block OSV API access and bypass the check. Mitigation: OSV is layer 4 of 4; the other three checks (blacklist, cooldown, typo) still apply. A targeted attacker would need to evade all four layers.
- **[Risk] No persistent blacklist updates** — blacklist is static in config.yaml. Mitigation: acceptable for P0; OSV provides dynamic threat intelligence. Persistent blacklist sync (from OSV dump) planned for P1.
- **[Trade-off] Path-based routing requires users to change registry URL** — not just proxy settings. Standard `--index-url` flag works for pip but npm uses `.npmrc`. This is more explicit setup than a transparent proxy but avoids TLS interception complexity.
