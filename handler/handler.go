package handler

import (
	"Blog/blog-BE/models"
	"Blog/blog-BE/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	user, err := models.FindUserByEmail(email)

	if err != nil || user.Password != password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to send verification code",
		})
		return
	}

	if err := models.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Registration failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
	})
}
