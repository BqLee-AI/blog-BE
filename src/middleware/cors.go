package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     splitAndTrim(os.Getenv("CORS_ALLOW_ORIGINS"), []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		AllowMethods:     splitAndTrim(os.Getenv("CORS_ALLOW_METHODS"), []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		AllowHeaders:     splitAndTrim(os.Getenv("CORS_ALLOW_HEADERS"), []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}),
		ExposeHeaders:    splitAndTrim(os.Getenv("CORS_EXPOSE_HEADERS"), []string{"Content-Length"}),
		AllowCredentials: strings.EqualFold(os.Getenv("CORS_ALLOW_CREDENTIALS"), "true"),
		MaxAge:           12 * time.Hour,
	})
}

func splitAndTrim(value string, defaultValues []string) []string {
	if strings.TrimSpace(value) == "" {
		return defaultValues
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValues
	}

	return result
}
