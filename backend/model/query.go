package model

// VendorQuery 厂商查询参数
type VendorQuery struct {
	Code string `schema:"code"` // 精确匹配
	Name string `schema:"name"` // 模糊匹配
}

// OrganizationQuery 机构查询参数
type OrganizationQuery struct {
	Code string `schema:"code"` // 精确匹配
	Name string `schema:"name"` // 模糊匹配
}

// ServiceQuery 接口查询参数
type ServiceQuery struct {
	ServiceID string `schema:"service_id"` // 模糊匹配
	Name      string `schema:"name"`       // 模糊匹配
	VendorID  uint64 `schema:"vendor_id"`  // 精确匹配
	OrgID     uint64 `schema:"org_id"`     // 精确匹配
}

// ScriptQuery 公共函数库查询参数
type ScriptQuery struct {
	Name        string `schema:"name"`        // 模糊匹配
	Description string `schema:"description"` // 模糊匹配
}

// HookScriptQuery Hook脚本查询参数
type HookScriptQuery struct {
	Name      string `schema:"name"`       // 模糊匹配
	HookPoint string `schema:"hook_point"` // 精确匹配
}
