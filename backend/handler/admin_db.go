package handler

import (
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/model"
	"github.com/ruke318/gateway/util"
	"github.com/savsgio/atreugo/v11"
)

// AdminDBHandler 数据库管理接口
type AdminDBHandler struct {
	adminToken string
}

func NewAdminDBHandler(adminToken string) *AdminDBHandler {
	return &AdminDBHandler{adminToken: adminToken}
}

// AuthMiddleware 认证中间件
func (h *AdminDBHandler) AuthMiddleware(ctx *atreugo.RequestCtx) error {
	token := string(ctx.Request.Header.Peek("X-Admin-Token"))
	if token != h.adminToken {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "unauthorized",
		}, 401)
	}
	return ctx.Next()
}

// ============ 厂商 ============

func (h *AdminDBHandler) ListVendors(ctx *atreugo.RequestCtx) error {
	var q model.VendorQuery
	util.BindQuery(ctx, &q)

	var vendors []model.Vendor
	query := database.DB.Model(&model.Vendor{})
	if q.Code != "" {
		query = query.Where("code = ?", q.Code)
	}
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if err := query.Find(&vendors).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, vendors)
}

func (h *AdminDBHandler) GetVendor(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var vendor model.Vendor
	if err := database.DB.First(&vendor, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "vendor not found")
	}
	return h.successJSON(ctx, vendor)
}

func (h *AdminDBHandler) CreateVendor(ctx *atreugo.RequestCtx) error {
	var vendor model.Vendor
	if err := util.BindJSON(ctx, &vendor); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&vendor).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, vendor)
}

func (h *AdminDBHandler) UpdateVendor(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var vendor model.Vendor
	if err := util.BindJSON(ctx, &vendor); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	vendor.ID = id
	if err := database.DB.Save(&vendor).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, vendor)
}

func (h *AdminDBHandler) DeleteVendor(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.Vendor{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 机构 ============

func (h *AdminDBHandler) ListOrganizations(ctx *atreugo.RequestCtx) error {
	var q model.OrganizationQuery
	util.BindQuery(ctx, &q)

	var orgs []model.Organization
	query := database.DB.Model(&model.Organization{})
	if q.Code != "" {
		query = query.Where("code = ?", q.Code)
	}
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if err := query.Find(&orgs).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, orgs)
}

func (h *AdminDBHandler) GetOrganization(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var org model.Organization
	if err := database.DB.First(&org, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "organization not found")
	}
	return h.successJSON(ctx, org)
}

func (h *AdminDBHandler) CreateOrganization(ctx *atreugo.RequestCtx) error {
	var org model.Organization
	if err := util.BindJSON(ctx, &org); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&org).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, org)
}

func (h *AdminDBHandler) UpdateOrganization(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var org model.Organization
	if err := util.BindJSON(ctx, &org); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	org.ID = id
	if err := database.DB.Save(&org).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, org)
}

func (h *AdminDBHandler) DeleteOrganization(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.Organization{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 接口 ============

func (h *AdminDBHandler) ListServices(ctx *atreugo.RequestCtx) error {
	var q model.ServiceQuery
	util.BindQuery(ctx, &q)

	var services []model.Service
	query := database.DB.Model(&model.Service{}).Preload("Vendor").Preload("Organization")
	if q.ServiceID != "" {
		query = query.Where("service_id LIKE ?", "%"+q.ServiceID+"%")
	}
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.VendorID > 0 {
		query = query.Where("vendor_id = ?", q.VendorID)
	}
	if q.OrgID > 0 {
		query = query.Where("org_id = ?", q.OrgID)
	}
	if err := query.Find(&services).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, services)
}

func (h *AdminDBHandler) GetService(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var svc model.Service
	if err := database.DB.Preload("Vendor").Preload("Organization").Preload("Hooks").First(&svc, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "service not found")
	}
	return h.successJSON(ctx, svc)
}

func (h *AdminDBHandler) CreateService(ctx *atreugo.RequestCtx) error {
	var svc model.Service
	if err := util.BindJSON(ctx, &svc); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&svc).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, svc)
}

func (h *AdminDBHandler) UpdateService(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var svc model.Service
	if err := util.BindJSON(ctx, &svc); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	svc.ID = id
	if err := database.DB.Save(&svc).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, svc)
}

