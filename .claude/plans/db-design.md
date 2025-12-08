# Gateway 管理后台表结构设计

## 设计理念

### 1. 核心实体关系

```
┌─────────┐       ┌─────────────┐       ┌─────────┐
│  厂商   │◄──────│    接口     │──────►│  机构   │
│ vendor  │  N:1  │   service   │  N:1  │   org   │
└─────────┘       └──────┬──────┘       └─────────┘
                         │
                         │ 1:N
                         ▼
                  ┌─────────────┐       ┌─────────────┐
                  │ 接口Hook关联 │──────►│  驱动脚本   │
                  │service_hook │  N:1  │ hook_script │
                  └─────────────┘       └─────────────┘
```

### 2. 设计原则

- **职责分离**: 厂商、机构、接口、脚本各司其职
- **隐式关联**: 机构和厂商通过接口表隐式关联，简单直接
- **灵活复用**: 脚本库支持复用，也支持内联脚本
- **配置覆盖**: 接口可覆盖厂商默认配置
- **唯一约束**: service_id 在机构内唯一

### 3. 调用流程

```
POST /gateway/v1/invoke
{
    "com_id": "vendor-001",     // 厂商编码
    "service_id": "query_user", // 接口标识
    "unit_id": "org-001",       // 机构编码
    "req": {...}
}

网关处理:
1. 根据 unit_id + service_id 查找接口配置
2. 验证 com_id 与接口关联的厂商匹配
3. 加载 DSL 转换配置和 Hook 脚本
4. 执行请求处理流程
```

---

## 表总览

| 序号 | 表名 | 说明 |
|-----|------|------|
| 1 | vendor | 厂商表 |
| 2 | organization | 机构表（含 config 配置） |
| 3 | script_library | 公共函数库表 |
| 4 | hook_script | 驱动脚本表 |
| 5 | service | 接口表 |
| 6 | service_hook | 接口 Hook 关联表 |

**hook_point 可选值**:
- BeforeAuth / AfterAuth
- BeforeRequestTransform / AfterRequestTransform
- BeforeForward / AfterForward
- BeforeResponseTransform / AfterResponseTransform
- OnError

---

## 表结构设计 (MySQL)

### 1. 厂商表 (vendor)

```sql
CREATE TABLE `vendor` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `code` VARCHAR(64) NOT NULL COMMENT '厂商编码',
  `name` VARCHAR(128) NOT NULL COMMENT '厂商名称',
  `base_url` VARCHAR(512) DEFAULT NULL COMMENT '基础URL',
  `description` TEXT COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='厂商表';
```

### 2. 机构表 (organization)

```sql
CREATE TABLE `organization` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `code` VARCHAR(64) NOT NULL COMMENT '机构编码',
  `name` VARCHAR(128) NOT NULL COMMENT '机构名称',
  `config` JSON COMMENT '机构配置，如认证信息、默认参数等',
  `description` TEXT COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机构表';
```

**config 字段示例**:
```json
{
  "auth": {
    "type": "api_key",
    "api_key": "xxx"
  },
  "default_headers": {
    "X-Org-ID": "org-001"
  },
  "timeout": 30000
}
```

### 3. 公共函数库表 (script_library)

```sql
CREATE TABLE `script_library` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` VARCHAR(128) NOT NULL COMMENT '函数名称',
  `namespace` VARCHAR(64) DEFAULT 'global' COMMENT '命名空间，便于分类',
  `script_content` TEXT NOT NULL COMMENT '函数代码',
  `description` TEXT COMMENT '函数说明',
  `example` TEXT COMMENT '使用示例',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_namespace_name` (`namespace`, `name`),
  KEY `idx_namespace` (`namespace`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公共函数库表';
```

**script_content 示例**:
```javascript
// 函数名: formatDate
// 命名空间: utils
function formatDate(timestamp, format) {
  var date = new Date(timestamp);
  // ... 格式化逻辑
  return formatted;
}
```

**使用方式**: 在 Hook 脚本中通过 `lib.utils.formatDate()` 调用

### 4. 驱动脚本表 (hook_script)

```sql
CREATE TABLE `hook_script` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` VARCHAR(128) NOT NULL COMMENT '脚本名称',
  `hook_point` VARCHAR(32) NOT NULL COMMENT 'Hook节点类型',
  `script_content` TEXT NOT NULL COMMENT '脚本内容',
  `description` TEXT COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_hook_point` (`hook_point`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='驱动脚本表';
```

### 5. 接口表 (service)

```sql
CREATE TABLE `service` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `service_id` VARCHAR(64) NOT NULL COMMENT '接口标识',
  `org_id` BIGINT UNSIGNED NOT NULL COMMENT '机构ID',
  `vendor_id` BIGINT UNSIGNED NOT NULL COMMENT '厂商ID',
  `name` VARCHAR(128) NOT NULL COMMENT '接口名称',
  `description` TEXT COMMENT '描述',
  `backend_url` VARCHAR(512) DEFAULT NULL COMMENT '后端URL，可覆盖厂商配置',
  `backend_path` VARCHAR(512) DEFAULT NULL COMMENT '后端路径',
  `backend_method` VARCHAR(16) DEFAULT 'POST' COMMENT '请求方法',
  `request_transform` JSON COMMENT '请求DSL转换配置',
  `response_transform` JSON COMMENT '响应DSL转换配置',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_service` (`org_id`, `service_id`),
  KEY `idx_vendor_id` (`vendor_id`),
  KEY `idx_org_id` (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='接口表';
```

### 6. 接口 Hook 关联表 (service_hook)

```sql
CREATE TABLE `service_hook` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `service_id` BIGINT UNSIGNED NOT NULL COMMENT '接口ID',
  `hook_point` VARCHAR(32) NOT NULL COMMENT 'Hook节点类型',
  `script_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '脚本ID，与inline_script二选一',
  `inline_script` TEXT COMMENT '内联脚本，与script_id二选一',
  `priority` INT NOT NULL DEFAULT 0 COMMENT '优先级，数字越小越先执行',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_service_hook` (`service_id`, `hook_point`, `priority`),
  KEY `idx_script_id` (`script_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='接口Hook关联表';
```

---

## 字段说明

### 接口表核心字段

| 字段 | 说明 |
|-----|------|
| service_id | 接口标识，在机构内唯一 |
| org_id | 关联机构 |
| vendor_id | 关联厂商 |
| backend_url | 覆盖厂商的 base_url |
| request_transform | JSON 格式的请求 DSL 配置 |
| response_transform | JSON 格式的响应 DSL 配置 |

### 脚本复用策略

| 模式 | 字段 | 适用场景 |
|-----|------|---------|
| 引用脚本 | script_id | 多接口共用相同脚本 |
| 内联脚本 | inline_script | 接口特定的定制逻辑 |

---

## 查询示例

### 根据调用参数查找接口配置

```sql
SELECT s.*, v.base_url as vendor_base_url, v.name as vendor_name
FROM service s
JOIN organization o ON s.org_id = o.id
JOIN vendor v ON s.vendor_id = v.id
WHERE o.code = 'org-001'
  AND s.service_id = 'query_user'
  AND v.code = 'vendor-001'
  AND s.status = 1;
```

### 查询接口的所有 Hook 脚本

```sql
SELECT sh.hook_point, sh.priority,
       COALESCE(hs.script_content, sh.inline_script) as script_content
FROM service_hook sh
LEFT JOIN hook_script hs ON sh.script_id = hs.id
WHERE sh.service_id = ?
  AND sh.status = 1
ORDER BY sh.hook_point, sh.priority;
```
