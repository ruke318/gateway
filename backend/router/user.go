package router

import (
	"github.com/ruke318/gateway/auth"
	"github.com/ruke318/gateway/handler"
	"github.com/ruke318/gateway/middleware"
	"github.com/savsgio/atreugo/v11"
)

// RegisterUserRoutes 注册用户管理和操作日志相关路由
func RegisterUserRoutes(app *atreugo.Atreugo) {
	userHandler := handler.NewUserHandler()
	logHandler := handler.NewOperationLogHandler()

	// ==================== 公开路由（无需认证） ====================
	registerPublicRoutes(app, userHandler)

	// ==================== 需要认证的路由 ====================
	registerAuthenticatedRoutes(app, userHandler, logHandler)
}

// registerPublicRoutes 注册公开路由
func registerPublicRoutes(app *atreugo.Atreugo, userHandler *handler.UserHandler) {
	authGroup := app.NewGroupPath("/api/auth")

	// 用户登录
	authGroup.POST("/login", userHandler.Login)
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(app *atreugo.Atreugo, userHandler *handler.UserHandler, logHandler *handler.OperationLogHandler) {
	apiGroup := app.NewGroupPath("/api")
	apiGroup.UseBefore(auth.AuthMiddleware) // 所有 /api 路由都需要 JWT 认证

	// ==================== 用户个人相关 ====================
	// 获取当前用户信息
	apiGroup.GET("/auth/me", userHandler.GetCurrentUserInfo)
	// 修改当前用户密码
	apiGroup.POST("/auth/change-password", userHandler.ChangePassword)

	// ==================== 用户管理（管理员） ====================
	registerUserManagementRoutes(apiGroup, userHandler)

	// ==================== 操作日志（管理员） ====================
	registerOperationLogRoutes(apiGroup, logHandler)
}

// registerUserManagementRoutes 注册用户管理路由（管理员权限）
func registerUserManagementRoutes(apiGroup *atreugo.Router, userHandler *handler.UserHandler) {
	usersGroup := apiGroup.NewGroupPath("/users")
	usersGroup.UseBefore(auth.AdminMiddleware)                // 需要管理员权限
	usersGroup.UseBefore(middleware.OperationLogMiddleware)   // 记录操作日志

	usersGroup.GET("", userHandler.ListUsers)                 // GET /api/users - 获取用户列表
	usersGroup.POST("", userHandler.CreateUser)               // POST /api/users - 创建用户
	usersGroup.PUT("/{id}", userHandler.UpdateUser)           // PUT /api/users/:id - 更新用户
	usersGroup.DELETE("/{id}", userHandler.DeleteUser)        // DELETE /api/users/:id - 删除用户
	usersGroup.POST("/{id}/reset-password", userHandler.ResetPassword) // POST /api/users/:id/reset-password - 重置密码
}

// registerOperationLogRoutes 注册操作日志路由（管理员权限）
func registerOperationLogRoutes(apiGroup *atreugo.Router, logHandler *handler.OperationLogHandler) {
	logsGroup := apiGroup.NewGroupPath("/operation-logs")
	logsGroup.UseBefore(auth.AdminMiddleware) // 需要管理员权限

	logsGroup.GET("", logHandler.ListOperationLogs)           // GET /api/operation-logs - 获取日志列表
	logsGroup.GET("/{id}", logHandler.GetOperationLog)        // GET /api/operation-logs/:id - 获取日志详情
	logsGroup.GET("/statistics", logHandler.GetOperationStatistics) // GET /api/operation-logs/statistics - 获取统计信息
}
