## 1. Dependencies

- [x] 1.1 Add `github.com/spf13/cobra` dependency

## 2. Config

- [x] 2.1 Add `PIDFile` field to `config.Config`
- [x] 2.2 Update `config.json` with `pid_file` default

## 3. Daemon Logic

- [x] 3.1 Implement `writePIDFile()` / `readPIDFile()` / `removePIDFile()`
- [x] 3.2 Implement `startDaemon()` — `os.StartProcess` 重新执行自身
- [x] 3.3 Implement `stopDaemon()` — SIGTERM → 5s → SIGKILL
- [x] 3.4 Implement `statusDaemon()` — signal 0 检测

## 4. CLI Refactor

- [x] 4.1 Refactor `main.go` to cobra: rootCmd with run, start, stop, status subcommands
- [x] 4.2 `start` subcommand runs daemon flow
- [x] 4.3 `stop` subcommand calls stopDaemon
- [x] 4.4 `status` subcommand calls statusDaemon
- [x] 4.5 `run` (default) subcommand keeps current foreground behavior
- [x] 4.6 Run all existing tests — confirm no regressions
