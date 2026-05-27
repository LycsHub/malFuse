# 📖 MalFuse 开源投毒阻断工具架构与实现路径设计文档

`MalFuse` 是一款基于 **Go 语言** 开发的轻量级、零配置本地“恶意包物理防火墙”。它通过在本地建立高性能 HTTP 代理，在开发人员或 CI/CD 环境执行 `pip/npm install` 时，于恶意代码（Tarball/Wheel）落地解压前实施 **Inline Block（流式强阻断）**。其名字取自 **Mal**icious（恶意）与 **Fuse**（保险丝/熔断），寓意为软件供应链筑起一道触毒即断的“安全保险丝”。

---

## 一、 系统架构设计 (System Architecture)

系统采用 **“前端 CLI 劫持 + 本地代理层（核心）+ 混合多模态分析引擎”** 的扁平化、无状态架构。

### 1. 架构拓扑图

```
                   [ 终端输入: pip/uv/npm install ]
                                  │
                                  ▼ (环境变量/配置 劫持)
┌──────────────────────────────────────────────────────────────────┐
│ 1. 流量劫持与路由层 (Traffic Interception & Proxy)                │
│    • 监听本地 http://127.0.0.1:8080                              │
│    • 自动解析 NPM Registry 协议 与 PyPI Simple (PEP 503) 协议    │
└─────────────────────────────────┬────────────────────────────────┘
                                  │ (并发解包/抽取: 包名, 版本, 压缩包流)
                                  ▼
┌──────────────────────────────────────────────────────────────────┐
│ 2. Go 高并发决策引擎 (Goroutine Decision Engine)                 │
│    • 利用 Channel 收集本地静态与云端情报的分数                    │
└───────────────────────┬───────────────────┬──────────────────────┘
                        │                   │
                        ▼                   ▼
┌────────────────────────────────┐ ┌────────────────────────────────┐
│ 3A. 本地静态研判 (Local Engine) │ │ 3B. 威胁情报与时间戳 (Threat)  │
├────────────────────────────────┤ ├────────────────────────────────┤
│ • 仿冒包检测 (Rune编辑距离)     │ │ • 动态安全冷却期 (Cooldown)    │
│ • 流式 Tarball 解压与指纹匹配   │ │ • 本地 bbolt 已知黑名单布隆过滤│
│   (TeeReader 边下载边扫描)     │ │   (异步对接 OpenSSF OSV API)   │
└───────────────────────┬────────┘ └────────────────┬───────────────┘
                        │                           │
                        └─────────────┬─────────────┘
                                      │ (任意一环命中即触发熔断)
                                      ▼
                        / \
                       /     \
                     /  放行?  \
                     \         /
                       \     /
                         \ /
                          │
             ┌────────────┴────────────┐
             │                         │
          ▼ YES                     ▼ NO (触发熔断)
┌───────────────────────────┐   ┌───────────────────────────┐
│ 向原生包管理器回传数据流  │   │ 立即中断连接，返回 HTTP 403│
│ 依赖无感知安装成功        │   │ 终端高显红字打印安全拦截报告│
└───────────────────────────┐   └───────────────────────────┘

```

### 2. 核心 Go 模块划分 (Go Module Layout)

为了保持 `CGO_ENABLED=0` 纯净交叉编译分发，项目严格杜绝引入需要编译 C 依赖的库。

```
malfuse/
├── cmd/
│   └── malfuse/          # 唯一主程序入口 (main.go)
├── internal/
│   ├── cli/              # CLI 命令行框架 (基于 spf13/cobra 驱动)
│   ├── proxy/            # 本地 HTTP 反向代理层 (基于 httputil.ReverseProxy)
│   ├── engine/           # 核心算法（编辑距离计算、时间戳冷却期判断、流式扫描）
│   └── storage/          # 本地 KV 缓存库 (纯 Go 实现的 etcd-io/bbolt)
└── pkg/
    └── osv/              # 封装开源 OpenSSF OSV 情报接口的客户端 SDK

```

---

## 二、 关键技术选型与难点攻克

### 1. 强阻断的无摩擦实现：明文代理转密文上游

