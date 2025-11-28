# 配置热更新使用示例

## 场景：不重启服务动态管理接口

假设你正在运行一个网关服务，现在需要：
1. 添加一个新的产品查询接口
2. 修改用户接口的 DSL 转换规则
3. 更新认证 Hook 的逻辑

传统方式需要修改配置文件并重启服务，现在可以通过管理 API 实时更新！

---

## 步骤 1: 启动网关服务

```bash
./gateway
```

输出：
```
2025/11/28 19:33:59 Gateway starting on :8080
```

网关已启动，此时只有配置文件中的路由。

---

## 步骤 2: 查看当前路由配置

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes | jq
```

响应：
```json
{
  "success": true,
  "data": [
    {
      "path": "/api/users",
      "method": "POST",
      "backendUrl": "http://localhost:9090",
      "backendPath": "/v1/users",
      "backendMethod": "",
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

## 步骤 3: 动态添加产品接口（无需重启）

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "route": {
      "path": "/api/products",
      "method": "GET",
      "backendUrl": "http://localhost:9091",
      "backendPath": "/products/list",
      "responseTransform": {
        "success": true,
        "total": "$.total",
        "items": {
          "json.path": "$.data",
          "id": "$.product_id",
          "name": "$.product_name",
          "price": "$.price"
        }
      }
    }
  }' \
  http://localhost:8080/admin/routes/add
```

响应：
```json
{
  "success": true,
  "message": "route added successfully"
}
```

**立即生效！** 现在客户端可以访问 `/api/products` 接口了。

---

## 步骤 4: 测试新添加的接口

```bash
# 假设后端服务返回：
# {
#   "total": 100,
#   "data": [
#     {"product_id": "P001", "product_name": "iPhone", "price": 999},
#     {"product_id": "P002", "product_name": "iPad", "price": 599}
#   ]
# }

curl http://localhost:8080/api/products
```

经过 DSL 转换后的响应：
```json
{
  "success": true,
  "total": 100,
  "items": [
    {
      "id": "P001",
      "name": "iPhone",
      "price": 999
    },
    {
      "id": "P002",
      "name": "iPad",
      "price": 599
    }
  ]
}
```

---

## 步骤 5: 修改用户接口的 DSL 规则（无需重启）

假设现在需要在响应中添加请求方法信息：

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
        "message": "success",
        "userId": "$.data.id",
        "userName": "$.data.name",
        "requestMethod": "@ctx.request.method",
        "requestPath": "@ctx.request.path"
      }
    }
  }' \
  http://localhost:8080/admin/routes/update
```

响应：
```json
{
  "success": true,
  "message": "route updated successfully"
}
```

**立即生效！** 下一个请求就会包含 `requestMethod` 和 `requestPath` 字段。

---

## 步骤 6: 动态更新认证 Hook（无需重启）

假设需要修改认证逻辑，添加租户信息：

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "hookPoint": "BeforeAuth",
    "script": "if (context.requestHeaders.Authorization) { const token = context.requestHeaders.Authorization.replace(\"Bearer \", \"\"); context.data.userId = \"user-123\"; context.data.tenantId = \"tenant-001\"; context.data.role = \"admin\"; console.log(\"Auth successful:\", context.data.userId, \"Tenant:\", context.data.tenantId); }"
  }' \
  http://localhost:8080/admin/hooks/update
```

响应：
```json
{
  "success": true,
  "message": "hook updated successfully"
}
```

**立即生效！** 下一个请求就会执行新的认证逻辑。

---

## 步骤 7: 验证配置更新

再次查看所有路由：

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes | jq
```

响应：
```json
{
  "success": true,
  "data": [
    {
      "path": "/api/users",
      "method": "POST",
      "backendUrl": "http://localhost:9090",
      "backendPath": "/v1/users",
      "responseTransform": {
        "code": "200",
        "message": "success",
        "userId": "$.data.id",
        "userName": "$.data.name",
        "requestMethod": "@ctx.request.method",
        "requestPath": "@ctx.request.path"
      }
    },
    {
      "path": "/api/products",
      "method": "GET",
      "backendUrl": "http://localhost:9091",
      "backendPath": "/products/list",
      "responseTransform": {
        "success": true,
        "total": "$.total",
        "items": {
          "json.path": "$.data",
          "id": "$.product_id",
          "name": "$.product_name",
          "price": "$.price"
        }
      }
    }
  ]
}
```

---

## 步骤 8: 删除不需要的路由（无需重启）

假设产品接口不再需要：

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

响应：
```json
{
  "success": true,
  "message": "route deleted successfully"
}
```

**立即生效！** 客户端访问 `/api/products` 会返回 404。

---

## 总结

整个过程中：
- ✅ **没有重启服务**
- ✅ 动态添加了新接口
- ✅ 修改了 DSL 转换规则
- ✅ 更新了 Hook 脚本
- ✅ 删除了不需要的路由
- ✅ 所有配置立即生效
- ✅ 正在处理的请求不受影响

这就是配置热更新的强大之处！

---

## 生产环境建议

### 1. 配置持久化

管理 API 只修改内存配置。建议配合数据库使用：

```go
// 启动时从数据库加载配置
func loadConfigFromDB() {
    routes := queryRoutesFromDB()
    for _, route := range routes {
        router.AddRoute(route)
    }

    hooks := queryHooksFromDB()
    for _, hook := range hooks {
        hookManager.UpdateHook(parseHookPoint(hook.Point), hook.Script)
    }
}

// 通过管理 API 更新时，同时更新数据库
// 这样服务重启时配置不会丢失
```

### 2. 安全加固

```yaml
# 建议在配置文件中管理 admin token
adminToken: "${GATEWAY_ADMIN_TOKEN}"  # 从环境变量读取

# 限制管理 API 只能从内网访问
# 使用防火墙或反向代理规则
```

### 3. 审计日志

```go
// 记录所有配置变更
log.Printf("[AUDIT] Route added by IP=%s, Route=%+v", clientIP, route)
log.Printf("[AUDIT] Hook updated by IP=%s, HookPoint=%s", clientIP, hookPoint)
```

### 4. 配置备份

```bash
# 定期备份配置
0 */6 * * * curl -H "X-Admin-Token: $ADMIN_TOKEN" \
  http://localhost:8080/admin/routes > /backup/routes_$(date +\%Y\%m\%d_\%H\%M\%S).json
```
