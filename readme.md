# Gateway 使用指南

## 项目简介

这是一个**可扩展的 API 网关**，核心特点：

- 🔀 **灵活的路由配置** - 支持路径匹配、方法转换、URL 重写
- 🎯 **声明式 DSL 转换** - 使用 JSONPath + Context 语法进行数据转换，无需编写代码
- 🔌 **JavaScript Hook 系统** - 在 9 个生命周期节点注入自定义逻辑
- 🔥 **配置热更新** - 通过管理 API 动态添加/修改路由和 Hook，无需重启服务
- 🚀 **零停机部署** - 所有配置变更立即生效

## 快速开始

### 1. 编译运行

```bash
# 安装依赖
go mod tidy

# 编译
go build -o gateway .

# 运行
./gateway
```

默认监听端口：`:8080`

### 2. 基础配置

创建 `config.yaml`：

```yaml
port: ":8080"
backendURL: "http://localhost:9090"  # 默认后端服务
authToken: "your-secret-token"       # 认证 Token

routes:
  - path: "/api/users"
    method: "POST"
    backendUrl: "http://localhost:9090"
    backendPath: "/v1/users"
    responseTransform:
      code: "200"
      message: "success"
      data: "$.data"
```

### 3. 发送请求

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

## 核心概念

### 路由配置

每个路由支持以下配置：

```yaml
routes:
  - path: "/api/users"              # 匹配路径（支持通配符 * ）
    method: "POST"                  # 匹配 HTTP 方法
    backendUrl: "http://localhost:9090"     # 后端服务 URL
    backendPath: "/v1/users"        # 转发到后端的路径
    backendMethod: "PUT"            # 转发到后端的 HTTP 方法
    requestTransform: { ... }       # 请求体转换（可选）
    responseTransform: { ... }      # 响应体转换（可选）
```

### DSL 转换

DSL 转换有三种数据来源：

| 数据来源 | 语法 | 示例 |
|---------|------|------|
| **固定值** | 直接写字符串 | `"200"` |
| **JSONPath** | `$.` 前缀 | `"$.data.id"` |
| **Context** | `@ctx.` 前缀 | `"@ctx.request.body.userId"` |

## DSL 语法详解

### 1. 基本字段映射

```yaml
responseTransform:
  # 固定值
  code: "200"
  message: "success"

  # 从响应 JSON 提取（使用 JSONPath）
  userId: "$.data.id"
  userName: "$.data.name"
  userEmail: "$.data.email"
```

**示例：**

输入响应：
```json
{
  "data": {
    "id": 123,
    "name": "John",
    "email": "john@example.com"
  }
}
```

输出结果：
```json
{
  "code": "200",
  "message": "success",
  "userId": 123,
  "userName": "John",
  "userEmail": "john@example.com"
}
```

### 2. 嵌套对象转换

```yaml
responseTransform:
  user:
    id: "$.data.userId"
    profile:
      name: "$.data.userName"
      email: "$.data.userEmail"
  metadata:
    timestamp: "$.timestamp"
```

### 3. 数组转换 ⭐️ 重点

使用 `json.path` 标识数组，然后对每个元素进行转换。

```yaml
responseTransform:
  code: "200"
  data:
    items:
      json.path: "$.data"        # 指定源数组
      id: "$.ID_SRV"             # 从数组元素提取
      name: "$.EXAMINE_NAME"
      type: "$.citem_type"
```

**示例：**

输入响应：
```json
{
  "data": [
    {
      "ID_SRV": "001",
      "EXAMINE_NAME": "Blood Test",
      "citem_type": "LAB"
    },
    {
      "ID_SRV": "002",
      "EXAMINE_NAME": "X-Ray",
      "citem_type": "IMG"
    }
  ]
}
```

输出结果：
```json
{
  "code": "200",
  "data": {
    "items": [
      {
        "id": "001",
        "name": "Blood Test",
        "type": "LAB"
      },
      {
        "id": "002",
        "name": "X-Ray",
        "type": "IMG"
      }
    ]
  }
}
```

