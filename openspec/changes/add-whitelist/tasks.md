## 1. Database Schema

- [x] 1.1 Add `whitelist` table to `migrate()` in schema.go
- [x] 1.2 Implement `InsertWhitelist(db, name, ecosystem, version)`
- [x] 1.3 Implement `DeleteWhitelist(db, name, ecosystem, version)`
- [x] 1.4 Implement `IsWhitelisted(db, name, ecosystem, version)` — same semantics as Lookup
- [x] 1.5 Unit tests for whitelist CRUD

## 2. Engine Pipeline

- [x] 2.1 Add `Skip bool` field to `engine.Result`
- [x] 2.2 Update `Check()` to stop on `Skip == true`
- [x] 2.3 Implement `WhitelistCheck(db)` returning `{Skip: true, Reason: "whitelist"}`
- [x] 2.4 Unit tests: whitelist match skips downstream, non-match continues

## 3. CLI

- [x] 3.1 Add cobra `allow add/remove/list` subcommands to main.go
- [x] 3.2 Wire DB open in CLI commands

## 4. Main Wiring

- [x] 4.1 Add `WhitelistCheck(malDB)` as first check in main.go pipeline
- [x] 4.2 Run all existing tests — confirm no regressions
