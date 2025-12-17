package handler

import (
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/model"
	"github.com/ruke318/gateway/util"
	"github.com/savsgio/atreugo/v11"
)

// AdminDBHandler 数据库管理接口处理器
// 负责处理所有管理后台的 CRUD 操作
// 路由前缀：/admin/db/
// 所有接口都需要通过 JWT 认证且需要管理员权限
// 包含以下模块：
// - 厂商管理（Vendor）
// - 机构管理（Organization）
// - 接口管理（Service）
// - Hook 脚本管理（HookScript）
// - 公共函数库管理（ScriptLibrary）
// - 字典配置管理（DictionaryConfig）
// - 接口 Hook 关联管理（ServiceHook）
type AdminDBHandler struct{}

// NewAdminDBHandler 创建 AdminDBHandler 实例
func NewAdminDBHandler() *AdminDBHandler {
	return &AdminDBHandler{}
}

// ============ 厂商管理 ============
// 厂商（Vendor）代表外部接口提供方
// 例如：支付宝、微信支付、银联等
// 字段：code（编码）、name（名称）、base_url（基础URL）、description（描述）

// ListVendors 查询厂商列表
// 支持按 code（精确匹配）和 name（模糊匹配）过滤
// GET /admin/db/vendors?code=xxx&name=xxx
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

// GetVendor 根据ID获取单个厂商详情
// GET /admin/db/vendors/:id
func (h *AdminDBHandler) GetVendor(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var vendor model.Vendor
	if err := database.DB.First(&vendor, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "vendor not found")
	}
	return h.successJSON(ctx, vendor)
}

// CreateVendor 创建新厂商
// POST /admin/db/vendors
// Body: {"code": "alipay", "name": "支付宝", "base_url": "https://api.alipay.com"}
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

// UpdateVendor 更新厂商信息
// PUT /admin/db/vendors/:id
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

// DeleteVendor 删除厂商
// DELETE /admin/db/vendors/:id
func (h *AdminDBHandler) DeleteVendor(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.Vendor{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, nil)
}

// ============ 机构管理 ============
// 机构（Organization）代表内部使用方
// 例如：总部、分公司A、分公司B等
// 字段：code（编码）、name（名称）、config（配置JSON，存储appId、secret等敏感信息）
//
// 提供标准 CRUD 接口：
// - ListOrganizations: GET /admin/db/organizations
// - GetOrganization: GET /admin/db/organizations/:id
// - CreateOrganization: POST /admin/db/organizations
// - UpdateOrganization: PUT /admin/db/organizations/:id
// - DeleteOrganization: DELETE /admin/db/organizations/:id

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

// ============ 接口管理 ============
// 接口（Service）代表具体的 API 接口配置和转换规则
// 三层架构的核心：Vendor（厂商） + Organization（机构） + Service（接口）
// 字段：
// - service_id: 接口标识
// - name: 接口名称
// - vendor_id: 关联厂商ID
// - org_id: 关联机构ID
// - backend_path: 厂商后端路径
// - backend_method: HTTP方法（GET/POST）
// - body_type: 请求体类型（json/form/xml）
// - request_transform: 请求 DSL 映射（JSON）
// - response_transform: 响应 DSL 映射（JSON）
//
// 提供标准 CRUD 接口：
// - ListServices: GET /admin/db/services（支持按 service_id, name, vendor_id, org_id 过滤）
// - GetService: GET /admin/db/services/:id（包含关联的 Vendor、Organization、Hooks）
// - CreateService: POST /admin/db/services
// - UpdateService: PUT /admin/db/services/:id
// - DeleteService: DELETE /admin/db/services/:id

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

// ============ 公共函数库管理 ============
// 公共函数库（ScriptLibrary）存储全局共享的 JavaScript 函数
// 这些函数在所有 Hook 脚本中都可以直接调用
// 例如：通用的签名算法、数据处理函数、工具函数等
//
// 字段：
// - name: 函数名称
// - code: JavaScript 代码
// - description: 描述
//
// 特性：
// - 创建/更新/删除后会自动调用 hook.ReloadLibrary() 重新加载
// - 无需重启服务即可生效
//
// 提供标准 CRUD 接口：
// - ListScripts: GET /admin/db/scripts
// - GetScript: GET /admin/db/scripts/:id
// - CreateScript: POST /admin/db/scripts（自动重载）
// - UpdateScript: PUT /admin/db/scripts/:id（自动重载）
// - DeleteScript: DELETE /admin/db/scripts/:id（自动重载）

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

