package models

import "gorm.io/gorm"

// User 对应用户表模型。
type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
}
