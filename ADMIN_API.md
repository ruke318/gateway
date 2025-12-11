# Admin API Reference

All admin endpoints require `X-Admin-Token` header authentication.

**Base Path:** `/admin/db`

## Vendor Management

Manage external API providers.

```bash
GET    /admin/db/vendors          # List all vendors
GET    /admin/db/vendor/{id}      # Get vendor details
POST   /admin/db/vendor           # Create vendor
PUT    /admin/db/vendor/{id}      # Update vendor
DELETE /admin/db/vendor/{id}      # Delete vendor
```

**Request Body:**
```json
{
  "code": "vendor001",
  "name": "Vendor Name",
  "base_url": "http://api.vendor.com",
  "description": "Optional description"
}
```

## Organization Management

Manage internal organizations.

```bash
GET    /admin/db/organizations       # List all organizations
GET    /admin/db/organization/{id}   # Get organization details
POST   /admin/db/organization        # Create organization
PUT    /admin/db/organization/{id}   # Update organization
DELETE /admin/db/organization/{id}   # Delete organization
```

**Request Body:**
```json
{
  "code": "org001",
  "name": "Organization Name",
  "config": {
    "apiKey": "xxx",
    "secret": "yyy"
  },
  "description": "Optional description"
}
```

## Service Management

Manage API service configurations.

```bash
GET    /admin/db/services        # List all services
GET    /admin/db/service/{id}    # Get service details
POST   /admin/db/service         # Create service
PUT    /admin/db/service/{id}    # Update service
DELETE /admin/db/service/{id}    # Delete service
```

**Request Body:**
```json
{
  "service_id": "getUserInfo",
  "org_id": 1,
  "vendor_id": 1,
  "name": "Get User Information",
  "description": "Optional description",
  "backend_url": "http://api.example.com",
  "backend_path": "/v1/user/{userId}",
  "backend_method": "POST",
  "body_type": "json",
  "request_transform": {
    "userId": "$.req.id"
  },
  "response_transform": {
    "code": "$.code",
    "data": "$.result"
  }
}
```

**Fields:**
- `backend_url`: Optional, overrides vendor's base_url
- `backend_path`: Backend API path, supports `{key}` placeholders
- `backend_method`: HTTP method (GET, POST, PUT, DELETE)
- `body_type`: Request body type (json, form, xml)
- `request_transform`: DSL configuration for request transformation
- `response_transform`: DSL configuration for response transformation

## Hook Script Management

Manage reusable hook scripts.

```bash
GET    /admin/db/hook-scripts       # List all hook scripts
GET    /admin/db/hook-script/{id}   # Get hook script details
POST   /admin/db/hook-script        # Create hook script
PUT    /admin/db/hook-script/{id}   # Update hook script
DELETE /admin/db/hook-script/{id}   # Delete hook script
```

**Request Body:**
```json
{
  "name": "Signature Script",
  "script_content": "var body = JSON.parse(ctx.request.body); body.sign = crypto.md5(body.data); ctx.request.body = JSON.stringify(body);",
  "description": "Optional description"
}
```

## Common Script Library Management

Manage shared JavaScript functions.

```bash
GET    /admin/db/scripts         # List all common scripts
GET    /admin/db/script/{id}     # Get script details
POST   /admin/db/script          # Create script
PUT    /admin/db/script/{id}     # Update script
DELETE /admin/db/script/{id}     # Delete script
POST   /admin/db/reload-library  # Reload script library
```

**Note:** After updating common scripts, call `/admin/db/reload-library` to reload them into memory.

## Service Hook Association

Associate hooks with services.

```bash
GET    /admin/db/service-hooks       # List all service hooks
GET    /admin/db/service-hook/{id}   # Get service hook details
POST   /admin/db/service-hook        # Create service hook
PUT    /admin/db/service-hook/{id}   # Update service hook
DELETE /admin/db/service-hook/{id}   # Delete service hook
```

**Request Body:**
```json
{
  "service_pk": 1,
  "hook_point": "BeforeForward",
  "script_id": 1,
  "inline_script": "",
  "priority": 0
}
```

**Fields:**
- `service_pk`: Service ID (primary key)
- `hook_point`: Hook point name (see below)
- `script_id`: Reference to hook script (optional if using inline_script)
- `inline_script`: Inline JavaScript code (optional if using script_id)
- `priority`: Execution priority (lower number = higher priority)

## Hook Points

| Hook Point | Description |
|------------|-------------|
| `BeforeAuth` | Before authentication |
| `AfterAuth` | After authentication |
| `BeforeRequestTransform` | Before request DSL transformation |
| `AfterRequestTransform` | After request DSL transformation |
| `BeforeForward` | Before forwarding to backend |
| `AfterForward` | After receiving backend response |
| `BeforeResponseTransform` | Before response DSL transformation |
| `AfterResponseTransform` | After response DSL transformation |
| `OnError` | On error occurrence |

## Error Responses

| Status Code | Description |
|-------------|-------------|
| 401 | Unauthorized (invalid or missing token) |
| 400 | Bad Request (invalid request format) |
| 404 | Not Found (resource does not exist) |
| 500 | Internal Server Error |

**Error Response Format:**
```json
{
  "error": "Error message description"
}
```
