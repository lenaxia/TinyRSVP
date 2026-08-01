package handlers

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

// TestCoverageConstructors verifies that every handler constructor returns a
// non-nil instance. These constructors simply store their dependencies, so nil
// dependencies are acceptable for the non-nil assertion.
func TestCoverageConstructors(t *testing.T) {
	cases := []struct {
		name string
		fn   func() interface{}
	}{
		{"AdminDashboard", func() interface{} { return NewAdminDashboardHandler(nil) }},
		{"UserManagement", func() interface{} { return NewUserManagementHandler(nil) }},
		{"EventCustomization", func() interface{} { return NewEventCustomizationHandlers(nil) }},
		{"Invite", func() interface{} { return NewInviteHandlers(nil, "") }},
		{"ImportInvite", func() interface{} { return NewImportInviteHandlers(nil, nil, "") }},
		{"DeleteInvite", func() interface{} { return NewDeleteInviteHandlers(nil, nil) }},
		{"GetInvite", func() interface{} { return NewGetInviteHandlers(nil, nil) }},
		{"ManualInvite", func() interface{} { return NewManualInviteHandlers(nil, nil, "") }},
		{"RegenerateInviteToken", func() interface{} { return NewRegenerateInviteTokenHandlers(nil, nil) }},
		{"RevokeInvite", func() interface{} { return NewRevokeInviteHandlers(nil, nil) }},
		{"SendInvite", func() interface{} { return NewSendInviteHandlers(nil, nil, nil, "") }},
		{"UpdateInvite", func() interface{} { return NewUpdateInviteHandlers(nil, nil) }},
		{"Metrics", func() interface{} { return NewMetricsHandler(nil) }},
		{"Question", func() interface{} { return NewQuestionHandlers(nil) }},
		{"Settings", func() interface{} { return NewSettingsHandler(&config.Config{}) }},
		{"TemplateEditor", func() interface{} { return NewTemplateEditorHandlers(nil) }},
		{"Template", func() interface{} { return NewTemplateHandlers(nil) }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.fn(); got == nil {
				t.Errorf("%s constructor returned nil", c.name)
			}
		})
	}
}

// TestCoverageSetTemplates verifies each SetTemplates method stores the template
// pointer on the handler. The test lives in the same package so it can read the
// unexported templates field directly.
func TestCoverageSetTemplates(t *testing.T) {
	tmpl := template.Must(template.New("t").Parse("hello"))

	adminH := NewAdminDashboardHandler(nil)
	adminH.SetTemplates(tmpl)
	if adminH.templates == nil {
		t.Error("AdminDashboardHandler.templates not set")
	}

	userH := NewUserManagementHandler(nil)
	userH.SetTemplates(tmpl)
	if userH.templates == nil {
		t.Error("UserManagementHandler.templates not set")
	}

	custH := NewEventCustomizationHandlers(nil)
	custH.SetTemplates(tmpl)
	if custH.templates == nil {
		t.Error("EventCustomizationHandlers.templates not set")
	}

	metricsH := NewMetricsHandler(nil)
	metricsH.SetTemplates(tmpl)
	if metricsH.templates == nil {
		t.Error("MetricsHandler.templates not set")
	}

	settingsH := NewSettingsHandler(&config.Config{})
	settingsH.SetTemplates(tmpl)
	if settingsH.templates == nil {
		t.Error("SettingsHandler.templates not set")
	}

	editorH := NewTemplateEditorHandlers(nil)
	editorH.SetTemplates(tmpl)
	if editorH.templates == nil {
		t.Error("TemplateEditorHandlers.templates not set")
	}
}

// TestCoverageRegisterRoutes verifies each RegisterRoutes method registers its
// routes on a chi router without panicking. The handlers' dependencies are not
// consulted during route registration, so nil dependencies are acceptable.
func TestCoverageRegisterRoutes(t *testing.T) {
	cases := []struct {
		name string
		fn   func(r chi.Router)
	}{
		{"EventCustomization", func(r chi.Router) { NewEventCustomizationHandlers(nil).RegisterRoutes(r) }},
		{"Invite", func(r chi.Router) { NewInviteHandlers(nil, "").RegisterRoutes(r) }},
		{"ImportInvite", func(r chi.Router) { NewImportInviteHandlers(nil, nil, "").RegisterRoutes(r) }},
		{"DeleteInvite", func(r chi.Router) { NewDeleteInviteHandlers(nil, nil).RegisterRoutes(r) }},
		{"GetInvite", func(r chi.Router) { NewGetInviteHandlers(nil, nil).RegisterRoutes(r) }},
		{"ManualInvite", func(r chi.Router) { NewManualInviteHandlers(nil, nil, "").RegisterRoutes(r) }},
		{"RegenerateInviteToken", func(r chi.Router) { NewRegenerateInviteTokenHandlers(nil, nil).RegisterRoutes(r) }},
		{"RevokeInvite", func(r chi.Router) { NewRevokeInviteHandlers(nil, nil).RegisterRoutes(r) }},
		{"SendInvite", func(r chi.Router) { NewSendInviteHandlers(nil, nil, nil, "").RegisterRoutes(r) }},
		{"UpdateInvite", func(r chi.Router) { NewUpdateInviteHandlers(nil, nil).RegisterRoutes(r) }},
		{"Question", func(r chi.Router) { NewQuestionHandlers(nil).RegisterRoutes(r) }},
		{"TemplateEditor", func(r chi.Router) { NewTemplateEditorHandlers(nil).RegisterRoutes(r) }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RegisterRoutes panicked: %v", r)
				}
			}()
			c.fn(chi.NewRouter())
		})
	}
}

