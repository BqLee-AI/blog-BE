package routers

import (
	"blog-BE/src/config"
	"blog-BE/src/handler"
	"blog-BE/src/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	// 创建 Gin 引擎
	router := gin.Default()
	if err := router.SetTrustedProxies(config.AppConfig.TrustedProxies); err != nil {
		log.Fatalf("failed to set trusted proxies: %v", err)
	}
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("/api/v1")
	{
		v1.GET("/categories", handler.GetCategories)
		v1.GET("/tags", handler.GetTags)

		articles := v1.Group("/articles")
		articles.GET("", handler.GetArticles)
		articles.GET("/:id", handler.GetArticle)

		authArticles := v1.Group("/articles")
		authArticles.Use(middleware.JWTAuth())
		authArticles.POST("", handler.CreateArticle)
		authArticles.PUT("/:id", handler.UpdateArticle)
		authArticles.DELETE("/:id", handler.DeleteArticle)

		// 用户认证相关路由
		auth := v1.Group("/auth")
		auth.POST("/login", handler.LoginHandler)
		auth.POST("/send-code", handler.SendVerificationCodeHandler)
		auth.POST("/verify-email", handler.VerifyEmailHandler)
		// 刷新 token 和获取当前用户信息需要认证
		auth.POST("/refresh", handler.RefreshTokenHandler)
		// 获取当前用户信息接口需要 JWT 认证中间件保护
		auth.GET("/me", middleware.JWTAuth(), handler.MeHandler)
		auth.POST("/register", handler.RegisterHandler)
	}

	return router
}
