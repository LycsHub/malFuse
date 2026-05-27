## Why

当前 script scanner 仅按文件名匹配（`setup.py`、`preinstall.js` 等），无法覆盖完整的恶意包投毒攻击面。实际攻击中：

- **JS**: 恶意代码藏在 `package.json` 的 `scripts` 字段（`preinstall`、`postinstall`、`install`、`prepare`），shell 命令在 `npm install` 时自动执行。仅匹配 `preinstall.js` 文件名会漏掉直接嵌入 JSON 的命令。
- **Python**: `.pth` 文件静默执行、`pyproject.toml` 的 `build-backend` 钩子、`__init__.py` 导入时执行都是常用投毒向量，当前完全不检测。

需要从「文件名匹配」升级为「结构化内容分析」，覆盖 JS 和 Python 生态的核心攻击面。

## What Changes

- 新增 `js_analyzer.go`：结构化解析 `package.json`，提取 `scripts` 对象所有字段值，逐一检测；如引用 `.js` 文件则额外提取并扫描
- 新增 `python_analyzer.go`：检测 `setup.py`/`__init__.py` 全文、`.pth` 文件 import 行、`pyproject.toml` 的 `build-backend`
- 删除 `isInstallScript()` 文件名匹配 → 改为按文件路径/类型分发到对应分析器
- 三检测器（熵/混淆/外连）保持不变，复用现有逻辑
- `ScanConfig` 保持不变

## Capabilities

### Modified Capabilities

- `stream-scanner`: **BREAKING** — `isInstallScript` filename matching removed. Replaced with structured content analysis via `js_analyzer.go` and `python_analyzer.go`.

## Impact

- 新增：`internal/scanner/js_analyzer.go`、`internal/scanner/python_analyzer.go`
- 重写：`internal/scanner/scanner.go`（删除 isInstallScript，改为按文件类型分发）
- 删除：文件名硬编码匹配逻辑
- 新增测试：package.json scripts 提取、.pth 检测、pyproject.toml build-backend 检测
- 无新外部依赖（JSON 用 `encoding/json`，TOML 用手动正则提取）
