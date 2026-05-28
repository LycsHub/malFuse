## Context

P0 有四层元数据检测（malicious-db、cooldown、typo、OSV），但无法发现隐藏在安装脚本中的恶意代码。攻击者常将 payload 用 base64/hex 编码后嵌入 `setup.py` 或 `postinstall.js`。需要在下游流中实时扫描脚本内容。

**约束：** 纯 Go stdlib，零外部依赖；`CGO_ENABLED=0`；不缓冲完整 tarball。

## Goals / Non-Goals

**Goals:**
- 流式解压 tar/gzip/zip，边下边扫
- 三检测器：熵值、混淆、外连
- 可独立开关、可配阈值
- Fail-open：扫描异常不阻断正常安装
- 5MB 单文件、50MB 总流防解压炸弹

**Non-Goals:**
- AST 语法树分析
- YARA/ClamAV 规则引擎
- 二进制文件逆向
- 全量文件扫描（仅安装脚本）

## Decisions

**1. `io.TeeReader` + goroutine 异步扫描**
- Chosen: `io.TeeReader(resp.Body, pw)` 克隆流，主路立即转发客户端，旁路 goroutine 扫描。
- Rationale: 零延迟。正常包不受任何影响。恶意包在管道中检测到后通过 context cancel 断流。
- Alternative: 先下载完再扫再转发。Rejected — 引入巨大延迟。

**2. `io.Pipe` + context 断流机制**
- Chosen: TeeReader 写端是 `io.Pipe`，读端给 scanner goroutine。检测命中 → `cancel()` → `io.Copy` 感知 context → 断开客户端。
- Rationale: context 是 Go 标准机制的惯用做法。

**3. 纯 Go Shannon 熵实现**
- Chosen: 自实现 15 行 Shannon 熵。字节频率统计 → `-Σ p·log₂(p)`。
- Rationale: 算法极简，不需要三方库。后期可升级为卡方分布。

**4. Scanner 独立于 engine，通过接口桥接**
- Chosen: `internal/scanner/` 独立包，engine 定义 `StreamChecker` 接口，
  scanner 实现它。engine 不依赖 scanner 的具体实现。
- Rationale: 隔离关注点。scanner 可独立测试。

## Risks / Trade-offs

- **[Risk] 解压炸弹 OOM** → Mitigation: `io.LimitReader` 硬限 5MB/文件、50MB/总流
- **[Risk] TeeReader 增加少量 CPU 开销** → Mitigation: 默认关闭，opt-in
- **[Trade-off] 异步扫描可能在检测到之前已部分写入客户端** → Mitigation: context cancel 立即断开 TCP，剩余部分不会到达客户端
- **[Trade-off] 高熵阈值可能误判正常压缩文件** → Mitigation: 仅扫描安装脚本文件（非 `.whl`/`.tar.gz` 整体），可配阈值
