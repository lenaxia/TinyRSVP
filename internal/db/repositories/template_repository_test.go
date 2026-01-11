package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestTemplateRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	textContent := "Event: {{.Event.Title}}"

	tests := []struct {
		name     string
		template *models.Template
		wantErr  bool
	}{
		{
			name: "valid template",
			template: &models.Template{
				Name:        "Test Template",
				Type:        models.TemplateTypeRSVPPage,
				Description: "A test template",
				HTMLContent: "<html>{{.Event.Title}}</html>",
				IsDefault:   false,
				IsActive:    true,
				CreatedBy:   user.ID,
				Category:    models.CategoryPlain,
			},
			wantErr: false,
		},
		{
			name: "valid email template with text content",
			template: &models.Template{
				Name:        "Email Template",
				Type:        models.TemplateTypeInviteEmail,
				HTMLContent: "<html>{{.Event.Title}}</html>",
				TextContent: &textContent,
				IsDefault:   false,
				IsActive:    true,
				CreatedBy:   user.ID,
				Category:    models.CategoryPlain,
			},
			wantErr: false,
		},
		{
			name: "invalid template missing name",
			template: &models.Template{
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				CreatedBy:   user.ID,
			},
			wantErr: true,
		},
		{
			name: "invalid created_by",
			template: &models.Template{
				Name:        "Test Template",
				Type:        models.TemplateTypeRSVPPage,
				HTMLContent: "<html></html>",
				CreatedBy:   99999,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.template.ID == 0 {
					t.Error("Expected template ID to be set")
				}
				if tt.template.Version != 1 {
					t.Errorf("Expected version 1, got %d", tt.template.Version)
				}
				if tt.template.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.template.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestTemplateRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "Test Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>{{.Event.Title}}</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing template",
			id:      template.ID,
			wantErr: false,
		},
		{
			name:    "non-existent template",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := repo.GetByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if retrieved.ID != template.ID {
					t.Errorf("Expected ID %d, got %d", template.ID, retrieved.ID)
				}
				if retrieved.Name != template.Name {
					t.Errorf("Expected name %s, got %s", template.Name, retrieved.Name)
				}
				if retrieved.Type != template.Type {
					t.Errorf("Expected type %s, got %s", template.Type, retrieved.Type)
				}
			}
		})
	}
}

func TestTemplateRepository_GetByEventAndType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	template := &models.Template{
		EventID:     &eventID,
		Name:        "Event Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>{{.Event.Title}}</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tests := []struct {
		name         string
		eventID      int64
		templateType models.TemplateType
		wantErr      bool
	}{
		{
			name:         "existing template",
			eventID:      eventID,
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      false,
		},
		{
			name:         "non-existent event",
			eventID:      99999,
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      true,
		},
		{
			name:         "wrong type",
			eventID:      eventID,
			templateType: models.TemplateTypeInviteEmail,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := repo.GetByEventAndType(context.Background(), tt.eventID, tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEventAndType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if retrieved.ID != template.ID {
					t.Errorf("Expected ID %d, got %d", template.ID, retrieved.ID)
				}
			}
		})
	}
}

func TestTemplateRepository_GetDefaultByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "Default Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>{{.Event.Title}}</html>",
		IsDefault:   true,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tests := []struct {
		name         string
		templateType models.TemplateType
		wantErr      bool
	}{
		{
			name:         "existing default template",
			templateType: models.TemplateTypeRSVPPage,
			wantErr:      false,
		},
		{
			name:         "non-existent default template",
			templateType: models.TemplateTypeInviteEmail,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := repo.GetDefaultByType(context.Background(), tt.templateType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !retrieved.IsDefault {
					t.Error("Expected template to be default")
				}
				if retrieved.Type != tt.templateType {
					t.Errorf("Expected type %s, got %s", tt.templateType, retrieved.Type)
				}
			}
		})
	}
}

func TestTemplateRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	templates := []*models.Template{
		{
			EventID:     &eventID,
			Name:        "Event Template 1",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>1</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			Name:        "Default Template",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>2</html>",
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			Name:        "Inactive Template",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>3</html>",
			IsDefault:   false,
			IsActive:    false,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(context.Background(), tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   *TemplateFilters
		wantCount int
	}{
		{
			name:      "no filters",
			filters:   nil,
			wantCount: 3,
		},
		{
			name: "filter by event",
			filters: &TemplateFilters{
				EventID: &eventID,
			},
			wantCount: 1,
		},
		{
			name: "filter by type",
			filters: &TemplateFilters{
				Type: func() *models.TemplateType { t := models.TemplateTypeRSVPPage; return &t }(),
			},
			wantCount: 2,
		},
		{
			name: "filter by is_default",
			filters: &TemplateFilters{
				IsDefault: func() *bool { b := true; return &b }(),
			},
			wantCount: 1,
		},
		{
			name: "filter by is_active",
			filters: &TemplateFilters{
				IsActive: func() *bool { b := true; return &b }(),
			},
			wantCount: 2,
		},
		{
			name: "filter by created_by",
			filters: &TemplateFilters{
				CreatedBy: &user.ID,
			},
			wantCount: 3,
		},
		{
			name: "with limit",
			filters: &TemplateFilters{
				Limit: 2,
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.List(context.Background(), tt.filters)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("Expected %d templates, got %d", tt.wantCount, len(results))
			}
		})
	}
}

func TestTemplateRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "Original Name",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>Original</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	originalVersion := template.Version
	time.Sleep(10 * time.Millisecond)

	template.Name = "Updated Name"
	template.HTMLContent = "<html>Updated</html>"

	err = repo.Update(context.Background(), template)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if template.Version != originalVersion+1 {
		t.Errorf("Expected version %d, got %d", originalVersion+1, template.Version)
	}

	retrieved, err := repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated template: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrieved.Name)
	}

	if retrieved.HTMLContent != "<html>Updated</html>" {
		t.Errorf("Expected updated HTML content, got %s", retrieved.HTMLContent)
	}
}

func TestTemplateRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "To Delete",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>Delete me</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	err = repo.Delete(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(context.Background(), template.ID)
	if err == nil {
		t.Error("Expected NotFoundError after delete")
	}

	err = repo.Delete(context.Background(), 99999)
	if err == nil {
		t.Error("Expected error when deleting non-existent template")
	}
}

func TestTemplateRepository_SetActive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "Test Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>Test</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	err = repo.SetActive(context.Background(), template.ID, false)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	retrieved, err := repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve template: %v", err)
	}

	if retrieved.IsActive {
		t.Error("Expected template to be inactive")
	}

	err = repo.SetActive(context.Background(), template.ID, true)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	retrieved, err = repo.GetByID(context.Background(), template.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve template: %v", err)
	}

	if !retrieved.IsActive {
		t.Error("Expected template to be active")
	}

	err = repo.SetActive(context.Background(), 99999, false)
	if err == nil {
		t.Error("Expected error when setting active on non-existent template")
	}
}

func TestTemplateRepository_IsTemplateInUse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	eventRepo := NewEventRepository(db)
	user := createTestUser(t, userRepo)

	template := &models.Template{
		Name:        "Test Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>Test</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tests := []struct {
		name       string
		templateID int64
		setupEvent bool
		wantInUse  bool
		wantErr    bool
	}{
		{
			name:       "template not in use",
			templateID: template.ID,
			setupEvent: false,
			wantInUse:  false,
			wantErr:    false,
		},
		{
			name:       "template in use by event",
			templateID: template.ID,
			setupEvent: true,
			wantInUse:  true,
			wantErr:    false,
		},
		{
			name:       "non-existent template",
			templateID: 99999,
			setupEvent: false,
			wantInUse:  false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEvent {
				endTime := time.Now().Add(26 * time.Hour)
				event := &models.Event{
					Title:     "Test Event",
					StartTime: time.Now().Add(24 * time.Hour),
					EndTime:   &endTime,
					Timezone:  "America/Los_Angeles",
					Status:    models.EventStatusDraft,
					CreatedBy: user.ID,
				}
				err := eventRepo.Create(context.Background(), event)
				if err != nil {
					t.Fatalf("Failed to create event: %v", err)
				}

				updateQuery := `UPDATE events SET template_id = ? WHERE id = ?`
				_, err = db.Exec(context.Background(), updateQuery, template.ID, event.ID)
				if err != nil {
					t.Fatalf("Failed to update event template: %v", err)
				}

				defer eventRepo.Delete(context.Background(), event.ID)
			}

			inUse, err := repo.IsTemplateInUse(context.Background(), tt.templateID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsTemplateInUse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if inUse != tt.wantInUse {
				t.Errorf("IsTemplateInUse() = %v, want %v", inUse, tt.wantInUse)
			}
		})
	}
}

func TestTemplateRepository_SetDefault(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	template1 := &models.Template{
		Name:        "Template 1",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>1</html>",
		IsDefault:   true,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	template2 := &models.Template{
		Name:        "Template 2",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>2</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	err := repo.Create(context.Background(), template1)
	if err != nil {
		t.Fatalf("Failed to create template1: %v", err)
	}

	err = repo.Create(context.Background(), template2)
	if err != nil {
		t.Fatalf("Failed to create template2: %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "set new default",
			id:      template2.ID,
			wantErr: false,
		},
		{
			name:    "non-existent template",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SetDefault(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				retrieved, err := repo.GetByID(context.Background(), tt.id)
				if err != nil {
					t.Fatalf("Failed to retrieve template: %v", err)
				}

				if !retrieved.IsDefault {
					t.Error("Expected template to be default")
				}

				oldDefault, err := repo.GetByID(context.Background(), template1.ID)
				if err != nil {
					t.Fatalf("Failed to retrieve old default: %v", err)
				}

				if oldDefault.IsDefault && oldDefault.Type == retrieved.Type {
					t.Error("Expected old default to be unset")
				}
			}
		})
	}
}
