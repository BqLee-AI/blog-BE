package models

import (
	"blog-BE/src/dao"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:50;not null;uniqueIndex:idx_categories_name_active,where:deleted_at IS NULL" json:"name"`
	Slug        string         `gorm:"size:50;uniqueIndex:idx_categories_slug_active,where:deleted_at IS NULL" json:"slug"`
	Description string         `gorm:"size:255" json:"description"`
	ParentID    *uint          `gorm:"index" json:"parent_id"`
	Parent      *Category      `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func GetCategories() ([]Category, error) {
	var categories []Category
	err := dao.DB.Order("parent_id ASC NULLS FIRST").Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func GetCategoryByID(id uint) (*Category, error) {
	var category Category
	err := dao.DB.First(&category, id).Error
	return &category, err
}

func CreateCategory(category *Category) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		slug, err := buildUniqueSlug(&Category{}, firstNonEmpty(category.Slug, category.Name), 50)
		if err != nil {
			return err
		}
		category.Slug = slug

		if err := dao.DB.Create(category).Error; err != nil {
			lastErr = err
			if isSlugUniqueConstraintError(err) {
				continue
			}
			return err
		}

		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to create category after retries: %w", lastErr)
	}

	return fmt.Errorf("failed to create category after retries")
}
