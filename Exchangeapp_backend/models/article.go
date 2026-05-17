package models

import "gorm.io/gorm"

// Article 对应论坛文章模型。
type Article struct {
	gorm.Model
	Title          string `binding:"required"`
	Content        string `binding:"required"`
	Preview        string `binding:"required"`
	AuthorUsername string `gorm:"index"`
}
