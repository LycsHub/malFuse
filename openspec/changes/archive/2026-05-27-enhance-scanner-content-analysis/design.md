## Context

Current scanner (`scanner.go`) uses `isInstallScript()` with hardcoded file names. This misses the primary JS attack vector (`package.json` scripts fields) and multiple Python vectors (`.pth`, `pyproject.toml`).

## Goals / Non-Goals

**Goals:**
- Parse `package.json` → extract `scripts` values → apply 3 detectors
- Track script-referenced `.js` files for secondary scan
- Scan `setup.py`, `__init__.py` full content
- Detect malicious `.pth` lines (import/exec patterns)
- Parse `pyproject.toml` `[build-system]` section
- Remove filename-based `isInstallScript()`

**Non-Goals:**
- Full AST analysis (too heavy for stream scanning)
- Scanning non-executable Python files (`.py` that aren't `setup.py`/`__init__.py`)
- Deep TOML/JSON validation (best-effort parsing, fail-open)

## Decisions

**1. JSON structured parsing over regex extraction**
- Parse `package.json` with `encoding/json`. Extract `scripts` map, iterate values.
- Rationale: Clean, type-safe, no regex edge cases. JSON is the standard format.

**2. TOML manual extraction over full parser**
- Use `regexp` to extract `[build-system]` section and `build-backend` key.
- Alternative: Import `BurntSushi/toml`. Rejected — adds dependency for single-field extraction.

**3. Two-pass scan for JS packages**
- Pass 1: scan `package.json` scripts values. Collect referenced `.js` files.
- Pass 2: scan those `.js` files if present in the archive.
- Rationale: Prevents attackers from hiding payload in separate files.

**4. `.pth` detection**
- Scan each line of `.pth` files. If line starts with `import` or contains `exec`/`eval`, scan with detectors.
- Standard `.pth` lines are just paths (no spaces), so `import` is a strong signal.

## Risks / Trade-offs

- **[Risk] Script-referenced JS files not found in archive** → Mitigation: best-effort, skip silently
- **[Risk] pyproject.toml regex edge cases** → Mitigation: fail-open, log warning
