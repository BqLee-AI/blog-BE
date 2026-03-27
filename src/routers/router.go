package routers

import (
	"blog-BE/src/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 创建 Gin 引擎
	router := gin.Default()

	vi := router.Group("/api/v1")
	{
		// 用户认证相关路由
		auth := vi.Group("/auth")
		auth.GET("/login", handler.LoginHandler)
		auth.POST("/register", handler.RegisterHandler)
	}

	return router
}
