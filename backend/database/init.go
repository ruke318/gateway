package database

import (
	"log"

	"github.com/ruke318/gateway/model"
)

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 检查是否已有管理员
	var count int64
	if err := DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果没有用户，创建默认管理员
	if count == 0 {
		admin := &model.User{
			Username: "admin",
			RealName: "系统管理员",
			Role:     "admin",
			Status:   1,
		}
		if err := admin.SetPassword("admin123"); err != nil {
			return err
		}

		if err := DB.Create(admin).Error; err != nil {
			return err
		}

		log.Println("✓ 创建默认管理员账号: admin / admin123")
	}

	return nil
}
