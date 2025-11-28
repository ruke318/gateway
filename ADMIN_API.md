# Gateway 管理 API 文档

## 概述

Gateway 提供了一套完整的管理 API，支持在运行时动态管理路由配置和 Hook 脚本，**无需重启服务**。

## 认证

所有管理 API 都需要通过 `X-Admin-Token` Header 进行认证：

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes
```

⚠️ **安全提示**：
- 建议在生产环境中使用强密码作为 admin token
- 建议通过配置文件管理 admin token
- 建议限制管理 API 只能从内网访问

---

## 路由管理 API

### 1. 查询所有路由

**请求：**
```bash
GET /admin/routes
```

**示例：**
```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes
```

**响应：**
```json
{
  "success": true,
  "data": [
    {
      "path": "/api/users",
      "method": "POST",
      "backendUrl": "http://localhost:9090",
      "backendPath": "/v1/users",
      "backendMethod": "PUT",
      "requestTransform": {},
      "responseTransform": {
        "code": "200",
        "data": "$.data"
      }
    }
  ]
}
```

---

### 2. 添加路由

**请求：**
```bash
POST /admin/routes/add
Content-Type: application/json
```

**请求体：**
```json
{
  "route": {
    "path": "/api/orders",
    "method": "POST",
    "backendUrl": "http://localhost:9090",
    "backendPath": "/v1/orders",
    "backendMethod": "POST",
    "responseTransform": {
      "code": "200",
      "message": "success",
      "orderId": "$.data.id"
    }
  }
}
```

**示例：**
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

**响应：**
```json
{
  "success": true,
  "message": "route added successfully"
}
```

---

### 3. 更新路由

**请求：**
```bash
POST /admin/routes/update
Content-Type: application/json
```

**说明：** 根据 `path` 和 `method` 匹配现有路由并更新

**请求体：**
```json
{
  "route": {
    "path": "/api/orders",
    "method": "POST",
    "backendUrl": "http://localhost:9092",
    "backendPath": "/v2/orders",
    "responseTransform": {
      "code": "200",
      "message": "updated",
      "orderId": "$.id"
    }
  }
}
```

**示例：**
```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "route": {
      "path": "/api/users",
      "method": "POST",
      "backendUrl": "http://localhost:9090",
      "backendPath": "/v2/users",
      "responseTransform": {
        "success": true,
        "userId": "$.data.id"
      }
    }
  }' \
  http://localhost:8080/admin/routes/update
```

**响应：**
```json
{
  "success": true,
  "message": "route updated successfully"
}
```

**错误响应（路由不存在）：**
```
HTTP 404 Not Found
failed to update route: route not found: POST /api/xxx
```

---

### 4. 删除路由

**请求：**
```bash
POST /admin/routes/delete
Content-Type: application/json
```

**请求体：**
```json
{
  "path": "/api/orders",
  "method": "POST"
}
```

**示例：**
```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api/products",
    "method": "GET"
  }' \
  http://localhost:8080/admin/routes/delete
```

**响应：**
```json
{
  "success": true,
  "message": "route deleted successfully"
}
```

---

## Hook 管理 API

### 1. 更新 Hook 脚本

**请求：**
```bash
POST /admin/hooks/update
Content-Type: application/json
```

**请求体：**
```json
{
  "hookPoint": "BeforeAuth",
  "script": "console.log('Auth hook'); context.data.userId = '123';"
}
```

**支持的 HookPoint 值：**
- `BeforeAuth` - 认证前
- `AfterAuth` - 认证后
- `BeforeRequestTransform` - 请求转换前
- `AfterRequestTransform` - 请求转换后
- `BeforeForward` - 转发前
- `AfterForward` - 转发后
- `BeforeResponseTransform` - 响应转换前
- `AfterResponseTransform` - 响应转换后
- `OnError` - 错误处理

**示例：**
```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "hookPoint": "BeforeAuth",
    "script": "if (context.requestHeaders.Authorization) { const token = context.requestHeaders.Authorization.replace(\"Bearer \", \"\"); context.data.userId = \"user-123\"; context.data.tenantId = \"tenant-001\"; console.log(\"User authenticated:\", context.data.userId); }"
  }' \
  http://localhost:8080/admin/hooks/update
