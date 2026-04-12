package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTags(c *gin.Context) {
	tags, err := models.GetTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"获取标签失败",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"success",
		tags,
		"",
	))
}
