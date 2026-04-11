package middleware

import (
	"net/http"
	"time"

	"blog-BE/src/logger"
	"blog-BE/src/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", utils.RequestIDFromContext(c)),
		}

		if len(c.Errors) > 0 || c.Writer.Status() >= http.StatusInternalServerError {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			logger.Log.Error("HTTP Request", fields...)
			return
		}

		logger.Log.Info("HTTP Request", fields...)
	}
}
