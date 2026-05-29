# Conda 配置 malFuse 代理

Conda 需修改 `~/.condarc` 文件。

## 配置

编辑 `~/.condarc`（Linux/macOS）或 `%USERPROFILE%/.condarc`（Windows）：

```yaml
channels:
  - http://127.0.0.1:8080/pypi/simple/
  - defaults
```

## 验证

```bash
conda search requests 2>&1 | grep malFuse
```

## 还原

```yaml
# 删除或注释 malFuse 通道
channels:
  - defaults
```

---

## English

Conda requires modifying the `~/.condarc` file.

### Setup

Edit `~/.condarc` (Linux/macOS) or `%USERPROFILE%/.condarc` (Windows):

```yaml
channels:
  - http://127.0.0.1:8080/pypi/simple/
  - defaults
```

### Verify

```bash
conda search requests 2>&1 | grep malFuse
```

### Restore

```yaml
# Remove or comment the malFuse channel
channels:
  - defaults
```
