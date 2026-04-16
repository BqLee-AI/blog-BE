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
	databaseConfig := config.Get().Database

	// 构建连接字符串，使用配置中的数据库参数
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		databaseConfig.Host,
		databaseConfig.User,
		databaseConfig.Password,
		databaseConfig.Name,
		databaseConfig.Port,
		databaseConfig.SSLMode,
		databaseConfig.TimeZone,
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

	return nil
}

func MyClose() {
	if DB == nil {
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.L().Error("获取数据库连接失败", zap.Error(err))
		return
	}
	if sqlDB == nil {
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.L().Error("数据库断开失败", zap.Error(err))
	}
}
