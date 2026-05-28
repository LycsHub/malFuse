# Yarn v2+（Berry）配置 malFuse 代理

Yarn v2+ 使用 `.yarnrc.yml` 配置文件，不支持 `config set` 命令。

## 配置

编辑项目根目录或 `~/.yarnrc.yml`：

```yaml
npmRegistryServer: "http://127.0.0.1:8080/npm/"
```

如果同时使用 npm Scope，加：

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
# 删除或注释 npmRegistryServer 行
yarn config unset npmRegistryServer --home
# 或直接编辑 ~/.yarnrc.yml 删除该行
```
