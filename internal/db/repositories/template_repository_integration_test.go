package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestTemplateRepository_Integration_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)

	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	ctx := context.Background()
	textContent := "Event: {{.Event.Title}}"
	cssContent := ".container { color: red; }"

	t.Run("full CRUD lifecycle", func(t *testing.T) {
		template := &models.Template{
			EventID:     &eventID,
			Name:        "Integration Test Template",
			Type:        models.TemplateTypeRSVPPage,
			Description: "Full CRUD test",
			HTMLContent: "<html>{{.Event.Title}}</html>",
			TextContent: &textContent,
			CSSContent:  &cssContent,
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if template.ID == 0 {
			t.Fatal("Expected template ID to be set")
		}

		retrieved, err := repo.GetByID(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.Name != template.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, template.Name)
		}

		template.Name = "Updated Name"
		template.Description = "Updated description"
		if err := repo.Update(ctx, template); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		if template.Version != 2 {
			t.Errorf("Expected version 2 after update, got %d", template.Version)
		}

		retrieved, err = repo.GetByID(ctx, template.ID)
		if err != nil {
			t.Fatalf("GetByID after update failed: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("Name not updated: got %s, want 'Updated Name'", retrieved.Name)
		}

		if err := repo.Delete(ctx, template.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, template.ID)
		if err == nil {
			t.Error("Expected NotFoundError after delete")
		}
	})
}

func TestTemplateRepository_Integration_EventAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)

	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	ctx := context.Background()

	t.Run("template associated with event", func(t *testing.T) {
		template := &models.Template{
			EventID:     &eventID,
			Name:        "Event-Specific Template",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Event specific</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByEventAndType(ctx, eventID, models.TemplateTypeRSVPPage)
		if err != nil {
			t.Fatalf("GetByEventAndType failed: %v", err)
		}

		if retrieved.ID != template.ID {
			t.Errorf("Expected template ID %d, got %d", template.ID, retrieved.ID)
		}

		if retrieved.EventID == nil || *retrieved.EventID != eventID {
			t.Error("Event association not preserved")
		}
	})
}

