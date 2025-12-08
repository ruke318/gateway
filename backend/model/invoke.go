package model

// InvokeRequest 统一调用入口请求
type InvokeRequest struct {
	ComID     string      `json:"com_id"`     // 厂商编码
	ServiceID string      `json:"service_id"` // 接口标识
	UnitID    string      `json:"unit_id"`    // 机构编码
	BizNo     string      `json:"biz_no"`     // 业务流水号，用于日志追踪
	Req       interface{} `json:"req"`        // 请求体
}

// InvokeResponse 统一调用入口响应
type InvokeResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	LogID   string      `json:"log_id,omitempty"` // 返回 LogID 便于追踪
	Data    interface{} `json:"data,omitempty"`
}