// --- metrics_adapter.go adapter tests ---

// mockEmailHealthChecker implements EmailHealthChecker for adapter tests.
type mockEmailHealthChecker struct {
	getStatusFunc func(ctx context.Context) (*email.HealthStatus, error)
}

func (m *mockEmailHealthChecker) GetStatus(ctx context.Context) (*email.HealthStatus, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx)
	}
	return &email.HealthStatus{Healthy: true}, nil
}

// mockMetricsDatabase implements db.Database for adapter tests. Only DB() is
// exercised by GetDBStats; the remaining methods return zero values.
type mockMetricsDatabase struct {
	db *sql.DB
}

func (m *mockMetricsDatabase) DB() *sql.DB { return m.db }
func (m *mockMetricsDatabase) Close() error { return nil }
func (m *mockMetricsDatabase) Ping(ctx context.Context) error { return nil }
func (m *mockMetricsDatabase) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	return nil
}
func (m *mockMetricsDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *mockMetricsDatabase) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *mockMetricsDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}

// mockAdminStatsSource implements AdminDashboardService for adapter tests.
type mockAdminStatsSource struct {
	getAdminStatsFunc func(ctx context.Context) (*AdminDashboardStats, error)
}

func (m *mockAdminStatsSource) GetAdminStats(ctx context.Context) (*AdminDashboardStats, error) {
	if m.getAdminStatsFunc != nil {
		return m.getAdminStatsFunc(ctx)
	}
	return &AdminDashboardStats{TotalUsers: 1, TotalEvents: 2, TotalInvites: 3}, nil
}

func TestCoverageMetricsDataSource_New(t *testing.T) {
	src := NewMetricsDataSource(&mockAdminStatsSource{}, &mockEmailHealthChecker{}, &mockMetricsDatabase{})
	if src == nil {
		t.Fatal("NewMetricsDataSource returned nil")
	}
}

func TestCoverageMetricsDataSource_GetAdminStats_Delegates(t *testing.T) {
	want := &AdminDashboardStats{TotalUsers: 7, TotalEvents: 3, TotalInvites: 12}
	adminSvc := &mockAdminStatsSource{
		getAdminStatsFunc: func(ctx context.Context) (*AdminDashboardStats, error) {
			return want, nil
		},
	}
	src := NewMetricsDataSource(adminSvc, &mockEmailHealthChecker{}, &mockMetricsDatabase{})

	got, err := src.GetAdminStats(context.Background())
	if err != nil {
		t.Fatalf("GetAdminStats returned error: %v", err)
	}
	if got.TotalUsers != 7 || got.TotalEvents != 3 || got.TotalInvites != 12 {
		t.Errorf("GetAdminStats mapped fields = %+v, want %+v", got, want)
	}
}

