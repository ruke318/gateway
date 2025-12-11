# Usage Examples

## Start the Gateway

```bash
cd backend
go build -o gateway .
./gateway
```

## Invoke API

Send a request through the gateway:

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

## Admin API Examples

### Create Vendor

```bash
curl -X POST http://localhost:8080/admin/db/vendor \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "vendor001",
    "name": "Example Vendor",
    "base_url": "http://api.example.com"
  }'
```

### Create Organization

```bash
curl -X POST http://localhost:8080/admin/db/organization \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "org001",
    "name": "Example Organization",
    "config": {
      "apiKey": "xxx",
      "secret": "yyy"
    }
  }'
```

### Create Service

```bash
curl -X POST http://localhost:8080/admin/db/service \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "getUserInfo",
    "org_id": 1,
    "vendor_id": 1,
    "name": "Get User Info",
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

### Create Hook Script

```bash
curl -X POST http://localhost:8080/admin/db/hook-script \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Signature Script",
    "script_content": "var body = JSON.parse(ctx.request.body); body.sign = crypto.md5(body.data + ctx.data.org_config.secret); ctx.request.body = JSON.stringify(body);"
  }'
```

### Associate Hook with Service

```bash
curl -X POST http://localhost:8080/admin/db/service-hook \
  -H "X-Admin-Token: your-admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "service_pk": 1,
    "hook_point": "BeforeForward",
    "script_id": 1,
    "priority": 0
  }'
```

## Hook Script Examples

### Add Signature

```javascript
// BeforeForward hook - Add signature to request
var body = JSON.parse(ctx.request.body);
var timestamp = util.now();
var sign = crypto.hmacSHA256(
  body.data + timestamp,
  ctx.data.org_config.secret
);
body.timestamp = timestamp;
body.sign = sign;
ctx.request.body = JSON.stringify(body);
```

### Fetch External Token

```javascript
// BeforeForward hook - Get token from external auth service
var resp = http.postJSON("http://auth.example.com/token", {
  appId: ctx.data.org_config.appId,
  secret: ctx.data.org_config.secret
});
if (resp.status === 200) {
  ctx.request.headers["Authorization"] = "Bearer " + resp.json.token;
}
```

### Modify Backend Route

```javascript
// BeforeForward hook - Dynamic routing
ctx.data.route.backendUrl = "http://backup-api.example.com";
ctx.data.route.backendPath = "/v2/user";
```

### Error Handling

```javascript
// OnError hook - Custom error response
ctx.response.body = JSON.stringify({
  code: "500",
  message: "Service temporarily unavailable",
  error: ctx.error
});
ctx.response.status = 500;
```

## DSL Transform Examples

### Request Transform

Transform incoming request to backend format:

```json
{
  "userId": "$.req.userId",
  "timestamp": "@ctx.timestamp",
  "channel": "gateway"
}
```

### Response Transform

Transform backend response to unified format:

```json
{
  "code": "$.code",
  "message": "$.msg",
  "data": {
    "id": "$.data.userId",
    "name": "$.data.userName",
    "org": "@ctx.org_config.orgName"
  }
}
```

### Array Transform

Transform array data with field mapping:

```json
{
  "total": "$.total",
  "items": {
    "json.path": "$.data.list",
    "id": "$.ID",
    "name": "$.NAME",
    "status": "$.STATUS"
  }
}
```
