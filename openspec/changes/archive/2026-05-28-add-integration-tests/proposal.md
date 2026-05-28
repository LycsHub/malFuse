## Why

当前只存在单元测试。缺少覆盖完整代理管道的端到端集成测试：启动真实 HTTP 服务器、通过 SQLite 数据库验证阻断、模拟脚本扫描断流、验证优雅关闭、验证容错行为。

## What Changes

新增端到端集成测试套件，不修改任何生产代码。

## Impact

- 新增文件：`internal/proxy/integration_test.go` 或项目级 `integration_test.go`
- 测试用临时 SQLite 数据库（自动创建/销毁）
- 测试用临时 HTTP 服务器（`httptest`）
- 覆盖 5 个关键场景
