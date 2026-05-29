# 📖 MalFuse — 恶意包物理防火墙

[English](README.md) | 中文

`MalFuse` 是一款基于 **Go 语言** 开发的本地 HTTP 代理，在 `pip install` / `npm install` 等包安装过程中，于恶意代码落地前实施 **Inline Block（流式阻断）**。名字取自 **Mal**icious + **Fuse**（保险丝/熔断）。

P0 + P1 已全部完成，P2 进行中（白名单已实现）。252,637 条恶意包记录，覆盖 11 个生态系统。

---

# 使用

## 快速开始

```bash
# 1. 生成恶意包数据库
./malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 2. 配置包管理器
./malfuse link

# 3. 启动代理
./malfuse start            # 后台 daemon 模式
# 或
./malfuse -config config.json  # 前台直接运行

# 4. 正常安装依赖（流量自动过代理）
pip install requests
npm install lodash
```

---

## CLI 命令

### `malfuse` — 代理 + 管理

```bash
# 运行
malfuse                         # 前台运行（默认 config.json）
malfuse -config /path/to/config.json  # 指定配置文件
malfuse start                   # 后台 daemon 启动
malfuse stop                    # 停止 daemon
malfuse status                  # 查看运行状态

# 包管理器配置
malfuse link                    # 配置所有已安装的包管理器
malfuse link --target pip       # 仅配置 pip
malfuse link --target npm       # 仅配置 npm
malfuse link --target pnpm      # 仅配置 pnpm
malfuse link --target yarn      # 仅配置 yarn v1
malfuse unlink                  # 还原所有
malfuse unlink --target pip     # 仅还原 pip

# 白名单管理
malfuse allow add requests --ecosystem pypi              # 放行所有版本
malfuse allow add lodash --ecosystem npm --version 4.17.21  # 放行特定版本
malfuse allow remove requests --ecosystem pypi             # 移除白名单
malfuse allow list                                         # 查看全部
malfuse allow list --ecosystem npm                         # 按生态过滤
```

### `malfuse-db` — 数据库管理

```bash
# 直接写入 SQLite（在线使用）
malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 生成 SQL 增量文件（离线/air-gapped 环境使用）
malfuse-db --mode sql --output updates-20260527.sql --repo ossf-malicious-packages

# 指定配置（读取 repo_proxy 用于 GitHub 加速）
malfuse-db --config config.json --mode direct
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--mode` | `direct` | `direct` 直接写 SQLite / `sql` 生成增量 SQL 文件 |
| `--db` | `malfuse.db` | SQLite 数据库路径 |
| `--repo` | `ossf-malicious-packages` | git 仓库缓存目录 |
| `--output` | `updates-YYYYMMDDHHmm.sql` | SQL 输出路径（仅 sql 模式） |
| `--config` | `config.json` | 配置文件路径（读取 `repo_proxy` 代理设置） |

**更新机制：** 首次运行执行全量扫描（克隆 + 解析所有 OSV JSON）。后续运行通过 `git fetch` + `git diff --name-status` 仅处理增量变更，秒级完成。

**输出说明：** `direct` 模式写入 `malfuse.db`（WAL 模式，可被代理同时读取）。`sql` 模式生成 `INSERT OR REPLACE` + `DELETE` 语句，可导入到无网络环境的数据库。

---

## 包管理器配置

### 一键配置（`malfuse link`）

| 工具 | 命令 | 实际执行的配置 |
|------|------|---------------|
| pip | `malfuse link --target pip` | `pip config set global.index-url http://127.0.0.1:8080/pypi/simple/` |
| npm | `malfuse link --target npm` | `npm config set registry http://127.0.0.1:8080/npm/` |
| pnpm | `malfuse link --target pnpm` | `pnpm config set registry http://127.0.0.1:8080/npm/` |
| yarn v1 | `malfuse link --target yarn` | `yarn config set registry http://127.0.0.1:8080/npm/` |

`malfuse link` 会备份原始值到 `~/.malfuse_backup.json`，`malfuse unlink` 从备份还原全部配置。

