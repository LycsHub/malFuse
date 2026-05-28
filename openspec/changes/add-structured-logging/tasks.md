## 1. Logger Package

- [x] 1.1 Add `github.com/sirupsen/logrus` dependency
- [x] 1.2 Implement `internal/logger/` with Init/Info/Warn/Error/Fatal/Debug
- [x] 1.3 Add unit tests for level filtering, JSON output, file output

## 2. Config Integration

- [x] 2.1 Add `LoggingConfig` to `internal/config/config.go`
- [x] 2.2 Update `config.json` with `logging` section (default level=info, format=text, output=stdout)

## 3. Migration

- [x] 3.1 Replace all `log.Fatalf` → `logger.Fatal` (main.go, cmd/malfuse-db)
- [x] 3.2 Replace all `log.Printf("[BLOCKED]")` → `logger.Warn(...)` with fields
- [x] 3.3 Replace all `log.Printf("[WARN]")` → `logger.Warn(...)` (12 instances)
- [x] 3.4 Replace all `log.Printf("[ERROR]")` → `logger.Error(...)` (3 instances)
- [x] 3.5 Replace all info-level logs → `logger.Info(...)` (5 instances)
- [x] 3.6 Initialize logger in main.go and cmd/malfuse-db from config
- [x] 3.7 Run all tests — confirm no regressions
