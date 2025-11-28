# DSL Context 访问参考文档

本文档说明在 DSL 转换中可以通过 `@ctx.` 前缀访问的所有 Context 字段。

## 概述

在 DSL 转换配置中，你可以使用三种方式获取数据：

1. **固定值** - 直接使用字符串：`"200"`
2. **JSONPath** - 从响应/请求 JSON 中提取：`"$.data.id"`
3. **Context 访问** - 从 HookContext 中获取：`"@ctx.request.method"`

## Context 字段结构

Gateway 会自动将以下信息添加到 `ctx.Data` 中，可以在 DSL 转换中访问：

### 1. 请求信息 (`@ctx.request.*`)

在请求开始时就可用，包含原始 HTTP 请求的信息。

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `@ctx.request.method` | string | HTTP 请求方法 | `"GET"`, `"POST"`, `"PUT"` |
| `@ctx.request.path` | string | 请求路径 | `"/api/users"` |
| `@ctx.request.query` | string | 查询字符串（原始） | `"id=123&name=test"` |
| `@ctx.request.host` | string | 请求的 Host | `"localhost:8080"` |
| `@ctx.request.header.*` | string | 请求头（需指定具体的 header 名称） | `@ctx.request.header.Authorization` |

**示例：**
```yaml
responseTransform:
  requestMethod: "@ctx.request.method"
  requestPath: "@ctx.request.path"
  clientHost: "@ctx.request.host"
  authToken: "@ctx.request.header.Authorization"
```

### 2. 路由信息 (`@ctx.route.*`)

当请求匹配到路由配置时可用。

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `@ctx.route.path` | string | 路由配置的路径模式 | `"/api/users"` |
| `@ctx.route.method` | string | 路由配置的 HTTP 方法 | `"POST"` |
| `@ctx.route.backendUrl` | string | 后端服务 URL | `"http://localhost:9090"` |
| `@ctx.route.backendPath` | string | 后端路径 | `"/v1/users"` |
| `@ctx.route.backendMethod` | string | 转发到后端的 HTTP 方法 | `"PUT"` |

**示例：**
```yaml
responseTransform:
  routePath: "@ctx.route.path"
  backendService: "@ctx.route.backendUrl"
```

### 3. 响应信息 (`@ctx.response.*`)

在收到后端响应后可用（仅在 `responseTransform` 中可用）。

| 字段 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `@ctx.response.status` | int | HTTP 响应状态码 | `200`, `404`, `500` |
| `@ctx.response.header.*` | string | 响应头（需指定具体的 header 名称） | `@ctx.response.header.Content-Type` |

**示例：**
```yaml
responseTransform:
  httpStatus: "@ctx.response.status"
  contentType: "@ctx.response.header.Content-Type"
```

### 4. 自定义数据 (`@ctx.*`)

可以在 JavaScript Hook 中设置自定义数据，然后在 DSL 中访问。

**在 JS Hook 中设置：**
```javascript
// scripts/examples/auth.js
context.data.userId = "user-123";
context.data.tenantId = "tenant-001";
context.data.user = {
  id: "user-123",
  name: "John Doe",
  role: "admin"
};
```

**在 DSL 中访问：**
```yaml
responseTransform:
  userId: "@ctx.userId"
  tenantId: "@ctx.tenantId"
  userName: "@ctx.user.name"      # 支持嵌套访问
  userRole: "@ctx.user.role"
```

## 完整示例

### 示例 1：基本 Context 访问

```yaml
routes:
  - path: "/api/users"
    method: "GET"
    backendUrl: "http://localhost:9090"
    backendPath: "/v1/users"
    responseTransform:
      # 固定值
      code: "200"
      message: "success"

      # 从响应 JSON 提取
      users: "$.data"
      total: "$.total"

      # 从 Context 获取
      requestMethod: "@ctx.request.method"
      requestPath: "@ctx.request.path"
      backendStatus: "@ctx.response.status"
```

### 示例 2：数组处理 + Context 访问

