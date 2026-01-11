package models

import (
	"testing"
)

func TestTemplate_Validate_SystemTheme(t *testing.T) {
	template := &Template{
		Name:        "System Theme",
		Type:        TemplateTypeRSVPPage,
		Description: "System theme description",
		HTMLContent: "<div>Content</div>",
		Category:    CategoryPlain,
		IsDefault:   true,
		IsActive:    true,
		Version:     1,
		CreatedBy:   0,
		Tags:        []string{"system"},
		SortOrder:   0,
	}

	err := template.Validate()
	if err != nil {
		t.Errorf("System theme with CreatedBy=0 should be valid, got error: %v", err)
	}
}

func TestTemplate_Validate_UserTheme(t *testing.T) {
	template := &Template{
		Name:        "User Theme",
		Type:        TemplateTypeRSVPPage,
		Description: "User theme description",
		HTMLContent: "<div>Content</div>",
		Category:    CategoryCard,
		IsDefault:   false,
		IsActive:    true,
		Version:     1,
		CreatedBy:   123,
		Tags:        []string{"custom"},
		SortOrder:   5,
	}

	err := template.Validate()
	if err != nil {
		t.Errorf("User theme with CreatedBy>0 should be valid, got error: %v", err)
	}
}