* **痛点**：传统网络拦截工具需要生成本地根证书（CA）并强迫用户信任，开发推广阻力极大。
* **MalFuse 方案**：
  `malfuse` 本地监听 `[http://127.0.0.1:8080](http://127.0.0.1:8080)`（明文 HTTP）。通过 CLI 自动修改依赖源配置，引导包管理器发送明文 HTTP 给本地代理。本地代理服务在请求外网（如阿里云、官方源）时，**在代码内部强制发起强加密的 HTTPS 请求**。
> **效果**：本地零证书摩擦，公网数据传输依然绝对安全。



### 2. 不落地流式预审（Streaming Tarball Scan）

* **难点**：包管理器在安装时会并发下载几十个压缩包。如果把包完全下载到磁盘或内存再扫描，会引入巨大的安装滞后感。
* **MalFuse 方案**：
  利用 Go 标准库的 `io.TeeReader`。当包管理器发起下载 `.tar.gz` 或 `.whl` 文件请求时，代理层在将数据流源源不断吐给包管理器的同时，克隆出一路并行数据流输入 Go 的 `compress/gzip` 和 `archive/tar` 流式解压器。
  一旦检索到 `setup.py` 或 `package.json` 的文本行，立刻通过高效正则/布隆过滤器匹配高危动态代码执行指纹（如大量的十六进制混淆 `\x65\x76\x61\x6c` 或恶意反弹 Shell命令）。一旦命中，**直接暴力 Close 掉 TCP 连接**，包管理器由于接收数据流中断会立即报错崩溃，从而实现不落地熔断。

---

## 三、 实现路径与阶段优先级 (Implementation Roadmap)

开发路线遵循 **“先骨架、后算法、再深度静态审计”** 的原则，前两阶段聚焦于投毒频发的 **Python (PyPI)** 和 **JavaScript (NPM)**。

```
P0 [核心骨架阶段]
 └── 本地反向代理基础 ──> CLI一键配置/还原 ──> 静态黑名单403阻断

P1 [情报与时间检测]
 └── 安全冷却期熔断(48h) ──> 名字劫持纠错算法 ──> bbolt本地缓存与OSV异步同步

P2 [深度流式与多语言]
 └── io.TeeReader不落地流式解包 ──> 动态代码指纹扫描 ──> 扩展Cargo/Go生态

```

### 📌 P0 阶段：核心代理与熔断骨架（第 1-2 周）

* **功能优先级**：
1. **ReverseProxy 骨架**：利用 Go 标准库的 `httputil.NewSingleHostReverseProxy` 构建，针对 NPM 的 JSON 路由解析（`/@scope/pkg`）与 PyPI 的 PEP 503 Simple 页面格式进行协议适配。
2. **CLI 自动化一键联动**：使用 `spf13/cobra` 编写 CLI。实现 `malfuse link` 与 `malfuse unlink`，一键重写当前 Shell 环境变量（`HTTP_PROXY`）或修改本地全局文件（`~/.npmrc`、`~/.config/pip/pip.conf`）。
3. **403 熔断验证**：建立一个本地硬编码的高危包黑名单。当匹配到对应请求时，代理拦截器直接调用 `w.WriteHeader(http.StatusForbidden)` 写入 `403`，阻断原生包管理器继续下载。



### 📌 P1 阶段：时间维度与情报的并发异步研判（第 3-5 周）

* **功能优先级**：
1. **安全冷却期（Cooldown Pipeline）**：利用高性能 JSON 解析器（如 `buger/jsonparser`）直接在流中提取上游官方源返回的元数据时间戳（NPM 提取 `time`，PyPI 提取 `upload_time`）。**若该版本发布时间距今不足 48 小时，立即熔断**。此项策略可过滤掉 90% 以上突发性的 0-Day 供应链投毒。
2. **本地 Typo-Squatting（高仿包）检测**：内存中常驻 Top 2000 明星包的 `map[string]struct{}` 字典。编写纯 Go 实现的编辑距离算法（Levenshtein），对所有请求进行计算。如发现用户正在下载 `requets` 或 `lodaash` 等与明星包极度相似的组件，抛出警告或阻断。
3. **OSV 开源情报本地二级缓存**：集成纯 Go、无需 CGO 的本地 KV 库 **`etcd-io/bbolt`**。后台启动一个异步 Goroutine 任务，定时从 OpenSSF 官方的 OSV 数据库同步全量的投毒哈希表到本地，实现 `<10ms` 的本地黑名单秒级判定。