func TestTemplateRepository_Integration_DefaultTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	ctx := context.Background()
	textContent := "Default email text"

	t.Run("retrieve default template by type", func(t *testing.T) {
		template := &models.Template{
			Name:        "Default RSVP Template",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Default RSVP</html>",
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetDefaultByType(ctx, models.TemplateTypeRSVPPage)
		if err != nil {
			t.Fatalf("GetDefaultByType failed: %v", err)
		}

		if retrieved.ID != template.ID {
			t.Errorf("Expected template ID %d, got %d", template.ID, retrieved.ID)
		}

		if !retrieved.IsDefault {
			t.Error("Expected template to be marked as default")
		}
	})

	t.Run("multiple default templates returns most recent", func(t *testing.T) {
		template1 := &models.Template{
			Name:        "Default Email 1",
			Type:        models.TemplateTypeInviteEmail,
			HTMLContent: "<html>Email 1</html>",
			TextContent: &textContent,
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template1); err != nil {
			t.Fatalf("Create template1 failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		template2 := &models.Template{
			Name:        "Default Email 2",
			Type:        models.TemplateTypeInviteEmail,
			HTMLContent: "<html>Email 2</html>",
			TextContent: &textContent,
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template2); err != nil {
			t.Fatalf("Create template2 failed: %v", err)
		}

		retrieved, err := repo.GetDefaultByType(ctx, models.TemplateTypeInviteEmail)
		if err != nil {
			t.Fatalf("GetDefaultByType failed: %v", err)
		}

		if retrieved.ID != template2.ID {
			t.Errorf("Expected most recent template ID %d, got %d", template2.ID, retrieved.ID)
		}
	})

	t.Run("inactive default template not returned", func(t *testing.T) {
		template := &models.Template{
			Name:        "Inactive Default",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>Inactive</html>",
			IsDefault:   true,
			IsActive:    false,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		_, err := repo.GetDefaultByType(ctx, models.TemplateTypeConfirmationPage)
		if err == nil {
			t.Error("Expected NotFoundError for inactive default template")
		}
	})
}

func TestTemplateRepository_Integration_Filtering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	ctx := context.Background()
	textContent := "Email text"

	templates := []*models.Template{
		{
			EventID:     &eventID,
			Name:        "Event RSVP",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Event RSVP</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			Name:        "Default RSVP",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Default RSVP</html>",
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			Name:        "Default Email",
			Type:        models.TemplateTypeInviteEmail,
			HTMLContent: "<html>Email</html>",
			TextContent: &textContent,
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			Name:        "Inactive Template",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>Inactive</html>",
			IsDefault:   false,
			IsActive:    false,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(ctx, tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	t.Run("filter by event ID", func(t *testing.T) {
		filters := &TemplateFilters{
			EventID: &eventID,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 template, got %d", len(results))
		}

		if results[0].Name != "Event RSVP" {
			t.Errorf("Expected 'Event RSVP', got %s", results[0].Name)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		templateType := models.TemplateTypeRSVPPage
		filters := &TemplateFilters{
			Type: &templateType,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 RSVP templates, got %d", len(results))
		}
	})

	t.Run("filter by is_default", func(t *testing.T) {
		isDefault := true
		filters := &TemplateFilters{
			IsDefault: &isDefault,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 default templates, got %d", len(results))
		}

		for _, tmpl := range results {
			if !tmpl.IsDefault {
				t.Error("Non-default template in results")
			}
		}
	})

	t.Run("filter by is_active", func(t *testing.T) {
		isActive := true
		filters := &TemplateFilters{
			IsActive: &isActive,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 active templates, got %d", len(results))
		}

		for _, tmpl := range results {
			if !tmpl.IsActive {
				t.Error("Inactive template in results")
			}
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		templateType := models.TemplateTypeRSVPPage
		isDefault := true
		filters := &TemplateFilters{
			Type:      &templateType,
			IsDefault: &isDefault,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 template, got %d", len(results))
		}

		if results[0].Name != "Default RSVP" {
			t.Errorf("Expected 'Default RSVP', got %s", results[0].Name)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		filters := &TemplateFilters{
			Limit:  2,
			Offset: 0,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 templates, got %d", len(results))
		}

		filters.Offset = 2
		results, err = repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List with offset failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 templates on second page, got %d", len(results))
		}
	})
}

func TestTemplateRepository_Integration_ForeignKeyConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	ctx := context.Background()

	t.Run("invalid event_id rejected", func(t *testing.T) {
		invalidEventID := int64(99999)
		template := &models.Template{
			EventID:     &invalidEventID,
			Name:        "Invalid Event Template",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Test</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		err := repo.Create(ctx, template)
		if err == nil {
			t.Error("Expected error for invalid event_id")
		}
	})

	t.Run("invalid created_by rejected", func(t *testing.T) {
		template := &models.Template{
			Name:        "Invalid Creator Template",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Test</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   99999,
			Category:    models.CategoryPlain,
		}

		err := repo.Create(ctx, template)
		if err == nil {
			t.Error("Expected error for invalid created_by")
		}
	})

	t.Run("null event_id allowed for default templates", func(t *testing.T) {
		template := &models.Template{
			EventID:     nil,
			Name:        "Global Default",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>Global</html>",
			IsDefault:   true,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create with null event_id failed: %v", err)
		}

		if template.EventID != nil {
			t.Error("Expected event_id to remain nil")
		}
	})
}

func TestTemplateRepository_Integration_ActiveStatusToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	ctx := context.Background()

	t.Run("inactive template not returned by GetByEventAndType", func(t *testing.T) {
		template := &models.Template{
			EventID:     &eventID,
			Name:        "Active Toggle Test",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>Test</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		}

		if err := repo.Create(ctx, template); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		retrieved, err := repo.GetByEventAndType(ctx, eventID, models.TemplateTypeConfirmationPage)
		if err != nil {
			t.Fatalf("GetByEventAndType failed: %v", err)
		}

		if retrieved.ID != template.ID {
			t.Error("Expected to retrieve active template")
		}

		if err := repo.SetActive(ctx, template.ID, false); err != nil {
			t.Fatalf("SetActive failed: %v", err)
		}

		_, err = repo.GetByEventAndType(ctx, eventID, models.TemplateTypeConfirmationPage)
		if err == nil {
			t.Error("Expected NotFoundError for inactive template")
		}
	})
}

func TestTemplateRepository_Integration_VersionIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)

	ctx := context.Background()

	template := &models.Template{
		Name:        "Version Test",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<html>Version 1</html>",
		IsDefault:   false,
		IsActive:    true,
		CreatedBy:   user.ID,
		Category:    models.CategoryPlain,
	}

	if err := repo.Create(ctx, template); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if template.Version != 1 {
		t.Errorf("Expected initial version 1, got %d", template.Version)
	}

	for i := 2; i <= 5; i++ {
		template.HTMLContent = "<html>Version " + string(rune('0'+i)) + "</html>"
		if err := repo.Update(ctx, template); err != nil {
			t.Fatalf("Update %d failed: %v", i, err)
		}

		if template.Version != i {
			t.Errorf("Expected version %d, got %d", i, template.Version)
		}
	}

	retrieved, err := repo.GetByID(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Version != 5 {
		t.Errorf("Expected final version 5, got %d", retrieved.Version)
	}
}

func TestTemplateRepository_Integration_MultipleTemplatesPerEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewTemplateRepository(db)
	userRepo := NewUserRepository(db)
	user := createTestUser(t, userRepo)
	eventID := createTestEvent(t, db)

	ctx := context.Background()
	textContent := "Email text"

	templates := []*models.Template{
		{
			EventID:     &eventID,
			Name:        "Event RSVP Page",
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<html>RSVP</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			EventID:     &eventID,
			Name:        "Event Invite Email",
			Type:        models.TemplateTypeInviteEmail,
			HTMLContent: "<html>Invite</html>",
			TextContent: &textContent,
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
		{
			EventID:     &eventID,
			Name:        "Event Confirmation",
			Type:        models.TemplateTypeConfirmationPage,
			HTMLContent: "<html>Confirmation</html>",
			IsDefault:   false,
			IsActive:    true,
			CreatedBy:   user.ID,
			Category:    models.CategoryPlain,
		},
	}

	for _, tmpl := range templates {
		if err := repo.Create(ctx, tmpl); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	t.Run("retrieve each template type for event", func(t *testing.T) {
		rsvpTemplate, err := repo.GetByEventAndType(ctx, eventID, models.TemplateTypeRSVPPage)
		if err != nil {
			t.Fatalf("GetByEventAndType RSVP failed: %v", err)
		}
		if rsvpTemplate.Name != "Event RSVP Page" {
			t.Errorf("Expected 'Event RSVP Page', got %s", rsvpTemplate.Name)
		}

		emailTemplate, err := repo.GetByEventAndType(ctx, eventID, models.TemplateTypeInviteEmail)
		if err != nil {
			t.Fatalf("GetByEventAndType Email failed: %v", err)
		}
		if emailTemplate.Name != "Event Invite Email" {
			t.Errorf("Expected 'Event Invite Email', got %s", emailTemplate.Name)
		}

		confirmTemplate, err := repo.GetByEventAndType(ctx, eventID, models.TemplateTypeConfirmationPage)
		if err != nil {
			t.Fatalf("GetByEventAndType Confirmation failed: %v", err)
		}
		if confirmTemplate.Name != "Event Confirmation" {
			t.Errorf("Expected 'Event Confirmation', got %s", confirmTemplate.Name)
		}
	})

	t.Run("list all templates for event", func(t *testing.T) {
		filters := &TemplateFilters{
			EventID: &eventID,
		}

		results, err := repo.List(ctx, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 templates for event, got %d", len(results))
		}
	})
}
