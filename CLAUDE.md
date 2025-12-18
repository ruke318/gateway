# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gateway is a lightweight API gateway for unifying external vendor API integrations. It provides protocol conversion, DSL-based field mapping, JavaScript Hook extensions, and multi-tenant management.

**Core Value:** Internal business systems call a unified gateway API, and the gateway handles protocol conversion, data mapping, and interface adaptation regardless of how external vendor APIs change.

## Development Commands

### Backend (Go)

```bash
# Build
cd backend
go build -o gateway .

# Run
./gateway

# Run with auto-reload (if using air)
air

# Test
go test ./...                    # All tests
go test ./transform              # Specific package
go test -v ./transform           # Verbose output

# Format and lint
go fmt ./...
go vet ./...
```

### Frontend (Vue 3)

```bash
cd front

# Development server (http://localhost:3000)
npm run dev

# Production build
npm run build

# Preview production build
npm run preview
```

### Configuration

Backend configuration: `backend/config/config.yaml`

```yaml
port: ":8081"                    # Server port
db:
  host: "localhost"
  port: 3306
  database: "gateway"
vmPool:
  size: 100                      # JavaScript VM pool size
```

Frontend proxy automatically forwards `/admin/*` and `/api/*` to `http://localhost:8081`.

## Architecture

### Three-Tier Data Model

```
Vendor (厂商) → Organization (机构) → Service (接口)
     ↓               ↓                    ↓
  支付宝、微信      总部、分公司        支付、退款、查询
```

**Query Pattern:** Service JOIN Organization JOIN Vendor
- Uses `vendor.code = com_id` and `organization.code = unit_id` and `service.service_id = service_id`

### Request Flow

```
Invoke Request → Parse → Load Config (3-table join)
                    ↓
          Build Hook Context (包含 request, route, org_config)
                    ↓
          认证 Hooks (BeforeAuth → AfterAuth)
                    ↓
          请求转换 (BeforeRequestTransform → DSL → AfterRequestTransform)
                    ↓
          转发前 Hook (BeforeForward) → HTTP Forward → 转发后 Hook (AfterForward)
                    ↓
          响应转换 (BeforeResponseTransform → DSL → AfterResponseTransform)
                    ↓
          Return Response
```

### Key Components

**handler/invoke.go** - Main request processor
- `Invoke()` - Handles `/gateway/v1/invoke` requests
- `HandleNotify()` - Handles vendor callbacks at `/gateway/v1/notify/{unit_id}/{service_id}/{channel}`
- `invokeContext` struct holds request state throughout the pipeline

**hook/executor.go** - JavaScript execution engine
- VM Pool: Pre-created Goja VMs for concurrency
- `ExecuteHooks()` - Executes Hook scripts at specific execution points
- `resetVM()` - Cleans VM state between requests (IMPORTANT: prevents data leakage)

**hook/dictionary.go** - Dictionary management (NEW)
- Manages organization-level field mappings
- `LoadDictionary()` - Loads all configs from DB into memory cache
- `GenerateDictionaryJS()` - Generates JavaScript dict functions injected into VM

**transform/dsl.go** - DSL transformation engine
- Supports JSONPath (`$.req.field`), Context injection (`@ctx.field`)
- **NEW:** Function calls (`@fn.dict.get("type", "key")`)
- `processFunctionCall()` - Parses and executes @fn. expressions using VM pool

**hook/builtin.go** - Built-in JavaScript modules
- `crypto` - MD5, SHA, HMAC, AES, DES, RSA
- `http` - GET, POST, custom requests
- `encoding` - Base64, Hex, JSON, XML, URL
- `util` - Timestamps, UUID, time formatting
- `dict` - Dictionary lookup functions (NEW)

**hook/library.go** - Global function library
- User-defined JavaScript functions stored in `script_libraries` table
- Accessible as `lib.namespace.functionName()` in all Hooks and DSL

### Hook Execution Points (9 total)

1. `BeforeAuth` / `AfterAuth` - Authentication
2. `BeforeRequestTransform` / `AfterRequestTransform` - Request transformation
3. `BeforeForward` / `AfterForward` - ⭐ Most commonly used (signing, Token, decryption)
4. `BeforeResponseTransform` / `AfterResponseTransform` - Response transformation
5. `OnError` - Error handling

### DSL Syntax

| Syntax | Description | Example |
|--------|-------------|---------|
| `"literal"` | String/number/boolean | `"success"`, `200`, `true` |
| `$.path` | JSONPath query | `$.req.order_no` |
| `@ctx.path` | Context injection | `@ctx.org_config.appId` |
| `@fn.func()` | Function call (NEW) | `@fn.dict.get("payment_method", "ALIPAY")` |

### Hook Context Structure

```javascript
context.data = {
    request: {
        body: {
            com_id: "alipay",
            unit_id: "org001",    // ⚠️ 字典函数从这里获取机构ID
            service_id: "pay",
            biz_no: "BIZ123",
            req: {...}
        }
    },
    route: {
        backendUrl: "...",
        backendPath: "...",
        backendMethod: "..."
    },
    org_config: {  // 机构配置（敏感信息）
        appId: "...",
        secret: "...",
        privateKey: "..."
    }
}
```

## Important Implementation Details

### Dictionary System (NEW)

