package dao

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func InitPgSql() (err error) {
	// 1. 配置连接字符串 (根据你的数据库修改参数)
	dsn := "host=localhost user=admin password=123456 dbname=mydb port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	// 2. 建立连接
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
		log.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatal("数据库断开失败：", err)
	}
}
