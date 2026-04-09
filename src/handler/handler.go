package handler

import (
	"blog-BE/src/middleware"
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

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
	Code     string `json:"code" form:"code"`
}

func LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid login payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	user, err := models.FindUserByEmail(req.Email)
	if err != nil || !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			c,
			"Invalid email or password",
			nil,
			"AUTH_FAILED",
		))
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, user.Username, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to generate token",
			nil,
			"TOKEN_GENERATION_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
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
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Refresh token is required",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	tokens, err := utils.RefreshTokenPair(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			c,
			"Invalid or expired refresh token",
			nil,
			"TOKEN_INVALID",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
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
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewResponse(
			c,
			"Unauthorized",
			nil,
			"TOKEN_MISSING",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
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

func RegisterHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid registration payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if _, err := models.FindUserByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, utils.NewResponse(
			c,
			"Email is already registered",
			nil,
			"EMAIL_ALREADY_REGISTERED",
		))
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Printf("failed to check email status for register: email=[redacted] err=%v\n", err)
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to check email status",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if err := service.RequireEmailVerified(req.Email); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailNotVerified):
			c.JSON(http.StatusBadRequest, utils.NewResponse(
				c,
				"Please verify your email before registering",
				nil,
				"EMAIL_NOT_VERIFIED",
			))
		default:
			fmt.Printf("failed to verify email status for register: email=[redacted] err=%v\n", err)
			c.JSON(http.StatusInternalServerError, utils.NewResponse(
				c,
				"Failed to verify email status",
				nil,
				"VERIFICATION_STATUS_CHECK_FAILED",
			))
		}
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		if errors.Is(err, utils.ErrPasswordTooLong) {
			c.JSON(http.StatusBadRequest, utils.NewResponse(
				c,
				"Password is too long",
				nil,
				"PASSWORD_TOO_LONG",
			))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(
				c,
				"Password hashing failed",
				nil,
				"PASSWORD_HASH_FAILED",
			))
		}
		return
	}
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		RoleID:   0,
	}

	if err := models.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Registration failed",
			nil,
			"REGISTRATION_FAILED",
		))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"Registration successful",
		gin.H{"user_id": user.ID},
		"",
	))
}
