# Gateway

轻量级、可扩展的 Go API 网关，支持声明式 DSL 数据转换、JavaScript Hook 系统和配置热更新。

## 特性

- **灵活路由** - 路径匹配、方法转换、URL 重写、通配符支持
- **DSL 转换** - 使用 JSONPath + Context 语法进行声明式数据转换
- **Hook 系统** - 基于 goja 的 JavaScript 引擎，9 个生命周期节点
- **热更新** - 通过管理 API 动态管理路由和 Hook，零停机
- **双层 Hook** - 支持全局 Hook 和接口级别 Hook
- **并发安全** - 读写锁机制，支持高并发场景

## 快速开始

```bash
# 编译
go build -o gateway .

# 运行
./gateway
```

默认监听 `:8080`

## 配置示例

```yaml
port: ":8080"
backendURL: "http://localhost:9090"
authToken: "your-secret-token"

routes:
  - path: "/api/users"
    method: "POST"
    backendUrl: "http://localhost:9090"
    backendPath: "/v1/users"
    responseTransform:
      code: "200"
      userId: "$.data.id"
      userName: "$.data.name"
```

## 请求处理流程

```
请求 → BeforeAuth → 认证 → AfterAuth → BeforeRequestTransform
     → 请求转换 → AfterRequestTransform → BeforeForward
     → 代理转发 → AfterForward → BeforeResponseTransform
     → 响应转换 → AfterResponseTransform → 响应
     (任意环节出错 → OnError)
```

## DSL 转换

三种数据来源：

| 类型 | 语法 | 示例 |
|-----|------|------|
| 固定值 | 直接写字符串 | `"200"` |
| JSONPath | `$.` 前缀 | `"$.data.id"` |
| Context | `@ctx.` 前缀 | `"@ctx.request.method"` |

### 基本映射

```yaml
responseTransform:
  code: "200"                    # 固定值
  userId: "$.data.id"            # JSONPath
  method: "@ctx.request.method"  # Context
```

### 数组转换

```yaml
responseTransform:
  items:
    json.path: "$.data"         # 指定源数组
    id: "$.ID_SRV"              # 从数组元素提取
    name: "$.EXAMINE_NAME"
```

### Context 访问

```yaml
responseTransform:
  # 请求信息
  method: "@ctx.request.method"
  path: "@ctx.request.path"
  userId: "@ctx.request.body.userId"

  # 响应信息
  status: "@ctx.response.status"

  # 路由信息
  backend: "@ctx.route.backendUrl"

  # 自定义数据（通过 Hook 设置）
  tenantId: "@ctx.tenantId"
```

## Hook 系统

### 接口级别 Hook

```yaml
routes:
  - path: "/api/users"
    method: "POST"
    backendUrl: "http://localhost:9090"
    hooks:
      BeforeAuth: |
        if (context.requestHeaders.Authorization) {
          context.data.userId = "user-123";
        }
      BeforeForward: |
        context.requestHeaders["X-User-ID"] = context.data.userId;
```

### 全局 Hook

```go
hookManager := hook.NewManager()
hookManager.RegisterScript(hook.BeforeAuth, "scripts/auth.js")
// 或从字符串注册
hookManager.RegisterScriptString(hook.BeforeAuth, scriptContent)
```

### Hook 节点

| 节点 | 说明 |
|-----|------|
| BeforeAuth | 认证前 |
| AfterAuth | 认证后 |
| BeforeRequestTransform | 请求转换前 |
| AfterRequestTransform | 请求转换后 |
| BeforeForward | 转发前 |
| AfterForward | 转发后 |
| BeforeResponseTransform | 响应转换前 |
| AfterResponseTransform | 响应转换后 |
| OnError | 错误处理 |

## 管理 API

所有接口需要 `X-Admin-Token` 认证。

### 查询路由

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes
```

### 添加路由

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"route": {"path": "/api/products", "method": "GET", ...}}' \
  http://localhost:8080/admin/routes/add
```

### 更新路由

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"route": {"path": "/api/users", "method": "POST", ...}}' \
  http://localhost:8080/admin/routes/update
```

### 删除路由

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"path": "/api/products", "method": "GET"}' \
  http://localhost:8080/admin/routes/delete
```

### 更新 Hook

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"hookPoint": "BeforeAuth", "script": "..."}' \
  http://localhost:8080/admin/hooks/update
```

## 项目结构

```
gateway/
├── main.go              # 入口文件
├── config.yaml          # 配置文件
├── config/              # 配置管理
├── handler/             # HTTP 处理器
│   ├── gateway.go       # 网关核心逻辑
│   └── admin.go         # 管理 API
├── hook/                # Hook 系统
│   ├── types.go         # 类型定义
│   ├── manager.go       # Hook 管理器
│   └── executor.go      # JS 执行器
├── middleware/          # 中间件
├── proxy/               # 代理转发
├── router/              # 路由匹配
├── transform/           # DSL 转换引擎
└── scripts/             # Hook 脚本示例
```

## 技术栈

- Go 1.18+
- Viper (配置管理)
- goja (JavaScript 引擎)
- jsonpath (JSONPath 查询)

## 测试

```bash
go test ./...
```

## 文档

- [管理 API 文档](./ADMIN_API.md)
- [DSL Context 参考](./DSL_CONTEXT_REFERENCE.md)
- [使用示例](./EXAMPLE.md)
- [并发安全说明](./CONCURRENCY_SAFETY.md)

## License

MIT