### 手动配置

以下工具不支持 CLI 配置命令，需手动操作：

| 工具 | 文档 |
|------|------|
| yarn v2+ (Berry) | [docs/yarn-v2-config.md](docs/yarn-v2-config.md) |
| uv | [docs/uv-config.md](docs/uv-config.md) |
| poetry | [docs/poetry-config.md](docs/poetry-config.md) |
| conda | [docs/conda-config.md](docs/conda-config.md) |

---

## 阻断粒度说明

不同包管理器的代理行为不同，因此版本匹配的精度也不同：

| 生态 | 粒度 | 机制 |
|------|------|------|
| **PyPI (pip)** | 版本精确 | Simple API 仅阻断 `version=NULL` 的条目。代理自动重写 HTML 中的下载链接，pip 下载 tarball 时返回代理，触发精确版本匹配 |
| **npm** | 包级别（全阻断） | npm 的 `pacote` 下载库绕过代理缓存，无法在下载阶段精确匹配。仅 `version=NULL` 的条目在 Simple API 阶段 403 阻断 |

### npm 版本精确阻断方案

如果需要 npm 版本级别的控制，使用白名单配合全阻断：

```bash
# 1. 标记包为全阻断（version=NULL）
sqlite3 malfuse.db "INSERT INTO malicious_packages VALUES ('bad-lib', NULL, 'npm', '', '');"

# 2. 将安全版本加入白名单
./malfuse allow add bad-lib --ecosystem npm --version 2.0.5
```

这样 `npm install bad-lib@2.0.5` 正常通过，其他版本全部阻断。

### 阻断行为对照表

| 场景 | pip | npm |
|------|-----|-----|
| DB 中 `version=NULL` | 所有版本阻断 | 所有版本阻断 |
| DB 中 `version="1.0"`, 安装 1.0 | 阻断（下载时精确匹配） | 阻断（全阻断，除非白名单放行） |
| DB 中 `version="1.0"`, 安装 2.0 | 放行 | 阻断（全阻断，除非白名单放行） |

---

## 检测管道

每个安装请求依次通过 6 层检测（含白名单），任一命中即停止后续检查：

| # | 检测项 | 数据源 | 结果 | 默认 |
|---|--------|--------|------|------|
| 0 | **白名单** | SQLite `whitelist` 表 | 命中 → 直接 PASS，跳过所有后续检测 | 启用 |
| 1 | **恶意包数据库** | SQLite `malicious_packages`（252,637 条） | 命中 → 403 Forbidden | 启用 |
| 2 | **安全冷却期** | `malicious_packages.published`（OSV 报告时间戳） | 报告时间 < 48h → 403 | 关闭 |
| 3 | **Typo-Squatting** | 内嵌 2790 流行包 + Levenshtein 编辑距离 | 名字相似 → 403 | 关闭 |
| 4 | **OSV 漏洞 API** | `api.osv.dev/v1/query` + 内存 TTL 缓存 | 命中且 `block_on_vuln=true` → 403 | 启用（不阻断） |
| 5 | **流式脚本扫描** | TeeReader 边下边扫（tar/gzip 解包） | 恶意脚本 → 中断连接 | 关闭 |

**故障策略：**
- 白名单/typo — 无故障可能（纯内存/SQLite）
- 恶意包DB — DB 缺失或损坏时跳过（日志 WARN）
- 冷却期 — 无 DB 或缺少 `published` 字段时跳过；依赖 DB 查询，无额外网络请求
- OSV API — 网络不可达则放行（fail-open）；`block_on_vuln=false` 时仅日志记录漏洞数
- 脚本扫描 — 解包/解析错误则放行（fail-open）

**注意：** 冷却期和 typo-squatting **默认关闭**，因误报风险和稳定性考虑。启用前请评估影响。

**脚本扫描（#5）覆盖的投毒向量：**

