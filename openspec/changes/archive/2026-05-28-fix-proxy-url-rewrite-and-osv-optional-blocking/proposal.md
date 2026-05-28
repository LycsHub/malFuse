## Why

两个遗留问题导致 version-aware blocking 不可用：
1. Pip 拿到 Simple API 响应后，tarball 下载链接指向上游直连 URL，不经过代理。下载阶段的版本匹配无从触发。
2. OSV 检查默认阻断所有有 CVE 的包，导致 `requests`、`pandas` 等合法包无法安装。

## What Changes

1. **URL 重写**：在 `ModifyResponse` 中重写 Simple API HTML 响应内的 tarball 下载链接，把上游 host 替换为代理路径，使 pip 下载也经过代理
2. **OSV 阻断可选**：新增 `block_on_vuln` 配置字段，默认 `false`（仅记录漏洞数，不阻断）

## Impact

- `proxy/proxy.go` — ModifyResponse 增加 URL 重写逻辑
- `config/config.go` — OSVConfig 新增 `BlockOnVuln` 字段
- `config.json` — OSV 配置增加 `block_on_vuln: false`
- `engine/osv.go` — OSVCheck 尊重 `block_on_vuln`
