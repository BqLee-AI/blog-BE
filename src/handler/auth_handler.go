package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/service"
	"blog-BE/src/utils"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type verificationCodeRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type verificationEmailRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
	Code  string `json:"code" form:"code" binding:"required"`
}

func SendVerificationCodeHandler(c *gin.Context) {
	var req verificationCodeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid verification code request payload",
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
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
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
			c.JSON(http.StatusTooManyRequests, utils.NewResponse(
				c,
				cooldownErr.Error(),
				gin.H{"retry_after_seconds": int(cooldownErr.Remaining.Seconds())},
				"VERIFICATION_COOLDOWN",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			fmt.Sprintf("Failed to send verification code: %v", err),
			nil,
			"VERIFICATION_SEND_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"Verification code sent successfully",
		gin.H{"retry_after_seconds": 60},
		"",
	))
}

func VerifyEmailHandler(c *gin.Context) {
	var req verificationEmailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid verification payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if service.VerifyCode(req.Email, req.Code) {
		c.JSON(http.StatusOK, utils.NewResponse(
			c,
			"Email verification successful",
			gin.H{"verified": true},
			"",
		))
		return
	}

	c.JSON(http.StatusBadRequest, utils.NewResponse(
		c,
		"Verification code is incorrect or expired",
		nil,
		"VERIFICATION_CODE_INVALID",
	))
}
