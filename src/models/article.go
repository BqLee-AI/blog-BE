package models

import (
	"blog-BE/src/dao"
	"errors"
	"strings"
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
	Author     *User          `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	CategoryID *uint          `gorm:"index" json:"category_id"`
	Category   *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags       []Tag          `gorm:"many2many:article_tags" json:"tags,omitempty"`
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
	if err := dao.DB.First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}

func GetArticleWithAuthorByID(id uint) (*Article, error) {
	var article Article
	if err := dao.DB.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}

func GetArticles(page, pageSize int, status, keyword, sortBy string, authorID uint, filterByAuthor bool) ([]Article, int64, error) {
	var articles []Article
	var total int64

	countQuery := applyArticleListFilters(dao.DB.Model(&Article{}), status, keyword, authorID, filterByAuthor)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := applyArticleListFilters(dao.DB.Model(&Article{}), status, keyword, authorID, filterByAuthor)

	offset := (page - 1) * pageSize
	if err := listQuery.
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Order(getArticleListOrder(sortBy)).
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func applyArticleListFilters(query *gorm.DB, status, keyword string, authorID uint, filterByAuthor bool) *gorm.DB {
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if filterByAuthor {
		query = query.Where("author_id = ?", authorID)
	}
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", likeKeyword, likeKeyword)
	}

	return query
}

func getArticleListOrder(sortBy string) string {
	switch strings.ToLower(sortBy) {
	case "view_count":
		return "view_count DESC, created_at DESC"
	default:
		return "created_at DESC"
	}
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

func UpdateArticleTags(articleID uint, tagIDs []uint) error {
	var article Article
	if err := dao.DB.First(&article, articleID).Error; err != nil {
		return err
	}

	uniqueTagIDs := make([]uint, 0, len(tagIDs))
	seen := make(map[uint]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		uniqueTagIDs = append(uniqueTagIDs, tagID)
	}

	tags := make([]Tag, 0, len(uniqueTagIDs))
	if len(uniqueTagIDs) > 0 {
		if err := dao.DB.Where("id IN ?", uniqueTagIDs).Find(&tags).Error; err != nil {
			return err
		}
		if len(tags) != len(uniqueTagIDs) {
			return errors.New("one or more tag IDs are invalid")
		}
	}

	return dao.DB.Model(&article).Association("Tags").Replace(tags)
}
