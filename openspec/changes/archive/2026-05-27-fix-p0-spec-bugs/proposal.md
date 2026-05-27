## Why

P0 的 4 个遗留问题与已批准的 spec 不一致，需要修复以保持代码与规范的一致性：

1. **malFuse 代理以读写模式打开 SQLite** — spec 要求 `mode=ro`
2. **`packages.txt` 运行时读取而非编译时嵌入** — spec 要求 `//go:embed`
3. **流行包列表仅 ~119 条** — spec 要求 Top 2000
4. **`update_state.last_updated` 未填充** — spec 要求记录更新时间

## What Changes

- 代理通过新增的 `OpenReadOnly()` 以只读模式打开 SQLite（跳过 WAL/migrate）
- `packages.txt` 扩展至 ~2000 条流行包，通过 `//go:embed` 编译嵌入
- `savePrevCommit` 填充 `last_updated` 为当前时间
- 无需修改 spec（代码修复以符合现有 spec）

## Capabilities

### Modified Capabilities

- `malicious-db-check`: 代理以真正的只读模式打开数据库
- `typo-check`: popular packages list embedded at build time via `//go:embed`
- `git-incremental-update`: `last_updated` timestamp properly set

## Impact

- `internal/db/schema/schema.go` — 新增 `OpenReadOnly()` 函数
- `internal/engine/packages.txt` — 扩展至 ~2000 条
- `internal/engine/typo.go` — 使用 `//go:embed` 替代 `os.ReadFile`
- `main.go` — 使用 `schema.OpenReadOnly()` 替代 `schema.Open()`
- `internal/db/output/update.go` — `savePrevCommit` 传递时间戳
