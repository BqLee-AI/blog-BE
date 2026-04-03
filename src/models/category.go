package models

import (
	"blog-BE/src/dao"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Slug        string         `gorm:"size:50;uniqueIndex" json:"slug"`
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
	err := dao.DB.Where("parent_id IS NULL").Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func GetCategoryByID(id uint) (*Category, error) {
	var category Category
	err := dao.DB.First(&category, id).Error
	return &category, err
}

func CreateCategory(category *Category) error {
	return dao.DB.Create(category).Error
}
