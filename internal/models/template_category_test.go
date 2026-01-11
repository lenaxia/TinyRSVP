package models

import (
	"errors"
	"strings"
	"testing"
)

func TestTemplateCategory_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		category TemplateCategory
		want     bool
	}{
		{
			name:     "valid plain category",
			category: CategoryPlain,
			want:     true,
		},
		{
			name:     "valid card category",
			category: CategoryCard,
			want:     true,
		},
		{
			name:     "valid modern category",
			category: CategoryModern,
			want:     true,
		},
		{
			name:     "valid classic category",
			category: CategoryClassic,
			want:     true,
		},
		{
			name:     "valid fun category",
			category: CategoryFun,
			want:     true,
		},
		{
			name:     "invalid category",
			category: "invalid_category",
			want:     false,
		},
		{
			name:     "empty category",
			category: "",
			want:     false,
		},
		{
			name:     "random string",
			category: "random",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.IsValid(); got != tt.want {
				t.Errorf("TemplateCategory.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category TemplateCategory
		want     string
	}{
		{
			name:     "plain category",
			category: CategoryPlain,
			want:     "plain",
		},
		{
			name:     "card category",
			category: CategoryCard,
			want:     "card",
		},
		{
			name:     "modern category",
			category: CategoryModern,
			want:     "modern",
		},
		{
			name:     "classic category",
			category: CategoryClassic,
			want:     "classic",
		},
		{
			name:     "fun category",
			category: CategoryFun,
			want:     "fun",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("TemplateCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_Validate_ThemeFields(t *testing.T) {
	thumbnailURL := "/static/images/themes/wedding-thumb.jpg"
	imageURL := "/static/images/themes/wedding-header.jpg"

	tests := []struct {
		name     string
		template *Template
		wantErr  bool
		errField string
	}{
		{
			name: "valid template with theme fields",
			template: &Template{
				Name:         "Wedding Elegance",
				Type:         TemplateTypeRSVPPage,
				HTMLContent:  "<html>{{.Event.Title}}</html>",
				CreatedBy:    1,
				Category:     CategoryCard,
				Description:  "Elegant wedding invitation with floral design",
				ThumbnailURL: &thumbnailURL,
				ImageURL:     &imageURL,
				Tags:         []string{"wedding", "formal", "floral"},
				SortOrder:    1,
			},
			wantErr: false,
		},
		{
			name: "valid template with plain category",
			template: &Template{
				Name:        "Plain Text",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryPlain,
				Description: "Simple plain text invitation",
				Tags:        []string{"plain", "simple"},
				SortOrder:   0,
			},
			wantErr: false,
		},
		{
			name: "valid template with nil image URLs",
			template: &Template{
				Name:         "Text Only",
				Type:         TemplateTypeRSVPPage,
				HTMLContent:  "<html>{{.Event.Title}}</html>",
				CreatedBy:    1,
				Category:     CategoryPlain,
				Description:  "Text only template",
				ThumbnailURL: nil,
				ImageURL:     nil,
				Tags:         []string{"plain"},
				SortOrder:    0,
			},
			wantErr: false,
		},
		{
			name: "valid template with empty tags",
			template: &Template{
				Name:        "Simple Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryModern,
				Description: "Modern template",
				Tags:        []string{},
				SortOrder:   0,
			},
			wantErr: false,
		},
		{
			name: "invalid category",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    "invalid_category",
				Description: "Test description",
				Tags:        []string{"test"},
				SortOrder:   0,
			},
			wantErr:  true,
			errField: "category",
		},
		{
			name: "description too long",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryCard,
				Description: strings.Repeat("a", 501),
				Tags:        []string{"test"},
				SortOrder:   0,
			},
			wantErr:  true,
			errField: "description",
		},
		{
			name: "negative sort order",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryCard,
				Description: "Test description",
				Tags:        []string{"test"},
				SortOrder:   -1,
			},
			wantErr:  true,
			errField: "sort_order",
		},
		{
			name: "description with exactly 500 characters",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryCard,
				Description: strings.Repeat("a", 500),
				Tags:        []string{"test"},
				SortOrder:   0,
			},
			wantErr: false,
		},
		{
			name: "sort order zero is valid",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryCard,
				Description: "Test description",
				Tags:        []string{"test"},
				SortOrder:   0,
			},
			wantErr: false,
		},
		{
			name: "large sort order is valid",
			template: &Template{
				Name:        "Test Template",
				Type:        TemplateTypeRSVPPage,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				CreatedBy:   1,
				Category:    CategoryCard,
				Description: "Test description",
				Tags:        []string{"test"},
				SortOrder:   9999,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Template.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errField != "" {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if valErr.Field != tt.errField {
					t.Errorf("Expected error for field %s, got %s", tt.errField, valErr.Field)
				}
			}
		})
	}
}

func TestTemplate_Validate_ThemeFields_EdgeCases(t *testing.T) {
	t.Run("empty description is valid", func(t *testing.T) {
		template := &Template{
			Name:        "Test Template",
			Type:        TemplateTypeRSVPPage,
			HTMLContent: "<html>{{.Event.Title}}</html>",
			CreatedBy:   1,
			Category:    CategoryCard,
			Description: "",
			Tags:        []string{},
			SortOrder:   0,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid template with empty description, got error: %v", err)
		}
	})

	t.Run("nil tags is valid", func(t *testing.T) {
		template := &Template{
			Name:        "Test Template",
			Type:        TemplateTypeRSVPPage,
			HTMLContent: "<html>{{.Event.Title}}</html>",
			CreatedBy:   1,
			Category:    CategoryCard,
			Description: "Test",
			Tags:        nil,
			SortOrder:   0,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid template with nil tags, got error: %v", err)
		}
	})

	t.Run("multiple tags are valid", func(t *testing.T) {
		template := &Template{
			Name:        "Test Template",
			Type:        TemplateTypeRSVPPage,
			HTMLContent: "<html>{{.Event.Title}}</html>",
			CreatedBy:   1,
			Category:    CategoryCard,
			Description: "Test",
			Tags:        []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
			SortOrder:   0,
		}
		if err := template.Validate(); err != nil {
			t.Errorf("Expected valid template with multiple tags, got error: %v", err)
		}
	})
}
