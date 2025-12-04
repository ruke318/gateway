# 管理 API 文档

Gateway 提供管理 API 用于运行时动态管理路由和 Hook，无需重启服务。

## 认证

所有接口需要 `X-Admin-Token` Header：

```bash
curl -H "X-Admin-Token: admin-secret-token" http://localhost:8080/admin/routes
```

## 路由管理

### 查询所有路由

```bash
GET /admin/routes
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
      "responseTransform": {"code": "200", "data": "$.data"}
    }
  ]
}
```

### 添加路由

```bash
POST /admin/routes/add
```

请求体：

```json
{
  "route": {
    "path": "/api/orders",
    "method": "POST",
    "backendUrl": "http://localhost:9090",
    "backendPath": "/v1/orders",
    "responseTransform": {
      "code": "200",
      "orderId": "$.data.id"
    },
    "hooks": {
      "BeforeAuth": "context.data.userId = 'user-123';",
      "AfterResponseTransform": "console.log('done');"
    }
  }
}
```

### 更新路由

```bash
POST /admin/routes/update
```

根据 `path` 和 `method` 匹配并更新。

### 删除路由

```bash
POST /admin/routes/delete
```

请求体：

```json
{
  "path": "/api/orders",
  "method": "POST"
}
```

## Hook 管理

### 更新 Hook

```bash
POST /admin/hooks/update
```

请求体：

```json
{
  "hookPoint": "BeforeAuth",
  "script": "context.data.userId = '123';"
}
```

支持的 hookPoint：

- BeforeAuth
- AfterAuth
- BeforeRequestTransform
- AfterRequestTransform
- BeforeForward
- AfterForward
- BeforeResponseTransform
- AfterResponseTransform
- OnError

### 清空 Hook

```bash
POST /admin/hooks/clear
```

请求体：

```json
{
  "hookPoint": "BeforeAuth"
}
```

## 错误响应

| 状态码 | 说明 |
|-------|------|
| 401 | 未认证 |
| 400 | 请求格式错误 |
| 404 | 路由不存在 |
| 500 | 服务器错误 |

## 安全建议

- 使用强密码作为 admin token
- 限制管理 API 只允许内网访问
- 记录所有配置变更的审计日志
- 定期备份配置
