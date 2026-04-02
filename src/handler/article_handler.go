package handler

import (
	"blog-BE/src/middleware"
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

type articleListResponse struct {
	ID         uint                   `json:"id"`
	Title      string                 `json:"title"`
	Summary    string                 `json:"summary"`
	CoverImage string                 `json:"cover_image"`
	AuthorID   uint                   `json:"author_id"`
	Author     *articleAuthorResponse `json:"author,omitempty"`
	Status     string                 `json:"status"`
	ViewCount  int                    `json:"view_count"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type articleDetailResponse struct {
	articleListResponse
	Content string `json:"content"`
}

func newArticleListResponseItem(article *models.Article) articleListResponse {
	response := articleListResponse{
		ID:         article.ID,
		Title:      article.Title,
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

func newArticleDetailResponse(article *models.Article) articleDetailResponse {
	return articleDetailResponse{
		articleListResponse: newArticleListResponseItem(article),
		Content:             article.Content,
	}
}

func newArticleListResponse(articles []models.Article) []articleListResponse {
	responses := make([]articleListResponse, 0, len(articles))
	for i := range articles {
		responses = append(responses, newArticleListResponseItem(&articles[i]))
	}

	return responses
}

func GetArticles(c *gin.Context) {
	var req request.ArticleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid article list payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	claims, hasClaims := getOptionalArticleClaims(c)
	filterByAuthor := false
	authorID := uint(0)

	if req.Status == "" {
		req.Status = "published"
	}

	if req.Status != "published" {
		if !hasClaims || claims == nil {
			req.Status = "published"
		} else if claims.RoleID == 0 {
			filterByAuthor = true
			authorID = claims.UserID
		}
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	articles, total, err := models.GetArticles(req.Page, req.PageSize, req.Status, authorID, filterByAuthor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to fetch article list",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
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
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid article ID",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	article, err := models.GetArticleWithAuthorByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.NewResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	claims, hasClaims := getOptionalArticleClaims(c)
	if article.Status != "published" && !canReadUnpublishedArticle(article, claims, hasClaims) {
		c.JSON(http.StatusNotFound, utils.NewResponse(
			c,
			"Article not found",
			nil,
			"NOT_FOUND",
		))
		return
	}

	if err := models.IncrementViewCount(uint(id)); err != nil {
		log.Printf("failed to increment article view count: article_id=%d err=%v", id, err)
	} else {
		article.ViewCount++
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"success",
		newArticleDetailResponse(article),
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

	var req request.CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
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
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to create article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if refreshed, err := models.GetArticleWithAuthorByID(article.ID); err == nil {
		article = *refreshed
	}

	c.JSON(http.StatusCreated, utils.NewResponse(
		c,
		"Article created successfully",
		newArticleDetailResponse(&article),
		"",
	))
}

func UpdateArticle(c *gin.Context) {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
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
			c.JSON(http.StatusNotFound, utils.NewResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if article.AuthorID != claims.UserID && claims.RoleID == 0 {
		c.JSON(http.StatusForbidden, utils.NewResponse(
			c,
			"You do not have permission to update this article",
			nil,
			"FORBIDDEN",
		))
		return
	}

	var req request.UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"Invalid article payload",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			c,
			"No fields to update",
			nil,
			"INVALID_REQUEST",
		))
		return
	}

	if err := models.UpdateArticle(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to update article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"Article updated successfully",
		nil,
		"",
	))
}

func DeleteArticle(c *gin.Context) {
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

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
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
			c.JSON(http.StatusNotFound, utils.NewResponse(
				c,
				"Article not found",
				nil,
				"NOT_FOUND",
			))
			return
		}

		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to fetch article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	if article.AuthorID != claims.UserID && claims.RoleID == 0 {
		c.JSON(http.StatusForbidden, utils.NewResponse(
			c,
			"You do not have permission to delete this article",
			nil,
			"FORBIDDEN",
		))
		return
	}

	if err := models.DeleteArticle(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			c,
			"Failed to delete article",
			nil,
			"DATABASE_ERROR",
		))
		return
	}

	c.JSON(http.StatusOK, utils.NewResponse(
		c,
		"Article deleted successfully",
		nil,
		"",
	))
}
