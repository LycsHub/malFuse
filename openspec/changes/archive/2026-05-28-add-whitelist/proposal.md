## Why

恶意包数据库有 252,637 条记录，存在误报可能。企业内部私有包、安全研究用途包可能被误判。需提供白名单机制，让用户手动放行特定包，跳过所有后续检测。

## What Changes

- SQLite 新增 `whitelist` 表（name, version, ecosystem）
- engine.Result 新增 `Skip` 字段 — 跳过后续检测但不阻断
- 新增 `WhitelistCheck` 作为管道首位检查
- `malfuse allow add/remove/list` cobra 子命令
- `IsWhitelisted` CRUD 函数

## Capabilities

### Modified Capabilities

- `engine-pipeline`: Result 新增 Skip 字段；Check() 在 Skip=true 时也停止管道

## Impact

- 修改：`internal/engine/engine.go`（Result + Check loop）
- 修改：`internal/db/schema/schema.go`（migrate + 白名单 DDL + CRUD）
- 新增：`internal/engine/whitelist.go`
- 修改：`main.go`（管道首位 + CLI）
