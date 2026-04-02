package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
	Code      string      `json:"code,omitempty"`
}

func NewResponse(c *gin.Context, message string, data interface{}, code string) Response {
	requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(c.GetHeader("X-Trace-ID"))
	}
	if requestID == "" {
		if value, exists := c.Get("request_id"); exists {
			if id, ok := value.(string); ok {
				requestID = strings.TrimSpace(id)
			}
		}
	}

	return Response{
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Code:      code,
	}
}
