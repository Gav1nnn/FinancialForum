package config

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// initDB 初始化 MySQL 连接池并执行自动迁移。
func initDB() {
	dsn := AppConfig.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to initialize the database, got error %v ", err)
	}
	// 配置连接池参数，避免默认值在高并发下成为瓶颈。
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(AppConfig.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(AppConfig.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		log.Fatalf("Failed to configure database, get error %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Article{}, &models.ExchangeRate{}); err != nil {
		log.Fatalf("Failed to auto migrate database schema, got error %v", err)
	}

	global.Db = db
}
