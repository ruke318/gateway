# DSL Context 参考

通过 `@ctx.` 前缀访问 Context 数据。

## 数据来源

| 类型 | 语法 | 示例 |
|-----|------|------|
| 固定值 | 直接写 | `"200"` |
| JSONPath | `$.` 前缀 | `"$.data.id"` |
| Context | `@ctx.` 前缀 | `"@ctx.request.method"` |

## 请求信息 (`@ctx.request.*`)

| 字段 | 说明 |
|-----|------|
| `@ctx.request.method` | HTTP 方法 |
| `@ctx.request.path` | 请求路径 |
| `@ctx.request.query` | 查询字符串 |
| `@ctx.request.host` | Host |
| `@ctx.request.header.*` | 请求头 |
| `@ctx.request.body.*` | 请求体字段 |

## 路由信息 (`@ctx.route.*`)

| 字段 | 说明 |
|-----|------|
| `@ctx.route.service_id` | 接口标识 |
| `@ctx.route.backendUrl` | 后端 URL |
| `@ctx.route.backendPath` | 后端路径 |
| `@ctx.route.backendMethod` | 后端方法 |

## 响应信息 (`@ctx.response.*`)

仅在 responseTransform 中可用。

| 字段 | 说明 |
|-----|------|
| `@ctx.response.status` | HTTP 状态码 |
| `@ctx.response.header.*` | 响应头 |

## 机构配置 (`@ctx.org_config.*`)

访问机构的 config JSON 字段。

```json
{
  "apiKey": "@ctx.org_config.apiKey",
  "secret": "@ctx.org_config.secret"
}
```

## 自定义数据 (`@ctx.*`)

通过 Hook 设置：

```javascript
context.data.userId = "user-123";
context.data.token = "xxx";
```

DSL 中访问：

```json
{
  "userId": "@ctx.userId",
  "token": "@ctx.token"
}
```

## 数组转换

```json
{
  "items": {
    "json.path": "$.data",
    "id": "$.ID",
    "name": "$.NAME",
    "tenant": "@ctx.tenantId"
  }
}
```

## 注意事项

- `@ctx.request.*` 在 requestTransform 和 responseTransform 中都可用
- `@ctx.response.*` 仅在 responseTransform 中可用
- 支持多层嵌套：`@ctx.user.profile.age`
- 路径不存在时返回 null
