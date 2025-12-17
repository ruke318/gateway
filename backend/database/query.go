package database

import (
	"fmt"

	"github.com/ruke318/gateway/model"
	"gorm.io/gorm"
)

// GetServiceConfig 根据调用参数获取完整的接口配置
func GetServiceConfig(unitID, serviceID, comID string) (*model.Service, error) {
	var svc model.Service

	err := DB.Preload("Vendor").
		Preload("Organization").
		Preload("Hooks", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.StatusEnabled).Order("priority")
		}).
		Preload("Hooks.Script").
		Joins("JOIN organization ON organization.id = service.org_id").
		Joins("JOIN vendor ON vendor.id = service.vendor_id").
		Where("organization.code = ? AND service.service_id = ? AND vendor.code = ?", unitID, serviceID, comID).
		Where("service.status = ? AND organization.status = ? AND vendor.status = ?",
			model.StatusEnabled, model.StatusEnabled, model.StatusEnabled).
		First(&svc).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("service not found: unit_id=%s, service_id=%s, com_id=%s", unitID, serviceID, comID)
	}
	if err != nil {
		return nil, fmt.Errorf("query service failed: %w", err)
	}

	return &svc, nil
}

// GetScriptLibrary 获取所有公共函数库
func GetScriptLibrary() (map[string]map[string]string, error) {
	var scripts []model.ScriptLibrary
	if err := DB.Where("status = ?", model.StatusEnabled).Find(&scripts).Error; err != nil {
		return nil, err
	}

	// namespace -> name -> script_content
	library := make(map[string]map[string]string)
	for _, s := range scripts {
		if library[s.Namespace] == nil {
			library[s.Namespace] = make(map[string]string)
		}
		library[s.Namespace][s.Name] = s.ScriptContent
	}
	return library, nil
}

// GetDictionaryConfigs 获取所有字典配置（用于加载到内存）
func GetDictionaryConfigs() ([]model.DictionaryConfig, error) {
	var configs []model.DictionaryConfig
	if err := DB.Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetDictionaryConfigsByOrg 根据机构ID获取字典配置
func GetDictionaryConfigsByOrg(orgID string) ([]model.DictionaryConfig, error) {
	var configs []model.DictionaryConfig
	if err := DB.Where("org_id = ?", orgID).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// GetDictionaryConfigsByOrgAndType 根据机构ID和字典类型获取配置
func GetDictionaryConfigsByOrgAndType(orgID, dictType string) ([]model.DictionaryConfig, error) {
	var configs []model.DictionaryConfig
	if err := DB.Where("org_id = ? AND dict_type = ?", orgID, dictType).Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}
