package main

import (
	"fmt"
	"log"

	"blog-BE/src/config"
	"blog-BE/src/dao"
	"blog-BE/src/models"
	"blog-BE/src/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置文件
	if err := config.LoadConfig(".env"); err != nil {
		log.Fatal("加载 .env 文件失败：", err)
	}
	if err := config.InitRedis(); err != nil {
		log.Fatal("Redis 连接失败：", err)
	}

	// 设置 Gin 模式
	gin.SetMode(config.AppConfig.GinMode)

	// 创建数据库连接
	if err := dao.InitPgSql(); err != nil {
		log.Fatal("数据库连接失败：", err)
	}
	// 断开数据库连接
	defer dao.MyClose()

	// 模型绑定
	if err := dao.DB.AutoMigrate(&models.User{}, &models.Category{}, &models.Tag{}, &models.Article{}); err != nil {
		log.Fatal("数据库迁移失败：", err)
	}

	router := routers.SetupRouter()

	// 启动服务
	addr := fmt.Sprintf(":%d", config.AppConfig.AppPort)
	log.Printf("服务启动于 http://localhost:%d\n", config.AppConfig.AppPort)
	router.Run(addr)
}
