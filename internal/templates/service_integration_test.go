package templates

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupServiceTestDB(t *testing.T) (db.Database, func()) {
	t.Helper()

	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		database.Close()
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx := context.Background()
	if err := migrator.Up(ctx); err != nil {
		database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		database.Close()
	}

	return database, cleanup
}

var userEmailCounter int64 = 0

func createServiceTestUser(t *testing.T, database db.Database, role models.UserRole) *models.User {
	t.Helper()
	userEmailCounter++

	user := &models.User{
		Email: fmt.Sprintf("test%d@example.com", userEmailCounter),
		Name:  fmt.Sprintf("Test User %d", userEmailCounter),
		Role:  role,
	}

	query := `
		INSERT INTO users (email, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := database.Exec(context.Background(), query, user.Email, user.Name, user.Role, now, now)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now

	return user
}

func createServiceTestEvent(t *testing.T, database db.Database, userID int64) *models.Event {
	t.Helper()

	event := &models.Event{
		Title:     "Test Event",
		StartTime: time.Now().Add(24 * time.Hour),
		Timezone:  "America/Los_Angeles",
		Status:    models.EventStatusDraft,
		CreatedBy: userID,
	}

	query := `
		INSERT INTO events (title, start_time, timezone, status, created_by, created_at, updated_at, version, ics_sequence, max_plus_ones)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, 0, 0)
	`

	now := time.Now()
	result, err := database.Exec(context.Background(), query,
		event.Title, event.StartTime, event.Timezone, event.Status, event.CreatedBy, now, now)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get event ID: %v", err)
	}

	event.ID = id
	return event
}

func TestTemplateService_Integration_FullCRUDFlow(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user := createServiceTestUser(t, database, models.RoleEventManager)
	ctx := auth.WithUser(context.Background(), user)

	textContent := "Event: {{.Event.Title}}"
	template := &models.Template{
		Name:        "Custom Invite",
		Type:        models.TemplateTypeInviteEmail,
		Description: "My custom template",
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err := service.CreateTemplate(ctx, template)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	if template.ID == 0 {
		t.Error("Expected template ID to be set")
	}

	if template.CreatedBy != user.ID {
		t.Errorf("CreatedBy = %d, want %d", template.CreatedBy, user.ID)
	}

	if !template.IsActive {
		t.Error("Expected template to be active")
	}

	retrieved, err := service.GetTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetTemplate() error = %v", err)
	}

	if retrieved.Name != template.Name {
		t.Errorf("Name = %s, want %s", retrieved.Name, template.Name)
	}

	template.Name = "Updated Custom Invite"
	template.HTMLContent = "<h1>Updated: {{.Event.Title}}</h1>"

	err = service.UpdateTemplate(ctx, template)
	if err != nil {
		t.Fatalf("UpdateTemplate() error = %v", err)
	}

	if template.Version != 2 {
		t.Errorf("Version = %d, want 2", template.Version)
	}

	retrieved, err = service.GetTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetTemplate() after update error = %v", err)
	}

	if retrieved.Name != "Updated Custom Invite" {
		t.Errorf("Updated name = %s, want 'Updated Custom Invite'", retrieved.Name)
	}

	err = service.DeleteTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("DeleteTemplate() error = %v", err)
	}

	_, err = service.GetTemplate(ctx, template.ID)
	if err == nil {
		t.Error("Expected NotFoundError after delete")
	}
}

func TestTemplateService_Integration_PermissionEnforcement(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user1 := createServiceTestUser(t, database, models.RoleEventManager)
	user2 := createServiceTestUser(t, database, models.RoleEventManager)
	admin := createServiceTestUser(t, database, models.RoleAdmin)

	ctx1 := auth.WithUser(context.Background(), user1)

	textContent := "Event: {{.Event.Title}}"
	template := &models.Template{
		Name:        "User1 Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err := service.CreateTemplate(ctx1, template)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	ctx2 := auth.WithUser(context.Background(), user2)
	template.Name = "User2 Update"
	err = service.UpdateTemplate(ctx2, template)
	if err == nil {
		t.Error("Expected error when user2 tries to update user1's template")
	}

	err = service.DeleteTemplate(ctx2, template.ID)
	if err == nil {
		t.Error("Expected error when user2 tries to delete user1's template")
	}

	ctxAdmin := auth.WithUser(context.Background(), admin)
	template.Name = "Admin Update"
	err = service.UpdateTemplate(ctxAdmin, template)
	if err != nil {
		t.Errorf("Admin should be able to update any template: %v", err)
	}

	err = service.DeleteTemplate(ctxAdmin, template.ID)
	if err != nil {
		t.Errorf("Admin should be able to delete any template: %v", err)
	}
}

func TestTemplateService_Integration_CannotDeleteDefaultTemplate(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	systemUser := createServiceTestUser(t, database, models.RoleAdmin)
	ctx := auth.WithUser(context.Background(), systemUser)

	err := seeder.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	defaultTemplate, err := repo.GetDefaultByType(context.Background(), models.TemplateTypeInviteEmail)
	if err != nil {
		t.Fatalf("GetDefaultByType() error = %v", err)
	}

	err = service.DeleteTemplate(ctx, defaultTemplate.ID)
	if err == nil {
		t.Error("Expected error when deleting default template")
	}

	if !containsString(err.Error(), "Cannot delete default") {
		t.Errorf("Error = %v, want to contain 'Cannot delete default'", err)
	}
}

func TestTemplateService_Integration_CannotDeleteTemplateInUse(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user := createServiceTestUser(t, database, models.RoleEventManager)
	ctx := auth.WithUser(context.Background(), user)

	textContent := "Event: {{.Event.Title}}"
	template := &models.Template{
		Name:        "Template In Use",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err := service.CreateTemplate(ctx, template)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	event := createServiceTestEvent(t, database, user.ID)

	updateQuery := `UPDATE events SET template_id = ? WHERE id = ?`
	_, err = database.Exec(context.Background(), updateQuery, template.ID, event.ID)
	if err != nil {
		t.Fatalf("Failed to link template to event: %v", err)
	}

	err = service.DeleteTemplate(ctx, template.ID)
	if err == nil {
		t.Error("Expected error when deleting template in use")
	}

	if !containsString(err.Error(), "in use") {
		t.Errorf("Error = %v, want to contain 'in use'", err)
	}
}

func TestTemplateService_Integration_SetDefault(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	admin := createServiceTestUser(t, database, models.RoleAdmin)
	ctx := auth.WithUser(context.Background(), admin)

	err := seeder.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	oldDefault, err := repo.GetDefaultByType(context.Background(), models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetDefaultByType() error = %v", err)
	}

	textContent := "Event: {{.Event.Title}}"
	newTemplate := &models.Template{
		Name:        "New Default",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>New Default: {{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err = service.CreateTemplate(ctx, newTemplate)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	err = service.SetDefault(ctx, newTemplate.ID)
	if err != nil {
		t.Fatalf("SetDefault() error = %v", err)
	}

	currentDefault, err := repo.GetDefaultByType(context.Background(), models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetDefaultByType() after SetDefault error = %v", err)
	}

	if currentDefault.ID != newTemplate.ID {
		t.Errorf("Current default ID = %d, want %d", currentDefault.ID, newTemplate.ID)
	}

	if !currentDefault.IsDefault {
		t.Error("Expected new template to be marked as default")
	}

	oldDefaultRetrieved, err := repo.GetByID(context.Background(), oldDefault.ID)
	if err != nil {
		t.Fatalf("GetByID() for old default error = %v", err)
	}

	if oldDefaultRetrieved.IsDefault {
		t.Error("Expected old default to no longer be marked as default")
	}
}

func TestTemplateService_Integration_GetTemplateForEvent(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user := createServiceTestUser(t, database, models.RoleEventManager)
	ctx := auth.WithUser(context.Background(), user)

	err := seeder.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	event := createServiceTestEvent(t, database, user.ID)

	defaultTemplate, err := service.GetTemplateForEvent(ctx, event.ID, models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetTemplateForEvent() error = %v", err)
	}

	if !defaultTemplate.IsDefault {
		t.Error("Expected to get default template when event has no custom template")
	}

	textContent := "Event: {{.Event.Title}}"
	eventID := event.ID
	customTemplate := &models.Template{
		EventID:     &eventID,
		Name:        "Custom Event Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>Custom: {{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err = service.CreateTemplate(ctx, customTemplate)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	customRetrieved, err := service.GetTemplateForEvent(ctx, event.ID, models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetTemplateForEvent() after custom template error = %v", err)
	}

	if customRetrieved.ID != customTemplate.ID {
		t.Errorf("Expected custom template ID %d, got %d", customTemplate.ID, customRetrieved.ID)
	}

	err = service.SetActive(ctx, customTemplate.ID, false)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	fallbackTemplate, err := service.GetTemplateForEvent(ctx, event.ID, models.TemplateTypeRSVPPage)
	if err != nil {
		t.Fatalf("GetTemplateForEvent() after deactivation error = %v", err)
	}

	if fallbackTemplate.ID != defaultTemplate.ID {
		t.Error("Expected to fall back to default template when custom is inactive")
	}
}

func TestTemplateService_Integration_ListTemplates(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	seeder := NewSeeder(repo, 1)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user := createServiceTestUser(t, database, models.RoleEventManager)
	ctx := auth.WithUser(context.Background(), user)

	err := seeder.SeedDefaults(context.Background())
	if err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	textContent := "Event: {{.Event.Title}}"
	for i := 0; i < 3; i++ {
		template := &models.Template{
			Name:        fmt.Sprintf("Custom Template %d", i+1),
			Type:        models.TemplateTypeRSVPPage,
			HTMLContent: "<h1>{{.Event.Title}}</h1>",
			TextContent: &textContent,
		}
		err := service.CreateTemplate(ctx, template)
		if err != nil {
			t.Fatalf("CreateTemplate() error = %v", err)
		}
	}

	allTemplates, err := service.ListTemplates(ctx, &repositories.TemplateFilters{})
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	if len(allTemplates) < 6 {
		t.Errorf("Expected at least 6 templates (3 defaults + 3 custom), got %d", len(allTemplates))
	}

	rsvpType := models.TemplateTypeRSVPPage
	rsvpTemplates, err := service.ListTemplates(ctx, &repositories.TemplateFilters{
		Type: &rsvpType,
	})
	if err != nil {
		t.Fatalf("ListTemplates() with type filter error = %v", err)
	}

	if len(rsvpTemplates) != 4 {
		t.Errorf("Expected 4 RSVP templates (1 default + 3 custom), got %d", len(rsvpTemplates))
	}

	userID := user.ID
	userTemplates, err := service.ListTemplates(ctx, &repositories.TemplateFilters{
		CreatedBy: &userID,
	})
	if err != nil {
		t.Fatalf("ListTemplates() with created_by filter error = %v", err)
	}

	if len(userTemplates) < 3 {
		t.Errorf("Expected at least 3 user templates, got %d", len(userTemplates))
	}
}

func TestTemplateService_Integration_ConcurrentOperations(t *testing.T) {
	database, cleanup := setupServiceTestDB(t)
	defer cleanup()

	repo := repositories.NewTemplateRepository(database)
	engine := NewEngine()
	validator := NewValidator(engine)
	service := NewService(repo, validator)

	user := createServiceTestUser(t, database, models.RoleEventManager)
	ctx := auth.WithUser(context.Background(), user)

	textContent := "Event: {{.Event.Title}}"
	template := &models.Template{
		Name:        "Concurrent Template",
		Type:        models.TemplateTypeRSVPPage,
		HTMLContent: "<h1>{{.Event.Title}}</h1>",
		TextContent: &textContent,
	}

	err := service.CreateTemplate(ctx, template)
	if err != nil {
		t.Fatalf("CreateTemplate() error = %v", err)
	}

	done := make(chan bool, 2)
	errors := make(chan error, 2)

	go func() {
		template1 := &models.Template{
			ID:          template.ID,
			Name:        "Update 1",
			Type:        template.Type,
			HTMLContent: "<h1>Update 1</h1>",
			TextContent: &textContent,
			CreatedBy:   user.ID,
		}
		err := service.UpdateTemplate(ctx, template1)
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	go func() {
		template2 := &models.Template{
			ID:          template.ID,
			Name:        "Update 2",
			Type:        template.Type,
			HTMLContent: "<h1>Update 2</h1>",
			TextContent: &textContent,
			CreatedBy:   user.ID,
		}
		err := service.UpdateTemplate(ctx, template2)
		if err != nil {
			errors <- err
		}
		done <- true
	}()

	<-done
	<-done
	close(errors)

	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Logf("Concurrent updates resulted in %d errors (expected due to version conflicts)", errorCount)
	}

	finalTemplate, err := service.GetTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetTemplate() after concurrent updates error = %v", err)
	}

	if finalTemplate.Version < 2 {
		t.Errorf("Expected version >= 2 after concurrent updates, got %d", finalTemplate.Version)
	}
}
