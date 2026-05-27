## Why

Software supply chain attacks via malicious PyPI/NPM packages are surging. Current defenses (audit tools, CI scanners) operate post-download — the tarball already hits disk before inspection. malFuse moves the defense upstream: a local HTTP proxy that blocks malicious packages in-transit, before they ever land on the developer's machine.

## What Changes

- New `malFuse` binary: a foreground HTTP proxy with built-in multi-layered package inspection
- Path-based registry routing: map URL prefixes (e.g., `/pypi/`, `/npm/`) to real upstream registries
- Sequential check pipeline: blacklist → cooldown → typo-squatting → OSV vulnerability lookup
- In-memory TTL cache for OSV API responses (no persistent storage in P0)
- YAML configuration file for routing table, blacklist entries, and check toggles
- Hardcoded 403 Forbidden response when any check triggers a block

## Capabilities

### New Capabilities

- `proxy-routing`: Path-based HTTP reverse proxy that strips prefix and forwards to real package registries over HTTPS
- `blacklist-check`: Static blacklist check against configurable name/version patterns
- `cooldown-check`: Blocks packages published less than 48 hours ago (opt-in, fail-closed)
- `typo-check`: Levenshtein-based typo-squatting detection against top-2000 popular packages
- `osv-check`: Live OSV vulnerability API query with in-memory TTL cache (fail-open)
- `engine-pipeline`: Sequential check orchestration with short-circuit and per-check enable/disable

### Modified Capabilities

<!-- No existing capabilities to modify — greenfield project. -->

## Impact

- New Go binary (`malFuse`) with zero external dependencies beyond Go stdlib
- New packages: `internal/proxy/`, `internal/engine/`, `internal/config/`, `internal/osv/`
- Users configure pip/npm to point registry URL at `http://127.0.0.1:8080/pypi/` or `/npm/`
- No persistent state in P0 (OSV cache is in-memory, cleared on restart)
