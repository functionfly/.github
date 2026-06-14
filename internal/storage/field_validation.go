package storage

import (
	"fmt"
)

func isValidFieldName(name string) bool {
	return validIdentifierChars.MatchString(name) && len(name) <= 64
}

func validateFieldNames(fields map[string]interface{}) error {
	for field := range fields {
		if !isValidFieldName(field) {
			return fmt.Errorf("invalid field name: %s", field)
		}
	}
	return nil
}

func validateFieldName(field string) error {
	if !isValidFieldName(field) {
		return fmt.Errorf("invalid field name: %s", field)
	}
	return nil
}

var changelogEntryAllowedFields = map[string]bool{
	"title":          true,
	"version":        true,
	"content":        true,
	"published":      true,
	"published_at":   true,
	"change_summary": true,
}

var changelogChangeAllowedFields = map[string]bool{
	"entry_id": true,
	"category": true,
	"icon":     true,
	"items":    true,
}

var blogPostAllowedFields = map[string]bool{
	"title":            true,
	"slug":             true,
	"content":          true,
	"excerpt":          true,
	"author_id":        true,
	"category_id":      true,
	"tags":             true,
	"featured_image":   true,
	"meta_title":       true,
	"meta_description": true,
	"published":        true,
	"published_at":     true,
	"status":           true,
}

var blogCategoryAllowedFields = map[string]bool{
	"title":       true,
	"slug":        true,
	"description": true,
	"color":       true,
	"icon":        true,
	"order":       true,
}

var blogAuthorAllowedFields = map[string]bool{
	"name":         true,
	"slug":         true,
	"bio":          true,
	"photo":        true,
	"email":        true,
	"website":      true,
	"social_links": true,
	"role":         true,
	"active":       true,
}

var blogSettingsAllowedFields = map[string]bool{
	"blog_title":       true,
	"posts_per_page":   true,
	"meta_description": true,
}

func filterAllowedFields(fields map[string]interface{}, allowed map[string]bool) map[string]interface{} {
	filtered := make(map[string]interface{})
	for field, value := range fields {
		if allowed[field] {
			filtered[field] = value
		}
	}
	return filtered
}

func validateChangelogEntryFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

func validateChangelogChangeFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

func validateBlogPostFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

func validateBlogCategoryFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

func validateBlogAuthorFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

func validateBlogSettingsFields(fields map[string]interface{}) error {
	return validateFieldNames(fields)
}

var incidentAllowedFields = map[string]bool{
	"title":       true,
	"status":      true,
	"severity":    true,
	"resolved_at": true,
	"resolved_by": true,
}

func validateIncidentFields(fields map[string]interface{}) error {
	for field := range fields {
		if !isValidFieldName(field) {
			return fmt.Errorf("invalid field name: %s", field)
		}
		if !incidentAllowedFields[field] {
			return fmt.Errorf("field name '%s' is not allowed for incident updates", field)
		}
	}
	return nil
}