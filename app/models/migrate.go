package models

import "gorm.io/gorm"

func Migrate(db *gorm.DB) {
	// 执行数据库迁移
	db.AutoMigrate(&Book{})
}
