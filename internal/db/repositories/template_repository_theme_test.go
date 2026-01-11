package repositories

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestTemplateRepository_GetTemplatesByCategory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	thumbnailURL := "/static/images/themes/wedding-thumb.jpg"
	imageURL := "/static/images/themes/wedding-header.jpg"

	templates := []*models.Template{
		{
			Name:         "Wedding Theme 1",
			Type:         models.TemplateTypeRSVPPage,
			HTMLContent:  "<html>Wedding 1</html>",
			IsDefault:    false,
			IsActive:     true,
			CreatedBy:    user.ID,
			Category:     models.CategoryCard,
			Description:  "Elegant wedding theme",
			ThumbnailURL: &thumbnailURL,
			ImageURL:     &imageURL,
			Tags:         []string{"wedding", "formal"},
			SortOrder:    1,
		},
		{
			Name:         "Wedding Theme 2",
			Type:         models.TemplateTypeRSVPPage,
			HTMLContent:  "<html>Wedding 2</html>",
			IsDefault:    false,
			IsActive:     true,
			CreatedBy:    user.ID,
			Category:     models.CategoryCard,
			Description:  "Modern wedding theme",
			ThumbnailURL: &thumbnailURL,
			ImageURL:     &imageURL,
			Tags:         []string{"wedding", "modern"},
			SortOrder:    2,
		},
		{
			Name:        "Plain Theme",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Plain</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
			Description: "Simple plain theme",
			Tags:        []string{"plain", "simple"},
			SortOrder:   0,
		},
		{
			Name:        "Birthday Theme",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Birthday</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryFun,
			Description: "Fun birthday theme",
			Tags:        []string{"birthday", "fun"},
			SortOrder:   1,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	tests := []struct {
		name      string
		category  models.TemplateCategory
		wantCount int
		wantFirst string
	}{
		{
			name:      "get card category templates",
			category:  models.CategoryCard,
			wantCount: 2,
			wantFirst: "Wedding Theme 1",
		},
		{
			name:      "get plain category templates",
			category:  models.CategoryPlain,
			wantCount: 1,
			wantFirst: "Plain Theme",
		},
		{
			name:      "get fun category templates",
			category:  models.CategoryFun,
			wantCount: 1,
			wantFirst: "Birthday Theme",
		},
		{
			name:      "get modern category templates (none)",
			category:  models.CategoryModern,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetTemplatesByCategory(context.Background(), tt.category)
			if err != nil {
				t.Errorf("GetTemplatesByCategory() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("Expected %d templates, got %d", tt.wantCount, len(results))
				return
			}

			if tt.wantCount > 0 && results[0].Name != tt.wantFirst {
				t.Errorf("Expected first template name %s, got %s", tt.wantFirst, results[0].Name)
			}

			for _, tmpl := range results {
				if tmpl.Category != tt.category {
					t.Errorf("Expected category %s, got %s", tt.category, tmpl.Category)
				}
			}
		})
	}
}

func TestTemplateRepository_GetTemplatesByCategory_SortOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	templates := []*models.Template{
		{
			Name:        "Theme C",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>C</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme C",
			Tags:        []string{"test"},
			SortOrder:   3,
		},
		{
			Name:        "Theme A",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>A</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme A",
			Tags:        []string{"test"},
			SortOrder:   1,
		},
		{
			Name:        "Theme B",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>B</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme B",
			Tags:        []string{"test"},
			SortOrder:   2,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	results, err := repo.GetTemplatesByCategory(context.Background(), models.CategoryCard)
	if err != nil {
		t.Fatalf("GetTemplatesByCategory() error = %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 templates, got %d", len(results))
	}

	expectedOrder := []string{"Theme A", "Theme B", "Theme C"}
	for i, expected := range expectedOrder {
		if results[i].Name != expected {
			t.Errorf("Expected template %d to be %s, got %s", i, expected, results[i].Name)
		}
	}
}

func TestTemplateRepository_ListThemes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	templates := []*models.Template{
		{
			Name:        "RSVP Card Theme",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>RSVP Card</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Card theme for RSVP",
			Tags:        []string{"rsvp", "card"},
			SortOrder:   1,
		},
		{
			Name:        "RSVP Plain Theme",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>RSVP Plain</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
			Description: "Plain theme for RSVP",
			Tags:        []string{"rsvp", "plain"},
			SortOrder:   0,
		},
		{
			Name:        "Confirmation Card Theme",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>Confirmation Card</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Card theme for confirmation",
			Tags:        []string{"confirmation", "card"},
			SortOrder:   1,
		},
		{
			Name:        "Email Theme",
			Type:        models.TemplateTypeInviteEmail,
			HTMLContent: "<html>Email</html>",
			TextContent: func() *string { s := "Email text"; return &s }(),
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
			Description: "Email theme",
			Tags:        []string{"email"},
			SortOrder:   0,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	tests := []struct {
		name         string
		templateType models.TemplateType
		category     *models.TemplateCategory
		wantCount    int
		wantFirst    string
	}{
		{
			name:         "list all RSVP themes",
			templateType: models.TemplateTypeRSVPPage,
			category:     nil,
			wantCount:    2,
			wantFirst:    "RSVP Plain Theme",
		},
		{
			name:         "list RSVP card themes",
			templateType: models.TemplateTypeRSVPPage,
			category:     func() *models.TemplateCategory { c := models.CategoryCard; return &c }(),
			wantCount:    1,
			wantFirst:    "RSVP Card Theme",
		},
		{
			name:         "list RSVP plain themes",
			templateType: models.TemplateTypeRSVPPage,
			category:     func() *models.TemplateCategory { c := models.CategoryPlain; return &c }(),
			wantCount:    1,
			wantFirst:    "RSVP Plain Theme",
		},
		{
			name:         "list confirmation themes",
			templateType: models.TemplateTypeConfirmationPage,
			category:     nil,
			wantCount:    1,
			wantFirst:    "Confirmation Card Theme",
		},
		{
			name:         "list email themes",
			templateType: models.TemplateTypeInviteEmail,
			category:     nil,
			wantCount:    1,
			wantFirst:    "Email Theme",
		},
		{
			name:         "list RSVP modern themes (none)",
			templateType: models.TemplateTypeRSVPPage,
			category:     func() *models.TemplateCategory { c := models.CategoryModern; return &c }(),
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.ListThemes(context.Background(), tt.templateType, tt.category)
			if err != nil {
				t.Errorf("ListThemes() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("Expected %d templates, got %d", tt.wantCount, len(results))
				return
			}

			if tt.wantCount > 0 {
				if results[0].Name != tt.wantFirst {
					t.Errorf("Expected first template name %s, got %s", tt.wantFirst, results[0].Name)
				}

				if results[0].Type != tt.templateType {
					t.Errorf("Expected type %s, got %s", tt.templateType, results[0].Type)
				}

				if tt.category != nil && results[0].Category != *tt.category {
					t.Errorf("Expected category %s, got %s", *tt.category, results[0].Category)
				}
			}
		})
	}
}

func TestTemplateRepository_ListThemes_SortOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	templates := []*models.Template{
		{
			Name:        "Theme Z",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Z</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme Z",
			Tags:        []string{"test"},
			SortOrder:   3,
		},
		{
			Name:        "Theme A",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>A</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme A",
			Tags:        []string{"test"},
			SortOrder:   1,
		},
		{
			Name:        "Theme M",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>M</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme M",
			Tags:        []string{"test"},
			SortOrder:   2,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	results, err := repo.ListThemes(context.Background(), models.TemplateTypeRSVPPage, nil)
	if err != nil {
		t.Fatalf("ListThemes() error = %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 templates, got %d", len(results))
	}

	expectedOrder := []string{"Theme A", "Theme M", "Theme Z"}
	for i, expected := range expectedOrder {
		if results[i].Name != expected {
			t.Errorf("Expected template %d to be %s, got %s", i, expected, results[i].Name)
		}
	}
}

func TestTemplateRepository_ListThemes_WithSameSortOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	templates := []*models.Template{
		{
			Name:        "Theme C",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>C</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme C",
			Tags:        []string{"test"},
			SortOrder:   1,
		},
		{
			Name:        "Theme A",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>A</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme A",
			Tags:        []string{"test"},
			SortOrder:   1,
		},
		{
			Name:        "Theme B",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>B</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryCard,
			Description: "Theme B",
			Tags:        []string{"test"},
			SortOrder:   1,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	results, err := repo.ListThemes(context.Background(), models.TemplateTypeRSVPPage, nil)
	if err != nil {
		t.Fatalf("ListThemes() error = %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 templates, got %d", len(results))
	}

	expectedOrder := []string{"Theme A", "Theme B", "Theme C"}
	for i, expected := range expectedOrder {
		if results[i].Name != expected {
			t.Errorf("Expected template %d to be %s, got %s", i, expected, results[i].Name)
		}
	}
}

func TestTemplateRepository_ThemeFields_Persistence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	thumbnailURL := "/static/images/themes/test-thumb.jpg"
	imageURL := "/static/images/themes/test-header.jpg"

	template := &models.Template{
		Name:         "Theme Persistence Test",
		Type:         models.TemplateTypeRSVPPage,
		HTMLContent:  "<html>Test</html>",
		IsDefault:    false,
		IsActive:     true,
		CreatedBy:    user.ID,
		Category:     models.CategoryCard,
		Description:  "Testing theme field persistence",
		ThumbnailURL: &thumbnailURL,
		ImageURL:     &imageURL,
		Tags:         []string{"test", "persistence", "theme"},
		SortOrder:    5,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve template: %v", err)
	}

	if retrieved.Category != models.CategoryCard {
		t.Errorf("Expected category %s, got %s", models.CategoryCard, retrieved.Category)
	}

	if retrieved.Description != "Testing theme field persistence" {
		t.Errorf("Expected description 'Testing theme field persistence', got %s", retrieved.Description)
	}

	if retrieved.ThumbnailURL == nil || *retrieved.ThumbnailURL != thumbnailURL {
		t.Errorf("Expected thumbnail URL %s, got %v", thumbnailURL, retrieved.ThumbnailURL)
	}

	if retrieved.ImageURL == nil || *retrieved.ImageURL != imageURL {
		t.Errorf("Expected image URL %s, got %v", imageURL, retrieved.ImageURL)
	}

	if len(retrieved.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrieved.Tags))
	}

	expectedTags := []string{"test", "persistence", "theme"}
	for i, expected := range expectedTags {
		if i >= len(retrieved.Tags) || retrieved.Tags[i] != expected {
			t.Errorf("Expected tag %d to be %s, got %v", i, expected, retrieved.Tags)
		}
	}

	if retrieved.SortOrder != 5 {
		t.Errorf("Expected sort order 5, got %d", retrieved.SortOrder)
	}
}

func TestTemplateRepository_ThemeFields_NullableFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:         "Plain Theme",
		Type:         models.TemplateTypeRSVPPage,
		HTMLContent:  "<html>Plain</html>",
		IsDefault:    false,
		IsActive:     true,
		CreatedBy:    user.ID,
		Category:     models.CategoryPlain,
		Description:  "Plain theme without images",
		ThumbnailURL: nil,
		ImageURL:     nil,
		Tags:         []string{},
		SortOrder:    0,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve template: %v", err)
	}

	if retrieved.ThumbnailURL != nil {
		t.Errorf("Expected nil thumbnail URL, got %v", retrieved.ThumbnailURL)
	}

	if retrieved.ImageURL != nil {
		t.Errorf("Expected nil image URL, got %v", retrieved.ImageURL)
	}

	if retrieved.Tags == nil {
		t.Error("Expected empty tags array, got nil")
	}

	if len(retrieved.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(retrieved.Tags))
	}
}

