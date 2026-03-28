package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/service"
	"blog-BE/src/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 统一响应结构
type Response struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Code      string      `json:"code,omitempty"`
}

func newResponse(message string, data interface{}, code string) Response {
	return Response{
		Message:   message,
		Data:      data,
		RequestID: "trace-id", // 可从 context 获取
		Code:      code,
	}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			"Invalid login payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	user, err := models.FindUserByEmail(req.Email)

	if err != nil || user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, newResponse(
			"Invalid email or password",
			nil,
			"AUTH_FAILED",
		))
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, user.Username, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			"Failed to generate token",
			nil,
			"TOKEN_GENERATION_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		"Login successful",
		gin.H{
			"user": gin.H{
				"user_id":  user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role_id":  user.RoleID,
			},
			"tokens": gin.H{
				"token_type":         "Bearer",
				"access_token":       tokens.AccessToken,
				"refresh_token":      tokens.RefreshToken,
				"access_expires_at":  tokens.AccessExpiresAt,
				"refresh_expires_at": tokens.RefreshExpiresAt,
			},
		},
		"",
	))
}

func RefreshTokenHandler(c *gin.Context) {
	refreshToken := strings.TrimSpace(c.PostForm("refresh_token"))
	if refreshToken == "" {
		refreshToken = utils.ExtractBearerToken(c.GetHeader("Authorization"))
	}
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, newResponse(
			"Refresh token is required",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	tokens, err := utils.RefreshTokenPair(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, newResponse(
			"Invalid or expired refresh token",
			nil,
			"TOKEN_INVALID",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		"Token refreshed successfully",
		gin.H{
			"token_type":         "Bearer",
			"access_token":       tokens.AccessToken,
			"refresh_token":      tokens.RefreshToken,
			"access_expires_at":  tokens.AccessExpiresAt,
			"refresh_expires_at": tokens.RefreshExpiresAt,
		},
		"",
	))
}

func MeHandler(c *gin.Context) {
	claims, ok := utilsClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, newResponse(
			"Unauthorized",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		"Token is valid",
		gin.H{
			"user_id":    claims.UserID,
			"username":   claims.Username,
			"role_id":    claims.RoleID,
			"token_type": claims.TokenType,
		},
		"",
	))
}

func utilsClaimsFromContext(c *gin.Context) (*utils.Claims, bool) {
	value, exists := c.Get("jwtClaims")
	if !exists {
		return nil, false
	}

	claims, ok := value.(*utils.Claims)
	return claims, ok
}

func RegisterHandler(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	code := c.PostForm("code")
	user := &models.User{
		Username: username,
		Email:    email,
		Password: password,
		RoleID:   0,
	}
	// 验证码
	realcode, err := service.SendMail("", email)
	if err != nil || realcode == "" || realcode != code {
		c.JSON(http.StatusInternalServerError, newResponse(
			"Failed to send verification code",
			nil,
			"VERIFICATION_FAILED",
		))
		return
	}

	if err := models.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			"Registration failed",
			nil,
			"REGISTRATION_FAILED",
		))
		return
	}
	c.JSON(http.StatusOK, newResponse(
		"Registration successful",
		gin.H{"user_id": user.ID},
		"",
	))
}