```

**响应：**
```json
{
  "success": true,
  "message": "hook updated successfully"
}
```

**复杂脚本示例（使用文件）：**

```bash
# 从文件读取脚本内容
SCRIPT=$(cat <<'EOF'
// 认证 Hook
if (context.requestHeaders.Authorization) {
  const token = context.requestHeaders.Authorization.replace('Bearer ', '');

  // 解析 token
  context.data.userId = "user-123";
  context.data.tenantId = "tenant-001";
  context.data.user = {
    id: "user-123",
    name: "John Doe",
    role: "admin"
  };

  console.log("User authenticated:", context.data.userId);
}
EOF
)

curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d "{\"hookPoint\": \"BeforeAuth\", \"script\": $(echo "$SCRIPT" | jq -Rs .)}" \
  http://localhost:8080/admin/hooks/update
```

---

### 2. 清空 Hook 脚本

**请求：**
```bash
POST /admin/hooks/clear
Content-Type: application/json
```

**请求体：**
```json
{
  "hookPoint": "BeforeAuth"
}
```

**示例：**
```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"hookPoint": "BeforeAuth"}' \
  http://localhost:8080/admin/hooks/clear
```

**响应：**
```json
{
  "success": true,
  "message": "hook cleared successfully"
}
```

---

## 实际应用场景

### 场景 1：动态添加新接口

```bash
# 不需要重启服务，直接添加新的路由配置
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "route": {
      "path": "/api/payments",
      "method": "POST",
      "backendUrl": "http://localhost:9093",
      "backendPath": "/v1/payments",
      "responseTransform": {
        "success": true,
        "paymentId": "$.data.payment_id",
        "status": "$.data.status"
      }
    }
  }' \
  http://localhost:8080/admin/routes/add
```

### 场景 2：修改 DSL 转换规则

```bash
# 不需要重启服务，直接更新路由的 DSL 配置
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
        "message": "success",
        "user": {
          "id": "$.data.id",
          "name": "$.data.name",
          "email": "$.data.email"
        },
        "requestMethod": "@ctx.request.method"
      }
    }
  }' \
  http://localhost:8080/admin/routes/update
```

### 场景 3：从数据库加载配置并热更新

```go
// 假设从数据库查询路由配置
routes := queryRoutesFromDatabase()

for _, route := range routes {
    // 通过 HTTP API 添加路由
    addRouteViaAPI(route)
}

// 从数据库查询 Hook 脚本
hooks := queryHooksFromDatabase()

for _, hook := range hooks {
    // 通过 HTTP API 更新 Hook
    updateHookViaAPI(hook.Point, hook.Script)
}
```

### 场景 4：配置版本控制和回滚

```bash
# 1. 保存当前配置
CURRENT_CONFIG=$(curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes)

# 2. 更新配置
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{ ... }' \
  http://localhost:8080/admin/routes/update

# 3. 如果出现问题，回滚配置
# 先删除错误配置
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"path": "/api/users", "method": "POST"}' \
  http://localhost:8080/admin/routes/delete

# 再添加之前的配置
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d "$CURRENT_CONFIG" \
  http://localhost:8080/admin/routes/add
```

---

## 安全建议

1. **Token 管理**
   - 使用强密码作为 admin token
   - 定期轮换 token
   - 不要在代码中硬编码 token

2. **网络隔离**
   - 管理 API 只允许内网访问
   - 使用防火墙限制访问来源

3. **审计日志**
   - 记录所有管理 API 的调用
   - 包括调用者 IP、时间、操作内容

4. **配置备份**
   - 定期备份路由和 Hook 配置
   - 支持快速回滚

---

## 错误处理

所有错误都会返回非 200 状态码和错误信息：

```
HTTP 401 Unauthorized
unauthorized
```

```
HTTP 400 Bad Request
invalid request: unexpected end of JSON input
```

```
HTTP 404 Not Found
failed to update route: route not found: POST /api/xxx
```

```
HTTP 500 Internal Server Error
failed to update hook: script execution error
```
