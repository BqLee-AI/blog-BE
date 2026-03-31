package models

import (
	"blog-BE/src/dao"
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Title      string         `gorm:"size:255;not null" json:"title"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Summary    string         `gorm:"size:500" json:"summary"`
	CoverImage string         `gorm:"size:255" json:"cover_image"`
	AuthorID   uint           `gorm:"not null;index" json:"author_id"`
	Author     User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Status     string         `gorm:"size:20;default:draft;index" json:"status"`
	ViewCount  int            `gorm:"default:0" json:"view_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Article) TableName() string {
	return "articles"
}

func CreateArticle(article *Article) error {
	return dao.DB.Create(article).Error
}

func GetArticleByID(id uint) (*Article, error) {
	var article Article
	if err := dao.DB.Preload("Author").First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}

func GetArticles(page, pageSize int, status string) ([]Article, int64, error) {
	var articles []Article
	var total int64

	countQuery := dao.DB.Model(&Article{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := dao.DB.Model(&Article{})
	if status != "" {
		listQuery = listQuery.Where("status = ?", status)
	}

	offset := (page - 1) * pageSize
	if err := listQuery.
		Preload("Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func UpdateArticle(id uint, updates map[string]interface{}) error {
	return dao.DB.Model(&Article{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteArticle(id uint) error {
	return dao.DB.Delete(&Article{}, id).Error
}

func IncrementViewCount(id uint) error {
	return dao.DB.Model(&Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
