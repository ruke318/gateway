package model

// DictionaryConfig 字典配置
// 用于存储机构级别的字段映射关系，支持机构内转换和跨机构转换
type DictionaryConfig struct {
	BaseModel
	OrgID       string `gorm:"size:64;not null;index:idx_org_type;comment:机构ID" json:"org_id"`
	DictType    string `gorm:"size:64;not null;index:idx_org_type;comment:字典类型(如: payment_method, order_status)" json:"dict_type"`
	DictKey     string `gorm:"size:128;not null;comment:字典键(标准键名)" json:"dict_key"`
	DictValue   string `gorm:"size:256;not null;comment:字典值(机构特定值)" json:"dict_value"`
	Description string `gorm:"type:text;comment:说明" json:"description"`
}

func (DictionaryConfig) TableName() string {
	return "dictionary_config"
}
