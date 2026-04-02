package handler

import (
	"blog-BE/src/models"
	"blog-BE/src/models/request"
	"blog-BE/src/utils"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type articleAuthorResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type articleResponse struct {
	ID         uint                   `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Summary    string                 `json:"summary"`
	CoverImage string                 `json:"cover_image"`
	AuthorID   uint                   `json:"author_id"`
	Author     *articleAuthorResponse `json:"author,omitempty"`
	Status     string                 `json:"status"`
	ViewCount  int                    `json:"view_count"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

func newArticleResponse(article *models.Article) articleResponse {
	response := articleResponse{
		ID:         article.ID,
		Title:      article.Title,
		Content:    article.Content,
		Summary:    article.Summary,
		CoverImage: article.CoverImage,
		AuthorID:   article.AuthorID,
		Status:     article.Status,
		ViewCount:  article.ViewCount,
		CreatedAt:  article.CreatedAt,
		UpdatedAt:  article.UpdatedAt,
	}

	if article.Author != nil {
		response.Author = &articleAuthorResponse{
			ID:       article.Author.ID,
			Username: article.Author.Username,
		}
	}

	return response
}

func newArticleListResponse(articles []models.Article) []articleResponse {
	responses := make([]articleResponse, 0, len(articles))
	for i := range articles {
		responses = append(responses, newArticleResponse(&articles[i]))
	}

	return responses
}

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

	if req.Status == "" {
		req.Status = "published"
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
			"items":     newArticleListResponse(articles),
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

	claims, hasClaims := getOptionalArticleClaims(c)
	if article.Status != "published" && !canReadUnpublishedArticle(article, claims, hasClaims) {
		c.JSON(http.StatusNotFound, newResponse(
			c,
			"Article not found",
			nil,
			"NOT_FOUND",
		))
		return
	}

	if err := models.IncrementViewCount(uint(id)); err != nil {
		log.Printf("failed to increment article view count: article_id=%d err=%v", id, err)
	}

	c.JSON(http.StatusOK, newResponse(
		c,
		"success",
		newArticleResponse(article),
		"",
	))
}

func getOptionalArticleClaims(c *gin.Context) (*utils.Claims, bool) {
	token := utils.ExtractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		return nil, false
	}

	claims, err := utils.ParseAccessToken(token)
	if err != nil {
		return nil, false
	}

	return claims, true
}

func canReadUnpublishedArticle(article *models.Article, claims *utils.Claims, hasClaims bool) bool {
	if !hasClaims || claims == nil {
		return false
	}

	if article.AuthorID == claims.UserID {
		return true
	}

	return claims.RoleID != 0
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
		newArticleResponse(&article),
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
