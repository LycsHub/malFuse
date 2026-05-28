# 📖 MalFuse — 恶意包物理防火墙

`MalFuse` 是一款基于 **Go 语言** 开发的本地 HTTP 代理，在 `pip install` / `npm install` 等包安装过程中，于恶意代码落地前实施 **Inline Block（流式阻断）**。名字取自 **Mal**icious + **Fuse**（保险丝/熔断）。

---

## 当前状态

P0 核心骨架已实现，包含两个二进制：

| 组件 | 说明 |
|------|------|
| `malfuse` | 本地 HTTP 代理（127.0.0.1:8080），拦截安装请求，4 层检测后放行或返回 403 |
| `malfuse-db` | 恶意包数据库管理 CLI，从 [ossf/malicious-packages](https://github.com/ossf/malicious-packages) 拉取 OSV 数据存入 SQLite |

**数据库规模：** 252,637 条恶意包记录，覆盖 11 个生态系统（npm, PyPI, NuGet, RubyGems, Go, Crates.io, Maven, Packagist, VSCode 等）。

---

## 快速开始

```bash
# 1. 生成恶意包数据库（需 git + 网络）
./malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 2. 启动代理
./malfuse -config config.json

# 3. 配置包管理器（一键）
./malfuse link
```

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

## 检测管道（顺序执行，任一命中即阻断）

| # | 检测项 | 数据源 | 失败策略 | 默认 |
|---|--------|--------|----------|------|
| 1 | **恶意包数据库** | SQLite（252,637 条） | 无 DB 则跳过 | 启用 |
| 2 | **安全冷却期** | 上游 Registry 元数据时间戳 | fail-closed（阻断） | 关闭 |
| 3 | **Typo-Squatting** | 内嵌 Top 2000 流行包 + Levenshtein 编辑距离 | — | 启用 |
| 4 | **OSV API** | 实时 OSV 漏洞查询 + 内存 TTL 缓存 | fail-open（放行） | 启用 |

---

## 目录结构

```
malFuse/
├── main.go                    # malfuse 代理入口
├── config.json                # 配置文件（路由、检测开关、db_path、repo_proxy）
├── cmd/
│   └── malfuse-db/            # 数据库管理 CLI 入口
├── internal/
│   ├── config/                # JSON 配置加载
│   ├── proxy/                 # HTTP 代理层（路由匹配、上游转发、health）
│   ├── engine/                # 检测管道（恶意库、冷却期、typo、OSV、流式扫描）
│   ├── osv/                   # OSV API 客户端 + 内存 TTL 缓存
│   ├── scanner/               # 流式脚本扫描（熵值/混淆/外连检测 + JS/Python 结构解析）
│   ├── logger/                # logrus 结构化日志
│   ├── daemon/                # 后台进程管理（PID、信号）
│   ├── linker/                # 包管理器配置联动（pip/npm/pnpm/yarn）
│   └── db/
│       ├── schema/            # SQLite DDL + CRUD（WAL 模式）
│       ├── ingest/            # OSV JSON 解析 + Git 操作
│       └── output/            # 直接写库 / 生成 SQL 增量文件
├── malfuse.db                 # SQLite 恶意包数据库（由 malfuse-db 生成）
└── ossf-malicious-packages/   # Git 仓库缓存
```

---

## malfuse-db 命令

```bash
# 直接写入 SQLite（在线增量更新）
./malfuse-db --mode direct --db malfuse.db --repo ossf-malicious-packages

# 生成 SQL 增量文件（离线使用）
./malfuse-db --mode sql --repo ossf-malicious-packages
```

**增量机制：** 首次全量扫描，后续通过 `git fetch` + `git diff --name-status` 仅处理变更文件。

---

## 配置文件 (config.json)

```json
{
  "port": "8080",
  "host": "127.0.0.1",
  "db_path": "malfuse.db",
  "repo_proxy": "ghfast.top",
  "routing": [
    {"prefix": "/pypi/", "upstream": "https://pypi.org", "ecosystem": "pypi"},
    {"prefix": "/npm/", "upstream": "https://registry.npmjs.org", "ecosystem": "npm"}
  ],
  "cooldown": {"enabled": false, "duration": "48h"},
  "typo": {"enabled": true, "threshold": 2},
  "osv": {"enabled": true, "ttl": "1h", "base_url": "https://api.osv.dev"}
}
```

- `repo_proxy`: GitHub 加速代理（如 `ghfast.top`），不填则不使用代理
- `db_path`: SQLite 数据库路径，不存在时自动跳过数据库检测
- `pid_file`: daemon 模式 PID 文件路径（默认 `malfuse.pid`）
- `logging`: 日志配置（level/format/output）
- `script_scan`: 流式脚本扫描配置（默认关闭，opt-in）
- 路由支持多 Registry，按 URL 前缀匹配

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
- [x] 4 层检测管道（malicious-db / cooldown / typo / OSV）
- [x] 流式脚本扫描器（熵值 + 混淆 + 外连检测）
- [x] package.json scripts 解析 + Python .pth / pyproject.toml 检测

### 🟡 P1 — 自动化与运维（已完成）

- [x] `malfuse link` / `malfuse unlink` 自动配置 pip/npm/pnpm/yarn
- [x] logrus 结构化日志（级别控制、JSON 格式、文件输出）
- [x] `/health` 健康检查端点（DB 状态 + uptime）
- [x] 后台 daemon 模式（`malfuse start/stop/status`）
- [x] 端到端集成测试套件

### 🟢 P2 — 深度扫描与生态扩展

- [ ] 更多生态路由（RubyGems、NuGet、Crates.io、Go modules）
- [ ] `malfuse allow` 白名单管理
- [ ] Docker 镜像分发
- [ ] CI/CD Pipeline（lint、test、build、release）
- [ ] 安装脚本 AST 语法分析

---

## 技术栈

- **Go 1.19+** stdlib 为主
- **modernc.org/sqlite** — 纯 Go SQLite，零 CGo
- **OSV Schema 1.5.0** — 恶意包报告格式
- **Levenshtein 编辑距离** — Typo-Squatting 检测
- **`net/http/httputil.ReverseProxy`** — HTTP 代理转发
