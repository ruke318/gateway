<div align="center">

# Gateway

**轻量级 API 网关 - 外部接口统一对接平台**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

</div>

---

## 📖 目录

- [这是什么？](#这是什么)
- [使用场景](#使用场景)
- [核心功能](#核心功能)
- [快速开始](#快速开始)
- [工作原理](#工作原理)
- [字典配置](#字典配置)
- [DSL 函数调用](#dsl-函数调用)
- [Hook 系统](#hook-系统)
- [回调功能](#回调功能)
- [管理界面](#管理界面)
- [API 参考](#api-参考)
- [技术架构](#技术架构)
- [常见问题](#常见问题)

---

## 这是什么？

Gateway 是一个**外部接口统一对接平台**，帮助你将多个外部厂商的不同接口格式，转换为内部统一的业务接口。

**核心价值：** 无论外部厂商接口格式如何变化，内部业务系统只需调用统一的网关接口，由网关负责协议转换、数据映射和接口适配。

## 使用场景

```
┌─────────────────┐     ┌─────────────┐     ┌─────────────────┐
│   内部业务系统    │────▶│   Gateway   │────▶│   外部厂商接口    │
│  (统一调用格式)   │◀────│   (转换层)   │◀────│  (各种不同格式)   │
└─────────────────┘     └─────────────┘     └─────────────────┘
```

**典型场景：**
- 对接多家支付渠道，内部统一支付接口
- 对接多家物流公司，内部统一物流查询接口
- 对接多家短信服务商，内部统一短信发送接口
- 对接多家银行接口，内部统一金融服务接口

---

## 核心功能

### 🔄 协议转换
- 自动转换请求/响应格式
- 支持 JSON、Form、XML 三种格式
- 支持 RESTful 路径参数

### 📝 DSL 映射
- 声明式字段映射，无需编码
- 支持 JSONPath 查询：`$.req.order_no`
- 支持 Context 注入：`@ctx.org_config.appId`
- **支持函数调用**：`@fn.dict.get("payment_method", "ALIPAY")`
- 支持嵌套对象和数组转换

### 🗂️ 字典配置（新增）
- **机构级字段映射**：不同机构可配置不同的字段值
- **机构内转换**：标准键 → 机构特定值
- **跨机构转换**：机构A的值 ↔ 机构B的值
- **在 Hook、DSL、驱动中统一使用**
- 可视化配置界面，支持批量导入
- 热加载：修改后立即生效

### 🪝 Hook 扩展
- JavaScript 脚本处理签名、加密、Token 获取
- 9 个执行点：认证前后、转换前后、转发前后、错误处理
- **内置模块：**
  - `crypto` - MD5、SHA、HMAC、AES、DES、RSA 加密/解密/签名
  - `http` - GET、POST、自定义 HTTP 请求
  - `encoding` - Base64、Hex、JSON、XML、URL 编码/解码
  - `util` - 时间戳、UUID 生成
  - `dict` - 字典查询和转换（新增）
  - `console` - 日志输出
- 全局函数库：可复用的 JavaScript 函数
- VM 池化：高性能并发执行

### 🏢 多租户
- 厂商/机构/接口三层架构
- 机构级配置隔离
- 支持同一厂商多机构对接
- 灵活的配置继承和覆盖

### 🎨 可视化管理
- Web 界面配置，无需重启即可生效
- 专业的 JavaScript 编辑器（语法高亮、代码提示）
- 可视化 JSON 编辑器（结构化编辑 DSL 映射）
- 字典配置管理界面（新增）
- 操作日志追踪

### ⚡ 高性能
- 基于 Atreugo (FastHTTP)，5倍于标准库性能
- JavaScript VM 池化，100倍性能提升
- 数据库连接池优化
- 字典内存缓存
- 并发安全设计

### 🔐 安全可靠
- 支持自定义认证 Hook
- 管理后台 Token 认证
- 用户系统 JWT 认证
- 敏感信息加密存储
- 请求日志完整追踪

---

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 5.7+
- Node.js 16+ (前端)

### 安装运行

**后端服务：**

```bash
# 克隆项目
git clone https://github.com/ruke318/gateway.git
cd gateway/backend

# 配置数据库
# 编辑 config/config.yaml

# 编译运行
go build -o gateway .
./gateway
```

**前端管理界面：**

```bash
cd ../front
npm install
npm run dev    # 开发模式 (http://localhost:3000)
npm run build  # 生产构建
```

**配置说明：**

创建 `backend/config/config.yaml`：

```yaml
port: ":8081"
adminToken: "your-admin-token-here"

db:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your-password"
  database: "gateway"
  maxIdleConns: 10
  maxOpenConns: 100

log:
  path: "./logs"
  level: "info"

vmPool:
  size: 100  # JavaScript VM 池大小

httpPool:
  maxIdleConns: 100
  maxConnsPerHost: 50

circuitBreaker:
  maxRequests: 100
  interval: 60
  timeout: 5
```

### 调用示例

内部系统统一调用格式：

```bash
POST /gateway/v1/invoke
Content-Type: application/json

{
  "com_id": "alipay",           # 厂商编码
  "unit_id": "org001",          # 机构编码
  "service_id": "pay",          # 接口标识
  "biz_no": "BIZ20231201001",   # 业务流水号
  "req": {                      # 业务参数（统一格式）
    "amount": 100,
    "order_no": "ORDER001",
    "pay_method": "ALIPAY"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "trade_no": "2023112922001400001234567890",
    "out_trade_no": "ORDER001"
  },
  "log_id": "LOG_BIZ20231201001_20231129143025_1234"
}
```

---

## 工作原理

### 数据模型

| 实体 | 说明 | 示例 |
|------|------|------|
| **厂商 (Vendor)** | 外部接口提供方 | 支付宝、微信、银联 |
| **机构 (Organization)** | 内部使用方，存储对接配置 | 总部、分公司A、分公司B |
| **接口 (Service)** | 具体的接口配置和转换规则 | 支付、退款、查询 |
| **字典 (Dictionary)** | 机构级字段映射关系（新增） | payment_method: ALIPAY → 01 |
| **Hook 脚本** | 可复用的业务逻辑脚本 | 签名、加密、Token 获取 |
| **函数库** | 全局共享的 JavaScript 函数 | 公共工具函数 |

### 请求流程

```
请求 → 解析参数 → 加载配置
     ↓
认证Hook (BeforeAuth → AfterAuth)
     ↓
请求转换 (BeforeRequestTransform → DSL → AfterRequestTransform)
     ↓
转发前Hook (BeforeForward) → HTTP转发 → 转发后Hook (AfterForward)
     ↓
响应转换 (BeforeResponseTransform → DSL → AfterResponseTransform)
     ↓
返回响应
```

**详细流程：**

1. **解析请求** - 提取 `com_id`、`unit_id`、`service_id`、`biz_no`、`req`
2. **加载配置** - 三层联合查询（Vendor + Organization + Service + Dictionary）
3. **构建上下文** - 构建 Hook Context（包含请求、路由、机构配置）
4. **执行认证 Hook** - `BeforeAuth` → `AfterAuth`
5. **请求转换** - `BeforeRequestTransform` → DSL 映射 → `AfterRequestTransform`
6. **转发前处理** - `BeforeForward` Hook（添加签名、Token）
7. **HTTP 转发** - 发送到厂商接口
8. **转发后处理** - `AfterForward` Hook（解密、数据清洗）
9. **响应转换** - `BeforeResponseTransform` → DSL 映射 → `AfterResponseTransform`
10. **返回响应** - 统一格式返回

### DSL 转换示例

**请求转换**（内部格式 → 厂商格式）：

```json
{
  "method": "alipay.trade.pay",
  "app_id": "@ctx.org_config.appId",
  "timestamp": "@fn.util.now()",
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "subject": "$.req.title",
  "pay_method": "@fn.dict.get('payment_method', $.req.pay_method)"
}
```

**响应转换**（厂商格式 → 内部格式）：

```json
{
  "success": "$.alipay_trade_pay_response.code",
  "trade_no": "$.alipay_trade_pay_response.trade_no",
  "out_trade_no": "$.alipay_trade_pay_response.out_trade_no",
  "status_name": "@fn.dict.reverseGet('order_status', $.alipay_trade_pay_response.status)"
}
```

**DSL 语法说明：**

| 语法 | 说明 | 示例 |
|------|------|------|
| `"字面量"` | 直接使用字符串、数字、布尔值 | `"success"`, `200`, `true` |
| `$.path` | JSONPath 查询源数据 | `$.req.order_no` |
| `@ctx.path` | Context 注入 | `@ctx.org_config.appId` |
| `@fn.func()` | 函数调用（新增） | `@fn.dict.get("type", "key")` |

---

## 字典配置

### 什么是字典配置？

字典配置用于管理**机构级别的字段映射关系**，解决不同机构使用不同字段值的问题。

**场景示例：**

```
支付方式字典 (payment_method)：
- 机构 A: ALIPAY → "01", WECHAT → "02"
- 机构 B: ALIPAY → "A001", WECHAT → "W001"

订单状态字典 (order_status)：
- 机构 A: SUCCESS → "10", FAILED → "20"
- 机构 B: SUCCESS → "S", FAILED → "F"
```

### 数据结构

| 字段 | 说明 | 示例 |
|------|------|------|
| `org_id` | 机构ID | org001 |
| `dict_type` | 字典类型 | payment_method, order_status |
| `dict_key` | 标准键（用于跨机构映射） | ALIPAY, SUCCESS |
| `dict_value` | 机构特定值 | 01, A001 |
| `description` | 说明 | 支付宝支付方式 |

### 配置方式

#### 1. 管理界面配置

访问 `http://localhost:3000/dictionary-config`：

- **新增配置**：逐条添加字典项
- **批量导入**：粘贴 JSON 数据快速导入
- **筛选查询**：按机构、字典类型筛选
- **重新加载**：修改后点击"重新加载字典"立即生效

#### 2. 批量导入格式

```json
[
  {
    "org_id": "org001",
    "dict_type": "payment_method",
    "dict_key": "ALIPAY",
    "dict_value": "01",
    "description": "支付宝"
  },
  {
    "org_id": "org001",
    "dict_type": "payment_method",
    "dict_key": "WECHAT",
    "dict_value": "02",
    "description": "微信支付"
  },
  {
    "org_id": "org001",
    "dict_type": "order_status",
    "dict_key": "SUCCESS",
    "dict_value": "10",
    "description": "成功"
  }
]
```

### 使用方式

#### 1. 在 Hook 脚本中使用

```javascript
// ✅ 机构内转换（自动使用当前机构 ID）
var code = dict.get("payment_method", "ALIPAY");
// 当前机构 org001: "ALIPAY" → "01"
// 当前机构 org002: "ALIPAY" → "A001"

// ✅ 反向查找（通过 value 查 key）
var key = dict.reverseGet("payment_method", "01");
// 返回: "ALIPAY"

// ✅ 跨机构转换（从当前机构转换到目标机构）
var targetStatus = dict.translate("org002", "order_status", "10");
// org001 的 "10" → 标准键 "SUCCESS" → org002 的 "S"

// ✅ 完整跨机构转换（手动指定源机构和目标机构）
var value = dict.translateFull("org001", "org002", "order_status", "10");

// ✅ 批量转换
var codes = dict.batchGet("payment_method", ["ALIPAY", "WECHAT"]);
// 返回: {"ALIPAY": "01", "WECHAT": "02"}

// ✅ 获取所有映射
var allMethods = dict.getAll("payment_method");
// 返回: {"ALIPAY": "01", "WECHAT": "02", ...}
```

**完整示例：**

```javascript
// BeforeForward Hook - 使用字典转换支付方式
var body = JSON.parse(context.requestBody);

// 将内部标准的支付方式转换为厂商编码
var payMethodCode = dict.get("payment_method", body.req.pay_method);
body.req.pay_method_code = payMethodCode;

context.requestBody = JSON.stringify(body);
console.log("支付方式转换: " + body.req.pay_method + " -> " + payMethodCode);
```

#### 2. 在 DSL 中使用

```json
{
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "pay_method_code": "@fn.dict.get('payment_method', $.req.pay_method)",
  "status_name": "@fn.dict.reverseGet('order_status', $.resp.status)",
  "cross_org_status": "@fn.dict.translate('org002', 'order_status', $.resp.status)"
}
```

#### 3. 在 Go 驱动代码中使用

```go
import "github.com/ruke318/gateway/hook"

// 机构内转换
dict := hook.GetDictionary()
code := dict.GetDictValue("org001", "payment_method", "ALIPAY")
// "ALIPAY" → "01"

// 反向查找
key := dict.ReverseGetDictKey("org001", "payment_method", "01")
// "01" → "ALIPAY"

// 跨机构转换
value := dict.CrossOrgTranslate("org001", "org002", "order_status", "10")
// org001 的 "10" → org002 的 "S"
```

### 应用场景

1. **支付方式映射** - 不同机构使用不同的支付方式编码
2. **订单状态映射** - 统一内部状态与厂商状态的对应关系
3. **银行编码映射** - 不同机构的银行编码标准不同
4. **业务类型映射** - 机构特定的业务类型编码
5. **跨机构数据同步** - 数据在不同机构间流转时自动转换

---

## DSL 函数调用

### 什么是 DSL 函数调用？

DSL 函数调用允许你在 DSL 映射中**直接调用 JavaScript 函数**，无需编写 Hook 脚本。

### 语法格式

```
@fn.函数路径(参数1, 参数2, ...)
```

**函数路径类型：**
- **字典函数**：`dict.get`, `dict.translate`, `dict.reverseGet` 等
- **全局函数库**：`lib.命名空间.函数名`
- **内置函数**：`crypto.md5`, `util.now`, `encoding.base64Encode` 等

**参数类型：**
- **JSONPath 变量**：`$.req.field` - 从源数据提取
- **Context 变量**：`@ctx.field` - 从上下文提取
- **字面量**：`"string"`, `123`, `true`, `null`

### 使用示例

#### 1. 字典转换

```json
{
  "pay_method_code": "@fn.dict.get('payment_method', $.req.pay_method)",
  "status": "@fn.dict.translate('org002', 'order_status', $.resp.status)",
  "status_key": "@fn.dict.reverseGet('order_status', $.resp.status_code)"
}
```

**执行过程：**
1. 提取 `$.req.pay_method` 的值（如 "ALIPAY"）
2. 调用 `dict.get("payment_method", "ALIPAY")`
3. 从当前机构的字典中查找对应值（如 "01"）
4. 返回结果填充到 `pay_method_code` 字段

#### 2. 加密签名

```json
{
  "data": "$.req.data",
  "sign": "@fn.crypto.md5($.req.data)",
  "token": "@fn.encoding.base64Encode($.req.app_id)"
}
```

#### 3. 全局函数库

先在"公共函数库"中定义函数：

**函数库配置：**
- 名称：`buildSignString`
- 命名空间：`payment`
- 代码：

```javascript
function buildSignString(params, secret) {
    var keys = Object.keys(params).sort();
    var signStr = "";
    for (var i = 0; i < keys.length; i++) {
        signStr += keys[i] + "=" + params[keys[i]] + "&";
    }
    signStr += "secret=" + secret;
    return signStr;
}
```

**在 DSL 中调用：**

```json
{
  "sign_string": "@fn.lib.payment.buildSignString($.req, @ctx.org_config.secret)",
  "sign": "@fn.crypto.md5(@fn.lib.payment.buildSignString($.req, @ctx.org_config.secret))"
}
```

#### 4. 组合使用

```json
{
  "out_trade_no": "$.req.order_no",
  "timestamp": "@fn.util.now()",
  "pay_method": "@fn.dict.get('payment_method', $.req.pay_method)",
  "amount_encoded": "@fn.encoding.base64Encode($.req.amount)",
  "custom_sign": "@fn.crypto.md5($.req.data)"
}
```

### 支持的函数

#### 字典函数（dict 模块）

```javascript
dict.get(dictType, key)              // 机构内转换
dict.reverseGet(dictType, value)     // 反向查找
dict.translate(toOrg, dictType, value)  // 跨机构转换（从当前机构）
dict.translateFull(fromOrg, toOrg, dictType, value)  // 完整跨机构转换
dict.batchGet(dictType, keys)        // 批量转换
dict.getAll(dictType)                // 获取所有映射
```

#### 加密函数（crypto 模块）

```javascript
crypto.md5(data)
crypto.sha1(data)
crypto.sha256(data)
crypto.hmacSHA256(data, key)
crypto.aesEncrypt(plaintext, key)
crypto.rsaSign(data, privateKey)
```

#### 编码函数（encoding 模块）

```javascript
encoding.base64Encode(data)
encoding.base64Decode(data)
encoding.hexEncode(data)
encoding.urlEncode(data)
encoding.jsonEncode(obj)
```

#### 工具函数（util 模块）

```javascript
util.now()                          // 当前时间戳
util.uuid()                         // 生成 UUID
util.formatTime(ts, "YYYY-MM-DD")  // 格式化时间
```

#### 全局函数库（lib 模块）

```javascript
lib.命名空间.函数名(参数...)
```

### 函数调用解析

DSL 引擎会自动：

1. **解析函数表达式** - 提取函数路径和参数
2. **解析参数值** - 处理 `$.`、`@ctx.` 变量和字面量
3. **执行函数** - 使用 Goja VM 执行 JavaScript 函数
4. **返回结果** - 将结果填充到 DSL 映射中

### 性能优化

- VM 池化技术，函数调用性能损耗极小
- 字典数据全部加载到内存
- 支持并发执行，不阻塞主流程

---

## Hook 系统

### Hook 执行点

系统定义了 **9 个 Hook 执行点**：

| 执行点 | 说明 | 使用场景 |
|--------|------|----------|
| `BeforeAuth` | 认证前 | 参数解密、预处理 |
| `AfterAuth` | 认证后 | Token 验证 |
| `BeforeRequestTransform` | 请求转换前 | 数据预处理 |
| `AfterRequestTransform` | 请求转换后 | 数据后处理 |
| `BeforeForward` | 转发前 ⭐ | **添加签名、获取 Token、修改请求头** |
| `AfterForward` | 转发后 ⭐ | **解密响应、数据清洗、状态转换** |
| `BeforeResponseTransform` | 响应转换前 | 响应预处理 |
| `AfterResponseTransform` | 响应转换后 | 响应后处理 |
| `OnError` | 错误处理 | 异常捕获、告警 |

### Hook Context

Hook 脚本中可以访问 `context` 对象：

```javascript
context = {
    requestBody: "...",          // 请求体（JSON 字符串）
    responseBody: "...",         // 响应体（JSON 字符串）
    requestHeaders: {...},       // 请求头
    responseHeaders: {...},      // 响应头
    data: {
        request: {
            method: "POST",
            path: "/gateway/v1/invoke",
            body: {              // 完整请求数据
                com_id: "alipay",
                unit_id: "org001",
                service_id: "pay",
                biz_no: "BIZ123",
                req: {...}
            }
        },
        route: {
            service_id: "pay",
            backendUrl: "https://api.alipay.com",
            backendPath: "/gateway.do",
            backendMethod: "POST"
        },
        org_config: {            // 机构配置（敏感信息）
            appId: "2021001234567890",
            secret: "abc123...",
            privateKey: "-----BEGIN RSA PRIVATE KEY-----..."
        }
    }
}
```

### 内置模块

#### crypto 模块

```javascript
// 哈希
crypto.md5("data")                    // 返回: "8d777f385d3dfec8815d20f7496026dc"
crypto.sha1("data")
crypto.sha256("data")
crypto.sha512("data")

// HMAC
crypto.hmacMD5("data", "key")
crypto.hmacSHA1("data", "key")
crypto.hmacSHA256("data", "key")

// AES 加密/解密
crypto.aesEncrypt("plaintext", "1234567890123456")       // 自动生成 IV
crypto.aesDecrypt("ciphertext", "1234567890123456")      // 自动提取 IV
crypto.aesCBCEncrypt("plaintext", "key16bytes", "iv16bytes")  // 指定 IV
crypto.aesCBCDecrypt("ciphertext", "key16bytes", "iv16bytes")
crypto.aesECBEncrypt("plaintext", "key16bytes")          // ECB 模式（不推荐）
crypto.aesECBDecrypt("ciphertext", "key16bytes")

// DES 加密/解密
crypto.desEncrypt("plaintext", "key8byte", "iv8bytes")
crypto.desDecrypt("ciphertext", "key8byte", "iv8bytes")
crypto.des3Encrypt("plaintext", "key24bytes", "iv8bytes")
crypto.des3Decrypt("ciphertext", "key24bytes", "iv8bytes")

// RSA 加密/解密/签名/验签
crypto.rsaEncrypt("plaintext", publicKeyPEM)
crypto.rsaDecrypt("ciphertext", privateKeyPEM)
crypto.rsaSign("data", privateKeyPEM)                    // SHA256 签名
crypto.rsaVerify("data", "signature", publicKeyPEM)      // 返回 true/false

// 随机数
crypto.randomBytes(16)  // 返回 32 位 hex 字符串
```

#### http 模块

```javascript
// GET 请求
var resp = http.get("https://api.example.com/token", {
    "Authorization": "Bearer xxx"
});

// POST 请求（原始 body）
var resp = http.post("https://api.example.com/data", "raw body", {
    "Content-Type": "text/plain"
});

// POST JSON（自动序列化）
var resp = http.postJSON("https://api.example.com/api", {
    key: "value"
}, {
    "Authorization": "Bearer xxx"
});

// POST Form（自动编码）
var resp = http.postForm("https://api.example.com/form", {
    key1: "value1",
    key2: "value2"
});

// 自定义请求
var resp = http.request("PUT", "https://api.example.com/resource", {
    data: "value"
}, {
    "Authorization": "Bearer xxx"
});

// 响应格式
resp = {
    status: 200,              // HTTP 状态码
    headers: {...},           // 响应头
    body: "...",              // 响应体（字符串）
    json: {...}               // 自动解析的 JSON（如果是 JSON 格式）
}
```

#### encoding 模块

```javascript
// Base64
encoding.base64Encode("data")           // 返回: "ZGF0YQ=="
encoding.base64Decode("ZGF0YQ==")       // 返回: "data"

// Hex
encoding.hexEncode("data")              // 返回: "64617461"
encoding.hexDecode("64617461")          // 返回: "data"

// JSON
encoding.jsonEncode({key: "value"})     // 返回: '{"key":"value"}'
encoding.jsonDecode('{"key":"value"}')  // 返回: {key: "value"}

// XML
encoding.xmlEncode({key: "value"})
encoding.xmlDecode("<root><key>value</key></root>")

// URL
encoding.urlEncode("hello world")       // 返回: "hello+world"
encoding.urlDecode("hello%20world")     // 返回: "hello world"
```

#### util 模块

```javascript
// UUID
util.uuid()  // 返回: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

// 时间戳
util.now()   // 返回: Unix 时间戳（秒）

// 格式化时间
util.formatTime(1701234567, "YYYY-MM-DD HH:mm:ss")
// 返回: "2023-11-29 12:34:27"

// 解析时间
util.parseTime("2023-11-29 12:34:27", "YYYY-MM-DD HH:mm:ss")
// 返回: 1701234567

// 休眠
util.sleep(1000)  // 休眠 1000 毫秒
```

#### dict 模块（新增）

```javascript
// 机构内转换（自动使用当前机构 ID）
dict.get("payment_method", "ALIPAY")
// 返回当前机构的编码，如 "01"

// 反向查找（通过 value 查 key）
dict.reverseGet("payment_method", "01")
// 返回标准键，如 "ALIPAY"

// 跨机构转换（从当前机构转换到目标机构）
dict.translate("org002", "order_status", "10")
// org001 的 "10" → "SUCCESS" → org002 的 "S"

// 完整跨机构转换（手动指定源和目标）
dict.translateFull("org001", "org002", "order_status", "10")

// 批量转换
dict.batchGet("payment_method", ["ALIPAY", "WECHAT"])
// 返回: {"ALIPAY": "01", "WECHAT": "02"}

// 获取所有映射
dict.getAll("payment_method")
// 返回: {"ALIPAY": "01", "WECHAT": "02", "BANK": "03", ...}
```

#### console 模块

```javascript
console.log("Info message")
console.log("值:", value, "对象:", obj)
```

### Hook 脚本示例

#### 示例 1: 添加签名

```javascript
// BeforeForward Hook
var body = JSON.parse(context.requestBody);
var secret = context.data.org_config.secret;

// 构建签名字符串
var signStr = body.order_no + body.amount + secret;
body.sign = crypto.md5(signStr);

// 更新请求体
context.requestBody = JSON.stringify(body);

console.log("签名完成: " + body.sign);
```

#### 示例 2: 获取 Access Token

```javascript
// BeforeForward Hook
var appId = context.data.org_config.appId;
var secret = context.data.org_config.secret;

// 调用厂商 Token 接口
var resp = http.postJSON("https://api.vendor.com/oauth/token", {
    app_id: appId,
    secret: secret,
    grant_type: "client_credentials"
});

var tokenData = resp.json;
var accessToken = tokenData.access_token;

// 添加到请求头
context.requestHeaders["Authorization"] = "Bearer " + accessToken;

console.log("Token 获取成功，有效期: " + tokenData.expires_in + "秒");
```

#### 示例 3: 响应解密

```javascript
// AfterForward Hook
var respData = JSON.parse(context.responseBody);
var encryptedData = respData.data;
var key = context.data.org_config.aesKey;
var iv = context.data.org_config.aesIv;

// AES 解密
var decrypted = crypto.aesCBCDecrypt(encryptedData, key, iv);
respData.data = JSON.parse(decrypted);

// 更新响应体
context.responseBody = JSON.stringify(respData);

console.log("响应解密完成");
```

#### 示例 4: 使用字典和全局函数

```javascript
// BeforeForward Hook
var body = JSON.parse(context.requestBody);

// 使用字典转换支付方式
var payMethodCode = dict.get("payment_method", body.req.pay_method);
body.req.pay_method_code = payMethodCode;

// 使用全局函数构建签名
var signStr = lib.payment.buildSignString(body.req, context.data.org_config.secret);
body.sign = crypto.md5(signStr);

context.requestBody = JSON.stringify(body);
console.log("支付方式: " + body.req.pay_method + " -> " + payMethodCode);
console.log("签名: " + body.sign);
```

#### 示例 5: 动态修改路由

```javascript
// BeforeForward Hook - 根据金额切换通道
var body = JSON.parse(context.requestBody);

if (body.req.amount > 10000) {
    // 大额走专用通道
    context.data.route.backendPath = "/api/v2/pay/large";
    console.log("切换到大额支付通道");
} else {
    // 小额走普通通道
    context.data.route.backendPath = "/api/v2/pay/normal";
    console.log("使用普通支付通道");
}
```

### 全局函数库

在"公共函数库"管理界面定义可复用的函数：

**定义函数库：**

```javascript
// 函数库名称: commonUtils
// 命名空间: global

function buildSignString(params, secret) {
    var keys = Object.keys(params).sort();
    var signStr = "";
    for (var i = 0; i < keys.length; i++) {
        signStr += keys[i] + "=" + params[keys[i]] + "&";
    }
    signStr += "secret=" + secret;
    return signStr;
}

function validateTimestamp(ts) {
    var now = util.now();
    var diff = Math.abs(now - ts);
    return diff < 300; // 5分钟内有效
}
```

**在 Hook 中使用：**

```javascript
var body = JSON.parse(context.requestBody);
var secret = context.data.org_config.secret;

// 直接调用全局函数
var signStr = lib.global.buildSignString(body, secret);
body.sign = crypto.md5(signStr);

// 验证时间戳
if (!lib.global.validateTimestamp(body.timestamp)) {
    throw new Error("请求已过期");
}

context.requestBody = JSON.stringify(body);
```

**在 DSL 中使用：**

```json
{
  "sign": "@fn.crypto.md5(@fn.lib.global.buildSignString($.req, @ctx.org_config.secret))"
}
```

---

## 回调功能

Gateway 支持接收厂商回调，并将回调数据转发到内部业务系统。回调功能**完全复用 invoke 流程**。

### 回调 URL 格式

```
POST /gateway/v1/notify/{unit_id}/{service_id}/{channel}
GET  /gateway/v1/notify/{unit_id}/{service_id}/{channel}
```

**参数说明：**
- `unit_id`: 机构ID（如 `org001`）
- `service_id`: 服务标识（如 `payNotify`, `refundNotify`）
- `channel`: 渠道标识
  - **数字**（如 `1`, `2`）：走默认处理器，不验签，直接转换
  - **字符串**（如 `alipay`, `wechat`）：使用专属处理器

**URL 示例：**

```bash
# 机构 org001 的支付宝支付回调
POST https://gateway.example.com/gateway/v1/notify/org001/payNotify/alipay

# 机构 org002 的微信支付回调
POST https://gateway.example.com/gateway/v1/notify/org002/payNotify/wechat

# 机构 org001 的渠道 1 退款回调
GET https://gateway.example.com/gateway/v1/notify/org001/refundNotify/1
```

### 配置步骤

#### 1. 配置回调接口

在接口管理中创建回调接口：

- **接口标识**：`payNotify`
- **厂商**：支付宝
- **机构**：总部
- **后端 URL**：`http://order-service:8080`
- **后端路径**：`/api/payment/notify`

**请求转换 DSL：**

```json
{
  "order_no": "$.req.out_trade_no",
  "trade_no": "$.req.trade_no",
  "amount": "$.req.total_amount",
  "status": "@fn.dict.get('order_status', $.req.trade_status)",
  "notify_time": "$.req.notify_time"
}
```

**响应转换 DSL：**

```json
{
  "code": "success"
}
```

#### 2. 配置验签 Hook（可选）

如需验证厂商签名：

```javascript
// BeforeAuth Hook - 支付宝验签
var body = context.data.request.body.req;

// 提取签名
var sign = body.sign;
var signType = body.sign_type;
delete body.sign;
delete body.sign_type;

// 构建待签名字符串（按 key 排序）
var keys = Object.keys(body).sort();
var signStr = "";
for (var i = 0; i < keys.length; i++) {
    if (body[keys[i]] !== "") {
        signStr += keys[i] + "=" + body[keys[i]] + "&";
    }
}
signStr = signStr.slice(0, -1);  // 去掉最后的 &

// RSA 验签
var publicKey = context.data.org_config.alipayPublicKey;
var isValid = crypto.rsaVerify(signStr, sign, publicKey);

if (!isValid) {
    throw new Error("签名验证失败");
}

console.log("签名验证成功");
```

#### 3. 在厂商后台配置回调 URL

在支付宝、微信等厂商后台配置回调地址：

```
https://your-gateway.com/gateway/v1/notify/org001/payNotify/alipay
```

### 回调处理流程

```
厂商回调 → NotifyProcessor → InvokeRequest → Hook → DSL → 转发内部系统 → 返回响应给厂商
```

**详细流程：**

1. 厂商发送回调请求到网关
2. NotifyProcessor 将回调数据转换为 `InvokeRequest`
3. 走完整的 invoke 流程（Hook、DSL、转发）
4. 将内部系统响应转换为厂商要求的格式返回

### 渠道处理器

#### 默认处理器

适用于数字渠道，直接转换数据，不做特殊处理。

#### 支付宝处理器

- 解析 Form 格式回调
- 提取 `out_trade_no` 作为业务流水号

#### 微信处理器

- 解析 XML 格式回调
- 提取 `out_trade_no` 作为业务流水号

#### 自定义处理器

可以注册自定义渠道处理器：

```go
import "github.com/ruke318/gateway/handler"

type CustomNotifyProcessor struct{}

func (p *CustomNotifyProcessor) Process(ctx *atreugo.RequestCtx, unitID, serviceID, channel string) (*model.InvokeRequest, error) {
    // 自定义处理逻辑
    return &model.InvokeRequest{
        ComID: channel,
        UnitID: unitID,
        ServiceID: serviceID,
        BizNo: "...",
        Req: reqData,
    }, nil
}

// 注册
handler.RegisterNotifyProcessor("custom", &CustomNotifyProcessor{})
```

---

## 管理界面

### 功能模块

#### 1. 厂商管理
- 新增/编辑/删除外部厂商
- 配置厂商基础信息（编码、名称、Base URL）

#### 2. 机构管理
- 管理内部使用方（总部、分公司等）
- 配置机构专属参数（appId、secret、证书等）
- JSON 编辑器支持结构化编辑配置

#### 3. 接口管理
- 可视化配置接口转换规则
- JSON 编辑器支持结构化编辑 DSL 映射
- 配置请求/响应转换规则
- 设置厂商后端路径和请求方法

#### 4. Hook 脚本管理
- JavaScript 编辑器提供语法高亮和代码提示
- 编写可复用的 Hook 脚本
- 配置执行点和优先级

#### 5. 公共函数库管理
- 编写全局共享的 JavaScript 函数
- 支持命名空间管理
- 在所有 Hook 和 DSL 中直接调用

#### 6. 字典配置管理（新增）
- 可视化配置机构级字段映射
- 支持批量导入 JSON 数据
- 按机构、字典类型筛选
- 重新加载字典立即生效

#### 7. 用户管理
- 用户增删改查
- 密码管理
- JWT 认证

#### 8. 操作日志
- 记录所有管理操作
- 可追溯历史变更
- 支持筛选和搜索

### 界面截图

#### 厂商管理
<img src="./doc/产商管理.png" width="800" alt="厂商列表">

#### 机构管理
<img src="./doc/机构管理.png" width="800" alt="机构列表">

#### 接口管理
<img src="./doc/接口管理.png" width="800" alt="接口列表">

#### Hook 管理
<img src="./doc/Hook管理.png" width="800" alt="Hook 脚本">

#### 脚本编辑器
<img src="./doc/脚本编辑器.png" width="800" alt="脚本编辑器">

#### JSON 编辑器
<img src="./doc/json编辑器.png" width="800" alt="JSON 编辑器">

---

## API 参考

### 统一调用接口

**请求：**

```http
POST /gateway/v1/invoke
Content-Type: application/json

{
  "com_id": "alipay",
  "unit_id": "org001",
  "service_id": "pay",
  "biz_no": "BIZ20231201001",
  "req": {
    "amount": 100,
    "order_no": "ORDER001",
    "pay_method": "ALIPAY"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "trade_no": "2023112922001400001234567890",
    "out_trade_no": "ORDER001"
  },
  "log_id": "LOG_BIZ20231201001_20231129143025_1234"
}
```

### 回调接口

**厂商回调请求：**

```http
POST /gateway/v1/notify/{unit_id}/{service_id}/{channel}
Content-Type: application/x-www-form-urlencoded

out_trade_no=ORDER001&trade_no=2023112922001400001234567890&sign=xxx&...
```

**返回给厂商的响应：**

```
success  (或厂商要求的其他格式)
```

### 管理后台 API

所有管理接口需要在请求头中添加：

```
X-Admin-Token: your-admin-token-here
```

#### 厂商管理

- `GET /admin/db/vendors` - 查询厂商列表
- `GET /admin/db/vendor/{id}` - 获取厂商详情
- `POST /admin/db/vendor` - 创建厂商
- `PUT /admin/db/vendor/{id}` - 更新厂商
- `DELETE /admin/db/vendor/{id}` - 删除厂商

#### 机构管理

- `GET /admin/db/organizations` - 查询机构列表
- `GET /admin/db/organization/{id}` - 获取机构详情
- `POST /admin/db/organization` - 创建机构
- `PUT /admin/db/organization/{id}` - 更新机构
- `DELETE /admin/db/organization/{id}` - 删除机构

#### 接口管理

- `GET /admin/db/services` - 查询接口列表
- `GET /admin/db/service/{id}` - 获取接口详情
- `POST /admin/db/service` - 创建接口
- `PUT /admin/db/service/{id}` - 更新接口
- `DELETE /admin/db/service/{id}` - 删除接口

#### Hook 脚本管理

- `GET /admin/db/hook-scripts` - 查询 Hook 脚本列表
- `GET /admin/db/hook-script/{id}` - 获取脚本详情
- `POST /admin/db/hook-script` - 创建脚本
- `PUT /admin/db/hook-script/{id}` - 更新脚本
- `DELETE /admin/db/hook-script/{id}` - 删除脚本

#### 公共函数库管理

- `GET /admin/db/scripts` - 查询函数库列表
- `GET /admin/db/script/{id}` - 获取函数库详情
- `POST /admin/db/script` - 创建函数库
- `PUT /admin/db/script/{id}` - 更新函数库
- `DELETE /admin/db/script/{id}` - 删除函数库
- `POST /admin/db/reload-library` - 重新加载函数库

#### 字典配置管理（新增）

- `GET /admin/db/dictionary-configs?org_id=xxx&dict_type=xxx` - 查询字典配置列表
- `GET /admin/db/dictionary-config/{id}` - 获取字典配置详情
- `POST /admin/db/dictionary-config` - 创建字典配置
- `PUT /admin/db/dictionary-config/{id}` - 更新字典配置
- `DELETE /admin/db/dictionary-config/{id}` - 删除字典配置
- `POST /admin/db/dictionary-configs/batch` - 批量导入字典配置
- `POST /admin/db/reload-dictionary` - 重新加载字典

#### 用户管理

- `GET /api/users` - 查询用户列表（需 JWT）
- `POST /api/users` - 创建用户
- `PUT /api/users/{id}` - 更新用户
- `DELETE /api/users/{id}` - 删除用户

#### 操作日志

- `GET /api/operation-logs` - 查询操作日志（需 JWT）
- `GET /api/operation-logs/statistics` - 操作统计

---

## 技术架构

### 技术栈

**后端：**
- **语言：** Go 1.21+
- **Web 框架：** Atreugo v11 (基于 FastHTTP 的高性能框架)
- **ORM：** GORM (MySQL 驱动)
- **JavaScript 引擎：** Goja (用于 Hook 脚本执行)
- **配置管理：** Viper
- **日志：** Zap (结构化日志)
- **JSON Path：** jsonpath (DSL 转换)

**前端：**
- **框架：** Vue 3
- **UI 库：** Element Plus 2.4+
- **路由：** Vue Router 4.2+
- **HTTP 客户端：** Axios
- **编辑器：**
  - Ace Editor (JavaScript 代码编辑器)
  - JSON Editor (可视化 JSON 编辑器)
- **构建工具：** Vite 5.0+

**数据库：**
- MySQL 5.7+ (支持自动迁移和连接池)

### 系统架构

```
+-------------------------------------------------------------+
|                      内部业务系统                             |
+---------------------------+---------------------------------+
                            | 统一格式请求
                            v
+-------------------------------------------------------------+
|                      Gateway 网关层                          |
|                                                             |
|  +----------+  +----------+  +----------+  +----------+     |
|  | Handler  |->|   Hook   |->|Transform |->|  Proxy   |     |
|  +----------+  +----------+  +----------+  +----------+     |
|                                                             |
|  +-------------------------------------------------------+  |
|  |                   Database (MySQL)                    |  |
|  |   Vendor | Organization | Service | Hook | Dictionary |  |
|  +-------------------------------------------------------+  |
+---------------------------+---------------------------------+
                            | 厂商格式请求
                            v
+-------------------------------------------------------------+
|            外部厂商 API (支付宝、微信、银联等)                   |
+-------------------------------------------------------------+
```

### 核心模块

#### 1. Handler 层
- HTTP 请求接收和解析
- 参数验证
- 统一响应格式封装

#### 2. Hook 系统
- JavaScript VM 池化管理（100 个预创建的 VM 实例）
- 内置函数库（crypto、http、encoding、util、dict）
- 全局函数库加载
- 字典配置加载（新增）

#### 3. Transform 层
- DSL 转换引擎
- JSONPath 查询
- Context 注入
- 函数调用支持（新增）

#### 4. Proxy 层
- HTTP 连接池管理
- 熔断器
- 请求转发

#### 5. Database 层
- 数据库连接池
- ORM 操作封装
- 事务管理

### 性能优化

| 技术 | 性能提升 | 说明 |
|------|---------|------|
| **FastHTTP 框架** | 5倍 | 相比标准库 `net/http` |
| **JavaScript VM 池化** | 100倍 | 避免重复创建销毁 VM |
| **数据库连接池** | 10-15% | 复用连接，减少开销 |
| **预加载关联数据** | 避免 N+1 查询 | GORM Preload |
| **结构化日志 (Zap)** | 4倍 | 相比标准库 `log` |
| **字典内存缓存** | 显著 | 全部加载到内存，避免查询 |

### 性能基准

| 场景 | QPS | P50 延迟 | P99 延迟 |
|------|-----|---------|---------|
| 简单 DSL 转换（无 Hook） | ~5000 | 18ms | 45ms |
| DSL + BeforeForward Hook | ~4000 | 23ms | 60ms |
| 完整链路（DSL + 3个Hook + 转发） | ~2000 | 45ms | 120ms |

### 并发安全

- **VM 隔离**：每个请求独立的 VM 实例
- **无状态设计**：Hook Context 每次重新创建
- **连接池限制**：防止连接耗尽
- **读写锁**：字典缓存使用 `sync.RWMutex`
- **并发测试**：1000 并发成功率 99.8%

---

## 项目结构

```
gateway/
├── backend/                  # 后端服务
│   ├── config/               # 配置文件
│   │   └── config.yaml       # 主配置文件
│   ├── handler/              # HTTP 处理器
│   │   ├── invoke.go         # 统一调用处理器
│   │   ├── admin_db.go       # 管理后台处理器
│   │   ├── notify_processor.go  # 回调处理器
│   │   └── user.go           # 用户管理处理器
│   ├── hook/                 # Hook 系统
│   │   ├── executor.go       # Hook 执行器 + VM 池
│   │   ├── builtin.go        # 内置函数库
│   │   ├── library.go        # 全局函数库管理
│   │   ├── dictionary.go     # 字典管理（新增）
│   │   └── types.go          # 类型定义
│   ├── model/                # 数据模型
│   │   ├── vendor.go         # 厂商模型
│   │   ├── organization.go   # 机构模型
│   │   ├── service.go        # 接口模型
│   │   ├── hook_script.go    # Hook 脚本模型
│   │   ├── script_library.go # 函数库模型
│   │   ├── dictionary_config.go  # 字典配置模型（新增）
│   │   ├── user.go           # 用户模型
│   │   └── operation_log.go  # 操作日志模型
│   ├── transform/            # DSL 转换引擎
│   │   ├── dsl.go            # DSL 转换器（支持函数调用）
│   │   └── dsl_test.go       # 单元测试
│   ├── proxy/                # HTTP 代理
│   │   └── forwarder.go      # 请求转发器（连接池 + 熔断器）
│   ├── router/               # 路由注册
│   │   ├── out.go            # 对外路由（invoke、notify）
│   │   ├── admin_db.go       # 管理后台路由
│   │   └── user.go           # 用户管理路由
│   ├── database/             # 数据库
│   │   ├── database.go       # 数据库初始化
│   │   ├── migrate.go        # 自动迁移
│   │   ├── query.go          # 查询函数
│   │   └── init.go           # 默认数据初始化
│   ├── logger/               # 日志
│   │   └── logger.go         # Zap 日志封装
│   ├── util/                 # 工具函数
│   ├── middleware/           # 中间件
│   └── main.go               # 程序入口
│
├── front/                    # 前端管理界面
│   ├── src/
│   │   ├── views/            # 页面
│   │   │   ├── Vendor.vue    # 厂商管理
│   │   │   ├── Organization.vue  # 机构管理
│   │   │   ├── Service.vue   # 接口管理
│   │   │   ├── HookScript.vue    # Hook 脚本管理
│   │   │   ├── Script.vue    # 公共函数库管理
│   │   │   ├── DictionaryConfig.vue  # 字典配置管理（新增）
│   │   │   ├── User.vue      # 用户管理
│   │   │   ├── OperationLog.vue  # 操作日志
│   │   │   └── Login.vue     # 登录页
│   │   ├── layouts/          # 布局
│   │   │   └── MainLayout.vue
│   │   ├── router/           # 路由
│   │   │   └── index.js
│   │   ├── api/              # API 接口
│   │   │   ├── index.js      # API 定义
│   │   │   └── request.js    # Axios 封装
│   │   └── App.vue
│   ├── vite.config.js        # Vite 配置
│   └── package.json
│
├── doc/                      # 文档和截图
└── README.md                 # 本文件
```

---

## 使用流程

### 1. 配置厂商

在厂商管理界面添加外部 API 提供商：

- **编码**：`alipay`
- **名称**：支付宝
- **Base URL**：`https://openapi.alipay.com`
- **描述**：支付宝开放平台

### 2. 配置机构

添加内部使用方，配置对接凭证：

- **编码**：`org001`
- **名称**：总部
- **配置**（JSON）：
  ```json
  {
    "appId": "2021001234567890",
    "secret": "your-secret-key",
    "privateKey": "-----BEGIN RSA PRIVATE KEY-----\n...",
    "alipayPublicKey": "-----BEGIN PUBLIC KEY-----\n..."
  }
  ```

### 3. 配置字典（可选）

如果需要字段映射，配置字典：

**方式 1：单条添加**
- 机构 ID：`org001`
- 字典类型：`payment_method`
- 标准键：`ALIPAY`
- 机构值：`01`

**方式 2：批量导入**

```json
[
  {"org_id": "org001", "dict_type": "payment_method", "dict_key": "ALIPAY", "dict_value": "01"},
  {"org_id": "org001", "dict_type": "payment_method", "dict_key": "WECHAT", "dict_value": "02"},
  {"org_id": "org001", "dict_type": "order_status", "dict_key": "SUCCESS", "dict_value": "10"}
]
```

### 4. 配置接口

创建接口，设置 DSL 映射规则：

**基本信息：**
- **接口标识**：`pay`
- **厂商**：支付宝
- **机构**：总部
- **后端 URL**：`https://openapi.alipay.com`
- **后端路径**：`/gateway.do`
- **请求方法**：POST
- **Body 类型**：json

**请求转换 DSL：**
```json
{
  "method": "alipay.trade.pay",
  "app_id": "@ctx.org_config.appId",
  "timestamp": "@fn.util.now()",
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "subject": "$.req.title",
  "pay_method": "@fn.dict.get('payment_method', $.req.pay_method)"
}
```

**响应转换 DSL：**
```json
{
  "success": "$.alipay_trade_pay_response.code",
  "trade_no": "$.alipay_trade_pay_response.trade_no",
  "out_trade_no": "$.alipay_trade_pay_response.out_trade_no",
  "status_name": "@fn.dict.reverseGet('order_status', $.alipay_trade_pay_response.trade_status)"
}
```

### 5. 编写 Hook（可选）

如需复杂逻辑（签名、Token），编写 Hook 脚本：

**创建 Hook 脚本：**
- **名称**：支付宝签名
- **执行点**：BeforeForward
- **代码**：

```javascript
var body = JSON.parse(context.requestBody);
var secret = context.data.org_config.secret;

// 构建签名字符串
var keys = Object.keys(body).filter(k => k !== 'sign').sort();
var signStr = "";
for (var i = 0; i < keys.length; i++) {
    signStr += keys[i] + "=" + body[keys[i]] + "&";
}
signStr = signStr.slice(0, -1);

// 计算签名
body.sign = crypto.md5(signStr + secret);

context.requestBody = JSON.stringify(body);
console.log("签名完成: " + body.sign);
```

### 6. 关联 Hook

在接口管理界面，将 Hook 脚本关联到接口：

- **接口**：支付
- **Hook 脚本**：支付宝签名
- **执行点**：BeforeForward
- **优先级**：1

### 7. 调用测试

使用统一格式调用网关接口：

```bash
curl -X POST http://localhost:8081/gateway/v1/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "com_id": "alipay",
    "unit_id": "org001",
    "service_id": "pay",
    "biz_no": "BIZ20231201001",
    "req": {
      "amount": 100,
      "order_no": "ORDER001",
      "title": "测试商品",
      "pay_method": "ALIPAY"
    }
  }'
```

### 8. 监控日志

查看请求日志，排查问题：

```
[INFO] [LOG_BIZ20231201001_20231129143025_1234] Invoke 请求开始
    unit_id=org001 service_id=pay com_id=alipay
[INFO] [LOG_BIZ20231201001_20231129143025_1234] LoadConfig 加载接口配置
[INFO] [LOG_BIZ20231201001_20231129143025_1234] RequestTransform DSL转换完成
[INFO] [LOG_BIZ20231201001_20231129143025_1234] BeforeForward Hook执行成功
    输出: 签名完成: abc123...
[INFO] [LOG_BIZ20231201001_20231129143025_1234] Forward 转发成功 status=200
[INFO] [LOG_BIZ20231201001_20231129143025_1234] ResponseTransform DSL转换完成
[INFO] [LOG_BIZ20231201001_20231129143025_1234] Response 请求完成 duration=125ms
```

---

## 常见问题

### Q: 如何处理厂商签名？

**A:** 在 `BeforeForward` Hook 中添加签名逻辑，使用内置的 `crypto` 模块。

```javascript
var body = JSON.parse(context.requestBody);
var signStr = buildSignString(body);
body.sign = crypto.md5(signStr + context.data.org_config.secret);
context.requestBody = JSON.stringify(body);
```

### Q: 如何获取厂商 Token？

**A:** 在 `BeforeForward` Hook 中使用 `http.postJSON()` 调用厂商 Token 接口。

```javascript
var resp = http.postJSON("https://api.vendor.com/token", {
    app_id: context.data.org_config.appId,
    secret: context.data.org_config.secret
});
var token = resp.json.access_token;
context.requestHeaders["Authorization"] = "Bearer " + token;
```

### Q: 如何处理响应解密？

**A:** 在 `AfterForward` Hook 中使用 `crypto.aesDecrypt()` 等函数解密。

```javascript
var respData = JSON.parse(context.responseBody);
var decrypted = crypto.aesDecrypt(respData.data, context.data.org_config.aesKey);
respData.data = JSON.parse(decrypted);
context.responseBody = JSON.stringify(respData);
```

### Q: 字典配置修改后如何生效？

**A:** 点击管理界面的"重新加载字典"按钮，或调用 `POST /admin/db/reload-dictionary` 接口。字典会立即重新加载到内存，并清空 VM 池，新请求会使用新配置。

### Q: DSL 函数调用和 Hook 的区别？

**A:**
- **DSL 函数调用**：适合简单的数据转换（如字典查询、编码转换、简单计算）
- **Hook 脚本**：适合复杂的业务逻辑（如签名算法、Token 获取、条件判断、循环处理）

**选择建议：**
- 能用 DSL 函数解决的，优先用 DSL 函数（更直观、性能更好）
- 需要复杂逻辑时，使用 Hook 脚本

### Q: 如何调试 Hook 脚本？

**A:** 使用 `console.log()` 输出日志，查看后端控制台或日志文件。

```javascript
console.log("Context data:", JSON.stringify(context.data));
console.log("Request body:", context.requestBody);
console.log("转换后的值:", payMethodCode);
```

### Q: 支持哪些厂商回调？

**A:** 支持所有 HTTP/HTTPS 回调，支持 GET、POST 方法，支持 JSON、Form、XML 格式。可以通过自定义 NotifyProcessor 支持任何厂商。

### Q: 如何实现跨机构数据同步？

**A:** 使用字典的跨机构转换功能：

```javascript
// Hook 中
var orgAStatus = "10";
var orgBStatus = dict.translateFull("org001", "org002", "order_status", orgAStatus);
// "10" → "SUCCESS" → "S"

// DSL 中
{
  "org_b_status": "@fn.dict.translateFull('org001', 'org002', 'order_status', $.org_a_status)"
}
```

### Q: Hook 执行失败怎么办？

**A:**
1. 查看日志中的错误信息
2. 检查 Hook 脚本语法
3. 验证 Context 数据是否正确
4. 使用 `console.log()` 调试

```javascript
try {
    // 你的代码
} catch (e) {
    console.log("错误:", e.message);
    throw e;
}
```

### Q: DSL 转换失败怎么办？

**A:**
1. 验证 JSONPath 语法是否正确
2. 检查源数据结构
3. 确认 Context 变量存在
4. 使用 `BeforeRequestTransform` Hook 打印源数据调试

### Q: 如何配置回调 URL？

**A:** 在厂商后台配置回调地址：

```
https://your-gateway.com/gateway/v1/notify/{unit_id}/{service_id}/{channel}
```

例如：
```
https://gateway.example.com/gateway/v1/notify/org001/payNotify/alipay
```

其中：
- `org001` 是机构 ID
- `payNotify` 是回调接口标识
- `alipay` 是渠道标识

### Q: 性能如何优化？

**A:**
1. **增大 VM 池大小**：`vmPool.size: 200`
2. **增大数据库连接池**：`maxOpenConns: 200`
3. **启用缓存**：可以在 Hook 中缓存 Token
4. **减少 Hook 数量**：能用 DSL 解决的不用 Hook
5. **优化 DSL**：减少嵌套层级

### Q: 如何水平扩展？

**A:** Gateway 是无状态服务，支持多实例部署：

```
        Nginx (负载均衡)
              |
   +----------+----------+
   |          |          |
Gateway-1  Gateway-2  Gateway-3
   |          |          |
   +----------+----------+
              |
           MySQL
```

**注意事项：**
- 数据库是唯一共享资源
- 建议使用数据库读写分离
- 可以引入 Redis 缓存 Token 和配置

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## License

[Apache License 2.0](LICENSE)

---

## 联系方式

如有问题或建议，请提交 [Issue](https://github.com/ruke318/gateway/issues)。