| JS 生态 | Python 生态 |
|---------|------------|
| `package.json` `scripts.preinstall/postinstall/install` | `setup.py` 全文 |
| `package.json` `scripts` 中引用的 `.js` 文件 | `__init__.py` 全文 |
| 独立 `.js` 文件 | `.pth` 文件 `import` 行 |
| — | `pyproject.toml` `build-system.build-backend` |

每个向量经三检测器判定：**Shannon 信息熵**（阈值 4.5）、**代码混淆**（base64/hex/eval 链）、**外连检测**（URL/IP）。

---

## /health 健康检查

```bash
$ curl http://127.0.0.1:8080/health
{"db":true,"status":"ok","uptime":"2h34m5s"}
```

| 字段 | 说明 |
|------|------|
| `status` | `ok`（正常）或 `degraded`（DB 不可用） |
| `db` | SQLite 连接是否正常 |
| `uptime` | 代理进程运行时长 |

---

## 配置文件 (config.json)

```json
{
  "port": "8080",
  "host": "127.0.0.1",
  "db_path": "malfuse.db",
  "pid_file": "malfuse.pid",
  "repo_proxy": "ghfast.top",
  "logging": {
    "level": "info",
    "format": "text",
    "output": "stdout"
  },
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.tuna.tsinghua.edu.cn", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmmirror.com", "ecosystem": "npm"}
  ],
  "cooldown": {
    "enabled": false,
    "duration": "48h"
  },
  "typo": {
    "enabled": true,
    "threshold": 2
  },
  "osv": {
    "enabled": true,
    "block_on_vuln": false,
    "ttl": "1h",
    "base_url": "https://api.osv.dev"
  },
  "script_scan": {
    "enabled": false,
    "max_file_size": 5242880,
    "max_total_size": 52428800,
    "entropy": { "enabled": true, "threshold": 4.5 },
    "obfuscation": { "enabled": true, "base64_min_length": 100, "hex_min_length": 20 },
    "network": { "enabled": true, "allow_private_ips": false }
  }
}
```

**配置项说明：**

| 配置节 | 字段 | 说明 |
|--------|------|------|
| 基础 | `port` / `host` | 代理监听地址 |
| 基础 | `db_path` | SQLite 路径，不存在时自动跳过数据库检测 |
| 基础 | `pid_file` | daemon 模式 PID 文件路径（默认 `malfuse.pid`） |
| 基础 | `repo_proxy` | GitHub 加速代理域名（如 `ghfast.top`），不填不代理 |
| `logging` | `level` | `debug` / `info` / `warn` / `error` |
| `logging` | `format` | `text` 或 `json`（JSON 适用于日志采集系统） |
| `logging` | `output` | `stdout` 或文件路径（文件模式下同时输出 stdout + 文件） |
| `routing` | `prefix` | URL 前缀，匹配请求路径 |
| `routing` | `upstream` | 真实 Registry URL（代理内部走 HTTPS） |
| `routing` | `ecosystem` | 生态标识（`pypi` / `npm`，用于 DB + OSV 查询） |
| `cooldown` | `duration` | 冷却期时长，距 OSV 报告发布小于此值则阻断 |
| `cooldown` | `enabled` | 默认关闭，需显式开启 |
| `typo` | `threshold` | Levenshtein 编辑距离 ≤ 此值则阻断 |
| `osv` | `block_on_vuln` | 发现漏洞时是否阻断（默认 `false`，仅日志记录） |
| `osv` | `ttl` | 查询结果缓存时间 |
| `osv` | `base_url` | OSV API 地址 |
| `script_scan` | `enabled` | 默认关闭，需显式开启 |
| `script_scan` | `max_file_size` | 单文件分析上限（字节），超出则跳过 |
| `script_scan` | `max_total_size` | 流总大小上限，超出则停止扫描 |
| `script_scan.entropy` | `threshold` | Shannon 熵阈值（4.5 = 英文文本上限附近） |
| `script_scan.obfuscation` | `base64_min_length` | 触发检测的 base64 字符串最小长度 |
| `script_scan.obfuscation` | `hex_min_length` | 触发检测的连续 `\xNN` 最小次数 |
| `script_scan.network` | `allow_private_ips` | 是否放行内网 IP（如 `10.x`、`192.168.x`） |