### 4. 保留原始数据

使用 `"$."` 可以保留完整的源数据。

```yaml
responseTransform:
  data:
    items:
      json.path: "$.data"
      id: "$.ID_SRV"
      name: "$.EXAMINE_NAME"
      origin: "$."               # 保留完整原始数据
```

### 5. 访问 Context 数据 ⭐️ 重点

#### 5.1 访问请求体数据

在 `responseTransform` 中访问原始请求体的数据：

```yaml
# 假设客户端请求：POST /api/examines
# Body: {"userId": "123", "action": "query", "filters": {"type": "blood"}}

responseTransform:
  code: "$.code"
  data: "$.data"

  # 从请求体获取数据
  requestUserId: "@ctx.request.body.userId"           # "123"
  requestAction: "@ctx.request.body.action"           # "query"
  requestFilters: "@ctx.request.body.filters"         # 整个对象
  requestFilterType: "@ctx.request.body.filters.type" # "blood"（支持嵌套）
```

#### 5.2 访问请求元数据

```yaml
responseTransform:
  result: "$.data"

  # 请求元数据
  requestMethod: "@ctx.request.method"      # "GET", "POST", etc.
  requestPath: "@ctx.request.path"          # "/api/users"
  requestQuery: "@ctx.request.query"        # "id=123&name=test"
  requestHost: "@ctx.request.host"          # "localhost:8080"
  authHeader: "@ctx.request.header.Authorization"
```

#### 5.3 访问路由信息

```yaml
responseTransform:
  result: "$.data"

  # 路由信息
  routePath: "@ctx.route.path"
  routeMethod: "@ctx.route.method"
  backendUrl: "@ctx.route.backendUrl"
```

#### 5.4 访问响应元数据

```yaml
responseTransform:
  result: "$.data"

  # 响应元数据
  httpStatus: "@ctx.response.status"        # 200, 404, 500
  contentType: "@ctx.response.header.Content-Type"
```

#### 5.5 访问自定义数据（通过 Hook 设置）

在 JavaScript Hook 中设置：

```javascript
// scripts/auth.js
context.data.tenantId = "tenant-001";
context.data.user = {
  id: "user-123",
  name: "John Doe"
};
```

在 DSL 中访问：

```yaml
responseTransform:
  result: "$.data"

  # 自定义数据
  tenantId: "@ctx.tenantId"
  userId: "@ctx.user.id"
  userName: "@ctx.user.name"
```

## Context 数据结构参考

```javascript
ctx.Data = {
  request: {
    method: "POST",                    // HTTP 方法
    path: "/api/users",                // 请求路径
    query: "id=123&name=test",         // 查询字符串
    host: "localhost:8080",            // Host
    header: {                          // 请求头
      "Authorization": "Bearer xxx",
      "Content-Type": "application/json"
    },
    body: {                            // 请求体（JSON 解析后）
      "userId": "123",
      "action": "query",
      "params": { ... }
    }
  },
  route: {                             // 匹配的路由信息
    path: "/api/users",
    method: "POST",
    backendUrl: "http://localhost:9090",
    backendPath: "/v1/users",
    backendMethod: "PUT"
  },
  response: {                          // 响应元数据
    status: 200,
    header: {
      "Content-Type": "application/json"
    }
  },
  // 自定义数据（通过 Hook 设置）
  tenantId: "tenant-001",
  user: { id: "123", name: "John" }
}
```

## 完整示例：综合使用

