package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/service"
	"net/http"

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

func LoginHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	user, err := models.FindUserByEmail(email)

	if err != nil || user.Password != password {
		c.JSON(http.StatusUnauthorized, newResponse(
			"Invalid email or password",
			nil,
			"AUTH_FAILED",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		"Login successful",
		gin.H{"user_id": user.ID},
		"",
	))
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
