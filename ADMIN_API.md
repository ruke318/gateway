# 管理 API 文档

所有接口前缀 `/admin/db`，需要 `X-Admin-Token` Header 认证。

## 厂商管理

```bash
GET    /admin/db/vendors          # 列表
GET    /admin/db/vendor/{id}      # 详情
POST   /admin/db/vendor           # 创建
PUT    /admin/db/vendor/{id}      # 更新
DELETE /admin/db/vendor/{id}      # 删除
```

请求体：
```json
{
  "code": "vendor001",
  "name": "厂商名称",
  "base_url": "http://api.vendor.com",
  "description": "描述"
}
```

## 机构管理

```bash
GET    /admin/db/organizations       # 列表
GET    /admin/db/organization/{id}   # 详情
POST   /admin/db/organization        # 创建
PUT    /admin/db/organization/{id}   # 更新
DELETE /admin/db/organization/{id}   # 删除
```

请求体：
```json
{
  "code": "org001",
  "name": "机构名称",
  "config": {"key": "value"},
  "description": "描述"
}
```

## 接口管理

```bash
GET    /admin/db/services        # 列表
GET    /admin/db/service/{id}    # 详情
POST   /admin/db/service         # 创建
PUT    /admin/db/service/{id}    # 更新
DELETE /admin/db/service/{id}    # 删除
```

请求体：
```json
{
  "service_id": "getUserInfo",
  "org_id": 1,
  "vendor_id": 1,
  "name": "获取用户信息",
  "backend_path": "/v1/user/{userId}",
  "backend_method": "POST",
  "body_type": "json",
  "request_transform": {"userId": "$.req.id"},
  "response_transform": {"code": "$.code", "data": "$.result"}
}
```

## Hook 脚本管理

```bash
GET    /admin/db/hook-scripts       # 列表
GET    /admin/db/hook-script/{id}   # 详情
POST   /admin/db/hook-script        # 创建
PUT    /admin/db/hook-script/{id}   # 更新
DELETE /admin/db/hook-script/{id}   # 删除
```

## 公共函数库管理

```bash
GET    /admin/db/scripts         # 列表
GET    /admin/db/script/{id}     # 详情
POST   /admin/db/script          # 创建
PUT    /admin/db/script/{id}     # 更新
DELETE /admin/db/script/{id}     # 删除
POST   /admin/db/reload-library  # 重载
```

## 接口 Hook 关联

```bash
GET    /admin/db/service-hooks       # 列表
GET    /admin/db/service-hook/{id}   # 详情
POST   /admin/db/service-hook        # 创建
PUT    /admin/db/service-hook/{id}   # 更新
DELETE /admin/db/service-hook/{id}   # 删除
```

请求体：
```json
{
  "service_pk": 1,
  "hook_point": "BeforeAuth",
  "script_id": 1,
  "inline_script": "",
  "priority": 0
}
```

## Hook 节点

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

## 错误响应

| 状态码 | 说明 |
|-------|------|
| 401 | 未认证 |
| 400 | 请求格式错误 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |
