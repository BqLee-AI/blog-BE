package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/service"
	"blog-BE/src/utils"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 统一响应结构
type Response struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId,omitempty"`
	Code      string      `json:"code,omitempty"`
}

func newResponse(c *gin.Context, message string, data interface{}, code string) Response {
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

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type verificationCodeRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type registerRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
	Code     string `json:"code" form:"code" binding:"required"`
}

func LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid login payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	user, err := models.FindUserByEmail(req.Email)

	if err != nil || user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, newResponse(
			c,
			"Invalid email or password",
			nil,
			"AUTH_FAILED",
		))
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, user.Username, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to generate token",
			nil,
			"TOKEN_GENERATION_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
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
			c,
			"Refresh token is required",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	tokens, err := utils.RefreshTokenPair(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, newResponse(
			c,
			"Invalid or expired refresh token",
			nil,
			"TOKEN_INVALID",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
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
			c,
			"Unauthorized",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
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

func SendVerificationCodeHandler(c *gin.Context) {
	var req verificationCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid verification code request payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if _, err := models.FindUserByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, newResponse(
			c,
			"Email is already registered",
			nil,
			"EMAIL_ALREADY_REGISTERED",
		))
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			fmt.Sprintf("Failed to check email status: %v", err),
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if err := service.SendVerificationCode(req.Email); err != nil {
		var cooldownErr *service.VerificationCooldownError
		if errors.As(err, &cooldownErr) {
			c.JSON(http.StatusTooManyRequests, newResponse(
				c,
				cooldownErr.Error(),
				gin.H{"retry_after_seconds": int(cooldownErr.Remaining.Seconds())},
				"VERIFICATION_COOLDOWN",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			fmt.Sprintf("Failed to send verification code: %v", err),
			nil,
			"VERIFICATION_SEND_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
		"Verification code sent successfully",
		gin.H{"retry_after_seconds": 60},
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
	var req registerRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid registration payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if _, err := models.FindUserByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, newResponse(
			c,
			"Email is already registered",
			nil,
			"EMAIL_ALREADY_REGISTERED",
		))
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			fmt.Sprintf("Failed to check email status: %v", err),
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if err := service.VerifyVerificationCode(req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, service.ErrVerificationCodeNotFound):
			c.JSON(http.StatusBadRequest, newResponse(
				c,
				"Verification code not found or expired",
				nil,
				"VERIFICATION_CODE_MISSING",
			))
		case errors.Is(err, service.ErrVerificationCodeExpired):
			c.JSON(http.StatusBadRequest, newResponse(
				c,
				"Verification code has expired, please request a new one",
				nil,
				"VERIFICATION_CODE_EXPIRED",
			))
		case errors.Is(err, service.ErrVerificationCodeInvalid):
			c.JSON(http.StatusBadRequest, newResponse(
				c,
				"Verification code is incorrect",
				nil,
				"VERIFICATION_CODE_INVALID",
			))
		default:
			c.JSON(http.StatusInternalServerError, newResponse(
				c,
				fmt.Sprintf("Failed to verify code: %v", err),
				nil,
				"VERIFICATION_CHECK_FAILED",
			))
		}
		return
	}
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		RoleID:   0,
	}

	if err := models.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Registration failed",
			nil,
			"REGISTRATION_FAILED",
		))
		return
	}
	c.JSON(http.StatusOK, newResponse(
		c,
		"Registration successful",
		gin.H{"user_id": user.ID},
		"",
	))
}