func TestTemplateRepository_Update_ThemeFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	thumbnailURL1 := "/static/images/themes/original-thumb.jpg"
	imageURL1 := "/static/images/themes/original-header.jpg"

	template := &models.Template{
		Name:         "Original Theme",
		Type:         models.TemplateTypeRSVPPage,
		HTMLContent:  "<html>Original</html>",
		IsDefault:    false,
		IsActive:     true,
		CreatedBy:    user.ID,
		Category:     models.CategoryCard,
		Description:  "Original description",
		ThumbnailURL: &thumbnailURL1,
		ImageURL:     &imageURL1,
		Tags:         []string{"original", "test"},
		SortOrder:    1,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	thumbnailURL2 := "/static/images/themes/updated-thumb.jpg"
	imageURL2 := "/static/images/themes/updated-header.jpg"

	template.Category = models.CategoryModern
	template.Description = "Updated description"
	template.ThumbnailURL = &thumbnailURL2
	template.ImageURL = &imageURL2
	template.Tags = []string{"updated", "modern", "test"}
	template.SortOrder = 2

	err = repo.Update(context.Background(), template)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated template: %v", err)
	}

	if retrieved.Category != models.CategoryModern {
		t.Errorf("Expected category %s, got %s", models.CategoryModern, retrieved.Category)
	}

	if retrieved.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got %s", retrieved.Description)
	}

	if retrieved.ThumbnailURL == nil || *retrieved.ThumbnailURL != thumbnailURL2 {
		t.Errorf("Expected thumbnail URL %s, got %v", thumbnailURL2, retrieved.ThumbnailURL)
	}

	if retrieved.ImageURL == nil || *retrieved.ImageURL != imageURL2 {
		t.Errorf("Expected image URL %s, got %v", imageURL2, retrieved.ImageURL)
	}

	if len(retrieved.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrieved.Tags))
	}

	if retrieved.SortOrder != 2 {
		t.Errorf("Expected sort order 2, got %d", retrieved.SortOrder)
	}
}
