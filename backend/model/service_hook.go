package model

// ServiceHook 接口 Hook 关联
type ServiceHook struct {
	BaseModel
	ServicePK    uint64  `gorm:"not null;index;comment:接口ID" json:"service_pk"`
	HookPoint    string  `gorm:"size:32;not null;comment:Hook节点类型" json:"hook_point"`
	ScriptID     *uint64 `gorm:"comment:脚本ID" json:"script_id"`
	InlineScript string  `gorm:"type:text;comment:内联脚本" json:"inline_script"`
	Priority     int     `gorm:"default:0;comment:优先级" json:"priority"`

	// 关联
	Script *HookScript `gorm:"foreignKey:ScriptID" json:"script,omitempty"`
}

func (ServiceHook) TableName() string {
	return "service_hook"
}

// GetScriptContent 获取脚本内容（优先内联脚本）
func (sh *ServiceHook) GetScriptContent() string {
	if sh.InlineScript != "" {
		return sh.InlineScript
	}
	if sh.Script != nil {
		return sh.Script.ScriptContent
	}
	return ""
}