// ============ Hook 脚本管理 ============
// Hook 脚本（HookScript）是可复用的 JavaScript 脚本
// 可以关联到多个接口的不同执行点
//
// 字段：
// - name: 脚本名称
// - code: JavaScript 代码
// - hook_point: Hook 执行点（BeforeAuth, AfterAuth, BeforeForward 等）
// - description: 描述
//
// Hook 执行点（model.Hook*常量）：
// - BeforeAuth: 认证前处理
// - AfterAuth: 认证后处理
// - BeforeRequestTransform: 请求转换前
// - AfterRequestTransform: 请求转换后
// - BeforeForward: 转发前（常用于添加签名、Token）
// - AfterForward: 转发后（常用于解密、数据清洗）
// - BeforeResponseTransform: 响应转换前
// - AfterResponseTransform: 响应转换后
// - OnError: 错误处理
//
// 提供标准 CRUD 接口：
// - ListHookScripts: GET /admin/db/hook-scripts
// - GetHookScript: GET /admin/db/hook-scripts/:id
// - CreateHookScript: POST /admin/db/hook-scripts
// - UpdateHookScript: PUT /admin/db/hook-scripts/:id
// - DeleteHookScript: DELETE /admin/db/hook-scripts/:id

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

// ============ 接口 Hook 关联管理 ============
// 接口 Hook 关联（ServiceHook）用于将 Hook 脚本绑定到具体的接口和执行点
// 一个接口可以关联多个 Hook 脚本，按 priority 顺序执行
//
// 字段：
// - service_pk: 接口主键ID（关联 Service.ID）
// - script_id: Hook 脚本ID（关联 HookScript.ID）
// - hook_point: Hook 执行点
// - priority: 执行优先级（数字越小越先执行）
// - status: 状态（1=启用，0=禁用）
//
// 典型用例：
// 1. 接口 A 在 BeforeForward 执行点关联 2 个脚本：
//    - priority=1: 获取 Access Token
//    - priority=2: 添加 MD5 签名
// 2. 接口 B 在 AfterForward 执行点关联 1 个脚本：
//    - priority=1: 响应解密
//
// 提供标准 CRUD 接口：
// - ListServiceHooks: GET /admin/db/service-hooks?service_id=xxx
// - GetServiceHook: GET /admin/db/service-hooks/:id
// - CreateServiceHook: POST /admin/db/service-hooks
// - UpdateServiceHook: PUT /admin/db/service-hooks/:id（支持部分字段更新）
// - DeleteServiceHook: DELETE /admin/db/service-hooks/:id

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
// 手动触发重新加载全局 JavaScript 函数库
// 通常在批量修改 ScriptLibrary 后调用，无需重启服务

// ReloadLibrary 重新加载全局函数库
// POST /admin/db/reload-library
// 调用 hook.ReloadLibrary() 从数据库重新加载所有 ScriptLibrary
func (h *AdminDBHandler) ReloadLibrary(ctx *atreugo.RequestCtx) error {
	if err := hook.ReloadLibrary(); err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, "library reloaded")
}

// ============ 字典配置管理 ============
// 字典配置用于存储机构级别的字段映射关系
// 支持机构内转换和跨机构转换
// 字段：org_id（机构ID）、dict_type（字典类型）、dict_key（标准键）、dict_value（机构特定值）

// ListDictionaryConfigs 查询字典配置列表
// 支持按 org_id 和 dict_type 过滤
// GET /admin/db/dictionary-configs?org_id=xxx&dict_type=xxx
func (h *AdminDBHandler) ListDictionaryConfigs(ctx *atreugo.RequestCtx) error {
	orgID := string(ctx.QueryArgs().Peek("org_id"))
	dictType := string(ctx.QueryArgs().Peek("dict_type"))

	var configs []model.DictionaryConfig
	query := database.DB.Model(&model.DictionaryConfig{})

	if orgID != "" {
		query = query.Where("org_id = ?", orgID)
	}
	if dictType != "" {
		query = query.Where("dict_type = ?", dictType)
	}

	if err := query.Order("org_id, dict_type, dict_key").Find(&configs).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, configs)
}

