# Poetry 配置 malFuse 代理

Poetry 需修改 `pyproject.toml` 中的 source 配置。

## 配置

编辑项目 `pyproject.toml`：

```toml
[[tool.poetry.source]]
name = "malFuse"
url = "http://127.0.0.1:8080/pypi/simple/"
default = true
```

## 验证

```bash
poetry add --dry-run requests 2>&1 | grep malFuse
```

## 还原

```toml
# 删除 [[tool.poetry.source]] 节，或设置 default = false
```
