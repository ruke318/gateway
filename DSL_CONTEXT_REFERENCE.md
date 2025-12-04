# DSL Context 参考

在 DSL 转换中通过 `@ctx.` 前缀访问 Context 数据。

## 数据来源

| 类型 | 语法 | 示例 |
|-----|------|------|
| 固定值 | 直接写字符串 | `"200"` |
| JSONPath | `$.` 前缀 | `"$.data.id"` |
| Context | `@ctx.` 前缀 | `"@ctx.request.method"` |

## 请求信息 (`@ctx.request.*`)

| 字段 | 类型 | 说明 |
|-----|------|------|
| `@ctx.request.method` | string | HTTP 方法 |
| `@ctx.request.path` | string | 请求路径 |
| `@ctx.request.query` | string | 查询字符串 |
| `@ctx.request.host` | string | Host |
| `@ctx.request.header.*` | string | 请求头 |
| `@ctx.request.body.*` | any | 请求体字段 |

示例：

```yaml
responseTransform:
  method: "@ctx.request.method"
  path: "@ctx.request.path"
  userId: "@ctx.request.body.userId"
  auth: "@ctx.request.header.Authorization"
```

## 路由信息 (`@ctx.route.*`)

| 字段 | 类型 | 说明 |
|-----|------|------|
| `@ctx.route.path` | string | 路由路径模式 |
| `@ctx.route.method` | string | 路由方法 |
| `@ctx.route.backendUrl` | string | 后端 URL |
| `@ctx.route.backendPath` | string | 后端路径 |
| `@ctx.route.backendMethod` | string | 后端方法 |

## 响应信息 (`@ctx.response.*`)

仅在 `responseTransform` 中可用。

| 字段 | 类型 | 说明 |
|-----|------|------|
| `@ctx.response.status` | int | HTTP 状态码 |
| `@ctx.response.header.*` | string | 响应头 |

## 自定义数据 (`@ctx.*`)

通过 Hook 设置的数据：

```javascript
// Hook 中设置
context.data.userId = "user-123";
context.data.tenantId = "tenant-001";
context.data.user = {
  id: "user-123",
  name: "John"
};
```

```yaml
# DSL 中访问
responseTransform:
  userId: "@ctx.userId"
  tenantId: "@ctx.tenantId"
  userName: "@ctx.user.name"
```

## 完整示例

```yaml
routes:
  - path: "/api/examines"
    method: "POST"
    responseTransform:
      code: "$.code"
      message: "$.message"

      # 请求信息
      requestMethod: "@ctx.request.method"
      requestUserId: "@ctx.request.body.userId"

      # 响应信息
      httpStatus: "@ctx.response.status"

      # 自定义数据
      tenantId: "@ctx.tenantId"

      # 数组转换
      data:
        items:
          json.path: "$.data"
          id: "$.ID_SRV"
          name: "$.EXAMINE_NAME"
          tenant: "@ctx.tenantId"
```

## 注意事项

- `@ctx.request.*` 在 requestTransform 和 responseTransform 中都可用
- `@ctx.response.*` 仅在 responseTransform 中可用
- 支持多层嵌套访问：`@ctx.user.profile.age`
- 路径不存在时返回 null
