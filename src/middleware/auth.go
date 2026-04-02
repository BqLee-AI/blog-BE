package middleware

import (
	"net/http"
	"strings"

	"blog-BE/src/utils"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "jwtClaims"

type authResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
	Code      string      `json:"code,omitempty"`
}

func newAuthResponse(c *gin.Context, message string, data interface{}, code string) authResponse {
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

	return authResponse{
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Code:      code,
	}
}

// JWTAuth 是一个 Gin 中间件，用于验证请求中的 JWT 访问令牌；如果令牌无效或过期，则返回 401 错误。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.ExtractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, newAuthResponse(
				c,
				"Missing bearer token",
				nil,
				"TOKEN_MISSING",
			))
			c.Abort()
			return
		}

		claims, err := utils.ParseAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, newAuthResponse(
				c,
				"Invalid or expired token",
				nil,
				"TOKEN_INVALID",
			))
			c.Abort()
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

// ClaimsFromContext 从 Gin 上下文中提取 JWT Claims；如果不存在或类型不匹配，则返回 nil 和 false。
func ClaimsFromContext(c *gin.Context) (*utils.Claims, bool) {
	value, exists := c.Get(claimsContextKey)
	if !exists {
		return nil, false
	}

	claims, ok := value.(*utils.Claims)
	return claims, ok
}
