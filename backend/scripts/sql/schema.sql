-- Gateway 管理后台数据库表结构
-- 数据库: MySQL 8.0+
-- 字符集: utf8mb4

-- ==========================================
-- 1. 厂商表 (vendor)
-- ==========================================
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

-- ==========================================
-- 2. 机构表 (organization)
-- ==========================================
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

-- ==========================================
-- 3. 公共函数库表 (script_library)
-- ==========================================
CREATE TABLE `script_library` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` VARCHAR(128) NOT NULL COMMENT '函数名称',
  `namespace` VARCHAR(64) NOT NULL DEFAULT 'global' COMMENT '命名空间，便于分类',
  `script_content` TEXT NOT NULL COMMENT '函数代码',
  `description` TEXT COMMENT '函数说明',
  `example` TEXT COMMENT '使用示例',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_namespace_name` (`namespace`, `name`),
  KEY `idx_namespace` (`namespace`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公共函数库表;'

-- ==========================================
-- 4. 驱动脚本表 (hook_script)
-- ==========================================
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

-- ==========================================
-- 5. 接口表 (service)
-- ==========================================
CREATE TABLE `service` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `service_id` VARCHAR(64) NOT NULL COMMENT '接口标识',
  `org_id` BIGINT UNSIGNED NOT NULL COMMENT '机构ID',
  `vendor_id` BIGINT UNSIGNED NOT NULL COMMENT '厂商ID',
  `name` VARCHAR(128) NOT NULL COMMENT '接口名称',
  `description` TEXT COMMENT '描述',
  `backend_url` VARCHAR(512) DEFAULT NULL COMMENT '后端URL，可覆盖厂商配置',
  `backend_path` VARCHAR(512) DEFAULT NULL COMMENT '后端路径',
  `backend_method` VARCHAR(16) NOT NULL DEFAULT 'POST' COMMENT '请求方法',
  `body_type` VARCHAR(16) NOT NULL DEFAULT 'json' COMMENT '请求体类型：json/form',
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

-- ==========================================
-- 6. 接口 Hook 关联表 (service_hook)
-- ==========================================
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
