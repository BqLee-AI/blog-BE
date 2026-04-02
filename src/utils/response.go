package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const RequestIDContextKey = "request_id"

type Response struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
	Code      string      `json:"code,omitempty"`
}

func NewResponse(c *gin.Context, message string, data interface{}, code string) Response {
	return Response{
		Message:   message,
		Data:      data,
		RequestID: RequestIDFromContext(c),
		Code:      code,
	}
}

// RequestIDFromContext 提取请求 ID，优先使用请求头，其次使用上下文中的值。
func RequestIDFromContext(c *gin.Context) string {
	requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(c.GetHeader("X-Trace-ID"))
	}
	if requestID == "" {
		if value, exists := c.Get(RequestIDContextKey); exists {
			if id, ok := value.(string); ok {
				requestID = strings.TrimSpace(id)
			}
		}
	}

	return requestID
}
