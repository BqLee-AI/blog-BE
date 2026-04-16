package middleware

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"blog-BE/src/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	corsHandlerMu sync.RWMutex
	corsHandler   gin.HandlerFunc
	corsSignature string
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		getCORSHandler()(c)
	}
}

func getCORSHandler() gin.HandlerFunc {
	corsConfig := config.Get().CORS
	signature := corsConfigSignature(corsConfig)

	corsHandlerMu.RLock()
	if corsHandler != nil && corsSignature == signature {
		handler := corsHandler
		corsHandlerMu.RUnlock()
		return handler
	}
	corsHandlerMu.RUnlock()

	corsHandlerMu.Lock()
	defer corsHandlerMu.Unlock()

	if corsHandler != nil && corsSignature == signature {
		return corsHandler
	}

	corsSignature = signature
	corsHandler = cors.New(buildCORSConfig(corsConfig))
	return corsHandler
}

func buildCORSConfig(corsConfig config.CORSConfig) cors.Config {
	return cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}
}

func corsConfigSignature(corsConfig config.CORSConfig) string {
	parts := []string{
		strings.Join(corsConfig.AllowOrigins, ","),
		strings.Join(corsConfig.AllowMethods, ","),
		strings.Join(corsConfig.AllowHeaders, ","),
		strings.Join(corsConfig.ExposeHeaders, ","),
		strconv.FormatBool(corsConfig.AllowCredentials),
	}

	return strings.Join(parts, "|")
}
