## 1. Project Setup

- [x] 1.1 Create `internal/scanner/` directory
- [x] 1.2 Add `ScriptScanConfig` to `internal/config/config.go`
- [x] 1.3 Update `config.json` with `script_scan` section (default disabled)

## 2. Entropy Detector

- [x] 2.1 Implement Shannon entropy calculation
- [x] 2.2 Implement entropy check with configurable threshold
- [x] 2.3 Unit tests: normal text, encrypted, empty

## 3. Obfuscation Detector

- [x] 3.1 Implement base64 pattern detection (configurable min length)
- [x] 3.2 Implement hex escape pattern detection (configurable min count)
- [x] 3.3 Implement eval/exec call chain detection
- [x] 3.4 Unit tests: base64 match, base64 short, hex match, eval match

## 4. Network Detector

- [x] 4.1 Implement URL extraction and detection
- [x] 4.2 Implement IPv4 address detection with private IP filtering
- [x] 4.3 Unit tests: URL match, IP match, private IP blocked/allowed

## 5. Stream Scanner

- [x] 5.1 Implement target file matching (setup.py, preinstall.js, etc.)
- [x] 5.2 Implement tar/gzip stream extraction with file size limits
- [x] 5.3 Implement zip stream extraction with file size limits
- [x] 5.4 Orchestrate three detectors per extracted file (first-hit blocks)
- [x] 5.5 Implement fail-open on scan errors
- [x] 5.6 Unit tests: tar with setup.py, tar with preinstall.js, corrupt archive

## 6. Engine Integration

- [x] 6.1 Define `StreamChecker` interface and `ScanResult` type in engine
- [x] 6.2 Implement `scanner.StreamChecker` adapting `scanner.Scan` to engine interface

## 7. Proxy Integration

- [x] 7.1 Add `streamChecker` field to `proxy.Handler`
- [x] 7.2 Implement TeeReader stream cloning in `forward()` method
- [x] 7.3 Implement context cancel on StreamCheck BLOCK

## 8. Main Wiring & Config

- [x] 8.1 Wire scanner into main.go (construct, pass to proxy)
- [x] 8.2 Run all existing tests — confirm no regressions
