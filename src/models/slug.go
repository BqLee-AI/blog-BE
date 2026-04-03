package models

import (
	"blog-BE/src/dao"
	"fmt"
	"strings"
)

func buildUniqueSlug(model any, rawValue string, maxLength int) (string, error) {
	slugBase := normalizeSlugBase(rawValue)
	if slugBase == "" {
		slugBase = "item"
	}

	slugBase = truncateSlug(slugBase, maxLength)
	slug := slugBase

	for suffix := 2; ; suffix++ {
		var count int64
		if err := dao.DB.Model(model).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}

		suffixText := fmt.Sprintf("-%d", suffix)
		baseLimit := maxLength - len(suffixText)
		if baseLimit < 1 {
			return "", fmt.Errorf("slug max length %d is too short", maxLength)
		}

		candidateBase := truncateSlug(slugBase, baseLimit)
		if candidateBase == "" {
			candidateBase = "item"
		}
		slug = candidateBase + suffixText
	}
}

func normalizeSlugBase(rawValue string) string {
	rawValue = strings.ToLower(strings.TrimSpace(rawValue))
	if rawValue == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(rawValue))
	lastWasHyphen := false

	for _, r := range rawValue {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastWasHyphen = false
		case r == ' ' || r == '-' || r == '_' || r == '.' || r == '/':
			if builder.Len() > 0 && !lastWasHyphen {
				builder.WriteByte('-')
				lastWasHyphen = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func truncateSlug(value string, maxLength int) string {
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}

	return strings.Trim(value[:maxLength], "-")
}
