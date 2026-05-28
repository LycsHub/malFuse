## Why

当前代理仅从 Simple API 路由提取包名，无法获取版本号。结果：DB 中 pandas 只标记了版本 1.1.3 是恶意的，但 `install pandas==2.2.0` 也被阻断。需支持从下载 URL 路径中提取版本号，实现精确版本匹配。

同时修复 NPM scoped package 的 bug：`@scope/pkg/2.0.0` 会把 "/2.0.0" 误解析到包名中。

## What Changes

- `extractPackageInfo` 增加 PyPI 和 NPM 的版本号提取
- PyPI: 从 `<pkg>-<version>.tar.gz` 和 `<pkg>/<version>/` 路径提取
- NPM: 从 `<pkg>/<version>` 和 `<pkg>-<version>.tgz` 路径提取  
- 修复 NPM scoped package 版本污染 bug

## Impact

- 修改: `internal/proxy/proxy.go` — 重写 extractPackageInfo + 新增 extractVersion
- 修改: `internal/proxy/proxy_test.go` — 8 个新测试
