package routers

import (
	"Blog/blog-BE/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 创建 Gin 引擎
	router := gin.Default()

	router.GET("/login", handler.LoginHandler)
	router.POST("/register", handler.RegisterHandler)

	return router
}
