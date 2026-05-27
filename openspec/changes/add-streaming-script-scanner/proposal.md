## Why

恶意包常通过安装脚本（`setup.py`、`postinstall.js` 等）在安装时执行恶意代码。攻击者使用加密/混淆技术隐藏载荷：大段 base64、hex 编码、高信息熵密文、以及内嵌的外连 URL/IP。P0 的元数据检测（包名/版本/OSV）无法发现这类隐蔽攻击。需要在下载流中实时扫描安装脚本内容。

## What Changes

- 新增 `internal/scanner/` 包：流式解压 → 提取安装脚本 → 三检测器分析
- 三检测器（均纯 Go stdlib，可独立开关、可配阈值）：
  - **信息熵检测**：Shannon 熵计算，字节分布异常识别加密/混淆载荷
  - **混淆检测**：base64 大段编码、连续 hex 模式、eval/exec 调用链
  - **外连检测**：URL/IP 正则提取，检测静态嵌入的网络请求
- engine 新增 `StreamChecker` 接口，第 5 号检测位
- proxy 转发时用 `io.TeeReader` 克隆下载流，goroutine 异步扫描
- 检测命中 → 取消 context → 关闭 TCP 连接 → 客户端断流
- 5MB 单文件、50MB 总流硬限防解压炸弹

## Capabilities

### New Capabilities

- `entropy-detector`: Shannon 信息熵计算，可配阈值，超出则阻断
- `obfuscation-detector`: base64/hex 编码检测 + eval/exec 调用链识别
- `network-detector`: 内嵌 URL/IP 正则提取，可配是否允许内网地址
- `stream-scanner`: Tar/gzip/zip 流式解压，目标脚本文件匹配，三检测器编排
- `stream-engine-check`: engine 新增 StreamChecker 接口，proxy TeeReader 集成

### Modified Capabilities

- `engine-pipeline`: 新增第 5 号检测位（script-scan），在 OSV 之后、转发之前
- `proxy-routing`: forward 方法改造为 TeeReader 分流模式，支持异步流式扫描

## Impact

- 新增包：`internal/scanner/`（~5 文件，纯 Go stdlib）
- 新增接口：`engine.StreamChecker`
- 修改：`internal/engine/engine.go`、`internal/proxy/proxy.go`
- 修改：`internal/config/config.go`（新增 `ScriptScanConfig`）
- 修改：`main.go`（构造 scanner、注入 engine 和 proxy）
- 修改：`config.json`（新增 `script_scan` 配置节，默认关闭）
- 无新外部依赖
