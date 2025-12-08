package model

// ScriptLibrary 公共函数库
type ScriptLibrary struct {
	BaseModel
	Name          string `gorm:"size:128;not null;comment:函数名称" json:"name"`
	Namespace     string `gorm:"size:64;default:global;not null;comment:命名空间" json:"namespace"`
	ScriptContent string `gorm:"type:text;not null;comment:函数代码" json:"script_content"`
	Description   string `gorm:"type:text;comment:函数说明" json:"description"`
	Example       string `gorm:"type:text;comment:使用示例" json:"example"`
}

func (ScriptLibrary) TableName() string {
	return "script_library"
}
