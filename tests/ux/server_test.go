package ux

import (
	"context"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
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

const (
	migrationPath = "../../migrations/sqlite"
	templateBase  = "../../templates/web"
	staticDir     = "../../static"
	emailTemplate = "../../templates/email"
)

type uxTestServer struct {
	server              *httptest.Server
	database            db.Database
	sessionMgr          auth.SessionManager
	userService         auth.UserService
	eventService        events.Service
	inviteService       invites.IndividualInviteService
	inviteImportService invites.InviteService
	adminUser           *models.User
}

func setupUXTestServer(t *testing.T) *uxTestServer {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "tinyrsvp-ux-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: tmpFile.Name(),
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	migrator, err := db.NewMigrator(database.DB(), migrationPath)
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
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
		t.Fatalf("Failed to create admin user: %v", err)
	}
	if err := userService.UpdateUserRole(ctx, adminUser.ID, models.RoleAdmin); err != nil {
		t.Fatalf("Failed to set admin role: %v", err)
	}
	adminUser.Role = models.RoleAdmin

	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)

	seeder := templates.NewSeeder(templateRepo, adminUser.ID)
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	if err := seeder.SeedDefaults(seedCtx); err != nil {
		t.Fatalf("Failed to seed default templates: %v", err)
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

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
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

	tmpStorageDir, err := os.MkdirTemp("", "tinyrsvp-ux-storage-*")
	if err != nil {
		t.Fatalf("Failed to create storage dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpStorageDir) })

	storageProvider, err := storage.NewProvider(&storage.Config{
		Type:     "local",
		BasePath: tmpStorageDir,
		BaseURL:  "http://localhost",
	})
	if err != nil {
		t.Fatalf("Failed to create storage provider: %v", err)
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

	funcMap := buildTemplateFuncMap()

	dashboardTemplates := mustParseTemplates(t, "dashboard.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/partials/page_header.html",
		templateBase+"/dashboard.html",
	)
	dashboardHandler.SetTemplates(dashboardTemplates)

	eventListTemplates := mustParseTemplates(t, "event_list.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/partials/page_header.html",
		templateBase+"/event_list.html",
	)

	eventFormTemplates := mustParseTemplates(t, "event_form.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/partials/page_header.html",
		templateBase+"/partials/datetime_picker_panel.html",
		templateBase+"/partials/rsvp_settings_panel.html",
		templateBase+"/partials/theme_picker.html",
		templateBase+"/partials/theme_preview_modal.html",
		templateBase+"/partials/image_upload.html",
		templateBase+"/partials/color_picker.html",
		templateBase+"/event_form.html",
	)

	eventDetailTemplates := mustParseTemplates(t, "event_detail.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/partials/page_header.html",
		templateBase+"/event_detail.html",
	)

	eventWebHandlers := handlers.NewEventWebHandlers(eventService, templateService, eventListTemplates, eventFormTemplates, eventDetailTemplates)

	customizationTemplates := mustParseTemplates(t, "event_customization.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/event_customization.html",
	)
	customizationHandlers.SetTemplates(customizationTemplates)

	inviteListTemplates := mustParseTemplates(t, "invite_list.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/invite_list.html",
	)
	inviteWebHandlers := handlers.NewInviteWebHandlers(inviteImportService, eventRepo)
	inviteWebHandlers.SetTemplates(inviteListTemplates)

	adminDashboardTemplates := mustParseTemplates(t, "admin_dashboard.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/partials/page_header.html",
		templateBase+"/admin_dashboard.html",
	)
	adminDashboardHandler.SetTemplates(adminDashboardTemplates)

	userManagementTemplates := mustParseTemplates(t, "user_management.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/user_management.html",
	)
	userManagementHandler.SetTemplates(userManagementTemplates)

	rsvpPageTemplates := mustParseTemplates(t, "rsvp_page.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/rsvp_page.html",
	)

	confirmationTemplates := mustParseTemplates(t, "confirmation.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/confirmation.html",
	)

	rsvpSummaryTemplates := mustParseTemplates(t, "rsvp_summary.html", funcMap,
		templateBase+"/partials/base.html",
		templateBase+"/partials/navigation.html",
		templateBase+"/rsvp_summary.html",
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
		TemplateHandlers:         templateHandlers,
		CustomizationHandlers:    customizationHandlers,
		AssetHandler:             assetHandler,
		CleanupHandler:           cleanupHandler,
		AuthMiddleware:           middlewareAdapter,
		StaticFileServer:         staticFS,
	})

	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close() })

	return &uxTestServer{
		server:              srv,
		database:            database,
		sessionMgr:          sessionMgr,
		userService:         userService,
		eventService:        eventService,
		inviteService:       individualInviteService,
		inviteImportService: inviteImportService,
		adminUser:           adminUser,
	}
}

func (s *uxTestServer) url(path string) string {
	return s.server.URL + path
}

func (s *uxTestServer) adminUserID() string {
	return fmt.Sprintf("%d", s.adminUser.ID)
}

func newChromedpCtx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)

	combinedCancel := func() {
		timeoutCancel()
		cancel()
		allocCancel()
	}

	t.Cleanup(combinedCancel)
	return ctx, combinedCancel
}

func asAdmin(userID string) chromedp.Action {
	return network.SetExtraHTTPHeaders(network.Headers{
		"X-Test-User-ID": userID,
	})
}

func mustParseTemplates(t *testing.T, name string, funcMap template.FuncMap, files ...string) *template.Template {
	t.Helper()
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		t.Fatalf("Failed to parse template %q: %v", name, err)
	}
	return tmpl
}

func buildTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"until": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"formatDateTime": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006 at 3:04 PM MST")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("3:04 PM MST")
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			d := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				d[key] = values[i+1]
			}
			return d, nil
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"timezoneAbbr": func(iana string) string {
			loc, err := time.LoadLocation(iana)
			if err != nil {
				return iana
			}
			return time.Now().In(loc).Format("MST")
		},
	}
}

type noopEmailService struct{}

func newNoopEmailService() email.Service {
	return &noopEmailService{}
}

func (n *noopEmailService) SendConfirmationEmail(ctx context.Context, token string, rsvpRecord *models.RSVP, invite *models.Invite, event *models.Event, answers []*models.RSVPAnswer) error {
	return nil
}
