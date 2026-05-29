# Yarn v2+（Berry）配置 malFuse 代理

Yarn v2+ 使用 `.yarnrc.yml` 配置文件，不支持 `config set` 命令。

## 配置

编辑项目根目录或 `~/.yarnrc.yml`：

```yaml
npmRegistryServer: "http://127.0.0.1:8080/npm/"
```

如果同时使用 npm Scope：

```yaml
npmScopes:
  my-org:
    npmRegistryServer: "http://127.0.0.1:8080/npm/"
```

## 验证

```bash
yarn npm info react 2>&1 | grep malFuse
```

## 还原

```bash
yarn config unset npmRegistryServer --home
# 或直接编辑 ~/.yarnrc.yml 删除该行
```

---

## English

Yarn v2+ uses `.yarnrc.yml`. No `config set` command.

### Setup

Edit project root or `~/.yarnrc.yml`:

```yaml
npmRegistryServer: "http://127.0.0.1:8080/npm/"
```

For npm scopes:

```yaml
npmScopes:
  my-org:
    npmRegistryServer: "http://127.0.0.1:8080/npm/"
```

### Verify

```bash
yarn npm info react 2>&1 | grep malFuse
```

### Restore

```bash
yarn config unset npmRegistryServer --home
# or edit ~/.yarnrc.yml to remove the line
```