### 📌 P2 阶段：深度扫描与多语言扩展（第 6 周+）

* **功能优先级**：
1. **不落地流式扫描器**：实现 `io.TeeReader` 的并发解压流模型，通过内嵌的高性能正则过滤引擎，对下载过程中的动态安装脚本文件（`setup.py` / `preinstall.js`）进行关键词及代码隐蔽混淆分析。
2. **交互式控制台与白名单**：支持 `malfuse allow <pkg_name>@<version>` 快速写入本地 `bbolt` 白名单表，用于解决企业内部私有组件或误报组件的单次放行需求。
3. **多语言生态推广**：平移代理路由层逻辑，扩展适配 Rust (Cargo)、Go (go mod) 的下载连接拦截。



---

## 四、 核心 Go 代码实现蓝图 (Code Blueprint)

以下是本地反向代理与强阻断在 Go 语言中的核心实现模型：

```go
package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type MalFuseGatekeeper struct {
	reverseProxy *httputil.ReverseProxy
}

func NewMalFuseGatekeeper(upstreamURL string) (*MalFuseGatekeeper, error) {
	remote, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// 配置 Director：接管并重写向下游发起的请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = remote.Host // 保持上游 Host 一致，防止厂商镜像源拒绝请求
	}

	return &MalFuseGatekeeper{reverseProxy: proxy}, nil
}

func (mg *MalFuseGatekeeper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 路由解析层：提取当前的包名与请求版本
	pkgName, version := mg.parsePackageMetadata(r.URL.Path)

	// 2. 并发决策引擎：本地与情报交叉验证
	if mg.shouldBlock(pkgName, version) {
		log.Printf("[MalFuse BLOCKED] 成功阻断潜在的投毒恶意组件: %s@%s", pkgName, version)
		
		// 🔴 执行强熔断：直接返回 403 破坏包管理器的安装流
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("[MalFuse Action] 该组件未通过保险丝安全策略，安装已被物理熔断！\n"))
		return
	}

	// 3. 🟢 策略通过：流量无感透传
	mg.reverseProxy.ServeHTTP(w, r)
}

func (mg *MalFuseGatekeeper) parsePackageMetadata(path string) (string, string) {
	// 实际开发中需针对 NPM 路由与 PyPI 标准 PEP 503 实现对应的正则切片解析
	return "test-malicious-package", "1.0.0"
}

func (mg *MalFuseGatekeeper) shouldBlock(pkg, version string) bool {
	// P0 阶段核心验证测试
	if pkg == "test-malicious-package" {
		return true
	}
	return false
}

```

---

## 五、 开发备忘与性能调优指南 (Go Specific)

1. **高并发下的连接复用**：
   由于包管理器会瞬间产生大量并发网络 I/O，必须在初始化时对 Go 默认的 `http.Transport` 进行调优，避免出现大量 TIME_WAIT 导致文件描述符耗尽：
```go
http.DefaultTransport.(*http.Transport).MaxIdleConns = 256
http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
http.DefaultTransport.(*http.Transport).IdleConnTimeout = 90 * time.Second

```


2. **杜绝 CGO，保障绝对的交叉编译**：
   不要使用原生的 `go-sqlite3` 库，推荐使用纯 Go 实现的嵌入式数据库 `bbolt` 或 `modernc.org/sqlite`。这能确保在构建时执行 `CGO_ENABLED=0 go build`，实现一份代码无缝编译出可以在各版本 Linux、Mac (M1/Intel)、Windows 上开箱即用的纯净二进制 `malfuse` 文件。
3. **精准的内存控制**：
   在 P2 阶段实施 `io.TeeReader` 压缩包流式解压时，必须限制读取单个文件的缓冲区大小（利用 `io.LimitReader`），防止攻击者构造“解压炸弹（Zip Bomb）”导致本地代理发生 OOM（内存溢出）崩溃。

```

```