<div align="center">

# Gateway

**轻量级、可扩展的 Go API 网关**

[English](./readme.md) | [简体中文](./readme_zh.md)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/yourusername/gateway/pulls)

声明式 DSL 数据转换 • JavaScript Hook 系统 • 多租户架构

[特性](#特性) • [快速开始](#快速开始) • [文档](#文档) • [界面截图](#界面截图)

</div>

---

## 特性

- 🏢 **多租户架构** - 厂商/机构/接口三层架构，支持多对多关系
- 🔄 **DSL 转换** - 使用 JSONPath + Context 语法进行声明式数据转换
- 🪝 **Hook 系统** - 基于 goja 的 JavaScript 引擎，9 个生命周期钩子
- 🛠️ **丰富内置函数** - 加密、HTTP、编码、工具等模块
- 📦 **多种内容类型** - 支持 JSON、Form、XML 格式转换
- 🎯 **动态路由** - 支持 `{key}` 占位符的路径模板
- 🔥 **热更新** - 通过管理 API 动态更新配置，零停机
- 🎨 **Web 管理界面** - 完整的 Web 可视化管理界面

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- MySQL 5.7+（或 PostgreSQL/SQLite）

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/gateway.git
cd gateway/backend

# 构建
go build -o gateway .

# 运行
./gateway
```

网关默认在 `:8080` 端口启动。

### 基本使用

向统一网关入口发送请求：

```bash
POST /gateway/v1/invoke
Content-Type: application/json
```

```json
{
  "com_id": "厂商编码",
  "unit_id": "机构编码",
  "service_id": "接口标识",
  "biz_no": "业务流水号",
  "req": {
    "your": "业务数据"
  }
}
```

## 架构设计

### 数据模型

| 实体 | 说明 |
|------|------|
| **Vendor（厂商）** | 外部 API 提供方 |
| **Organization（机构）** | 内部机构，与厂商多对多关系 |
| **Service（接口）** | API 端点，配置转换规则和钩子 |

### 请求流程

```
请求 → BeforeAuth → AfterAuth → BeforeRequestTransform
    → DSL 请求转换 → AfterRequestTransform → BeforeForward
    → 代理转发 → AfterForward → BeforeResponseTransform
    → DSL 响应转换 → AfterResponseTransform → 响应
    (出错时 → OnError)
```

## DSL 转换

使用三种语法类型进行声明式数据转换：

| 类型 | 语法 | 示例 |
|------|------|------|
| **字面量** | 直接值 | `"200"` |
| **JSONPath** | `$.` 前缀 | `"$.data.id"` |
| **Context** | `@ctx.` 前缀 | `"@ctx.request.method"` |

### 示例

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

## Hook 系统

使用 JavaScript 编写自定义逻辑，可访问内置函数：

### 内置模块

| 模块 | 函数 |
|------|------|
| **crypto** | md5, sha256, hmacSHA256, aesEncrypt/Decrypt, rsaEncrypt/Decrypt, rsaSign/Verify |
| **http** | get, post, postJSON, postForm, request |
| **encoding** | base64Encode/Decode, jsonEncode/Decode, urlEncode/Decode |
| **util** | uuid, now, formatTime, parseTime, sleep |

### Hook 示例

```javascript
// BeforeForward 钩子
function beforeForward(ctx) {
  // 添加自定义请求头
  ctx.request.headers["X-Custom-Token"] = crypto.md5(ctx.request.body);

  // 修改请求体
  var body = JSON.parse(ctx.request.body);
  body.timestamp = util.now();
  ctx.request.body = JSON.stringify(body);

  return ctx;
}
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

## 管理 API

所有管理端点需要 `X-Admin-Token` 请求头认证。

**基础路径：** `/admin/db`

| 资源 | 端点 |
|------|------|
| 厂商 | `GET/POST /vendors`, `GET/PUT/DELETE /vendor/{id}` |
| 机构 | `GET/POST /organizations`, `GET/PUT/DELETE /organization/{id}` |
| 接口 | `GET/POST /services`, `GET/PUT/DELETE /service/{id}` |
| Hook 脚本 | `GET/POST /hook-scripts`, `GET/PUT/DELETE /hook-script/{id}` |
| 公共脚本库 | `GET/POST /scripts`, `GET/PUT/DELETE /script/{id}` |
| 接口 Hook | `GET/POST /service-hooks`, `GET/PUT/DELETE /service-hook/{id}` |

## 项目结构

```
backend/
├── main.go           # 应用入口
├── handler/          # HTTP 处理器
├── hook/             # Hook 系统 + 内置函数
├── model/            # 数据模型
├── transform/        # DSL 转换引擎
├── proxy/            # HTTP 代理
├── database/         # 数据库操作
└── router/           # 路由注册

front/
├── src/
│   ├── views/        # Vue 页面
│   ├── components/   # Vue 组件
│   └── router/       # 前端路由
└── package.json
```

## 技术栈

- **Web 框架：** [atreugo](https://github.com/savsgio/atreugo)（基于 fasthttp）
- **ORM：** [GORM](https://gorm.io/)
- **JavaScript 引擎：** [goja](https://github.com/dop251/goja)
- **日志：** [zap](https://github.com/uber-go/zap)
- **前端：** Vue 3 + Element Plus

## 文档

- [管理 API 参考](./ADMIN_API.md)
- [DSL Context 参考](./DSL_CONTEXT_REFERENCE.md)
- [使用示例](./EXAMPLE.md)
- [并发安全](./CONCURRENCY_SAFETY.md)

## 贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 支持

- 📖 [文档](./docs)
- 🐛 [问题追踪](https://github.com/yourusername/gateway/issues)
- 💬 [讨论区](https://github.com/yourusername/gateway/discussions)

---

<div align="center">
Made with ❤️ by the Gateway Team
</div>
