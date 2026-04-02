package request

type CreateArticleReq struct {
	Title      string `json:"title" binding:"required,min=1,max=255"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary" binding:"omitempty,max=500"`
	CoverImage string `json:"cover_image"`
	Status     string `json:"status" binding:"omitempty,oneof=draft published"`
}

type UpdateArticleReq struct {
	Title      *string `json:"title" binding:"omitempty,min=1,max=255"`
	Content    *string `json:"content" binding:"omitempty,min=1"`
	Summary    *string `json:"summary" binding:"omitempty,max=500"`
	CoverImage *string `json:"cover_image"`
	Status     *string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type ArticleListReq struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=draft published archived"`
}
