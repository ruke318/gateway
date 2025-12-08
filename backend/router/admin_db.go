package router

import (
	"github.com/ruke318/gateway/handler"
	"github.com/savsgio/atreugo/v11"
)

// RegisterAdminDBRoutes 注册数据库管理路由
func RegisterAdminDBRoutes(app *atreugo.Atreugo, h *handler.AdminDBHandler) {
	// 管理接口分组，带认证中间件
	admin := app.NewGroupPath("/admin/db")
	admin.UseBefore(h.AuthMiddleware)

	// 厂商
	admin.GET("/vendors", h.ListVendors)
	admin.GET("/vendor/{id}", h.GetVendor)
	admin.POST("/vendor", h.CreateVendor)
	admin.PUT("/vendor/{id}", h.UpdateVendor)
	admin.DELETE("/vendor/{id}", h.DeleteVendor)

	// 机构
	admin.GET("/organizations", h.ListOrganizations)
	admin.GET("/organization/{id}", h.GetOrganization)
	admin.POST("/organization", h.CreateOrganization)
	admin.PUT("/organization/{id}", h.UpdateOrganization)
	admin.DELETE("/organization/{id}", h.DeleteOrganization)

	// 接口
	admin.GET("/services", h.ListServices)
	admin.GET("/service/{id}", h.GetService)
	admin.POST("/service", h.CreateService)
	admin.PUT("/service/{id}", h.UpdateService)
	admin.DELETE("/service/{id}", h.DeleteService)

	// 公共函数库
	admin.GET("/scripts", h.ListScripts)
	admin.GET("/script/{id}", h.GetScript)
	admin.POST("/script", h.CreateScript)
	admin.PUT("/script/{id}", h.UpdateScript)
	admin.DELETE("/script/{id}", h.DeleteScript)

	// Hook 脚本
	admin.GET("/hook-scripts", h.ListHookScripts)
	admin.GET("/hook-script/{id}", h.GetHookScript)
	admin.POST("/hook-script", h.CreateHookScript)
	admin.PUT("/hook-script/{id}", h.UpdateHookScript)
	admin.DELETE("/hook-script/{id}", h.DeleteHookScript)

	// 接口 Hook 关联
	admin.GET("/service-hooks", h.ListServiceHooks)
	admin.GET("/service-hook/{id}", h.GetServiceHook)
	admin.POST("/service-hook", h.CreateServiceHook)
	admin.PUT("/service-hook/{id}", h.UpdateServiceHook)
	admin.DELETE("/service-hook/{id}", h.DeleteServiceHook)

	// 重载函数库
	admin.POST("/reload-library", h.ReloadLibrary)
}
