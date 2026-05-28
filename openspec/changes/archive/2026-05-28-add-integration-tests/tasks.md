## 1. Full Pipeline Tests

- [x] 1.1 Test proxy blocks malicious package found in SQLite (403 + reason)
- [x] 1.2 Test proxy passes safe package through to upstream
- [x] 1.3 Test proxy skips DB check gracefully when DB file missing

## 2. Streaming Scanner Integration

- [x] 2.1 Test proxy blocks tarball with malicious preinstall in package.json
- [x] 2.2 Test proxy passes tarball with clean package.json

## 3. Lifecycle & Error Handling

- [x] 3.1 Test graceful shutdown (SIGTERM → server stops within 5s)
- [x] 3.2 Test unmatched route returns 502
- [x] 3.3 Test config loading from file