**Database:** `dictionary_config` table
- Structure: `org_id`, `dict_type`, `dict_key`, `dict_value`
- Cache: All loaded into memory on startup (`hook.LoadDictionary()`)
- Thread-safe: Uses `sync.RWMutex`

**Auto-reload mechanism:**
- When config changes, call `hook.ReloadDictionary()`
- This clears VM pool to inject new dictionary data
- No service restart needed

**JavaScript API:**
```javascript
dict.get(dictType, key)                          // Auto uses current org_id
dict.reverseGet(dictType, value)                 // Value → Key
dict.translate(toOrg, dictType, value)           // Cross-org translation
dict.translateFull(fromOrg, toOrg, dictType, value)
```

### VM Pool Architecture

- **Pool Size:** Configurable (default 100)
- **VM Lifecycle:** Get from pool → Execute → Reset (clear context) → Return to pool
- **Thread Safety:** Each request gets isolated VM instance
- **Injection:** On VM creation, inject builtin modules + global library + dictionary

⚠️ **CRITICAL:** Always call `resetVM()` before returning to pool to prevent data leakage between requests.

### DSL Function Call Execution

When DSL encounters `@fn.funcPath(args...)`:
1. Parse function expression using regex
2. Resolve arguments (supports `$.`, `@ctx.`, literals)
3. Get VM from pool
4. Set `context.data` to VM
5. Execute: `(funcPath).apply(null, [args])`
6. Return result to DSL mapping

This allows calling any JavaScript function (dict, crypto, lib.*, etc.) directly in DSL without writing Hooks.

### Callback System

**URL Format:** `/gateway/v1/notify/{unit_id}/{service_id}/{channel}`

**Flow:** Vendor Callback → NotifyProcessor → InvokeRequest → Full invoke pipeline

**Processors:**
- Numeric channel (e.g., "1") → `DefaultNotifyProcessor`
- String channel (e.g., "alipay") → Custom processor (e.g., `AlipayNotifyProcessor`)

## Authentication

**Management API:** JWT authentication + admin role required
- Middleware chain: `auth.AuthMiddleware` → `auth.AdminMiddleware` → `middleware.OperationLogMiddleware`
- Routes: `/admin/db/*`

**User API:** JWT authentication
- Middleware: `auth.AuthMiddleware`
- Routes: `/api/*`

**Public API:** No authentication
- Routes: `/gateway/v1/invoke`, `/gateway/v1/notify/*`

## Database Schema

Key tables:
- `vendor` - External API providers
- `organization` - Internal clients (stores config JSON: appId, secret, keys)
- `service` - Interface configs (DSL mappings, backend URL/path)
- `hook_script` - JavaScript Hook code
- `service_hook` - Service ↔ Hook associations (with priority, execution_point)
- `script_library` - Global JavaScript functions
- `dictionary_config` - Organization-level field mappings (NEW)
- `user` - User accounts (bcrypt passwords)
- `operation_log` - Audit trail

**Auto-migration:** `database.AutoMigrate()` on startup

## Common Patterns

### Adding New Hook Module

1. Define Go functions in `hook/builtin.go`
2. Register in `RegisterBuiltins()`:
   ```go
   vm.Set("myModule", map[string]interface{}{
       "myFunc": goFunction,
   })
   ```

### Adding New Dictionary Function

1. Add JavaScript function in `hook/dictionary.go` → `GenerateDictionaryJS()`
2. Function auto-injected to all VMs on creation

### Adding New DSL Function Support

Functions are automatically callable in DSL via `@fn.` prefix if they exist in VM.
No code changes needed - just ensure function is registered in VM.

### Adding New Notify Processor

```go
type MyProcessor struct{}

func (p *MyProcessor) Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error) {
    // Convert callback data to InvokeRequest
}

// Register in notify_processor.go
var notifyProcessors = map[string]NotifyProcessor{
    "mychannel": &MyProcessor{},
}
```

## Testing

### Backend Tests

```bash
cd backend
go test ./transform/dsl_test.go -v       # DSL transformation tests
```

### Manual API Testing

```bash
# Login to get JWT token
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Use token for admin APIs
curl -X GET http://localhost:8081/admin/db/vendors \
  -H "Authorization: Bearer <token>"

# Invoke API (no auth required)
curl -X POST http://localhost:8081/gateway/v1/invoke \
  -H "Content-Type: application/json" \
  -d '{"com_id":"alipay","unit_id":"org001","service_id":"pay","biz_no":"BIZ123","req":{}}'
```

## Critical Notes

1. **VM Pool Safety:** Always reset VM context after use to prevent data leakage
2. **Dictionary Reload:** After modifying dictionary configs, call `ReloadDictionary()` which clears VM pool
3. **Context Structure:** Dictionary functions expect `context.data.request.body.unit_id` to exist
4. **DSL Function Execution:** Uses VM pool, so dictionary data must be injected on VM creation
5. **Hook vs DSL Functions:** Prefer DSL functions for simple transformations, Hooks for complex logic
6. **Authentication:** All `/admin/db/*` routes require JWT + admin role (unified, no more X-Admin-Token)

## Default Credentials

- Username: `admin`
- Password: `admin123`
- Created by: `database.InitDefaultData()`

## Port Configuration

- Backend: `:8081` (configurable in `config/config.yaml`)
- Frontend dev server: `3000` (proxies to backend)
