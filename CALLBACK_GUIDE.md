# 厂商回调功能使用指南

## 概述

Gateway 支持接收厂商回调，并将回调数据转发到内部业务系统。回调功能**完全复用 invoke 流程**，包括 Hook、DSL 转换、代理转发等所有功能。

## 回调 URL 格式

```
POST /gateway/v1/notify/{service_id}/{channel}
GET  /gateway/v1/notify/{service_id}/{channel}
```

### 参数说明

- `service_id`: 服务标识（如 `payNotify`, `refundNotify`）
- `channel`: 渠道标识
  - **数字**（如 `1`, `2`）：走默认处理器，不验签，直接转换
  - **字符串**（如 `alipay`, `wechat`）：使用专属处理器

### URL 示例

```bash
# 数字渠道（走默认逻辑）
POST https://your-gateway.com/gateway/v1/notify/payNotify/1

# 支付宝回调
POST https://your-gateway.com/gateway/v1/notify/payNotify/alipay

# 微信回调
POST https://your-gateway.com/gateway/v1/notify/refundNotify/wechat
```

---

## 配置步骤

### 1. 在数据库中配置接口

```sql
-- 示例：配置支付宝支付回调接口
INSERT INTO service (
    service_id,
    vendor_id,      -- 支付宝厂商ID
    org_id,         -- 机构ID
    name,
    backend_url,    -- 内部系统地址
    backend_path,   -- 内部系统路径
    backend_method,
    request_transform
) VALUES (
    'payNotify',
    1,  -- 假设支付宝厂商ID=1
    1,  -- 假设机构ID=1
    '支付成功回调',
    'http://order-service:8080',
    '/api/payment/notify',
    'POST',
    '{
        "order_no": "$.req.out_trade_no",
        "trade_no": "$.req.trade_no",
        "amount": "$.req.total_amount",
        "status": "$.req.trade_status",
        "notify_time": "$.req.notify_time"
    }'
);
```

### 2. 配置 Hook（可选）

#### 验签 Hook

```javascript
// BeforeAuth Hook - 支付宝验签
var body = context.data.request.body;

// 提取签名
var sign = body.sign;
var signType = body.sign_type;
delete body.sign;
delete body.sign_type;

// 构造待签名字符串
var keys = Object.keys(body).sort();
var signStr = keys.map(function(k) {
    return k + '=' + body[k];
}).join('&');

// 验证签名
var publicKey = context.data.org_config.alipay_public_key;
var valid = crypto.rsaVerify(signStr, sign, publicKey, 'SHA256');

if (!valid) {
    throw new Error('支付宝签名验证失败');
}

console.log('支付宝签名验证通过');
```

#### 自定义响应格式 Hook

```javascript
// AfterResponseTransform Hook - 返回支付宝要求的格式
var response = JSON.parse(context.responseBody);

if (response.code === 200) {
    // 成功：返回支付宝要求的格式
    context.responseBody = JSON.stringify({"code": "SUCCESS"});
} else {
    // 失败
    context.responseBody = JSON.stringify({"code": "FAIL", "msg": response.message});
}
```

### 3. 配置机构参数

```sql
-- 在 organization 表中配置支付宝公钥等参数
UPDATE organization SET config = '{
    "alipay_public_key": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
    "app_id": "2021000000000001"
}' WHERE id = 1;
```

---

## 调用流程

### 场景 1：数字渠道（默认逻辑）

```bash
# 1. 厂商发起回调
POST https://your-gateway.com/gateway/v1/notify/payNotify/1
Content-Type: application/json

{
  "order_no": "ORDER001",
  "amount": 100,
  "status": "success"
}

# 2. Gateway 处理流程
DefaultNotifyProcessor 转换:
  ↓
InvokeRequest {
  com_id: "1",
  unit_id: "default",
  service_id: "payNotify",
  biz_no: "ORDER001",
  req: {"order_no": "ORDER001", "amount": 100, "status": "success"}
}
  ↓
查询配置 (vendor_id=1, service_id=payNotify)
  ↓
执行 Hook（可选）
  ↓
DSL 转换
  ↓
转发到内部: POST http://order-service:8080/api/payment/notify
  ↓
返回厂商: {"code": "SUCCESS", "message": "success"}
```

### 场景 2：支付宝回调（专属处理器）

```bash
# 1. 支付宝发起回调
POST https://your-gateway.com/gateway/v1/notify/payNotify/alipay
Content-Type: application/x-www-form-urlencoded

out_trade_no=ORDER001&trade_no=2024XXX&total_amount=100&
trade_status=TRADE_SUCCESS&notify_time=2024-01-01 12:00:00&
sign=xxx&sign_type=RSA2

# 2. Gateway 处理流程
AlipayNotifyProcessor 转换:
  ↓
InvokeRequest {
  com_id: "alipay",
  unit_id: "default",
  service_id: "payNotify",
  biz_no: "ORDER001",
  req: {
    "out_trade_no": "ORDER001",
    "trade_no": "2024XXX",
    "total_amount": "100",
    "trade_status": "TRADE_SUCCESS",
    "sign": "xxx"
  }
}
  ↓
查询配置 (vendor.code='alipay', service_id=payNotify)
  ↓
BeforeAuth Hook: 验证支付宝签名 ✓
  ↓
DSL 转换 (Form → 内部 JSON)
  ↓
转发到内部系统
  ↓
AfterResponseTransform Hook: 自定义响应格式
  ↓
返回支付宝: {"code": "SUCCESS"}
```

