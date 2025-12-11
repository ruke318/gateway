<div align="center">

# Gateway

**Lightweight API Gateway - Unified External Interface Integration Platform**

[简体中文](./readme.md) | [English](./readme_en.md)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

</div>

---

## What is this?

Gateway is a **unified external interface integration platform** that helps you transform multiple external vendor APIs with different formats into a unified internal business interface.

**Core Value:** No matter how external vendor interfaces change, internal business systems only need to call the unified gateway interface. The gateway handles protocol conversion, data mapping, and interface adaptation.

## Use Cases

```
┌─────────────────┐     ┌─────────────┐     ┌─────────────────┐
│ Internal System │────▶│   Gateway   │────▶│  External APIs  │
│ (Unified Format)│◀────│ (Transform) │◀────│(Various Formats)│
└─────────────────┘     └─────────────┘     └─────────────────┘
```

**Typical Scenarios:**
- Integrate multiple payment channels with a unified payment interface
- Integrate multiple logistics providers with a unified tracking interface
- Integrate multiple SMS providers with a unified messaging interface
- Integrate multiple bank APIs with a unified financial service interface

## Core Features

- 🔄 **Protocol Conversion** - Auto-convert request/response formats, supports JSON, Form, XML
- 📝 **DSL Mapping** - Declarative field mapping, no coding required for data transformation
- 🪝 **Hook Extension** - JavaScript scripts for signatures, encryption, token fetching, etc.
- 🏢 **Multi-Tenant** - Vendor/Organization/Service three-tier architecture
- 🎨 **Visual Management** - Web UI configuration, takes effect without restart

## Quick Start

### Prerequisites

- Go 1.21+
- MySQL 5.7+

### Installation

```bash
git clone https://github.com/yourusername/gateway.git
cd gateway/backend
go build -o gateway .
./gateway
```

### Usage Example

Unified internal system call format:

```bash
POST /gateway/v1/invoke
Content-Type: application/json

{
  "com_id": "alipay",           # Vendor code
  "unit_id": "org001",          # Organization code
  "service_id": "pay",          # Service identifier
  "biz_no": "BIZ20231201001",   # Business transaction ID
  "req": {                      # Business parameters (unified format)
    "amount": 100,
    "order_no": "ORDER001"
  }
}
```

The gateway automatically converts the request to the vendor's required format and transforms the response back to a unified format.

## How It Works

### Data Model

| Entity | Description | Example |
|--------|-------------|---------|
| **Vendor** | External API provider | Alipay, WeChat Pay, UnionPay |
| **Organization** | Internal consumer, stores integration config | HQ, Branch A, Branch B |
| **Service** | Specific API config and transformation rules | Pay, Refund, Query |

### Request Flow

```
Request → Auth → Request Transform (DSL) → Hook → Forward to Vendor → Response Transform (DSL) → Return
```

### DSL Transform Example

Convert unified internal format to vendor format:

```json
{
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "timestamp": "@ctx.timestamp",
  "sign": "@ctx.sign"
}
```

### Hook Script Example

Handle complex logic like signatures:

```javascript
// BeforeForward - Add signature
var body = JSON.parse(ctx.request.body);
body.sign = crypto.md5(body.data + ctx.data.org_config.secret);
ctx.request.body = JSON.stringify(body);
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

## Project Structure

```
backend/                 # Backend service
├── handler/            # HTTP handlers
├── hook/               # Hook system + built-in functions
├── model/              # Data models
├── transform/          # DSL transformation engine
├── proxy/              # HTTP proxy
└── router/             # Route registration

front/                   # Frontend management UI
└── src/
    ├── views/          # Pages
    └── components/     # Components
```

## Tech Stack

- **Backend:** Go + atreugo + GORM + goja
- **Frontend:** Vue 3 + Element Plus
- **Database:** MySQL

## Documentation

- [Admin API Reference](./ADMIN_API.md)
- [DSL Context Reference](./DSL_CONTEXT_REFERENCE.md)
- [Usage Examples](./EXAMPLE.md)
- [Concurrency Safety](./CONCURRENCY_SAFETY.md)

## License

[Apache License 2.0](LICENSE)
