<div align="center">

# Gateway

**A lightweight, extensible API Gateway built with Go**

[English](./readme.md) | [简体中文](./readme_zh.md)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/yourusername/gateway/pulls)

Declarative DSL data transformation • JavaScript Hook system • Multi-tenant architecture

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Screenshots](#screenshots)

</div>

---

## Features

- 🏢 **Multi-Tenant Architecture** - Three-tier architecture (Vendor/Organization/Service) with many-to-many relationships
- 🔄 **DSL Transformation** - Declarative data transformation using JSONPath + Context syntax
- 🪝 **Hook System** - JavaScript engine powered by goja with 9 lifecycle hooks
- 🛠️ **Rich Built-in Functions** - Crypto, HTTP, encoding, utilities and more
- 📦 **Multiple Content Types** - Support for JSON, Form, and XML format conversion
- 🎯 **Dynamic Routing** - Path templates with `{key}` placeholders
- 🔥 **Hot Reload** - Zero-downtime configuration updates via Admin API
- 🎨 **Web UI** - Complete web-based management interface

## Quick Start

### Prerequisites

- Go 1.21 or higher
- MySQL 5.7+ (or PostgreSQL/SQLite)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/gateway.git
cd gateway/backend

# Build
go build -o gateway .

# Run
./gateway
```

The gateway will start on `:8080` by default.

### Basic Usage

Send requests to the unified gateway endpoint:

```bash
POST /gateway/v1/invoke
Content-Type: application/json
```

```json
{
  "com_id": "vendor_code",
  "unit_id": "organization_code",
  "service_id": "service_identifier",
  "biz_no": "business_transaction_id",
  "req": {
    "your": "business_data"
  }
}
```

## Architecture

### Data Model

| Entity | Description |
|--------|-------------|
| **Vendor** | External API providers |
| **Organization** | Internal organizations with many-to-many vendor relationships |
| **Service** | API endpoints with transformation rules and hooks |

### Request Flow

```
Request → BeforeAuth → AfterAuth → BeforeRequestTransform
        → DSL Request Transform → AfterRequestTransform → BeforeForward
        → Proxy Forward → AfterForward → BeforeResponseTransform
        → DSL Response Transform → AfterResponseTransform → Response
        (OnError on any failure)
```

## DSL Transformation

Transform data declaratively using three syntax types:

| Type | Syntax | Example |
|------|--------|---------|
| **Literal** | Direct value | `"200"` |
| **JSONPath** | `$.` prefix | `"$.data.id"` |
| **Context** | `@ctx.` prefix | `"@ctx.request.method"` |

### Example

```json
{
  "code": "200",
  "userId": "$.data.id",
  "items": {
    "json.path": "$.data",
    "id": "$.ID",
    "name": "$.NAME"
  }
}
```

## Hook System

Write custom logic in JavaScript with access to built-in functions:

### Built-in Modules

| Module | Functions |
|--------|-----------|
| **crypto** | md5, sha256, hmacSHA256, aesEncrypt/Decrypt, rsaEncrypt/Decrypt, rsaSign/Verify |
| **http** | get, post, postJSON, postForm, request |
| **encoding** | base64Encode/Decode, jsonEncode/Decode, urlEncode/Decode |
| **util** | uuid, now, formatTime, parseTime, sleep |

### Hook Example

```javascript
// BeforeForward hook
function beforeForward(ctx) {
  // Add custom headers
  ctx.request.headers["X-Custom-Token"] = crypto.md5(ctx.request.body);

  // Modify request
  var body = JSON.parse(ctx.request.body);
  body.timestamp = util.now();
  ctx.request.body = JSON.stringify(body);

  return ctx;
}
```

## Screenshots

### Vendor Management
<img src="./doc/产商管理.png" width="800" alt="Vendor List">
<img src="./doc/新增编辑厂商.png" width="800" alt="Vendor Edit">

### Organization Management
<img src="./doc/机构管理.png" width="800" alt="Organization List">
<img src="./doc/新增编辑机构.png" width="800" alt="Organization Edit">

### Service Management
<img src="./doc/接口管理.png" width="800" alt="Service List">
<img src="./doc/接口管理编辑.png" width="800" alt="Service Edit">

### Hook Management
<img src="./doc/Hook管理.png" width="800" alt="Hook Scripts">
<img src="./doc/接口Hook管理.png" width="800" alt="Service Hooks">

### Editors
<img src="./doc/脚本编辑器.png" width="800" alt="Script Editor">
<img src="./doc/json编辑器.png" width="800" alt="JSON Editor">

## Admin API

All admin endpoints require `X-Admin-Token` header authentication.

**Base Path:** `/admin/db`

| Resource | Endpoints |
|----------|-----------|
| Vendors | `GET/POST /vendors`, `GET/PUT/DELETE /vendor/{id}` |
| Organizations | `GET/POST /organizations`, `GET/PUT/DELETE /organization/{id}` |
| Services | `GET/POST /services`, `GET/PUT/DELETE /service/{id}` |
| Hook Scripts | `GET/POST /hook-scripts`, `GET/PUT/DELETE /hook-script/{id}` |
| Common Scripts | `GET/POST /scripts`, `GET/PUT/DELETE /script/{id}` |
| Service Hooks | `GET/POST /service-hooks`, `GET/PUT/DELETE /service-hook/{id}` |

## Project Structure

```
backend/
├── main.go           # Application entry point
├── handler/          # HTTP handlers
├── hook/             # Hook system + built-in functions
├── model/            # Data models
├── transform/        # DSL transformation engine
├── proxy/            # HTTP proxy
├── database/         # Database operations
└── router/           # Route registration

front/
├── src/
│   ├── views/        # Vue pages
│   ├── components/   # Vue components
│   └── router/       # Frontend routing
└── package.json
```

## Tech Stack

- **Web Framework:** [atreugo](https://github.com/savsgio/atreugo) (fasthttp-based)
- **ORM:** [GORM](https://gorm.io/)
- **JavaScript Engine:** [goja](https://github.com/dop251/goja)
- **Logging:** [zap](https://github.com/uber-go/zap)
- **Frontend:** Vue 3 + Element Plus

## Documentation

- [Admin API Reference](./ADMIN_API.md)
- [DSL Context Reference](./DSL_CONTEXT_REFERENCE.md)
- [Usage Examples](./EXAMPLE.md)
- [Concurrency Safety](./CONCURRENCY_SAFETY.md)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](./docs)
- 🐛 [Issue Tracker](https://github.com/yourusername/gateway/issues)
- 💬 [Discussions](https://github.com/yourusername/gateway/discussions)

---

<div align="center">
Made with ❤️ by the Gateway Team
</div>
