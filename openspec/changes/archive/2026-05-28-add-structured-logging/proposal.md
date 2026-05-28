## Why

目前全项目使用 Go 标准库 `log.Printf`/`log.Fatalf`（26 处），无日志级别、无结构化字段、无格式切换。生产环境中无法按级别过滤（调试 vs 阻断事件），无法输出 JSON 格式供日志系统采集。需要替换为 logrus 结构化日志。

## What Changes

- 新增 `internal/logger/` 包，包装 `github.com/sirupsen/logrus`
- 全项目 26 处 `log.*` 替换为 `logger.Info/Warn/Error/Fatal`
- `config.json` 新增 `logging` 配置节（level/format/output）
- 日志调用增加结构化字段（package、reason、ecosystem）
- 引入 1 个外部依赖（logrus）

## Capabilities

### New Capabilities

- `structured-logging`: Configurable logrus wrapper with level filtering, text/JSON format, stdout/file output

## Impact

- 新增：`internal/logger/logger.go`
- 修改：8 个文件共 26 处日志调用
- 修改：`config.json`（新增 logging 节）
- 修改：`internal/config/config.go`（新增 LoggingConfig）
- 新增依赖：`github.com/sirupsen/logrus`
