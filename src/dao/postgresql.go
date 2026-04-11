package dao

import (
	"fmt"

	"blog-BE/src/config"
	"blog-BE/src/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InitPgSql() (err error) {
	// 构建连接字符串，使用配置中的数据库参数
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.AppConfig.DBHost,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
		config.AppConfig.DBPort,
	)

	// 建立连接
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	if err = sqlDB.Ping(); err != nil {
		return err
	}

	// 迁移模型
	DB.AutoMigrate()

	return nil
}

func MyClose() {
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Error("获取数据库连接失败", zap.Error(err))
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.Log.Error("数据库断开失败", zap.Error(err))
	}
}
