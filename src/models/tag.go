package models

import (
	"blog-BE/src/dao"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:30;not null;uniqueIndex" json:"name"`
	Slug      string         `gorm:"size:30;uniqueIndex" json:"slug"`
	Color     string         `gorm:"size:7;default:#6366f1" json:"color"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func GetTags() ([]Tag, error) {
	var tags []Tag
	err := dao.DB.Order("name ASC").Find(&tags).Error
	return tags, err
}

func GetTagByID(id uint) (*Tag, error) {
	var tag Tag
	err := dao.DB.First(&tag, id).Error
	return &tag, err
}

func CreateTag(tag *Tag) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		slug, err := buildUniqueSlug(&Tag{}, firstNonEmpty(tag.Slug, tag.Name), 30)
		if err != nil {
			return err
		}
		tag.Slug = slug

		if err := dao.DB.Create(tag).Error; err != nil {
			lastErr = err
			if isUniqueConstraintError(err) {
				continue
			}
			return err
		}

		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("failed to create tag after retries")
}

func GetOrCreateTags(names []string) ([]Tag, error) {
	tags := make([]Tag, 0, len(names))
	seen := make(map[string]struct{}, len(names))

	for _, name := range names {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}

		normalizedName := strings.ToLower(trimmedName)
		if _, ok := seen[normalizedName]; ok {
			continue
		}
		seen[normalizedName] = struct{}{}

		var tag Tag
		err := dao.DB.Where("LOWER(name) = LOWER(?)", trimmedName).First(&tag).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			tag = Tag{Name: trimmedName}
			if err := CreateTag(&tag); err != nil {
				return nil, err
			}
		}

		tags = append(tags, tag)
	}

	return tags, nil
}
