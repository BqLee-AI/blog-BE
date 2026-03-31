package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/models/request"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetArticles(c *gin.Context) {
	var req request.ArticleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article list payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	articles, total, err := models.GetArticles(req.Page, req.PageSize, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to fetch article list",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
		"success",
		gin.H{
			"items":     articles,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
		"",
	))
}

func GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article ID",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	article, err := models.GetArticleByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, newResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	go func(articleID uint) {
		_ = models.IncrementViewCount(articleID)
	}(uint(id))

	c.JSON(http.StatusOK, newResponse(
		c,
		"success",
		article,
		"",
	))
}

func CreateArticle(c *gin.Context) {
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

	var req request.CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	article := models.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		AuthorID:   claims.UserID,
		Status:     req.Status,
	}
	if article.Status == "" {
		article.Status = "draft"
	}

	if err := models.CreateArticle(&article); err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to create article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if refreshed, err := models.GetArticleByID(article.ID); err == nil {
		article = *refreshed
	}

	c.JSON(http.StatusCreated, newResponse(
		c,
		"Article created successfully",
		article,
		"",
	))
}

func UpdateArticle(c *gin.Context) {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article ID",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	article, err := models.GetArticleByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, newResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if article.AuthorID != claims.UserID && claims.RoleID == 0 {
		c.JSON(http.StatusForbidden, newResponse(
			c,
			"You do not have permission to update this article",
			nil,
			"FORBIDDEN",
		))
		return
	}

	var req request.UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"No fields to update",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if err := models.UpdateArticle(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to update article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
		"Article updated successfully",
		nil,
		"",
	))
}

func DeleteArticle(c *gin.Context) {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, newResponse(
			c,
			"Invalid article ID",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	article, err := models.GetArticleByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, newResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if article.AuthorID != claims.UserID && claims.RoleID == 0 {
		c.JSON(http.StatusForbidden, newResponse(
			c,
			"You do not have permission to delete this article",
			nil,
			"FORBIDDEN",
		))
		return
	}

	if err := models.DeleteArticle(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, newResponse(
			c,
			"Failed to delete article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, newResponse(
		c,
		"Article deleted successfully",
		nil,
		"",
	))
}
