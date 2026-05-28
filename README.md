# 📖 MalFuse — 恶意包物理防火墙

`MalFuse` 是一款基于 **Go 语言** 开发的本地 HTTP 代理，在 `pip install` / `npm install` 等包安装过程中，于恶意代码落地前实施 **Inline Block（流式阻断）**。名字取自 **Mal**icious + **Fuse**（保险丝/熔断）。

---

## 当前状态

P0 + P1 已全部完成，P2 进行中：

| 组件 | 说明 |
|------|------|
| `malfuse` | 本地 HTTP 代理（127.0.0.1:8080），5 层检测管道 + daemon + 健康检查 + 白名单 |
| `malfuse-db` | 恶意包数据库管理 CLI，从 [ossf/malicious-packages](https://github.com/ossf/malicious-packages) 拉取 OSV 数据存入 SQLite |

**数据库规模：** 252,637 条恶意包记录，覆盖 11 个生态系统。

---

## 快速开始

```bash
# 1. 生成恶意包数据库（需 git + 网络）
./malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 2. 配置包管理器（一键）
./malfuse link

# 3. 启动代理
./malfuse start                 # 后台 daemon 模式
./malfuse -config config.json   # 或前台直接运行

# 4. 安装依赖（走代理）
pip install requests
npm install lodash
```

---

## CLI 命令

```bash
malfuse start                   # 后台启动 daemon
malfuse stop                    # 停止 daemon
malfuse status                  # 查看运行状态
malfuse link [--target pip|npm] # 配置包管理器
malfuse unlink [--target pip]   # 还原包管理器配置
malfuse allow add <pkg> --ecosystem pypi [--version 1.0]  # 添加白名单
malfuse allow remove <pkg> --ecosystem npm                 # 移除白名单
malfuse allow list [--ecosystem pypi]                      # 查看白名单
malfuse -config config.json      # 前台直接运行
```

---

## 检测管道（顺序执行，任一命中即停止）

| # | 检测项 | 数据源 | 失败策略 | 默认 |
|---|--------|--------|----------|------|
| 0 | **白名单** | SQLite whitelist 表 | — | 启用 |
| 1 | **恶意包数据库** | SQLite（252,637 条） | 无 DB 则跳过 | 启用 |
| 2 | **安全冷却期** | 上游 Registry 元数据时间戳 | fail-closed（阻断） | 关闭 |
| 3 | **Typo-Squatting** | 内嵌 2790 流行包 + Levenshtein | — | 启用 |
| 4 | **OSV API** | 实时漏洞查询 + 内存 TTL 缓存 | fail-open（放行） | 启用 |
| 5 | **流式脚本扫描** | TeeReader 边下边扫 | fail-open（放行） | 关闭 |

管道之外还有 `/health` 健康检查端点（`GET /health → {"status":"ok","db":true,"uptime":"2h34m"}`）。

---

## 包管理器配置

| 工具 | 一键配置 | 手动配置文档 |
|------|---------|-------------|
| pip | `malfuse link --target pip` | — |
| npm | `malfuse link --target npm` | — |
| pnpm | `malfuse link --target pnpm` | — |
| yarn v1 | `malfuse link --target yarn` | — |
| yarn v2+ | — | [docs/yarn-v2-config.md](docs/yarn-v2-config.md) |
| uv | — | [docs/uv-config.md](docs/uv-config.md) |
| poetry | — | [docs/poetry-config.md](docs/poetry-config.md) |
| conda | — | [docs/conda-config.md](docs/conda-config.md) |

还原：`malfuse unlink`

---

## 目录结构

```
malFuse/
├── main.go                    # malfuse 代理入口（cobra CLI）
├── config.json                # 配置文件
├── cmd/
│   └── malfuse-db/            # 数据库管理 CLI 入口
├── internal/
│   ├── config/                # JSON 配置加载
│   ├── proxy/                 # HTTP 代理层（路由、转发、health）
│   ├── engine/                # 检测管道 + 白名单
│   ├── scanner/               # 流式脚本扫描（熵/混淆/外连 + JS/Python 解析）
│   ├── osv/                   # OSV API 客户端 + TTL 缓存
│   ├── logger/                # logrus 结构化日志
│   ├── daemon/                # 后台进程管理（PID、信号）
│   ├── linker/                # 包管理器配置联动
│   └── db/
│       ├── schema/            # SQLite DDL + WAL
│       ├── ingest/            # OSV JSON 解析 + Git 操作
│       └── output/            # 直接写库 / SQL 增量文件
├── docs/                      # 手动配置文档
├── malfuse.db                 # SQLite 数据库
└── ossf-malicious-packages/   # Git 仓库缓存
```

---

## 配置文件 (config.json)

```json
{
  "port": "8080",
  "host": "127.0.0.1",
  "db_path": "malfuse.db",
  "pid_file": "malfuse.pid",
  "repo_proxy": "ghfast.top",
  "logging": { "level": "info", "format": "text", "output": "stdout" },
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.org", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmjs.org", "ecosystem": "npm"}
  ],
  "cooldown": { "enabled": false, "duration": "48h" },
  "typo": { "enabled": true, "threshold": 2 },
  "osv": { "enabled": true, "ttl": "1h", "base_url": "https://api.osv.dev" },
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

| 字段 | 说明 |
|------|------|
| `repo_proxy` | GitHub 加速代理，不填则不使用 |
| `db_path` | DB 不存在时自动跳过数据库检测 |
| `pid_file` | daemon 模式 PID 文件路径 |
| `logging` | 日志级别/格式/输出（text/json，stdout/文件） |
| `script_scan` | 流式脚本扫描（默认关闭，opt-in） |

---

## 构建

```bash
CGO_ENABLED=0 go build -o malfuse .
CGO_ENABLED=0 go build -o malfuse-db ./cmd/malfuse-db/
```

纯 Go 实现，无 CGo 依赖，交叉编译生成 Linux/macOS/Windows 纯净二进制。

---

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

---

## 技术栈

- **Go 1.19+** stdlib 为主
- **github.com/spf13/cobra** — CLI 框架
- **github.com/sirupsen/logrus** — 结构化日志
- **modernc.org/sqlite** — 纯 Go SQLite，零 CGo
- **net/http/httputil.ReverseProxy** — HTTP 代理转发
- **OSV Schema 1.5.0** — 恶意包报告格式
- **Levenshtein 编辑距离** — Typo-Squatting 检测
