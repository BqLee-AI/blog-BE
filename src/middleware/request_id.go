package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"blog-BE/src/utils"

	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.RequestIDFromContext(c)
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(utils.RequestIDContextKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buffer[:])
}
