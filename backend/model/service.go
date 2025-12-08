package model

import "encoding/json"

// Service 接口配置
type Service struct {
	BaseModel
	ServiceID         string          `gorm:"size:64;not null;comment:接口标识" json:"service_id"`
	OrgID             uint64          `gorm:"not null;index;comment:机构ID" json:"org_id"`
	VendorID          uint64          `gorm:"not null;index;comment:厂商ID" json:"vendor_id"`
	Name              string          `gorm:"size:128;not null;comment:接口名称" json:"name"`
	Description       string          `gorm:"type:text;comment:描述" json:"description"`
	BackendURL        string          `gorm:"size:512;comment:后端URL" json:"backend_url"`
	BackendPath       string          `gorm:"size:512;comment:后端路径" json:"backend_path"`
	BackendMethod     string          `gorm:"size:16;default:POST;comment:请求方法" json:"backend_method"`
	BodyType          string          `gorm:"size:16;default:json;comment:请求体类型" json:"body_type"`
	RequestTransform  json.RawMessage `gorm:"type:json;comment:请求DSL转换配置" json:"request_transform"`
	ResponseTransform json.RawMessage `gorm:"type:json;comment:响应DSL转换配置" json:"response_transform"`

	// 关联
	Vendor       *Vendor       `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
	Organization *Organization `gorm:"foreignKey:OrgID" json:"organization,omitempty"`
	Hooks        []ServiceHook `gorm:"foreignKey:ServicePK" json:"hooks,omitempty"`
}

func (Service) TableName() string {
	return "service"
}

// GetBackendURL 获取后端 URL（优先接口配置，否则用厂商配置）
func (s *Service) GetBackendURL() string {
	if s.BackendURL != "" {
		return s.BackendURL
	}
	if s.Vendor != nil {
		return s.Vendor.BaseURL
	}
	return ""
}

// GetHooksMap 获取 Hook 脚本 map
func (s *Service) GetHooksMap() map[string][]string {
	hooks := make(map[string][]string)
	for _, h := range s.Hooks {
		content := h.GetScriptContent()
		if content != "" {
			hooks[h.HookPoint] = append(hooks[h.HookPoint], content)
		}
	}
	return hooks
}

// GetRequestTransformMap 获取请求转换配置
func (s *Service) GetRequestTransformMap() (map[string]interface{}, error) {
	if len(s.RequestTransform) == 0 {
		return nil, nil
	}

	// 先尝试直接解析为 map
	var result map[string]interface{}
	if err := json.Unmarshal(s.RequestTransform, &result); err == nil {
		return result, nil
	}

	// 如果失败，可能是字符串形式的 JSON，先解析为字符串再解析
	var str string
	if err := json.Unmarshal(s.RequestTransform, &str); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetResponseTransformMap 获取响应转换配置
func (s *Service) GetResponseTransformMap() (map[string]interface{}, error) {
	if len(s.ResponseTransform) == 0 {
		return nil, nil
	}

	// 先尝试直接解析为 map
	var result map[string]interface{}
	if err := json.Unmarshal(s.ResponseTransform, &result); err == nil {
		return result, nil
	}

	// 如果失败，可能是字符串形式的 JSON，先解析为字符串再解析
	var str string
	if err := json.Unmarshal(s.ResponseTransform, &str); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		return nil, err
	}
	return result, nil
}