```yaml
routes:
  - path: "/api/examines"
    method: "POST"
    backendUrl: "http://localhost:9090"
    backendPath: "/examines"
    responseTransform:
      # 固定值
      code_success: "200"

      # 从响应获取
      code: "$.code"
      message: "$.message"

      # 从请求体获取
      requestUserId: "@ctx.request.body.userId"
      requestAction: "@ctx.request.body.action"

      # 从请求元数据获取
      requestMethod: "@ctx.request.method"
      requestPath: "@ctx.request.path"

      # 从响应元数据获取
      httpStatus: "@ctx.response.status"

      # 数组转换
      data:
        total: "$.total"
        items:
          json.path: "$.data"
          id: "$.ID_SRV"
          name: "$.EXAMINE_NAME"
          type: "$.citem_type"
          # 每个数组元素中也添加请求信息
          requestedBy: "@ctx.request.body.userId"
          originalData: "$."
```

## JavaScript Hook 系统

### Hook 节点

系统支持在 9 个生命周期节点注入 JavaScript 代码：

```
1. BeforeAuth              - 认证前
2. AfterAuth               - 认证后
3. BeforeRequestTransform  - 请求转换前
4. AfterRequestTransform   - 请求转换后
5. BeforeForward           - 转发前
6. AfterForward            - 转发后
7. BeforeResponseTransform - 响应转换前
8. AfterResponseTransform  - 响应转换后
9. OnError                 - 错误处理
```

### Hook 示例

**scripts/examples/auth.js**

```javascript
// 在 context 中设置自定义数据
if (context.requestHeaders.Authorization) {
  const token = context.requestHeaders.Authorization.replace('Bearer ', '');

  // 解析 token（这里简化处理）
  context.data.userId = "user-123";
  context.data.tenantId = "tenant-001";
  context.data.user = {
    id: "user-123",
    name: "John Doe",
    role: "admin"
  };
  context.data.timestamp = new Date().toISOString();
}

console.log("Auth hook executed");
```

### 注册 Hook

系统支持两种 Hook 注册方式：

#### 1. 从文件注册（适合本地开发和静态脚本）

在 `main.go` 中从文件加载 Hook：

```go
hookManager := hook.NewManager()

// 从项目中的脚本文件注册
hookManager.RegisterScript(hook.BeforeAuth, "scripts/examples/auth.js")
hookManager.RegisterScript(hook.AfterRequestTransform, "scripts/examples/transform.js")
hookManager.RegisterScript(hook.OnError, "scripts/examples/error.js")
```

#### 2. 从字符串注册（适合数据库存储和动态脚本）⭐️ 推荐

当你想把 JavaScript 脚本存储在数据库中时，使用 `RegisterScriptString` 方法：

```go
hookManager := hook.NewManager()

// 从数据库加载脚本内容（示例）
scriptContent := `
// 认证 Hook 脚本
if (context.requestHeaders.Authorization) {
  const token = context.requestHeaders.Authorization.replace('Bearer ', '');

  // 解析 token 并设置用户信息
  context.data.userId = "user-123";
  context.data.tenantId = "tenant-001";
  context.data.user = {
    id: "user-123",
    name: "John Doe",
    role: "admin"
  };

  console.log("User authenticated:", context.data.userId);
}
`

// 直接注册字符串形式的脚本
err := hookManager.RegisterScriptString(hook.BeforeAuth, scriptContent)
if err != nil {
  log.Fatal("Failed to register hook:", err)
}
```

**实际使用场景（从数据库加载）：**

```go
// 假设从数据库查询脚本
type HookScript struct {
  HookPoint string
  Content   string
}

// 从数据库查询所有 Hook 脚本
scripts := []HookScript{
  {HookPoint: "BeforeAuth", Content: "...JS code from DB..."},
  {HookPoint: "AfterAuth", Content: "...JS code from DB..."},
  {HookPoint: "OnError", Content: "...JS code from DB..."},
}

// 注册所有脚本
hookManager := hook.NewManager()
for _, script := range scripts {
  var hookPoint hook.HookPoint

  switch script.HookPoint {
  case "BeforeAuth":
    hookPoint = hook.BeforeAuth
  case "AfterAuth":
    hookPoint = hook.AfterAuth
  case "OnError":
    hookPoint = hook.OnError
  // ... 其他 Hook 节点
  }

  err := hookManager.RegisterScriptString(hookPoint, script.Content)
  if err != nil {
    log.Printf("Failed to register hook %s: %v", script.HookPoint, err)
  }
}
```

