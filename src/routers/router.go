package routers

import (
	"blog-BE/src/handler"
	"blog-BE/src/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 创建 Gin 引擎
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	vi := router.Group("/api/v1")
	{
		// 用户认证相关路由
		auth := vi.Group("/auth")
		// 登录接口支持 GET 和 POST 两种方式，方便前端调试和兼容不同请求习惯
		auth.GET("/login", handler.LoginHandler)
		auth.POST("/login", handler.LoginHandler)
		// 刷新 token 和获取当前用户信息需要认证
		auth.POST("/refresh", handler.RefreshTokenHandler)
		// 获取当前用户信息接口需要 JWT 认证中间件保护
		auth.GET("/me", middleware.JWTAuth(), handler.MeHandler)
		auth.POST("/register", handler.RegisterHandler)
	}

	return router
}
