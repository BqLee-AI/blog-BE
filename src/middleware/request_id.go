package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = strings.TrimSpace(c.GetHeader("X-Trace-ID"))
		}
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(requestIDContextKey, requestID)
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