func TestCoverageMetricsDataSource_GetAdminStats_Error(t *testing.T) {
	wantErr := errors.New("admin stats unavailable")
	adminSvc := &mockAdminStatsSource{
		getAdminStatsFunc: func(ctx context.Context) (*AdminDashboardStats, error) {
			return nil, wantErr
		},
	}
	src := NewMetricsDataSource(adminSvc, &mockEmailHealthChecker{}, &mockMetricsDatabase{})

	got, err := src.GetAdminStats(context.Background())
	if !errors.Is(err, wantErr) {
		t.Errorf("GetAdminStats error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("GetAdminStats result = %v, want nil on error", got)
	}
}

func TestCoverageMetricsDataSource_GetEmailQueueStatus_Delegates(t *testing.T) {
	emailChk := &mockEmailHealthChecker{
		getStatusFunc: func(ctx context.Context) (*email.HealthStatus, error) {
			return &email.HealthStatus{
				Healthy:      true,
				QueueSize:    4,
				SendingCount: 2,
				FailedCount:  1,
			}, nil
		},
	}
	src := NewMetricsDataSource(&mockAdminStatsSource{}, emailChk, &mockMetricsDatabase{})

	got, err := src.GetEmailQueueStatus(context.Background())
	if err != nil {
		t.Fatalf("GetEmailQueueStatus returned error: %v", err)
	}
	if got.QueueSize != 4 || got.SendingCount != 2 || got.FailedCount != 1 || !got.Healthy {
		t.Errorf("GetEmailQueueStatus mapped fields = %+v", got)
	}
}

func TestCoverageMetricsDataSource_GetEmailQueueStatus_Error(t *testing.T) {
	wantErr := errors.New("email checker unavailable")
	emailChk := &mockEmailHealthChecker{
		getStatusFunc: func(ctx context.Context) (*email.HealthStatus, error) {
			return nil, wantErr
		},
	}
	src := NewMetricsDataSource(&mockAdminStatsSource{}, emailChk, &mockMetricsDatabase{})

	got, err := src.GetEmailQueueStatus(context.Background())
	if !errors.Is(err, wantErr) {
		t.Errorf("GetEmailQueueStatus error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Errorf("GetEmailQueueStatus result = %v, want nil on error", got)
	}
}

func TestCoverageMetricsDataSource_GetDBStats_Delegates(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available: %v", err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(3)

	database := &mockMetricsDatabase{db: sqlDB}
	src := NewMetricsDataSource(&mockAdminStatsSource{}, &mockEmailHealthChecker{}, database)

	got, err := src.GetDBStats()
	if err != nil {
		t.Fatalf("GetDBStats returned error: %v", err)
	}
	if got == nil {
		t.Fatal("GetDBStats returned nil metrics")
	}
	if got.MaxOpenConnections != 3 {
		t.Errorf("MaxOpenConnections = %d, want 3", got.MaxOpenConnections)
	}
}

// --- settings.go tests ---

func TestCoverageSettings_NewHandler(t *testing.T) {
	h := NewSettingsHandler(&config.Config{})
	if h == nil {
		t.Fatal("NewSettingsHandler returned nil")
	}
}

func TestCoverageSettings_SetTemplates(t *testing.T) {
	tmpl := template.Must(template.New("t").Parse("settings"))
	h := NewSettingsHandler(&config.Config{})
	h.SetTemplates(tmpl)
	if h.templates == nil {
		t.Error("SettingsHandler.templates not set")
	}
}

func TestCoverageSettings_Page_WithoutTemplates(t *testing.T) {
	h := NewSettingsHandler(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
	user := &models.User{ID: 1, Email: "admin@example.com", Name: "Admin", Role: models.RoleAdmin}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	h.SettingsPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (nil templates are an internal error)", w.Code, http.StatusInternalServerError)
	}
}

func TestCoverageSettings_Page_WithTemplates(t *testing.T) {
	tmpl := template.Must(template.New("admin_settings.html").Parse(`<p>{{.ActivePage}}</p>`))
	h := NewSettingsHandler(&config.Config{})
	h.SetTemplates(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
	user := &models.User{ID: 1, Email: "admin@example.com", Name: "Admin", Role: models.RoleAdmin}
	req = req.WithContext(auth.WithUser(req.Context(), user))

	w := httptest.NewRecorder()
	h.SettingsPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCoverageSettings_Page_Unauthorized(t *testing.T) {
	h := NewSettingsHandler(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
	req.Header.Set("Accept", "application/json")

	w := httptest.NewRecorder()
	h.SettingsPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// --- events_id_resolver.go: resolveEventIDFromRepo ---

// covEventRepo embeds mockRSVPEventRepository so it still satisfies
// repositories.EventRepository, while overriding the fallthrough lookups that
// the base mock hardcodes to (nil, nil).
type covEventRepo struct {
	*mockRSVPEventRepository
	getByPublicIDFunc     func(ctx context.Context, publicID string) (*models.Event, error)
	getByFriendlyNameFunc func(ctx context.Context, friendlyName string) (*models.Event, error)
}

func (m *covEventRepo) GetByPublicID(ctx context.Context, publicID string) (*models.Event, error) {
	if m.getByPublicIDFunc != nil {
		return m.getByPublicIDFunc(ctx, publicID)
	}
	return nil, &models.NotFoundError{Resource: "Event", ID: publicID}
}

func (m *covEventRepo) GetByFriendlyName(ctx context.Context, friendlyName string) (*models.Event, error) {
	if m.getByFriendlyNameFunc != nil {
		return m.getByFriendlyNameFunc(ctx, friendlyName)
	}
	return nil, &models.NotFoundError{Resource: "Event", ID: friendlyName}
}

func TestCoverageResolveEventIDFromRepo(t *testing.T) {
	repo := &covEventRepo{
		mockRSVPEventRepository: &mockRSVPEventRepository{
			getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
				if id == 123 {
					return &models.Event{ID: 123, Title: "Found by ID"}, nil
				}
				return nil, &models.NotFoundError{Resource: "Event", ID: id}
			},
		},
	}

	t.Run("found by numeric id", func(t *testing.T) {
		event, err := resolveEventIDFromRepo(context.Background(), repo, "123")
		if err != nil {
			t.Fatalf("resolveEventIDFromRepo error = %v", err)
		}
		if event == nil || event.ID != 123 {
			t.Errorf("resolveEventIDFromRepo event = %+v, want ID 123", event)
		}
	})

	t.Run("not found", func(t *testing.T) {
		event, err := resolveEventIDFromRepo(context.Background(), repo, "nonexistent-event")
		if err == nil {
			t.Error("resolveEventIDFromRepo expected error, got nil")
		}
		if event != nil {
			t.Errorf("resolveEventIDFromRepo event = %+v, want nil", event)
		}
	})
}

// --- rsvp.go: GetCalendar ---

func TestCoverageGetCalendar(t *testing.T) {
	inviteSvc := &mockRSVPInviteService{
		getInviteByTokenFunc: func(ctx context.Context, token string) (*models.Invite, error) {
			return &models.Invite{ID: 1, EventID: 10}, nil
		},
	}
	eventRepo := &mockRSVPEventRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Event, error) {
			return &models.Event{
				ID:        10,
				Title:     "Test Event",
				StartTime: time.Now().Add(24 * time.Hour),
				Timezone:  "UTC",
			}, nil
		},
	}
	h := NewRSVPHandler(inviteSvc, eventRepo, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/rsvp/abc123/calendar", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "abc123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetCalendar(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/calendar; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/calendar", ct)
	}
}

func TestCoverageGetCalendar_NoToken(t *testing.T) {
	h := NewRSVPHandler(&mockRSVPInviteService{}, &mockRSVPEventRepository{}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/rsvp//calendar", nil)
	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetCalendar(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- template_editor.go: parseTemplateIDFromPath ---

func TestCoverageParseTemplateIDFromPath(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantID  int64
		wantErr bool
	}{
		{"valid", "42", 42, false},
		{"non-numeric", "abc", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-5", 0, true},
		{"empty", "", 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			id, err := parseTemplateIDFromPath(c.input)
			if (err != nil) != c.wantErr {
				t.Errorf("parseTemplateIDFromPath(%q) error = %v, wantErr %v", c.input, err, c.wantErr)
			}
			if id != c.wantID {
				t.Errorf("parseTemplateIDFromPath(%q) id = %d, want %d", c.input, id, c.wantID)
			}
		})
	}
}

// --- templates.go: SetDefault ---

func TestCoverageTemplateHandlers_SetDefault(t *testing.T) {
	var calledID int64
	svc := &mockTemplateService{
		SetDefaultFunc: func(ctx context.Context, id int64) error {
			calledID = id
			return nil
		},
	}
	h := NewTemplateHandlers(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/templates/5/set-default", nil)
	req = req.WithContext(auth.WithUser(context.Background(), &models.User{
		ID: 1, Email: "mgr@example.com", Name: "Manager", Role: models.RoleEventManager,
	}))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.SetDefault(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if calledID != 5 {
		t.Errorf("service.SetDefault called with id %d, want 5", calledID)
	}
}

// --- event_customization.go: CustomizationPage ---

func TestCoverageCustomizationPage_WithTemplates(t *testing.T) {
	tmpl := template.Must(template.New("event_customization.html").Parse(`event:{{.EventID}}`))
	h := NewEventCustomizationHandlers(nil)
	h.SetTemplates(tmpl)

	req := httptest.NewRequest(http.MethodGet, "/events/42/template/customization", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.CustomizationPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCoverageCustomizationPage_NoTemplates(t *testing.T) {
	h := NewEventCustomizationHandlers(nil)

	req := httptest.NewRequest(http.MethodGet, "/events/42/template/customization", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.CustomizationPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// Compile-time assertions that the adapter mocks satisfy their interfaces.
var (
	_ EmailHealthChecker = (*mockEmailHealthChecker)(nil)
	_ db.Database        = (*mockMetricsDatabase)(nil)
	_ AdminDashboardService = (*mockAdminStatsSource)(nil)
)