---

## 构建

```bash
CGO_ENABLED=0 go build -o malfuse .
CGO_ENABLED=0 go build -o malfuse-db ./cmd/malfuse-db/
```

纯 Go，零 CGo 依赖，一处编译可在 Linux / macOS (Intel + Apple Silicon) / Windows 运行。

---

# 开发

## 目录结构

```
malFuse/
├── main.go                    # malfuse 代理入口（cobra CLI）
├── config.json                # 配置文件
├── cmd/
│   └── malfuse-db/            # 数据库管理 CLI 入口
├── internal/
│   ├── config/                # JSON 配置加载 + 验证
│   ├── proxy/                 # HTTP 代理层（路由匹配、上游转发、health）
│   ├── engine/                # 检测管道（白名单、恶意库、冷却期、typo、OSV） + StreamChecker 接口
│   ├── scanner/               # 流式脚本扫描（熵值/混淆/外连 + JS/Python 结构解析）
│   ├── osv/                   # OSV API 客户端 + 内存 TTL 缓存
│   ├── logger/                # logrus 结构化日志封装
│   ├── daemon/                # 后台进程管理（PID 文件、信号）
│   ├── linker/                # 包管理器配置联动（pip/npm/pnpm/yarn）
│   └── db/
│       ├── schema/            # SQLite DDL + CRUD（WAL 模式，DBExec 接口）
│       ├── ingest/            # OSV JSON 1.5.0 解析 + Git 操作
│       └── output/            # 直接写库 / SQL 增量文件生成
├── docs/                      # 手动配置文档
├── malfuse.db                 # SQLite 数据库（由 malfuse-db 生成）
└── ossf-malicious-packages/   # Git 仓库缓存（gitignore）
```

## 运行测试

```bash
# 所有单元测试 + 集成测试
go test ./internal/...

# 指定包
go test -v ./internal/scanner/
go test -v ./internal/engine/
```

当前 110+ 个测试，覆盖全部包。

## 实现路线 (Roadmap)

### ✅ P0 — 核心骨架（已完成）

- [x] HTTP 反向代理 + 路由匹配
- [x] 恶意包 SQLite 数据库（252,637 条，11 个生态系统）
- [x] `malfuse-db` CLI（git 增量爬取 + SQL 离线模式）
- [x] 检测管道（malicious-db / cooldown / typo / OSV）
- [x] 流式脚本扫描器（熵值 / 混淆 / 外连 + JS/Python 结构解析）

### ✅ P1 — 自动化与运维（已完成）

- [x] `malfuse link` / `malfuse unlink`（pip / npm / pnpm / yarn）
- [x] logrus 结构化日志（级别控制、JSON 格式、文件输出）
- [x] `/health` 健康检查端点
- [x] 后台 daemon 模式（`malfuse start/stop/status`）
- [x] 端到端集成测试套件

### 🟢 P2 — 深度扫描与生态扩展

- [x] `malfuse allow` 白名单管理
- [ ] 更多生态路由（RubyGems、NuGet、Crates.io、Go modules）
- [ ] Docker 镜像分发
- [ ] CI/CD Pipeline（lint、test、build、release）
- [ ] 安装脚本 AST 语法分析

## 技术栈

| 组件 | 库 |
|------|-----|
| CLI 框架 | `github.com/spf13/cobra` |
| 结构化日志 | `github.com/sirupsen/logrus` |
| SQLite | `modernc.org/sqlite`（纯 Go，零 CGo） |
| HTTP 代理 | `net/http/httputil.ReverseProxy`（stdlib） |
| 恶意包数据格式 | [OSV Schema 1.5.0](https://ossf.github.io/osv-schema/) |
| Typo 检测 | 自实现 Levenshtein 编辑距离 |
| 信息熵 | 自实现 Shannon Entropy |
| 混淆检测 | regexp（stdlib） |
