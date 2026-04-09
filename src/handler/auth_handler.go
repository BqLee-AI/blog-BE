package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/service"
	"blog-BE/src/utils"
	"errors"
	"log"
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
		log.Printf("failed to check email status for send-code: request_id=%s err=%v", utils.RequestIDFromContext(c), err)
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to check email status",
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
			"Failed to send verification code",
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

	if err := service.VerifyAndMarkEmailVerified(req.Email, req.Code); err == nil {

		c.JSON(http.StatusOK, utils.NewResponse(
			c,
			"Email verification successful",
			gin.H{"verified": true},
			"",
		))
		return
	} else if errors.Is(err, service.ErrVerificationCodeNotFound) {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Verification code not found",
			nil,
			"VERIFICATION_CODE_MISSING",
		))
		return
	} else if errors.Is(err, service.ErrVerificationCodeInvalid) {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Verification code is incorrect or expired",
			nil,
			"VERIFICATION_CODE_INVALID",
		))
		return
	}

	c.JSON(http.StatusInternalServerError, utils.NewResponse(
		c,
		"Failed to verify code",
		nil,
		"VERIFICATION_CHECK_FAILED",
	))
}
