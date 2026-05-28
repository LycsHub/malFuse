# uv 配置 malFuse 代理

uv 原生不支持 `config set` 命令，需手动配置。

## 方法一：环境变量（推荐）

```bash
export UV_INDEX_URL=http://127.0.0.1:8080/pypi/simple/
uv pip install requests
```

## 方法二：配置文件

编辑 `~/.config/uv/uv.toml`（Linux/macOS）或 `%APPDATA%/uv/uv.toml`（Windows）：

```toml
[pip]
index-url = "http://127.0.0.1:8080/pypi/simple/"
```

## 验证

```bash
uv pip install --dry-run requests 2>&1 | grep malFuse
```

## 还原

```bash
unset UV_INDEX_URL
# 或删除 uv.toml 中的 [pip] 节
```
