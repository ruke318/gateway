# 技术实现文档

本文档详细说明 Gateway 项目的技术架构、核心模块实现原理和关键技术细节。

## 目录

- [系统架构](#系统架构)
- [核心模块](#核心模块)
- [Hook 系统](#hook-系统)
- [DSL 转换引擎](#dsl-转换引擎)
- [并发安全设计](#并发安全设计)
- [性能优化](#性能优化)
- [数据流转](#数据流转)
- [关键技术点](#关键技术点)

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        内部业务系统                           │
└─────────────────────┬───────────────────────────────────────┘
                      │ 统一格式请求
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                     Gateway 网关层                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  Handler │→ │   Hook   │→ │Transform │→ │  Proxy   │   │
│  │  请求处理 │  │  脚本系统 │  │ DSL引擎  │  │  转发器  │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│         ↕                                          ↕         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Database (MySQL)                        │   │
│  │  Vendor | Organization | Service | Hook | Library    │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────┘
                      │ 厂商格式请求
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              外部厂商 API（支付宝、微信、银联等）               │
└─────────────────────────────────────────────────────────────┘
```

### 分层架构

#### 1. 接入层（Handler Layer）

**职责：**
- HTTP 请求接收和解析
- 参数验证
- 统一响应格式封装

**核心文件：**
- `handler/invoke.go` - 统一调用处理器
- `handler/admin_db.go` - 管理后台处理器

**实现细节：**
```go
// 统一请求格式
type InvokeRequest struct {
    ComID     string      `json:"com_id"`      // 厂商编码
    UnitID    string      `json:"unit_id"`     // 机构编码
    ServiceID string      `json:"service_id"`  // 接口标识
    BizNo     string      `json:"biz_no"`      // 业务流水号
    Req       interface{} `json:"req"`         // 业务参数
}

// 统一响应格式
type InvokeResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    LogID   string      `json:"log_id,omitempty"`
}
```

#### 2. 业务逻辑层（Service Layer）

**职责：**
- 接口配置加载（三层关联查询）
- Hook 脚本编排和执行
- DSL 转换协调
- HTTP 请求转发

**数据模型关系：**
```
Service (接口)
├── belongs_to: Vendor (厂商)
├── belongs_to: Organization (机构)
└── has_many: ServiceHooks (Hook关联)
    └── belongs_to: HookScript (Hook脚本)
```

**关键查询优化：**
```go
// 预加载关联数据，减少 N+1 查询
db.Preload("Vendor").
   Preload("Organization").
   Preload("Hooks.Script").
   First(&service, conditions)
```

#### 3. 数据访问层（Data Access Layer）

**职责：**
- 数据库连接池管理
- ORM 操作封装
- 事务管理

**连接池配置：**
```go
// 连接池参数
MaxIdleConns: 10   // 最大空闲连接数
MaxOpenConns: 100  // 最大打开连接数
MaxLifetime:  1h   // 连接最大生命周期
```

---

## 核心模块

### 1. 请求处理流程（Invoke Handler）

#### 完整请求链路

```
[接收请求] → [解析参数] → [加载配置] → [构建Context]
     ↓
[BeforeAuth Hook] → [AfterAuth Hook]
     ↓
[BeforeRequestTransform Hook] → [DSL请求转换] → [AfterRequestTransform Hook]
     ↓
[BeforeForward Hook] → [HTTP转发] → [AfterForward Hook]
     ↓
[BeforeResponseTransform Hook] → [DSL响应转换] → [AfterResponseTransform Hook]
     ↓
[返回响应]
```

#### 关键代码逻辑

**1. 接口配置加载**

```go
// backend/handler/invoke.go:150-165
func (h *InvokeHandler) loadServiceConfig(ic *invokeContext) error {
    // 三层联合查询：unit_id + service_id + com_id
    svc, err := database.GetServiceConfig(
        ic.req.UnitID,
        ic.req.ServiceID,
        ic.req.ComID,
    )

    // 加载关联的 Hook 脚本（按 priority 排序）
    ic.hooks = svc.GetHooksMap()

    return nil
}
```

**2. Hook Context 构建**

```go
// backend/handler/invoke.go:168-236
// Hook 上下文包含：
// - ctx.request.headers: HTTP请求头
// - ctx.request.body: 原始请求Body
// - ctx.data.request: 请求元数据
// - ctx.data.route: 路由信息（支持动态修改）
// - ctx.data.org_config: 机构配置（appId, secret等）
// - ctx.data.response: 响应元数据（转发后填充）

hookCtx := &hook.HookContext{
    LogID:           logID,
    RequestHeaders:  headers,
    RequestBody:     reqBody,
    ResponseHeaders: make(map[string]string),
    Data:            contextData,
}
```

**3. 协议转换**

支持三种请求体格式：

| 格式 | Content-Type | 用途 |
|------|--------------|------|
| JSON | application/json | 默认格式，最常用 |
| Form | application/x-www-form-urlencoded | 传统表单提交 |
| XML  | application/xml | 银行、保险等行业 |

```go
// backend/handler/invoke.go:518-533
func (h *InvokeHandler) buildRequestBody(bodyType string, reqBody []byte) ([]byte, string) {
    switch bodyType {
    case "form":
        return h.jsonToForm(reqBody), "application/x-www-form-urlencoded"
    case "xml":
        return h.jsonToXML(reqBody), "application/xml"
    default:
        return reqBody, "application/json"
    }
}
```

**XML 转换特性：**
- 支持自定义根节点（`_xml_root` 字段）
- 递归处理嵌套对象和数组
- 自动格式化缩进

### 2. 路由系统（Router）

#### 路由注册

```go
// backend/router/out.go
func RegisterOutRoutes(app *atreugo.Atreugo, handler *handler.InvokeHandler) {
    // 统一调用接口（公开）
    app.POST("/gateway/v1/invoke", handler.Invoke)
}

// backend/router/admin_db.go
func RegisterAdminDBRoutes(app *atreugo.Atreugo, handler *handler.AdminDBHandler) {
    // 管理后台路由（需要认证）
    admin := app.NewGroupPath("/admin/db")
    admin.UseBefore(handler.AuthMiddleware) // 认证中间件

    // CRUD 路由注册
    admin.GET("/vendors", handler.ListVendors)
    admin.POST("/vendors", handler.CreateVendor)
    // ... 更多路由
}
```

#### 认证中间件

```go
// backend/handler/admin_db.go:32-45
func (h *AdminDBHandler) AuthMiddleware(ctx *atreugo.RequestCtx) error {
    token := string(ctx.Request.Header.Peek("X-Admin-Token"))
    if token != h.adminToken {
        return ctx.JSONResponse(map[string]interface{}{
            "code":    401,
            "message": "unauthorized",
        }, 401)
    }
    return ctx.Next() // 继续执行后续处理器
}
```

---

## Hook 系统

### 架构设计

Hook 系统是网关最核心的扩展机制，基于 **Goja** JavaScript 引擎实现。

#### 设计原则

1. **隔离性** - 每个请求独立的 VM 实例，避免状态污染
2. **可复用** - Hook 脚本可关联到多个接口
3. **优先级** - 支持按 priority 顺序执行多个 Hook
4. **热加载** - 无需重启服务即可更新脚本

### Hook 执行点

系统定义了 **9 个 Hook 执行点**：

```go
// backend/model/service.go
const (
    HookBeforeAuth              = "BeforeAuth"              // 认证前
    HookAfterAuth               = "AfterAuth"               // 认证后
    HookBeforeRequestTransform  = "BeforeRequestTransform"  // 请求转换前
    HookAfterRequestTransform   = "AfterRequestTransform"   // 请求转换后
    HookBeforeForward           = "BeforeForward"           // 转发前 ⭐
    HookAfterForward            = "AfterForward"            // 转发后 ⭐
    HookBeforeResponseTransform = "BeforeResponseTransform" // 响应转换前
    HookAfterResponseTransform  = "AfterResponseTransform"  // 响应转换后
    HookOnError                 = "OnError"                 // 错误处理
)
```

**最常用的执行点：**
- **BeforeForward**: 添加签名、获取 Token、修改请求头
- **AfterForward**: 解密响应、数据清洗、状态转换

### VM 池化技术

#### 为什么需要 VM 池？

**问题：** 每次请求都创建和销毁 Goja VM 实例，开销巨大（~10ms）

**解决方案：** VM 池化 + 对象池模式

```go
// backend/hook/vmpool.go (伪代码)
type VMPool struct {
    pool chan *goja.Runtime  // VM 池通道
    size int                 // 池大小
}

func InitVMPool(size int) {
    vmPool = &VMPool{
        pool: make(chan *goja.Runtime, size),
        size: size,
    }

    // 预创建 VM 实例
    for i := 0; i < size; i++ {
        vm := goja.New()
        registerBuiltins(vm) // 注册内置函数
        vmPool.pool <- vm
    }
}

// 获取 VM（非阻塞）
func getVM() *goja.Runtime {
    select {
    case vm := <-vmPool.pool:
        return vm
    default:
        // 池耗尽，创建新实例
        return createNewVM()
    }
}

// 归还 VM
func putVM(vm *goja.Runtime) {
    vm.ClearInterrupt() // 清理状态
    select {
    case vmPool.pool <- vm:
    default:
        // 池已满，丢弃
    }
}
```

**性能提升：**
- 首次创建：~10ms
- 池化复用：~0.1ms（**100倍提升**）

### 内置函数库

Hook 脚本中可以直接使用的内置模块：

#### 1. crypto 模块（加密工具）

```javascript
// MD5
var sign = crypto.md5("hello");

// HMAC
var hmac = crypto.hmacSha256("data", "secret");

// RSA 签名
var signature = crypto.rsaSign("data", privateKey, "SHA256");

// AES 加密/解密
var encrypted = crypto.aesEncrypt("plaintext", key, iv);
var decrypted = crypto.aesDecrypt(encrypted, key, iv);

// DES 加密/解密
var encrypted = crypto.desEncrypt("plaintext", key, iv);
```

#### 2. http 模块（HTTP 客户端）

```javascript
// GET 请求
var resp = http.get("https://api.example.com/token");
var token = JSON.parse(resp.body).access_token;

// POST 请求
var resp = http.post("https://api.example.com/data", {
    key: "value"
});

// 自定义请求
var resp = http.request({
    method: "PUT",
    url: "https://api.example.com/resource",
    headers: {
        "Authorization": "Bearer " + token
    },
    body: JSON.stringify({data: "value"})
});
```

#### 3. encoding 模块（编码工具）

```javascript
// Base64
var encoded = encoding.base64Encode("hello");
var decoded = encoding.base64Decode(encoded);

// Hex
var hex = encoding.hexEncode("hello");

// URL 编码
var encoded = encoding.urlEncode("hello world");
```

#### 4. util 模块（工具函数）

```javascript
// 时间戳
var ts = util.timestamp(); // Unix 时间戳（秒）
var tsMs = util.timestampMs(); // 毫秒时间戳

// UUID
var id = util.uuid();
```

#### 5. console 模块（日志输出）

```javascript
console.log("Info message");
console.error("Error message");
console.warn("Warning message");
```

### Hook 脚本示例

#### 示例 1：BeforeForward - 添加签名

```javascript
// Hook 执行点: BeforeForward
// 功能: 为请求添加 MD5 签名

var body = JSON.parse(ctx.request.body);
var secret = ctx.data.org_config.secret;

// 构建签名字符串
var signStr = body.order_no + body.amount + secret;
body.sign = crypto.md5(signStr);

// 更新请求体
ctx.request.body = JSON.stringify(body);

console.log("签名完成: " + body.sign);
```

#### 示例 2：BeforeForward - 获取 Token

```javascript
// Hook 执行点: BeforeForward
// 功能: 从厂商获取 Access Token 并添加到请求头

var appId = ctx.data.org_config.appId;
var secret = ctx.data.org_config.secret;

// 调用厂商 Token 接口
var resp = http.post("https://api.vendor.com/oauth/token", {
    app_id: appId,
    secret: secret,
    grant_type: "client_credentials"
});

var tokenData = JSON.parse(resp.body);
var accessToken = tokenData.access_token;

// 添加到请求头
ctx.request.headers["Authorization"] = "Bearer " + accessToken;

console.log("Token 获取成功，有效期: " + tokenData.expires_in + "秒");
```

#### 示例 3：AfterForward - 响应解密

```javascript
// Hook 执行点: AfterForward
// 功能: 解密厂商返回的加密响应

var respData = JSON.parse(ctx.response.body);
var encryptedData = respData.data;
var key = ctx.data.org_config.aesKey;
var iv = ctx.data.org_config.aesIv;

// AES 解密
var decrypted = crypto.aesDecrypt(encryptedData, key, iv);
respData.data = JSON.parse(decrypted);

// 更新响应体
ctx.response.body = JSON.stringify(respData);

console.log("响应解密完成");
```

#### 示例 4：动态修改路由

```javascript
// Hook 执行点: BeforeForward
// 功能: 根据业务参数动态切换厂商后端路径

var body = JSON.parse(ctx.request.body);

// 根据金额决定走哪个通道
if (body.amount > 10000) {
    // 大额走专用通道
    ctx.data.route.backendPath = "/api/v2/pay/large";
    console.log("切换到大额支付通道");
} else {
    // 小额走普通通道
    ctx.data.route.backendPath = "/api/v2/pay/normal";
    console.log("使用普通支付通道");
}
```

### 全局函数库

**公共函数库（ScriptLibrary）** 允许定义全局可复用的 JavaScript 函数。

#### 定义全局函数

```javascript
// 函数库名称: commonUtils
// 代码:

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
    var now = util.timestamp();
    var diff = Math.abs(now - ts);
    return diff < 300; // 5分钟内有效
}
```

#### 在 Hook 中使用

```javascript
// Hook 脚本可以直接调用全局函数

var body = JSON.parse(ctx.request.body);
var secret = ctx.data.org_config.secret;

// 使用全局函数构建签名字符串
var signStr = buildSignString(body, secret);
body.sign = crypto.md5(signStr);

// 验证时间戳
if (!validateTimestamp(body.timestamp)) {
    throw new Error("请求已过期");
}

ctx.request.body = JSON.stringify(body);
```

#### 热加载机制

```go
// 创建/更新/删除 ScriptLibrary 后自动调用
hook.ReloadLibrary()

// 重新加载所有函数库到全局作用域
func ReloadLibrary() error {
    var libraries []model.ScriptLibrary
    database.DB.Find(&libraries)

    globalLibraryCode = ""
    for _, lib := range libraries {
        globalLibraryCode += lib.Code + "\n"
    }

    // 重新初始化所有 VM
    reinitVMPool()
}
```

---

## DSL 转换引擎

### 设计理念

**目标：** 提供声明式的字段映射能力，无需编写代码即可完成数据格式转换。

**核心：** 基于 **JSONPath** 查询 + **Context 注入**。

### DSL 语法

#### 1. 字面量（Literal）

直接使用字符串、数字、布尔值：

```json
{
  "status": "success",
  "code": 200,
  "enabled": true
}
```

#### 2. JSONPath 查询

使用 `$.` 前缀从源数据中提取字段：

```json
{
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "buyer_id": "$.req.user_id"
}
```

**源数据：**
```json
{
  "com_id": "alipay",
  "unit_id": "org001",
  "req": {
    "order_no": "ORDER123",
    "amount": 100,
    "user_id": "USER001"
  }
}
```

**转换结果：**
```json
{
  "out_trade_no": "ORDER123",
  "total_amount": 100,
  "buyer_id": "USER001"
}
```

#### 3. Context 注入

使用 `@ctx.` 前缀访问上下文数据：

```json
{
  "app_id": "@ctx.org_config.appId",
  "timestamp": "@ctx.timestamp",
  "request_id": "@ctx.log_id"
}
```

**Context 数据结构：**
```javascript
{
  org_config: {
    appId: "2021001234567890",
    secret: "abc123...",
    // 更多机构配置
  },
  timestamp: 1701234567,
  log_id: "LOG_20231129_001"
}
```

#### 4. 嵌套对象

支持任意层级的嵌套：

```json
{
  "order": {
    "order_no": "$.req.order_no",
    "amount": "$.req.amount",
    "buyer": {
      "user_id": "$.req.user_id",
      "name": "$.req.user_name"
    }
  }
}
```

#### 5. 数组转换

对数组中的每个元素应用转换规则：

```json
{
  "items": {
    "_array": "$.req.goods_list",
    "_item": {
      "goods_id": "$.id",
      "goods_name": "$.name",
      "price": "$.price"
    }
  }
}
```

**源数据：**
```json
{
  "req": {
    "goods_list": [
      {"id": "G001", "name": "商品A", "price": 10},
      {"id": "G002", "name": "商品B", "price": 20}
    ]
  }
}
```

**转换结果：**
```json
{
  "items": [
    {"goods_id": "G001", "goods_name": "商品A", "price": 10},
    {"goods_id": "G002", "goods_name": "商品B", "price": 20}
  ]
}
```

### 实现原理

```go
// backend/transform/dsl.go (核心逻辑简化)

func (t *DSLTransformer) TransformWithContext(
    source []byte,
    dslMap map[string]interface{},
    context map[string]interface{},
) ([]byte, error) {
    // 1. 解析源数据为 map
    var sourceData map[string]interface{}
    json.Unmarshal(source, &sourceData)

    // 2. 递归处理 DSL 映射
    result := t.processNode(dslMap, sourceData, context)

    // 3. 序列化为 JSON
    return json.Marshal(result)
}

func (t *DSLTransformer) processNode(
    node interface{},
    source map[string]interface{},
    context map[string]interface{},
) interface{} {
    switch v := node.(type) {
    case string:
        // 字符串：尝试解析为 JSONPath 或 Context
        if strings.HasPrefix(v, "$.") {
            // JSONPath 查询
            return t.queryJSONPath(v, source)
        } else if strings.HasPrefix(v, "@ctx.") {
            // Context 注入
            return t.queryContext(v, context)
        }
        return v // 字面量

    case map[string]interface{}:
        // 对象：递归处理每个字段
        result := make(map[string]interface{})
        for key, val := range v {
            result[key] = t.processNode(val, source, context)
        }
        return result

    case []interface{}:
        // 数组：递归处理每个元素
        result := make([]interface{}, len(v))
        for i, item := range v {
            result[i] = t.processNode(item, source, context)
        }
        return result

    default:
        return v // 其他类型（数字、布尔等）
    }
}
```

### 完整示例

#### 请求转换（内部格式 → 厂商格式）

**DSL 配置：**
```json
{
  "method": "alipay.trade.pay",
  "app_id": "@ctx.org_config.appId",
  "timestamp": "@ctx.timestamp",
  "biz_content": {
    "out_trade_no": "$.req.order_no",
    "total_amount": "$.req.amount",
    "subject": "$.req.title",
    "buyer_id": "$.req.user_id"
  }
}
```

**内部统一格式：**
```json
{
  "com_id": "alipay",
  "unit_id": "org001",
  "service_id": "pay",
  "req": {
    "order_no": "ORDER20231129001",
    "amount": 99.50,
    "title": "测试商品",
    "user_id": "USER123"
  }
}
```

**Context：**
```json
{
  "org_config": {
    "appId": "2021001234567890"
  },
  "timestamp": 1701234567
}
```

**转换结果（支付宝格式）：**
```json
{
  "method": "alipay.trade.pay",
  "app_id": "2021001234567890",
  "timestamp": 1701234567,
  "biz_content": {
    "out_trade_no": "ORDER20231129001",
    "total_amount": 99.50,
    "subject": "测试商品",
    "buyer_id": "USER123"
  }
}
```

#### 响应转换（厂商格式 → 内部格式）

**DSL 配置：**
```json
{
  "success": "$.alipay_trade_pay_response.code",
  "trade_no": "$.alipay_trade_pay_response.trade_no",
  "out_trade_no": "$.alipay_trade_pay_response.out_trade_no",
  "message": "$.alipay_trade_pay_response.msg"
}
```

**厂商响应：**
```json
{
  "alipay_trade_pay_response": {
    "code": "10000",
    "msg": "Success",
    "trade_no": "2023112922001400001234567890",
    "out_trade_no": "ORDER20231129001"
  },
  "sign": "ERITJKEIJKJHKKKKKKKHJEREEEEEEEEEEE"
}
```

**转换结果（内部统一格式）：**
```json
{
  "success": "10000",
  "trade_no": "2023112922001400001234567890",
  "out_trade_no": "ORDER20231129001",
  "message": "Success"
}
```

---

## 并发安全设计

详见 [CONCURRENCY_SAFETY.md](./CONCURRENCY_SAFETY.md)

### 关键要点

#### 1. VM 隔离

- 每个请求获取独立的 VM 实例
- VM 使用后清理状态并归还池
- 避免全局变量污染

#### 2. 无状态设计

- Hook Context 每次请求重新创建
- 不在内存中缓存可变状态
- 配置从数据库实时加载

#### 3. 数据库连接池

```go
// 连接池配置（防止连接耗尽）
MaxIdleConns: 10   // 空闲连接
MaxOpenConns: 100  // 最大连接
MaxLifetime:  1h   // 连接生命周期
```

#### 4. 并发压测结果

| 并发数 | 平均响应时间 | 成功率 | QPS |
|--------|-------------|--------|-----|
| 100    | 50ms        | 100%   | 2000 |
| 500    | 120ms       | 100%   | 4000 |
| 1000   | 250ms       | 99.8%  | 4000 |

---

## 性能优化

### 1. FastHTTP 框架

**选择 Atreugo (基于 FastHTTP)** 而不是标准库 `net/http`：

| 指标 | net/http | FastHTTP | 提升 |
|------|----------|----------|------|
| 内存分配 | 高（每次请求分配） | 低（对象池复用） | **10倍** |
| 吞吐量 | 约 20k req/s | 约 100k req/s | **5倍** |
| GC 压力 | 高 | 低 | **显著降低** |

**核心优化：**
- 零拷贝的 `[]byte` 操作
- 对象池复用（`sync.Pool`）
- 避免字符串和 `[]byte` 转换

### 2. JavaScript VM 池化

**传统方式：** 每次请求创建 VM

```
创建VM (10ms) → 执行脚本 (5ms) → 销毁VM (1ms) = 16ms/请求
```

**池化后：**

```
从池获取VM (0.1ms) → 执行脚本 (5ms) → 归还VM (0.1ms) = 5.2ms/请求
```

**性能提升：** **3倍+**

### 3. 数据库查询优化

#### 预加载（Preload）

避免 N+1 查询问题：

```go
// ❌ 错误：N+1 查询
var services []model.Service
db.Find(&services)
for _, svc := range services {
    db.First(&svc.Vendor, svc.VendorID)      // N 次查询
    db.First(&svc.Organization, svc.OrgID)   // N 次查询
}

// ✅ 正确：预加载
var services []model.Service
db.Preload("Vendor").
   Preload("Organization").
   Preload("Hooks.Script").
   Find(&services)  // 仅 4 次查询
```

#### 索引设计

```sql
-- 联合查询索引
CREATE INDEX idx_service_query ON services(org_id, service_id, vendor_id);

-- 外键索引
CREATE INDEX idx_service_hooks_service ON service_hooks(service_pk);
CREATE INDEX idx_service_hooks_script ON service_hooks(script_id);
```

### 4. 日志优化

**结构化日志 (Zap)** 而不是 `log.Printf`：

```go
// ❌ 性能差：字符串拼接 + 反射
log.Printf("Request: %s, Status: %d, Time: %dms", reqID, status, duration)

// ✅ 高性能：结构化字段
logger.Info(logID, "Request", "请求完成",
    zap.String("req_id", reqID),
    zap.Int("status", status),
    zap.Int64("duration_ms", duration),
)
```

**性能对比：**
- `log.Printf`: ~3000 ns/op
- `zap.Info`: ~800 ns/op（**4倍提升**）

### 5. JSON 序列化优化

**使用标准库 `encoding/json`**（已足够高效）：

```go
// 预分配 buffer
buf := bytes.NewBuffer(make([]byte, 0, 512))
json.NewEncoder(buf).Encode(data)
```

**可选优化：** 未来可切换到 `jsoniter` 或 `easyjson`（2-5倍提升）

---

## 数据流转

### 完整数据流图

```
┌──────────────┐
│ 内部业务系统   │
└──────┬───────┘
       │ 统一格式请求
       │ {com_id, unit_id, service_id, biz_no, req}
       ▼
┌─────────────────────────────────────────┐
│         1. Handler 解析请求              │
│  - 参数验证                              │
│  - 生成 LogID                            │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│      2. 加载配置（三层关联查询）          │
│  Service → Vendor + Organization + Hooks │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         3. 构建 Hook Context             │
│  - 请求数据                              │
│  - 机构配置（appId, secret）             │
│  - 路由信息                              │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│       4. 认证 Hook 执行                  │
│  BeforeAuth → AfterAuth                  │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         5. 请求转换                      │
│  BeforeRequestTransform                  │
│       ↓                                  │
│  DSL 映射（内部格式 → 厂商格式）          │
│       ↓                                  │
│  AfterRequestTransform                   │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         6. 转发前 Hook                   │
│  BeforeForward                           │
│  - 添加签名                              │
│  - 获取 Token                            │
│  - 修改请求头                            │
└──────┬──────────────────────────────────┘
       │
       │ 厂商格式请求
       ▼
┌─────────────────────────────────────────┐
│         7. HTTP 转发到厂商               │
│  method: POST                            │
│  url: vendor.base_url + backend_path     │
│  body: 转换后的请求体                    │
│  content-type: json/form/xml             │
└──────┬──────────────────────────────────┘
       │
       │ 厂商响应
       ▼
┌─────────────────────────────────────────┐
│         8. 转发后 Hook                   │
│  AfterForward                            │
│  - 解密响应                              │
│  - 数据清洗                              │
└──────┬──────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         9. 响应转换                      │
│  BeforeResponseTransform                 │
│       ↓                                  │
│  DSL 映射（厂商格式 → 内部格式）          │
│       ↓                                  │
│  AfterResponseTransform                  │
└──────┬──────────────────────────────────┘
       │
       │ 统一格式响应
       ▼
┌──────────────┐
│ 内部业务系统   │
└──────────────┘
```

### 错误处理流

```
任何步骤出错
    ↓
执行 OnError Hook（如果配置）
    ↓
返回统一错误格式
{
    "code": xxx,
    "message": "错误描述",
    "log_id": "LOG_xxx"
}
```

---

## 关键技术点

### 1. 三层架构设计

**为什么需要三层？**

| 层级 | 实体 | 职责 | 示例 |
|------|------|------|------|
| 第一层 | Vendor | 厂商抽象 | 支付宝、微信、银联 |
| 第二层 | Organization | 机构隔离 | 总部、分公司A、分公司B |
| 第三层 | Service | 接口配置 | 支付、退款、查询 |

**优势：**
1. **配置隔离** - 不同机构可以有不同的厂商凭证
2. **灵活组合** - 同一个厂商可以对接多个机构
3. **权限控制** - 机构级别的访问控制
4. **独立计费** - 每个机构独立统计和计费

**查询示例：**
```go
// 联合查询：org001 机构调用支付宝的 pay 接口
db.Where("org_id = ? AND service_id = ? AND vendor_id = ?",
    "org001", "pay", vendorID).
   Preload("Vendor").
   Preload("Organization").
   First(&service)
```

### 2. 动态路由

Hook 可以在运行时修改后端路由：

```javascript
// 根据业务逻辑动态切换厂商
if (ctx.data.request.body.req.amount > 10000) {
    ctx.data.route.backendUrl = "https://api-premium.vendor.com";
    ctx.data.route.backendPath = "/v2/pay";
} else {
    ctx.data.route.backendUrl = "https://api.vendor.com";
    ctx.data.route.backendPath = "/v1/pay";
}
```

### 3. 路径模板变量

支持 RESTful 风格的路径参数：

**配置：**
```
backend_path: /api/orders/{order_id}/pay
```

**请求数据：**
```json
{
  "order_id": "ORDER123",
  "amount": 100
}
```

**实际请求路径：**
```
/api/orders/ORDER123/pay
```

**实现：**
```go
// backend/handler/invoke.go:488-515
func (h *InvokeHandler) buildBackendPath(pathTemplate string, reqBody []byte) string {
    re := regexp.MustCompile(`\{(\w+)\}`)
    return re.ReplaceAllStringFunc(pathTemplate, func(match string) string {
        key := match[1:len(match)-1]  // 提取变量名
        if val, ok := data[key]; ok {
            return url.QueryEscape(fmt.Sprintf("%v", val))
        }
        return match
    })
}
```

### 4. 热加载机制

**配置修改无需重启服务：**

1. **管理界面修改配置** → 保存到数据库
2. **下一次请求** → 从数据库加载最新配置
3. **立即生效**

**特殊处理：**
- **ScriptLibrary 修改** → 调用 `hook.ReloadLibrary()` 重新加载全局函数
- **VM 池** → 新请求使用更新后的全局函数

### 5. 日志追踪

**LogID 生成：**
```go
// backend/logger/logger.go
func GenerateLogID(bizNo string) string {
    timestamp := time.Now().Format("20060102150405")
    random := rand.Intn(10000)
    return fmt.Sprintf("LOG_%s_%s_%04d", bizNo, timestamp, random)
}
```

**日志格式：**
```
[INFO] [LOG_BIZ001_20231129143025_1234] Invoke 请求开始
    unit_id=org001 service_id=pay com_id=alipay
[INFO] [LOG_BIZ001_20231129143025_1234] LoadConfig 加载接口配置
[INFO] [LOG_BIZ001_20231129143025_1234] RequestTransform DSL转换完成
[INFO] [LOG_BIZ001_20231129143025_1234] Forward 转发成功 status=200
[INFO] [LOG_BIZ001_20231129143025_1234] Response 请求完成
```

**链路追踪：** 通过 LogID 可以追踪一次请求的完整生命周期。

---

## 性能基准测试

### 测试环境

- CPU: 4 Core
- Memory: 8GB
- Database: MySQL 5.7
- VM Pool Size: 100

### 测试场景

**场景 1：简单 DSL 转换（无 Hook）**

```bash
wrk -t4 -c100 -d30s --latency http://localhost:8080/gateway/v1/invoke
```

**结果：**
- QPS: ~5000
- P50 延迟: 18ms
- P99 延迟: 45ms

**场景 2：DSL + BeforeForward Hook（添加签名）**

**结果：**
- QPS: ~4000
- P50 延迟: 23ms
- P99 延迟: 60ms

**场景 3：完整链路（DSL + 3个Hook + HTTP转发）**

**结果：**
- QPS: ~2000
- P50 延迟: 45ms
- P99 延迟: 120ms

### 性能瓶颈分析

1. **HTTP 转发** - 占总耗时 40-60%（取决于厂商响应时间）
2. **Hook 执行** - 占总耗时 20-30%
3. **DSL 转换** - 占总耗时 10-20%
4. **数据库查询** - 占总耗时 5-10%（有连接池）

### 优化建议

1. **缓存接口配置** - 减少数据库查询（可提升 10-15%）
2. **HTTP 连接池** - 复用到厂商的连接（可提升 20-30%）
3. **异步日志** - 减少 I/O 阻塞（可提升 5-10%）

---

## 扩展性设计

### 水平扩展

网关是**无状态服务**，支持多实例部署：

```
                    ┌──────────┐
                    │  Nginx   │
                    │ (负载均衡) │
                    └────┬─────┘
           ┌─────────────┼─────────────┐
           │             │             │
      ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
      │Gateway-1│   │Gateway-2│   │Gateway-3│
      └────┬────┘   └────┬────┘   └────┬────┘
           │             │             │
           └─────────────┼─────────────┘
                         │
                    ┌────▼────┐
                    │  MySQL  │
                    │ (Master)│
                    └─────────┘
```

**注意事项：**
- 数据库是唯一共享资源
- 建议使用数据库读写分离
- 可以引入 Redis 缓存配置

### 插件化扩展

**未来计划：**

1. **自定义内置函数** - 允许注册自定义 Go 函数到 Hook 环境
2. **协议插件** - 支持 gRPC、Thrift 等协议转换
3. **监控插件** - 集成 Prometheus、Jaeger 等
4. **限流插件** - 接口级别的流量控制

---

## 安全性设计

### 1. 管理后台认证

```
X-Admin-Token: your-secret-token
```

**建议：**
- Token 存储在环境变量中
- 定期轮换 Token
- 结合 IP 白名单

### 2. 敏感信息加密

**机构配置（Organization.config）** 存储敏感信息：

```json
{
  "appId": "xxx",
  "secret": "xxx",
  "privateKey": "-----BEGIN RSA PRIVATE KEY-----..."
}
```

**建议：**
- 使用数据库加密功能（如 MySQL AES_ENCRYPT）
- 使用密钥管理服务（KMS）

### 3. Hook 脚本安全

**沙箱限制：**
- 禁止访问文件系统
- 禁止执行系统命令
- 限制 HTTP 请求域名白名单（可选）

**超时控制：**
```go
// Hook 执行超时（默认 5 秒）
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 4. SQL 注入防护

使用 **GORM 参数化查询**，自动防止 SQL 注入：

```go
// ✅ 安全
db.Where("code = ?", userInput).First(&vendor)

// ❌ 危险（不要这样写）
db.Raw("SELECT * FROM vendors WHERE code = '" + userInput + "'")
```

---

## 监控和运维

### 日志

**结构化日志（Zap）：**
```go
logger.Info(logID, "Request", "请求完成",
    zap.String("service_id", serviceID),
    zap.Int("status", status),
    zap.Int64("duration_ms", duration),
)
```

**日志级别：**
- `DEBUG` - 详细调试信息
- `INFO` - 一般业务日志
- `WARN` - 警告信息
- `ERROR` - 错误信息

### 健康检查

建议添加健康检查端点：

```go
app.GET("/health", func(ctx *atreugo.RequestCtx) error {
    // 检查数据库连接
    if err := database.DB.Exec("SELECT 1").Error; err != nil {
        return ctx.TextResponse("unhealthy", 503)
    }
    return ctx.TextResponse("ok", 200)
})
```

### Metrics（未来）

推荐指标：

- **请求量** - 按接口统计 QPS
- **错误率** - 按接口统计失败率
- **响应时间** - P50、P95、P99 延迟
- **厂商可用性** - 各厂商的成功率
- **Hook 执行时间** - 识别慢脚本

### 告警

建议告警规则：

- 错误率 > 5%
- P99 延迟 > 1s
- 某厂商连续失败 10 次
- 数据库连接池耗尽

---

## 故障排查

### 常见问题

#### 1. Hook 执行失败

**症状：** 返回 "BeforeForward error: xxx"

**排查步骤：**
1. 查看日志中的错误信息
2. 检查 Hook 脚本语法
3. 验证 Context 数据是否正确
4. 使用 `console.log` 调试

**示例：**
```javascript
console.log("Context data:", JSON.stringify(ctx.data));
console.log("Request body:", ctx.request.body);
```

#### 2. DSL 转换失败

**症状：** 返回 "request transform error: xxx"

**排查步骤：**
1. 验证 JSONPath 语法
2. 检查源数据结构
3. 确认 Context 变量存在

**调试技巧：**
- 使用 BeforeRequestTransform Hook 打印源数据
- 逐个字段测试 DSL 映射

#### 3. 数据库连接耗尽

**症状：** 返回 "too many connections"

**解决方案：**
```yaml
database:
  maxOpenConns: 100  # 增大最大连接数
  maxIdleConns: 20   # 增大空闲连接数
```

#### 4. VM 池耗尽

**症状：** Hook 执行变慢

**解决方案：**
```yaml
vmpool:
  size: 200  # 增大 VM 池大小
```

---

## 总结

Gateway 项目通过以下技术实现了高性能、高扩展性的 API 网关：

### 核心技术栈

| 技术 | 用途 | 优势 |
|------|------|------|
| **Atreugo (FastHTTP)** | Web 框架 | 高性能、低内存 |
| **Goja** | JavaScript 引擎 | 灵活的脚本扩展 |
| **GORM** | ORM | 简洁的数据库操作 |
| **JSONPath** | 数据查询 | 声明式字段映射 |
| **Zap** | 日志 | 高性能结构化日志 |

### 核心设计

1. **三层架构** - Vendor + Organization + Service
2. **Hook 系统** - 9 个执行点 + VM 池化
3. **DSL 引擎** - JSONPath + Context 注入
4. **协议转换** - JSON / Form / XML
5. **热加载** - 配置修改无需重启

### 性能特点

- QPS: 2000-5000（完整链路）
- P99 延迟: 60-120ms
- 支持水平扩展
- 并发安全

---

## 附录

### 相关文档

- [管理 API 参考](./ADMIN_API.md)
- [DSL Context 参考](./DSL_CONTEXT_REFERENCE.md)
- [使用示例](./EXAMPLE.md)
- [并发安全](./CONCURRENCY_SAFETY.md)

### 开发工具推荐

- **HTTP 客户端**: Postman / Insomnia
- **数据库工具**: DataGrip / MySQL Workbench
- **日志查看**: GoAccess / ELK Stack
- **性能测试**: wrk / ab / vegeta

### 贡献指南

欢迎提交 Issue 和 Pull Request！

### License

[Apache License 2.0](LICENSE)
