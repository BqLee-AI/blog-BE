package main

import (
	"fmt"

	"blog-BE/src/config"
	"blog-BE/src/dao"
	"blog-BE/src/logger"
	"blog-BE/src/models"
	"blog-BE/src/routers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置文件
	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	serverConfig := config.Get().Server
	logger.InitLogger(serverConfig.Mode)
	defer logger.Sync()

	if err := config.InitRedis(); err != nil {
		logger.L().Fatal("Redis 连接失败", zap.Error(err))
	}

	// 设置 Gin 模式
	gin.SetMode(serverConfig.Mode)

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
	addr := fmt.Sprintf(":%d", serverConfig.Port)
	logger.L().Info("服务启动", zap.String("addr", "http://localhost"+addr), zap.Int("port", serverConfig.Port))
	if err := router.Run(addr); err != nil {
		logger.L().Fatal("服务启动失败", zap.Error(err))
	}
}
