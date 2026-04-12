package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	categories, err := models.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"获取分类失败",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"success",
		categories,
		"",
	))
}