**优势对比：**

| 方式 | 适用场景 | 优点 | 缺点 |
|-----|---------|------|------|
| `RegisterScript` | 本地开发、静态脚本 | 简单直接、版本控制友好 | 需要重启部署才能更新 |
| `RegisterScriptString` | 生产环境、动态管理 | 支持数据库存储、热更新、集中管理 | 需要额外的存储和管理系统 |

## 常见问题

### 1. 如何访问响应数据？

**直接用 `$.` 即可**，不需要 `@ctx.response.body`：

```yaml
# ✅ 正确
responseTransform:
  userId: "$.data.id"

# ❌ 错误（不要这样）
responseTransform:
  userId: "@ctx.response.body.data.id"
```

### 2. 如何在响应中访问请求体数据？

使用 `@ctx.request.body.*`：

```yaml
responseTransform:
  result: "$.data"
  requestUserId: "@ctx.request.body.userId"
```

### 3. 数组转换时如何添加固定值？

直接在模板中写固定值即可：

```yaml
responseTransform:
  data:
    items:
      json.path: "$.data"
      id: "$.ID_SRV"
      name: "$.EXAMINE_NAME"
      type: "fixed-type"              # 固定值
      pageNo: "1"                     # 固定值
```

### 4. 如何处理请求方法转换？

在路由配置中指定 `backendMethod`：

```yaml
routes:
  - path: "/api/orders"
    method: "POST"                    # 客户端用 POST
    backendMethod: "PUT"              # 转发到后端用 PUT
```

### 5. 通配符路由如何使用？

使用 `*` 匹配任意路径：

```yaml
routes:
  - path: "/api/products/*"           # 匹配 /api/products/xxx
    method: "GET"
    backendUrl: "http://localhost:9091"
```

### 6. 应该使用文件注册还是字符串注册 Hook？

**开发环境** - 使用 `RegisterScript`（文件方式）：
- 脚本可以用 Git 版本管理
- IDE 有语法高亮和代码提示
- 调试方便

**生产环境** - 使用 `RegisterScriptString`（字符串方式）：
- 支持不重启网关动态更新脚本
- 集中化管理（数据库存储）
- 支持多环境配置（测试/生产脚本分离）
- 便于权限控制和审计

**混合使用**：
```go
// 基础 Hook 从文件加载（稳定不变）
hookManager.RegisterScript(hook.OnError, "scripts/error_handler.js")

// 业务 Hook 从数据库加载（可动态更新）
for _, script := range loadScriptsFromDB() {
  hookManager.RegisterScriptString(script.HookPoint, script.Content)
}
```

### 7. 配置热更新是否线程安全？

**是的，完全线程安全！**

- Router 使用 `sync.RWMutex` 保护路由表
- Hook Manager 使用 `sync.RWMutex` 保护 Hook 注册表
- 读操作使用读锁，允许并发
- 写操作使用写锁，保证数据一致性

在高并发场景下，动态更新配置不会影响正在处理的请求。

### 8. 配置更新后会影响正在处理的请求吗？

**不会！**

- 正在处理的请求使用的是更新前的配置副本
- 新请求会使用更新后的配置
- 配置更新是原子操作，不会出现部分更新的情况

### 9. 如何持久化热更新的配置？

管理 API 只修改内存中的配置。如果需要持久化，建议：

**方案 1：使用数据库存储配置**
```go
// 从数据库加载配置
routes := loadRoutesFromDB()
for _, route := range routes {
  router.AddRoute(route)
}

// 通过管理 API 更新时，同时更新数据库
// 服务重启时从数据库重新加载
```

**方案 2：定期导出配置到文件**
```bash
# 定期导出当前配置
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes > routes_backup.json
```