---

## 扩展自定义渠道处理器

如果需要接入新的渠道（如银联、PayPal），可以实现自定义处理器：

```go
// backend/handler/notify_processor.go

// 银联回调处理器
type UnionpayNotifyProcessor struct{}

func (p *UnionpayNotifyProcessor) Process(ctx *atreugo.RequestCtx, serviceID string, channel string) (*model.InvokeRequest, error) {
    // 1. 解析银联特有的请求格式
    reqData := make(map[string]interface{})
    // ... 解析逻辑 ...

    // 2. 构造 InvokeRequest
    return &model.InvokeRequest{
        ComID:     "unionpay",
        UnitID:    "default",
        ServiceID: serviceID,
        BizNo:     reqData["orderId"].(string),
        Req:       reqData,
    }, nil
}

// 注册处理器
func init() {
    RegisterNotifyProcessor("unionpay", &UnionpayNotifyProcessor{})
}
```

---

## 测试示例

### 测试数字渠道

```bash
curl -X POST http://localhost:8080/gateway/v1/notify/payNotify/1 \
  -H "Content-Type: application/json" \
  -d '{
    "order_no": "ORDER001",
    "amount": 100,
    "status": "success"
  }'
```

### 测试支付宝回调

```bash
curl -X POST http://localhost:8080/gateway/v1/notify/payNotify/alipay \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d 'out_trade_no=ORDER001&trade_no=2024XXX&total_amount=100&trade_status=TRADE_SUCCESS&sign=xxx&sign_type=RSA2'
```

### 测试 GET 回调

```bash
curl "http://localhost:8080/gateway/v1/notify/refundNotify/1?order_no=ORDER001&status=success"
```

---

## 常见问题

### Q1: 如何自定义响应格式？

A: 在 `AfterResponseTransform` Hook 中修改 `context.responseBody`：

```javascript
// 自定义响应格式
context.responseBody = JSON.stringify({
    "code": "SUCCESS",
    "msg": "处理成功"
});
```

### Q2: 如何实现幂等性（防止重复通知）？

A: 在 `BeforeForward` Hook 中检查：

```javascript
// 检查订单号是否已处理
var orderNo = JSON.parse(context.requestBody).order_no;
var cacheKey = 'notify:' + orderNo;

// 使用 http 模块调用 Redis 检查
var response = http.get('http://redis-service/get?key=' + cacheKey);
if (response.statusCode === 200) {
    // 已处理过，直接返回成功
    throw new Error('DUPLICATE');
}

// 标记已处理
http.post('http://redis-service/set', {
    key: cacheKey,
    value: '1',
    ttl: 86400  // 24小时
});
```

### Q3: 数字渠道和字符串渠道有什么区别？

A:
- **数字渠道**（如 `/notify/payNotify/1`）：使用默认处理器，不做特殊处理
- **字符串渠道**（如 `/notify/payNotify/alipay`）：使用专属处理器，可以自定义解析逻辑

### Q4: 回调地址应该配置到哪里？

A: 将 Gateway 的回调地址配置到厂商后台：

- **支付宝**：商家中心 → 开发设置 → 接口加签方式 → 异步通知地址
- **微信支付**：商户平台 → 产品中心 → 开发配置 → 支付配置 → 通知URL

### Q5: 如何调试回调？

A: 查看日志：

```bash
tail -f backend/logs/gateway.log | grep "Notify"
```

日志会记录：
- 收到的原始请求（method, headers, body）
- 处理器转换结果（InvokeRequest）
- Hook 执行情况
- 转发结果
- 返回的响应

---

## 完整配置示例

### 数据库配置

```sql
-- 1. 厂商
INSERT INTO vendor (code, name, base_url) VALUES ('alipay', '支付宝', 'https://openapi.alipay.com');

-- 2. 机构
INSERT INTO organization (code, name, config) VALUES (
    'default',
    '默认机构',
    '{"alipay_public_key": "MIIBIjAN..."}'
);

-- 3. 接口
INSERT INTO service (
    service_id, vendor_id, org_id, name,
    backend_url, backend_path, backend_method,
    request_transform
) VALUES (
    'payNotify', 1, 1, '支付回调',
    'http://order-service:8080', '/api/payment/notify', 'POST',
    '{"order_no": "$.req.out_trade_no", "amount": "$.req.total_amount"}'
);

-- 4. Hook 脚本
INSERT INTO hook_script (name, content) VALUES (
    'alipay_verify_sign',
    '/* 验签脚本 */'
);

-- 5. 关联 Hook
INSERT INTO service_hook (service_pk, script_id, hook_point) VALUES (
    1, 1, 'BeforeAuth'
);
```

### 给厂商的回调地址

```
支付宝异步通知地址:
https://your-domain.com/gateway/v1/notify/payNotify/alipay

微信支付通知地址:
https://your-domain.com/gateway/v1/notify/payNotify/wechat
```

---

## 总结

回调功能的核心优势：

✅ **完全复用 invoke 流程** - Hook、DSL、转发全部复用
✅ **极简配置** - 无需新增表字段，同一套配置
✅ **灵活扩展** - 数字渠道走默认逻辑，字符串渠道自定义处理
✅ **统一管理** - 出向调用和回调用同一套系统管理

需要帮助？查看源码：
- `backend/handler/notify_processor.go` - 处理器实现
- `backend/handler/invoke.go` - HandleNotify 方法
- `backend/router/out.go` - 路由注册
