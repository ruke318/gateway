<div align="center">

# Gateway

**轻量级 API 网关 - 外部接口统一对接平台**

[简体中文](./readme.md) | [English](./readme_en.md)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

</div>

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

## 核心功能

- 🔄 **协议转换** - 自动转换请求/响应格式，支持 JSON、Form、XML
- 📝 **DSL 映射** - 声明式字段映射，无需编码即可完成数据转换
- 🪝 **Hook 扩展** - JavaScript 脚本处理签名、加密、Token 获取等复杂逻辑
- 🏢 **多租户** - 厂商/机构/接口三层架构，灵活管理多方对接关系
- 🎨 **可视化管理** - Web 界面配置，无需重启即可生效

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 5.7+

### 安装运行

```bash
git clone https://github.com/yourusername/gateway.git
cd gateway/backend
go build -o gateway .
./gateway
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
    "order_no": "ORDER001"
  }
}
```

网关自动将请求转换为厂商要求的格式，并将响应转换回统一格式返回。

## 工作原理

### 数据模型

| 实体 | 说明 | 示例 |
|------|------|------|
| **厂商 (Vendor)** | 外部接口提供方 | 支付宝、微信、银联 |
| **机构 (Organization)** | 内部使用方，存储对接配置 | 总部、分公司A、分公司B |
| **接口 (Service)** | 具体的接口配置和转换规则 | 支付、退款、查询 |

### 请求流程

```
请求 → 认证 → 请求转换(DSL) → Hook处理 → 转发厂商 → 响应转换(DSL) → 返回
```

### DSL 转换示例

将内部统一格式转换为厂商格式：

```json
{
  "out_trade_no": "$.req.order_no",
  "total_amount": "$.req.amount",
  "timestamp": "@ctx.timestamp",
  "sign": "@ctx.sign"
}
```

### Hook 脚本示例

处理签名等复杂逻辑：

```javascript
// BeforeForward - 添加签名
var body = JSON.parse(ctx.request.body);
body.sign = crypto.md5(body.data + ctx.data.org_config.secret);
ctx.request.body = JSON.stringify(body);
```

## 界面截图

### 厂商管理
<img src="./doc/产商管理.png" width="800" alt="厂商列表">
<img src="./doc/新增编辑厂商.png" width="800" alt="编辑厂商">

### 机构管理
<img src="./doc/机构管理.png" width="800" alt="机构列表">
<img src="./doc/新增编辑机构.png" width="800" alt="编辑机构">

### 接口管理
<img src="./doc/接口管理.png" width="800" alt="接口列表">
<img src="./doc/接口管理编辑.png" width="800" alt="编辑接口">

### Hook 管理
<img src="./doc/Hook管理.png" width="800" alt="Hook 脚本">
<img src="./doc/接口Hook管理.png" width="800" alt="接口 Hook">

### 编辑器
<img src="./doc/脚本编辑器.png" width="800" alt="脚本编辑器">
<img src="./doc/json编辑器.png" width="800" alt="JSON 编辑器">

## 项目结构

```
backend/                 # 后端服务
├── handler/            # HTTP 处理器
├── hook/               # Hook 系统 + 内置函数
├── model/              # 数据模型
├── transform/          # DSL 转换引擎
├── proxy/              # HTTP 代理
└── router/             # 路由注册

front/                   # 前端管理界面
└── src/
    ├── views/          # 页面
    └── components/     # 组件
```

## 技术栈

- **后端：** Go + atreugo + GORM + goja
- **前端：** Vue 3 + Element Plus
- **数据库：** MySQL

## 文档

- [管理 API 参考](./ADMIN_API.md)
- [DSL Context 参考](./DSL_CONTEXT_REFERENCE.md)
- [使用示例](./EXAMPLE.md)
- [并发安全](./CONCURRENCY_SAFETY.md)

## License

[Apache License 2.0](LICENSE)