func (h *AdminDBHandler) DeleteService(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.Service{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 公共函数库 ============

func (h *AdminDBHandler) ListScripts(ctx *atreugo.RequestCtx) error {
	var q model.ScriptQuery
	util.BindQuery(ctx, &q)

	var scripts []model.ScriptLibrary
	query := database.DB.Model(&model.ScriptLibrary{})
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Description != "" {
		query = query.Where("description LIKE ?", "%"+q.Description+"%")
	}
	if err := query.Find(&scripts).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, scripts)
}

func (h *AdminDBHandler) GetScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var script model.ScriptLibrary
	if err := database.DB.First(&script, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "script not found")
	}
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) CreateScript(ctx *atreugo.RequestCtx) error {
	var script model.ScriptLibrary
	if err := util.BindJSON(ctx, &script); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&script).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	hook.ReloadLibrary()
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) UpdateScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var script model.ScriptLibrary
	if err := util.BindJSON(ctx, &script); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	script.ID = id
	if err := database.DB.Save(&script).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	hook.ReloadLibrary()
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) DeleteScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.ScriptLibrary{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	hook.ReloadLibrary()
	return h.successJSON(ctx, nil)
}

// ============ Hook 脚本 ============

func (h *AdminDBHandler) ListHookScripts(ctx *atreugo.RequestCtx) error {
	var q model.HookScriptQuery
	util.BindQuery(ctx, &q)

	var scripts []model.HookScript
	query := database.DB.Model(&model.HookScript{})
	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.HookPoint != "" {
		query = query.Where("hook_point = ?", q.HookPoint)
	}
	if err := query.Find(&scripts).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, scripts)
}

func (h *AdminDBHandler) GetHookScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var script model.HookScript
	if err := database.DB.First(&script, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "hook script not found")
	}
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) CreateHookScript(ctx *atreugo.RequestCtx) error {
	var script model.HookScript
	if err := util.BindJSON(ctx, &script); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&script).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) UpdateHookScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var script model.HookScript
	if err := util.BindJSON(ctx, &script); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	script.ID = id
	if err := database.DB.Save(&script).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, script)
}

func (h *AdminDBHandler) DeleteHookScript(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.HookScript{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 接口 Hook 关联 ============

func (h *AdminDBHandler) ListServiceHooks(ctx *atreugo.RequestCtx) error {
	serviceID := util.GetUint64(ctx, "service_id")
	var hooks []model.ServiceHook
	query := database.DB.Preload("Script")
	if serviceID > 0 {
		query = query.Where("service_pk = ?", serviceID)
	}
	if err := query.Find(&hooks).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, hooks)
}

func (h *AdminDBHandler) GetServiceHook(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var sh model.ServiceHook
	if err := database.DB.Preload("Script").First(&sh, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "service hook not found")
	}
	return h.successJSON(ctx, sh)
}

func (h *AdminDBHandler) CreateServiceHook(ctx *atreugo.RequestCtx) error {
	var sh model.ServiceHook
	if err := util.BindJSON(ctx, &sh); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&sh).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, sh)
}

func (h *AdminDBHandler) UpdateServiceHook(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var input model.ServiceHook
	if err := util.BindJSON(ctx, &input); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}

	// 构建更新字段（只更新非零值）
	updates := make(map[string]interface{})
	if input.HookPoint != "" {
		updates["hook_point"] = input.HookPoint
	}
	if input.ScriptID != nil {
		updates["script_id"] = input.ScriptID
	}
	// priority 和 status 可能为 0，需要判断是否真的要更新
	// 简单判断：如果有传其他字段，说明是完整更新；否则是单独更新 status
	if len(updates) > 0 {
		// 完整更新，包括 priority
		updates["priority"] = input.Priority
	}
	// status 总是更新（因为可能是切换操作）
	updates["status"] = input.Status

	if err := database.DB.Model(&model.ServiceHook{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}

	// 返回更新后的数据
	var sh model.ServiceHook
	database.DB.Preload("Script").First(&sh, id)
	return h.successJSON(ctx, sh)
}

func (h *AdminDBHandler) DeleteServiceHook(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.ServiceHook{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 重载函数库 ============

func (h *AdminDBHandler) ReloadLibrary(ctx *atreugo.RequestCtx) error {
	if err := hook.ReloadLibrary(); err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, "library reloaded")
}

// ============ 工具方法 ============

func (h *AdminDBHandler) successJSON(ctx *atreugo.RequestCtx, data interface{}) error {
	return ctx.JSONResponse(map[string]interface{}{
		"code": 0,
		"data": data,
	})
}

func (h *AdminDBHandler) errorJSON(ctx *atreugo.RequestCtx, code int, message string) error {
	return ctx.JSONResponse(map[string]interface{}{
		"code":    code,
		"message": message,
	}, code)
}
