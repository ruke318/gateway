package model

// HookScript 驱动脚本
type HookScript struct {
	BaseModel
	Name          string `gorm:"size:128;not null;comment:脚本名称" json:"name"`
	HookPoint     string `gorm:"size:32;not null;index;comment:Hook节点类型" json:"hook_point"`
	ScriptContent string `gorm:"type:text;not null;comment:脚本内容" json:"script_content"`
	Description   string `gorm:"type:text;comment:描述" json:"description"`
}

func (HookScript) TableName() string {
	return "hook_script"
}

// Hook 节点类型常量
const (
	HookBeforeAuth              = "BeforeAuth"
	HookAfterAuth               = "AfterAuth"
	HookBeforeRequestTransform  = "BeforeRequestTransform"
	HookAfterRequestTransform   = "AfterRequestTransform"
	HookBeforeForward           = "BeforeForward"
	HookAfterForward            = "AfterForward"
	HookBeforeResponseTransform = "BeforeResponseTransform"
	HookAfterResponseTransform  = "AfterResponseTransform"
	HookOnError                 = "OnError"
)
