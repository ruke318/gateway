# 使用示例

演示如何使用管理 API 动态管理网关配置。

## 启动服务

```bash
./gateway
# Gateway starting on :8080
```

## 查看当前路由

```bash
curl -H "X-Admin-Token: admin-secret-token" \
  http://localhost:8080/admin/routes | jq
```

## 动态添加接口

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

立即生效，客户端可访问 `/api/products`。

## 修改 DSL 转换

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

## 更新 Hook

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "hookPoint": "BeforeAuth",
    "script": "if (context.requestHeaders.Authorization) { context.data.userId = \"user-123\"; context.data.tenantId = \"tenant-001\"; }"
  }' \
  http://localhost:8080/admin/hooks/update
```

## 删除路由

```bash
curl -X POST \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"path": "/api/products", "method": "GET"}' \
  http://localhost:8080/admin/routes/delete
```

## 生产环境建议

### 配置持久化

```go
// 启动时从数据库加载
func loadConfigFromDB() {
    routes := queryRoutesFromDB()
    for _, route := range routes {
        router.AddRoute(route)
    }
}
```

### 配置备份

```bash
# 定期备份
curl -H "X-Admin-Token: $TOKEN" \
  http://localhost:8080/admin/routes > routes_backup.json
```

### 安全加固

```yaml
# 从环境变量读取 token
adminToken: "${GATEWAY_ADMIN_TOKEN}"
```