// GetDictionaryConfig 根据ID获取单个字典配置详情
// GET /admin/db/dictionary-configs/:id
func (h *AdminDBHandler) GetDictionaryConfig(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var config model.DictionaryConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "dictionary config not found")
	}
	return h.successJSON(ctx, config)
}

// CreateDictionaryConfig 创建新字典配置
// POST /admin/db/dictionary-configs
// Body: {"org_id": "org001", "dict_type": "payment_method", "dict_key": "ALIPAY", "dict_value": "01"}
func (h *AdminDBHandler) CreateDictionaryConfig(ctx *atreugo.RequestCtx) error {
	var config model.DictionaryConfig
	if err := util.BindJSON(ctx, &config); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}
	if err := database.DB.Create(&config).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}

	// 创建后自动重新加载字典
	hook.ReloadDictionary()

	return h.successJSON(ctx, config)
}

// UpdateDictionaryConfig 更新字典配置
// PUT /admin/db/dictionary-configs/:id
func (h *AdminDBHandler) UpdateDictionaryConfig(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	var config model.DictionaryConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return h.errorJSON(ctx, 404, "dictionary config not found")
	}

	var updates model.DictionaryConfig
	if err := util.BindJSON(ctx, &updates); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}

	if err := database.DB.Model(&config).Updates(&updates).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}

	// 更新后自动重新加载字典
	hook.ReloadDictionary()

	return h.successJSON(ctx, config)
}

// DeleteDictionaryConfig 删除字典配置
// DELETE /admin/db/dictionary-configs/:id
func (h *AdminDBHandler) DeleteDictionaryConfig(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")
	if err := database.DB.Delete(&model.DictionaryConfig{}, id).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}

	// 删除后自动重新加载字典
	hook.ReloadDictionary()

	return h.successJSON(ctx, "deleted")
}

// BatchCreateDictionaryConfigs 批量创建字典配置
// POST /admin/db/dictionary-configs/batch
// Body: [{"org_id": "org001", "dict_type": "payment_method", "dict_key": "ALIPAY", "dict_value": "01"}, ...]
func (h *AdminDBHandler) BatchCreateDictionaryConfigs(ctx *atreugo.RequestCtx) error {
	var configs []model.DictionaryConfig
	if err := util.BindJSON(ctx, &configs); err != nil {
		return h.errorJSON(ctx, 400, err.Error())
	}

	if err := database.DB.Create(&configs).Error; err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}

	// 批量创建后自动重新加载字典
	hook.ReloadDictionary()

	return h.successJSON(ctx, configs)
}

// ReloadDictionary 重新加载字典配置
// POST /admin/db/reload-dictionary
// 调用 hook.ReloadDictionary() 从数据库重新加载所有字典配置
func (h *AdminDBHandler) ReloadDictionary(ctx *atreugo.RequestCtx) error {
	if err := hook.ReloadDictionary(); err != nil {
		return h.errorJSON(ctx, 500, err.Error())
	}
	return h.successJSON(ctx, "dictionary reloaded")
}

// ============ 工具方法 ============

// successJSON 返回成功响应
// 统一格式：{"code": 0, "data": xxx}
func (h *AdminDBHandler) successJSON(ctx *atreugo.RequestCtx, data interface{}) error {
	return ctx.JSONResponse(map[string]interface{}{
		"code": 0,
		"data": data,
	})
}

// errorJSON 返回错误响应
// 统一格式：{"code": xxx, "message": "xxx"}
// HTTP 状态码设置为 code 参数的值
func (h *AdminDBHandler) errorJSON(ctx *atreugo.RequestCtx, code int, message string) error {
	return ctx.JSONResponse(map[string]interface{}{
		"code":    code,
		"message": message,
	}, code)
}
