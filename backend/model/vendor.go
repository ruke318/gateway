package model

// Vendor 厂商
type Vendor struct {
	BaseModel
	Code        string `gorm:"size:64;uniqueIndex;not null;comment:厂商编码" json:"code"`
	Name        string `gorm:"size:128;not null;comment:厂商名称" json:"name"`
	BaseURL     string `gorm:"size:512;comment:基础URL" json:"base_url"`
	Description string `gorm:"type:text;comment:描述" json:"description"`
}

func (Vendor) TableName() string {
	return "vendor"
}
