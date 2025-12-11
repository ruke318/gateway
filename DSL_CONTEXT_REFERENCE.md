# DSL Context Reference

Access context data using the `@ctx.` prefix in DSL transformations.

## Data Sources

| Type | Syntax | Example |
|------|--------|---------|
| **Literal** | Direct value | `"200"` |
| **JSONPath** | `$.` prefix | `"$.data.id"` |
| **Context** | `@ctx.` prefix | `"@ctx.request.method"` |

## Request Information (`@ctx.request.*`)

Available in both request and response transformations.

| Field | Description |
|-------|-------------|
| `@ctx.request.method` | HTTP method (GET, POST, etc.) |
| `@ctx.request.path` | Request path |
| `@ctx.request.query` | Query string |
| `@ctx.request.host` | Host header |
| `@ctx.request.header.*` | Request headers |
| `@ctx.request.body.*` | Request body fields |

## Route Information (`@ctx.route.*`)

| Field | Description |
|-------|-------------|
| `@ctx.route.service_id` | Service identifier |
| `@ctx.route.backendUrl` | Backend URL |
| `@ctx.route.backendPath` | Backend path |
| `@ctx.route.backendMethod` | Backend HTTP method |

## Response Information (`@ctx.response.*`)

Only available in response transformations.

| Field | Description |
|-------|-------------|
| `@ctx.response.status` | HTTP status code |
| `@ctx.response.header.*` | Response headers |

## Organization Config (`@ctx.org_config.*`)

Access organization's config JSON fields.

```json
{
  "apiKey": "@ctx.org_config.apiKey",
  "secret": "@ctx.org_config.secret"
}
```

## Custom Data (`@ctx.*`)

Set custom data in hooks:

```javascript
ctx.data.userId = "user-123";
ctx.data.token = "xxx";
```

Access in DSL:

```json
{
  "userId": "@ctx.userId",
  "token": "@ctx.token"
}
```

## Array Transformation

Use `json.path` to iterate over arrays:

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

## Notes

- `@ctx.request.*` is available in both request and response transformations
- `@ctx.response.*` is only available in response transformations
- Supports nested paths: `@ctx.user.profile.age`
- Returns null if path does not exist
