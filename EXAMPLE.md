# 使用示例

## 启动服务

```bash
cd backend
./gateway
```

## 调用接口

```bash
curl -X POST http://localhost:8080/gateway/v1/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "com_id": "vendor001",
    "unit_id": "org001",
    "service_id": "getUserInfo",
    "biz_no": "BIZ20231201001",
    "req": {
      "userId": "12345"
    }
  }'
```

## 管理 API 示例

### 创建厂商

```bash
curl -X POST http://localhost:8080/admin/db/vendor \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "vendor001",
    "name": "示例厂商",
    "base_url": "http://api.example.com"
  }'
```

### 创建机构

```bash
curl -X POST http://localhost:8080/admin/db/organization \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "org001",
    "name": "示例机构",
    "config": {"apiKey": "xxx", "secret": "yyy"}
  }'
```

### 创建接口

```bash
curl -X POST http://localhost:8080/admin/db/service \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "getUserInfo",
    "org_id": 1,
    "vendor_id": 1,
    "name": "获取用户信息",
    "backend_path": "/v1/user/{userId}",
    "backend_method": "GET",
    "body_type": "json",
    "response_transform": {
      "code": "$.code",
      "data": {
        "id": "$.data.id",
        "name": "$.data.name"
      }
    }
  }'
```

### 创建 Hook 脚本

```bash
curl -X POST http://localhost:8080/admin/db/hook-script \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "签名脚本",
    "script_content": "var body = JSON.parse(context.requestBody); body.sign = crypto.md5(body.data + context.data.org_config.secret); context.requestBody = JSON.stringify(body);"
  }'
```

### 关联 Hook 到接口

```bash
curl -X POST http://localhost:8080/admin/db/service-hook \
  -H "X-Admin-Token: admin-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service_pk": 1,
    "hook_point": "BeforeForward",
    "script_id": 1,
    "priority": 0
  }'
```

## Hook 脚本示例

### 添加签名

```javascript
var body = JSON.parse(context.requestBody);
var timestamp = util.now();
var sign = crypto.hmacSHA256(
  body.data + timestamp,
  context.data.org_config.secret
);
body.timestamp = timestamp;
body.sign = sign;
context.requestBody = JSON.stringify(body);
```

### 调用外部接口获取 Token

```javascript
var resp = http.postJSON("http://auth.example.com/token", {
  appId: context.data.org_config.appId,
  secret: context.data.org_config.secret
});
if (resp.status === 200) {
  context.requestHeaders["Authorization"] = "Bearer " + resp.json.token;
}
```

### 修改后端路由

```javascript
context.data.route.backendUrl = "http://backup-api.example.com";
context.data.route.backendPath = "/v2/user";
```