**方案 3：使用配置中心**（如 etcd、Consul）
- 从配置中心加载配置
- 通过管理 API 更新时，同时更新配置中心
- 支持配置版本控制和回滚

## 配置热更新 🔥 重要

Gateway 支持通过管理 API **在运行时动态管理配置**，无需重启服务！

### 为什么需要热更新？

**传统方式的问题：**
- ❌ 每次添加新接口都要修改 config.yaml 并重启服务
- ❌ 修改 DSL 转换规则需要重启
- ❌ 更新 Hook 脚本需要重启
- ❌ 重启导致服务中断

**热更新的优势：**
- ✅ 动态添加/修改/删除路由配置
- ✅ 动态更新 DSL 转换规则
- ✅ 动态更新 JavaScript Hook 脚本
- ✅ 零停机，配置立即生效
- ✅ 支持从数据库加载配置

### 管理 API 快速上手

所有管理 API 都需要通过 `X-Admin-Token` Header 认证。

#### 1. 动态添加路由

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "route": {
      "path": "/api/products",
      "method": "GET",
      "backendUrl": "http://localhost:9091",
      "backendPath": "/products",
      "responseTransform": {
        "success": true,
        "items": "$.data"
      }
    }
  }' \
  http://localhost:8080/admin/routes/add
```

**立即生效！** 客户端可以马上访问 `/api/products` 接口。

#### 2. 动态更新 DSL 配置

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "route": {
      "path": "/api/users",
      "method": "POST",
      "backendUrl": "http://localhost:9090",
      "backendPath": "/v1/users",
      "responseTransform": {
        "code": "200",
        "userId": "$.data.id",
        "userName": "$.data.name",
        "requestMethod": "@ctx.request.method"
      }
    }
  }' \
  http://localhost:8080/admin/routes/update
```

**立即生效！** 下一个请求就会使用新的 DSL 配置。

#### 3. 动态更新 Hook 脚本

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "hookPoint": "BeforeAuth",
    "script": "console.log(\"New auth logic\"); context.data.userId = \"user-456\";"
  }' \
  http://localhost:8080/admin/hooks/update
```

**立即生效！** 下一个请求就会执行新的 Hook 脚本。

#### 4. 查看当前所有路由

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes
```

### 完整管理 API 文档

详细的 API 文档请查看：[ADMIN_API.md](./ADMIN_API.md)

包含：
- 路由管理（查询、添加、更新、删除）
- Hook 管理（更新、清空）
- 错误处理
- 安全建议
- 实际应用场景


## 项目结构

```
gateway/
├── main.go                    # 入口文件
├── config.yaml                # 配置文件
├── config/                    # 配置管理
│   └── config.go
├── handler/                   # HTTP 处理器
│   └── gateway.go
├── hook/                      # Hook 系统
│   ├── types.go              # Hook 接口定义
│   ├── manager.go            # Hook 管理器
│   └── executor.go           # JavaScript 执行器
├── middleware/                # 中间件
│   ├── auth.go               # 认证中间件
│   ├── transform.go          # 转换中间件
│   └── error.go              # 错误处理中间件
├── proxy/                     # 代理转发
│   └── forwarder.go
├── router/                    # 路由匹配
│   └── router.go
├── transform/                 # DSL 转换引擎
│   ├── dsl.go
│   └── dsl_test.go
└── scripts/                   # Hook 脚本
    └── examples/
        ├── auth.js
        ├── transform.js
        └── error.js
```

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行 DSL 转换测试
go test -v ./transform/

# 查看测试覆盖率
go test -cover ./transform/
```

## 总结

Gateway 提供了三层灵活性：

1. **路由配置** - 声明式配置 URL、方法转换
2. **DSL 转换** - 无代码的数据转换
3. **Hook 系统** - JavaScript 动态逻辑注入

从简单到复杂，你可以根据需求选择合适的方式来实现业务逻辑。
