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
	"github.com/ruke318/gateway/router"
	"github.com/ruke318/gateway/transform"
	"github.com/savsgio/atreugo/v11"
)

func main() {
	cfg := config.Load()

	// 初始化日志
	if err := logger.Init(cfg.Log.Path, cfg.Log.Level); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync()

	// 初始化数据库
	dbConfig := &database.Config{
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Database:     cfg.DB.Database,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MaxLifetime:  time.Hour,
		Debug:        cfg.DB.Debug,
	}
	if err := database.Init(dbConfig); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 自动迁移表结构
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 加载公共函数库
	if err := hook.LoadLibrary(); err != nil {
		log.Printf("Warning: Failed to load script library: %v", err)
	}

	// 初始化 JS VM 池
	hook.InitVMPool(cfg.VMPool.Size)

	forwarder := proxy.NewForwarder(cfg.BackendURL)
	dslTransformer := transform.NewDSLTransformer()

	// 创建 Handler
	invokeHandler := handler.NewInvokeHandler(forwarder, dslTransformer)
	adminDBHandler := handler.NewAdminDBHandler(cfg.AdminToken)

	// 创建 atreugo 实例
	app := atreugo.New(atreugo.Config{
		Addr: cfg.Port,
	})

	// 注册路由
	router.RegisterAdminDBRoutes(app, adminDBHandler)
	router.RegisterOutRoutes(app, invokeHandler)

	log.Printf("Gateway starting on %s", cfg.Port)
	if err := app.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
