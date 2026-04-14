package main

import (
	"fmt"
	"os"

	"blog-BE/src/config"
	"blog-BE/src/dao"
	"blog-BE/src/logger"
	"blog-BE/src/models"
	"blog-BE/src/routers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger(os.Getenv("GIN_MODE"))
	defer logger.Sync()

	// 加载配置文件
	if err := config.LoadConfig(".env"); err != nil {
		logger.L().Fatal("加载 .env 文件失败", zap.Error(err))
	}
	logger.InitLogger(config.AppConfig.GinMode)
	if err := config.InitRedis(); err != nil {
		logger.L().Fatal("Redis 连接失败", zap.Error(err))
	}

	// 设置 Gin 模式
	gin.SetMode(config.AppConfig.GinMode)

	// 创建数据库连接
	if err := dao.InitPgSql(); err != nil {
		logger.L().Fatal("数据库连接失败", zap.Error(err))
	}
	// 断开数据库连接
	defer dao.MyClose()

	// 模型绑定
	if err := dao.DB.AutoMigrate(&models.User{}, &models.Category{}, &models.Tag{}, &models.Article{}); err != nil {
		logger.L().Fatal("数据库迁移失败", zap.Error(err))
	}

	router := routers.SetupRouter()

	// 启动服务
	addr := fmt.Sprintf(":%d", config.AppConfig.AppPort)
	logger.L().Info("服务启动", zap.String("addr", "http://localhost"+addr), zap.Int("port", config.AppConfig.AppPort))
	if err := router.Run(addr); err != nil {
		logger.L().Fatal("服务启动失败", zap.Error(err))
	}
}