```yaml
routes:
  - path: "/api/examines"
    method: "GET"
    responseTransform:
      code_success: "200"
      code: "$.code"
      msg: "$.message"

      # 请求信息
      requestMethod: "@ctx.request.method"
      requestPath: "@ctx.request.path"

      # 自定义数据（需在 Hook 中设置）
      userId: "@ctx.user.id"
      tenantId: "@ctx.tenantId"

      data:
        pages: "1"
        zd_list:
          json.path: "$.data"
          item_id: "$.ID_SRV"
          item_name: "$.EXAMINE_NAME"
          # 数组元素中也可以访问 Context
          tenant_id: "@ctx.tenantId"
          request_method: "@ctx.request.method"
          origin_data: "$."
```

### 示例 3：请求转换中使用 Context

```yaml
routes:
  - path: "/api/orders"
    method: "POST"
    requestTransform:
      order:
        id: "$.orderId"
        items: "$.items"
        # 添加请求信息到转换后的数据
        clientHost: "@ctx.request.host"
        requestPath: "@ctx.request.path"
      # 添加自定义数据
      userId: "@ctx.user.id"
      tenantId: "@ctx.tenantId"
```

## 使用场景

### 1. 添加请求追踪信息

```yaml
responseTransform:
  data: "$.data"
  metadata:
    requestMethod: "@ctx.request.method"
    requestPath: "@ctx.request.path"
    backendStatus: "@ctx.response.status"
    timestamp: "@ctx.timestamp"  # 需在 Hook 中设置
```

### 2. 多租户数据隔离

```javascript
// 在 Hook 中设置租户 ID
context.data.tenantId = extractTenantFromToken(context.requestHeaders.Authorization);
```

```yaml
responseTransform:
  code: "$.code"
  data:
    items:
      json.path: "$.data"
      id: "$.id"
      name: "$.name"
      tenant_id: "@ctx.tenantId"  # 为每个数据项添加租户 ID
```

### 3. 请求日志和审计

```yaml
responseTransform:
  result: "$.data"
  audit:
    userId: "@ctx.user.id"
    action: "@ctx.request.method"
    resource: "@ctx.request.path"
    status: "@ctx.response.status"
    timestamp: "@ctx.timestamp"
```

## 注意事项

1. **可用性**：
   - `@ctx.request.*` - 在 `requestTransform` 和 `responseTransform` 中都可用
   - `@ctx.route.*` - 仅当请求匹配到路由时可用
   - `@ctx.response.*` - 仅在 `responseTransform` 中可用
   - 自定义数据 - 取决于在哪个 Hook 中设置

2. **嵌套访问**：
   - 支持多层嵌套：`@ctx.user.profile.age`
   - 如果路径不存在，返回 `null`

3. **类型保持**：
   - Context 中的数据类型会被保留（string、int、bool 等）
   - 不会自动转换为字符串

4. **性能**：
   - Context 访问是内存操作，性能开销很小
   - 建议将常用数据缓存到 Context 中，避免重复计算

## Hook 与 DSL 配合使用

### 在 Hook 中设置数据

```javascript
// scripts/examples/auth.js
// BeforeAuth Hook
if (context.requestHeaders.Authorization) {
  const token = context.requestHeaders.Authorization.replace('Bearer ', '');
  const decoded = decodeToken(token);

  context.data.userId = decoded.userId;
  context.data.tenantId = decoded.tenantId;
  context.data.user = {
    id: decoded.userId,
    name: decoded.userName,
    role: decoded.role
  };
  context.data.timestamp = new Date().toISOString();
}
```

### 在 DSL 中使用

```yaml
responseTransform:
  code: "$.code"
  data: "$.data"
  # 使用 Hook 中设置的数据
  userId: "@ctx.userId"
  userName: "@ctx.user.name"
  userRole: "@ctx.user.role"
  tenantId: "@ctx.tenantId"
  timestamp: "@ctx.timestamp"
```

## 总结

通过 `@ctx.` 前缀，你可以在 DSL 转换中访问：

- ✅ 请求信息（方法、路径、头部等）
- ✅ 路由配置信息
- ✅ 响应信息（状态码、头部等）
- ✅ JavaScript Hook 中设置的自定义数据

这使得 DSL 转换更加灵活和强大，可以根据请求上下文动态生成响应数据。
