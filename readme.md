# Gateway

轻量级、可扩展的 Go API 网关，支持声明式 DSL 数据转换、JavaScript Hook 系统和多租户架构。

## 特性

- **多租户架构** - 厂商、机构、接口三层架构，支持多对多关系
- **DSL 转换** - 使用 JSONPath + Context 语法进行声明式数据转换
- **Hook 系统** - 基于 goja 的 JavaScript 引擎，9 个生命周期节点
- **丰富内置函数** - crypto、http、encoding、util 等模块
- **多请求体类型** - 支持 JSON、Form、XML 格式转换
- **动态路径** - 支持 `{key}` 占位符的路径模板
- **热更新** - 通过管理 API 动态管理配置，零停机

## 快速开始

```bash
cd backend
go build -o gateway .
./gateway
```

默认监听 `:8080`

## 统一调用入口

```
POST /gateway/v1/invoke
```

```json
{
  "com_id": "厂商编码",
  "unit_id": "机构编码",
  "service_id": "接口标识",
  "biz_no": "业务流水号",
  "req": { "业务参数": "..." }
}
```

## 数据模型

| 模型 | 说明 |
|-----|------|
| Vendor | 厂商 - 外部接口提供方 |
| Organization | 机构 - 内部机构，与厂商多对多 |
| Service | 接口 - 关联厂商和机构，配置转换规则和 Hook |

## 请求处理流程

```
请求 → BeforeAuth → AfterAuth → BeforeRequestTransform
     → DSL请求转换 → AfterRequestTransform → BeforeForward
     → 代理转发 → AfterForward → BeforeResponseTransform
     → DSL响应转换 → AfterResponseTransform → 响应
     (出错 → OnError)
```

## DSL 转换

| 类型 | 语法 | 示例 |
|-----|------|------|
| 固定值 | 直接写 | `"200"` |
| JSONPath | `$.` 前缀 | `"$.data.id"` |
| Context | `@ctx.` 前缀 | `"@ctx.request.method"` |

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

## Hook 内置函数

| 模块 | 函数 |
|-----|------|
| crypto | md5, sha256, hmacSHA256, aesEncrypt/Decrypt, rsaEncrypt/Decrypt, rsaSign/Verify |
| http | get, post, postJSON, postForm, request |
| encoding | base64Encode/Decode, jsonEncode/Decode, urlEncode/Decode |
| util | uuid, now, formatTime, parseTime, sleep |

## 管理 API

前缀 `/admin/db`，需要 `X-Admin-Token` 认证。

| 资源 | 路径 |
|-----|------|
| 厂商 | /vendors, /vendor/{id} |
| 机构 | /organizations, /organization/{id} |
| 接口 | /services, /service/{id} |
| Hook 脚本 | /hook-scripts, /hook-script/{id} |
| 公共函数库 | /scripts, /script/{id} |
| 接口 Hook | /service-hooks, /service-hook/{id} |

## 项目结构

```
backend/
├── main.go           # 入口
├── handler/          # HTTP 处理器
├── hook/             # Hook 系统 + 内置函数
├── model/            # 数据模型
├── transform/        # DSL 转换引擎
├── proxy/            # HTTP 代理
├── database/         # 数据库操作
└── router/           # 路由注册
```

## 技术栈

atreugo (Web) / GORM (ORM) / goja (JS) / zap (日志)

## 文档

- [管理 API](./ADMIN_API.md)
- [DSL Context 参考](./DSL_CONTEXT_REFERENCE.md)
- [使用示例](./EXAMPLE.md)
- [并发安全](./CONCURRENCY_SAFETY.md)
