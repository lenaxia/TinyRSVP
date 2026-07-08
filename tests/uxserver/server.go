// Package uxserver provides a shared in-process HTTP test server for UX
// (browser-based) tests. Both the chromedp tests in tests/ux and the
// Playwright tests in tests/ux_playwright import this package to avoid
// duplicating the ~270 lines of handler/service/repository wiring.
//
// The server wires the full router with real handlers against a file-backed
// SQLite database, seeds an admin user and the default templates, and
// returns a Server value exposing the services that tests need to set up
// fixture data (events, invites).
package uxserver

import (
	"context"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
	"github.com/lenaxia/tinyrsvp/internal/storage"
	"github.com/lenaxia/tinyrsvp/internal/templates"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

// Paths are relative to the repository root. Both tests/ux and
// tests/ux_playwright are two levels deep, so "../.." reaches the root.
const (
	DefaultMigrationPath = "../../migrations/sqlite"
	DefaultTemplateBase  = "../../templates/web"
	DefaultStaticDir     = "../../static"
)

// Options lets callers override paths (useful when the package is imported
// from a different depth). Zero-value fields fall back to the Defaults above.
type Options struct {
	MigrationPath string
	TemplateBase  string
	StaticDir     string
}

// Server is the shared in-process test server.
type Server struct {
	HTTPServer          *httptest.Server
	Database            db.Database
	SessionMgr          auth.SessionManager
	UserService         auth.UserService
	EventService        events.Service
	InviteService       invites.IndividualInviteService
	InviteImportService invites.InviteService
	AdminUser           *models.User
}

// URL returns the absolute URL for a path on the test server.
func (s *Server) URL(path string) string {
	return s.HTTPServer.URL + path
}

// AdminUserID returns the admin user's ID as a string (for the
// X-Test-User-ID auth-bypass header).
func (s *Server) AdminUserID() string {
	return intToStr(s.AdminUser.ID)
}

func intToStr(i int64) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// Setup builds and returns a Server with the full router wired. The caller
// is responsible for closing the server (handled automatically via t.Cleanup
// when Setup is passed a *testing.T — but see SetupTest which does this).
func Setup(opts Options) (*Server, func(), error) {
	migrationPath := opts.MigrationPath
	if migrationPath == "" {
		migrationPath = DefaultMigrationPath
	}
	templateBase := opts.TemplateBase
	if templateBase == "" {
		templateBase = DefaultTemplateBase
	}
	staticDir := opts.StaticDir
	if staticDir == "" {
		staticDir = DefaultStaticDir
	}

	tmpFile, err := os.CreateTemp("", "tinyrsvp-ux-*.db")
	if err != nil {
		return nil, nil, err
	}
	tmpFile.Close()

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: tmpFile.Name(),
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	migrator, err := db.NewMigrator(database.DB(), migrationPath)
	if err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	emailQueueRepo := repositories.NewEmailQueueRepository(database)
	templateRepo := repositories.NewTemplateRepository(database)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	adminUser, err := userService.GetOrCreateUser(ctx, "ux-admin@test.example.com", "UX Test Admin", nil)
	if err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}
	if err := userService.UpdateUserRole(ctx, adminUser.ID, models.RoleAdmin); err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}
	adminUser.Role = models.RoleAdmin

	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)

	seeder := templates.NewSeeder(templateRepo, adminUser.ID)
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	if err := seeder.SeedDefaults(seedCtx); err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, nil, eventValidator, authChecker)

	customizationService := events.NewCustomizationService(eventRepo, templateRepo, authChecker)
	dashboardService := events.NewDashboardService(eventRepo, inviteRepo, rsvpRepo)
	questionValidator := events.NewQuestionValidator()
	questionService := events.NewQuestionService(eventRepo, questionRepo, questionValidator, authChecker)

	tokenSecret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tokenSecretBytes, _ := hex.DecodeString(tokenSecret)
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteImportService := invites.NewInviteService(tokenGenerator, inviteRepo)
	individualInviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

	requireAuth := middleware.TestRequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	eventHandlers := handlers.NewEventHandlers(eventService)
	questionHandlers := handlers.NewQuestionHandlers(questionService)
	templateHandlers := handlers.NewTemplateHandlers(templateService)
	customizationHandlers := handlers.NewEventCustomizationHandlers(customizationService)
	userHandler := handlers.NewUserHandler(userService, authChecker)

	adminService := admin.NewAdminService(userService, eventRepo, inviteRepo)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(adminService)
	userManagementHandler := handlers.NewUserManagementHandler(userService)

	// Settings + metrics handlers (added for Epic 10 stories 10-10 and 10-11).
	// Wired here so both chromedp and Playwright UX tests can exercise them.
	testConfig := &config.Config{
		Server:   config.ServerConfig{BaseURL: "http://localhost"},
		Database: config.DatabaseConfig{Type: "sqlite"},
	}
	settingsHandler := handlers.NewSettingsHandler(testConfig)

	emailChecker := email.NewHealthChecker(emailQueueRepo, nil)
	metricsDataSource := handlers.NewMetricsDataSource(adminService, emailChecker, database)
	adminMetricsHandler := handlers.NewMetricsHandler(metricsDataSource)

	metricsCollector := middleware.NewPrometheusMetrics()
	promMetricsHandler := middleware.MetricsHandler(metricsCollector)
	promMiddleware := middleware.PrometheusMetrics(metricsCollector)

	tmpStorageDir, err := os.MkdirTemp("", "tinyrsvp-ux-storage-*")
	if err != nil {
		database.Close()
		cleanup()
		return nil, nil, err
	}

	storageProvider, err := storage.NewProvider(&storage.Config{
		Type:     "local",
		BasePath: tmpStorageDir,
		BaseURL:  "http://localhost",
	})
	if err != nil {
		os.RemoveAll(tmpStorageDir)
		database.Close()
		cleanup()
		return nil, nil, err
	}
	imageService := assets.NewImageService(storageProvider)
	imageHandlers := handlers.NewImageHandlers(imageService, eventService, authChecker)
	assetHandler := handlers.NewAssetHandler(storageProvider)

	baseURL := "http://localhost"

	inviteHandlers := handlers.NewInviteHandlers(individualInviteService, baseURL)
	importInviteHandlers := handlers.NewImportInviteHandlers(inviteImportService, eventRepo, baseURL)
	manualInviteHandlers := handlers.NewManualInviteHandlers(inviteImportService, eventRepo, baseURL)
	revokeInviteHandlers := handlers.NewRevokeInviteHandlers(inviteImportService, eventRepo)
	regenerateInviteHandlers := handlers.NewRegenerateInviteTokenHandlers(inviteImportService, eventRepo)
	listInviteHandlers := handlers.NewListInviteHandlers(inviteImportService, eventRepo)
	getInviteHandlers := handlers.NewGetInviteHandlers(inviteImportService, eventRepo)
	updateInviteHandlers := handlers.NewUpdateInviteHandlers(inviteImportService, eventRepo)
	deleteInviteHandlers := handlers.NewDeleteInviteHandlers(inviteImportService, eventRepo)
	sendInviteHandlers := handlers.NewSendInviteHandlers(inviteImportService, eventRepo, emailQueueRepo, baseURL)
	cleanupHandler := handlers.NewCleanupHandler(inviteImportService)

	funcMap := BuildTemplateFuncMap()

	dashboardTemplates := mustParse("dashboard.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "dashboard.html"),
	)
	dashboardHandler.SetTemplates(dashboardTemplates)

	eventListTemplates := mustParse("event_list.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "event_list.html"),
	)

	eventFormTemplates := mustParse("event_form.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "partials/datetime_picker_panel.html"),
		filepath.Join(templateBase, "partials/rsvp_settings_panel.html"),
		filepath.Join(templateBase, "partials/theme_picker.html"),
		filepath.Join(templateBase, "partials/theme_preview_modal.html"),
		filepath.Join(templateBase, "partials/image_upload.html"),
		filepath.Join(templateBase, "partials/color_picker.html"),
		filepath.Join(templateBase, "event_form.html"),
	)

	eventDetailTemplates := mustParse("event_detail.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "event_detail.html"),
	)

	eventWebHandlers := handlers.NewEventWebHandlers(eventService, templateService, eventListTemplates, eventFormTemplates, eventDetailTemplates)

	customizationTemplates := mustParse("event_customization.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "event_customization.html"),
	)
	customizationHandlers.SetTemplates(customizationTemplates)

	inviteListTemplates := mustParse("invite_list.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "invite_list.html"),
	)
	inviteWebHandlers := handlers.NewInviteWebHandlers(inviteImportService, eventRepo)
	inviteWebHandlers.SetTemplates(inviteListTemplates)

	adminDashboardTemplates := mustParse("admin_dashboard.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_dashboard.html"),
	)
	adminDashboardHandler.SetTemplates(adminDashboardTemplates)

	adminSettingsTemplates := mustParse("admin_settings.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_settings.html"),
	)
	settingsHandler.SetTemplates(adminSettingsTemplates)

	adminMetricsTemplates := mustParse("admin_metrics.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_metrics.html"),
	)
	adminMetricsHandler.SetTemplates(adminMetricsTemplates)

	userManagementTemplates := mustParse("user_management.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "user_management.html"),
	)
	userManagementHandler.SetTemplates(userManagementTemplates)

	rsvpPageTemplates := mustParse("rsvp_page.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "rsvp_page.html"),
	)

	confirmationTemplates := mustParse("confirmation.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "confirmation.html"),
	)

	rsvpSummaryTemplates := mustParse("rsvp_summary.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "rsvp_summary.html"),
	)

	rsvpService := rsvp.NewServiceWithEmail(database, inviteImportService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, newNoopEmailService())

	rsvpHandler := handlers.NewRSVPHandler(inviteImportService, eventRepo, rsvpRepo, questionRepo)
	rsvpHandler.SetTemplates(rsvpPageTemplates)
	rsvpHandler.SetConfirmationTemplates(confirmationTemplates)
	rsvpHandler.SetRSVPService(rsvpService)
	rsvpHandler.SetAnswerRepository(answerRepo)
	rsvpHandler.SetTemplateRepository(templateRepo)
	rsvpHandler.SetTemplateService(templateService)
	rsvpHandler.SetCustomizationService(customizationService)

	rsvpSummaryHandler := handlers.NewRSVPSummaryHandler(eventRepo, rsvpRepo, questionRepo, answerRepo)
	rsvpSummaryHandler.SetTemplates(rsvpSummaryTemplates)

	staticFS := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))

	router := handlers.NewRouter(&handlers.RouterHandlers{
		HealthHandler:            handlers.NewHealthHandler("test"),
		DashboardHandler:         dashboardHandler,
		EventHandlers:            eventHandlers,
		EventWebHandlers:         eventWebHandlers,
		QuestionHandlers:         questionHandlers,
		InviteHandlers:           inviteHandlers,
		InviteWebHandlers:        inviteWebHandlers,
		ImportInviteHandlers:     importInviteHandlers,
		ManualInviteHandlers:     manualInviteHandlers,
		RevokeInviteHandlers:     revokeInviteHandlers,
		RegenerateInviteHandlers: regenerateInviteHandlers,
		ListInviteHandlers:       listInviteHandlers,
		GetInviteHandlers:        getInviteHandlers,
		UpdateInviteHandlers:     updateInviteHandlers,
		DeleteInviteHandlers:     deleteInviteHandlers,
		SendInviteHandlers:       sendInviteHandlers,
		ImageHandlers:            imageHandlers,
		RSVPHandler:              rsvpHandler,
		RSVPSummaryHandler:       rsvpSummaryHandler,
		UserHandler:              userHandler,
		AdminDashboardHandler:    adminDashboardHandler,
		UserManagementHandler:    userManagementHandler,
		SettingsHandler:          settingsHandler,
		AdminMetricsHandler:      adminMetricsHandler,
		TemplateHandlers:         templateHandlers,
		CustomizationHandlers:    customizationHandlers,
		AssetHandler:             assetHandler,
		CleanupHandler:           cleanupHandler,
		AuthMiddleware:           middlewareAdapter,
		MetricsHandler:           promMetricsHandler,
		MetricsMiddleware:        promMiddleware,
		StaticFileServer:         staticFS,
	})

	srv := httptest.NewServer(router)

	finalCleanup := func() {
		srv.Close()
		os.RemoveAll(tmpStorageDir)
		database.Close()
		os.Remove(tmpFile.Name())
	}

	return &Server{
		HTTPServer:          srv,
		Database:            database,
		SessionMgr:          sessionMgr,
		UserService:         userService,
		EventService:        eventService,
		InviteService:       individualInviteService,
		InviteImportService: inviteImportService,
		AdminUser:           adminUser,
	}, finalCleanup, nil
}

func mustParse(name string, funcMap template.FuncMap, files ...string) *template.Template {
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		panic("uxserver: failed to parse " + name + ": " + err.Error())
	}
	return tmpl
}

// noopEmailService is an email.Service stub that does nothing — used so the
// RSVP service can be wired without a real SMTP sender.
type noopEmailService struct{}

func newNoopEmailService() email.Service { return &noopEmailService{} }

func (n *noopEmailService) SendConfirmationEmail(_ context.Context, _ string, _ *models.RSVP, _ *models.Invite, _ *models.Event, _ []*models.RSVPAnswer) error {
	return nil
}
