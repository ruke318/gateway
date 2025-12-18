// Package main 是 Gateway API 网关的主入口
// Gateway 是一个轻量级的 API 网关，用于统一对接外部厂商接口
// 主要功能：协议转换、DSL 映射、Hook 扩展、多租户管理
package main

import (
	"log"
	"time"

	"github.com/ruke318/gateway/config"
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/handler"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/logger"
	"github.com/ruke318/gateway/proxy"
	"github.com/ruke318/gateway/redis"
	"github.com/ruke318/gateway/router"
	"github.com/ruke318/gateway/transform"
	"github.com/savsgio/atreugo/v11"
)

// main 是程序的主入口函数
// 初始化流程：
// 1. 加载配置（config.yaml 或环境变量）
// 2. 初始化日志系统（结构化日志）
// 3. 初始化数据库连接池（MySQL + GORM）
// 4. 自动迁移数据库表结构
// 5. 加载全局 JavaScript 脚本库
// 6. 初始化 JavaScript VM 池（用于 Hook 脚本并发执行）
// 7. 创建 HTTP 代理转发器
// 8. 创建 DSL 转换引擎
// 9. 创建请求处理器（invoke 和 admin）
// 10. 注册路由并启动 HTTP 服务
func main() {
	// 加载配置：优先级 环境变量 > config.yaml > 默认值
	cfg := config.Load()

	// 初始化日志系统（zap 结构化日志）
	// 日志将同时输出到文件和控制台
	if err := logger.Init(cfg.Log.Path, cfg.Log.Level); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync() // 程序退出时刷新日志缓冲区

	// 初始化数据库连接池
	// 使用 GORM + MySQL，支持连接池配置
	dbConfig := &database.Config{
		Host:         cfg.DB.Host,         // 数据库主机地址
		Port:         cfg.DB.Port,         // 数据库端口
		User:         cfg.DB.User,         // 数据库用户名
		Password:     cfg.DB.Password,     // 数据库密码
		Database:     cfg.DB.Database,     // 数据库名称
		MaxIdleConns: cfg.DB.MaxIdleConns, // 最大空闲连接数
		MaxOpenConns: cfg.DB.MaxOpenConns, // 最大打开连接数
		MaxLifetime:  time.Hour,           // 连接最大生命周期（1小时）
		Debug:        cfg.DB.Debug,        // 是否开启 SQL 调试日志
	}
	if err := database.Init(dbConfig); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 自动迁移表结构
	// 根据 model/ 中的实体定义，自动创建或更新数据库表
	// 包括：users, operation_logs, vendors, organizations, services 等
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化默认数据（管理员账号等）
	if err := database.InitDefaultData(); err != nil {
		log.Fatalf("Failed to init default data: %v", err)
	}

	// 初始化 Redis 连接（可选，如果未配置则跳过）
	// Redis 用于缓存 Token、配置数据等
	if cfg.Redis.Addr != "" {
		if err := redis.Init(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.PoolSize); err != nil {
			log.Printf("Warning: Failed to init redis: %v", err)
		} else {
			log.Printf("Redis connected: %s", cfg.Redis.Addr)
		}
	}

	// 加载全局 JavaScript 函数库
	// 从数据库 script_libraries 表加载全局共享的 JS 函数
	// 这些函数可以在所有 Hook 脚本中直接调用
	if err := hook.LoadLibrary(); err != nil {
		log.Printf("Warning: Failed to load script library: %v", err)
	}

	// 加载字典配置
	// 从数据库 dictionary_config 表加载所有字典配置到内存
	// 字典用于机构内转换和跨机构字段映射
	if err := hook.LoadDictionary(); err != nil {
		log.Printf("Warning: Failed to load dictionary: %v", err)
	}

	// 初始化 JavaScript VM 池
	// 预创建指定数量的 Goja VM 实例，用于并发执行 Hook 脚本
	// VM 池避免了频繁创建销毁 VM 的开销，提高并发性能
	hook.InitVMPool(cfg.VMPool.Size)

	// 创建 HTTP 代理转发器
	// 按厂商（host）管理连接池和熔断器
	forwarder := proxy.NewForwarder(&cfg.HTTPPool, &cfg.CircuitBreaker)

	// 创建 DSL 转换引擎
	// 用于执行声明式的字段映射（JSONPath 查询和 Context 注入）
	dslTransformer := transform.NewDSLTransformer()

	// 创建统一调用处理器
	// 负责处理 /gateway/v1/invoke 接口的请求
	// 流程：解析请求 → 加载配置 → Hook(OnAuth) → DSL转换 → Hook(BeforeForward) → 转发 → Hook(AfterForward) → 返回
	invokeHandler := handler.NewInvokeHandler(forwarder, dslTransformer)

	// 创建管理后台处理器
	// 负责处理 /admin/db/* 接口的 CRUD 操作
	// 包括：厂商、机构、接口、Hook脚本、函数库、字典配置的增删改查
	// 使用 JWT 认证 + 管理员权限验证
	adminDBHandler := handler.NewAdminDBHandler()

	// 创建 atreugo Web 框架实例
	// atreugo 是基于 FastHTTP 的高性能 HTTP 框架
	app := atreugo.New(atreugo.Config{
		Addr: cfg.Port, // 监听地址，格式：":8080"
	})

	// 注册管理后台路由（需要 X-Admin-Token 认证）
	// 路由前缀：/admin/db/
	router.RegisterAdminDBRoutes(app, adminDBHandler)

	// 注册用户管理和操作日志路由（需要 JWT 认证）
	// 路由前缀：/api/
	router.RegisterUserRoutes(app)

	// 注册统一调用路由（公开接口）
	// 路由：POST /gateway/v1/invoke
	router.RegisterOutRoutes(app, invokeHandler)

	// 启动 HTTP 服务
	log.Printf("Gateway starting on %s", cfg.Port)
	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
