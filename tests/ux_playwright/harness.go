package ux_playwright

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	pw "github.com/mxschmitt/playwright-go"

	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/templates"
)

const (
	migrationPath = "../../migrations/sqlite"
	templateBase  = "../../templates/web"
	staticDir     = "../../static"
)

// PWServer wraps an httptest.Server with the admin user ID for auth bypass.
type PWServer struct {
	Server     *httptest.Server
	AdminID    int64
	Database   db.Database
	EventSvc   events.Service
}

func (s *PWServer) URL(path string) string {
	return s.Server.URL + path
}

func (s *PWServer) AdminIDStr() string {
	return strconv.FormatInt(s.AdminID, 10)
}

// SetupTestServer builds an in-process server with the full router wired
// (auth, dashboard, admin, settings, metrics, events, RSVP) and a real
// SQLite database seeded with one admin user.
func SetupTestServer(t *testing.T) *PWServer {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "tinyrsvp-pw-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: tmpFile.Name(),
	})
	if err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	migrator, err := db.NewMigrator(database.DB(), migrationPath)
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}
	migCtx, migCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer migCancel()
	if err := migrator.Up(migCtx); err != nil {
		t.Fatalf("run migrations: %v", err)
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
	_ = questionRepo
	_ = rsvpRepo
	_ = answerRepo
	_ = inviteRepo

	userService := auth.NewUserService(userRepo)
	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	authChecker := auth.NewAuthorizationChecker()

	migCtx2, migCancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer migCancel2()
	adminUser, err := userService.GetOrCreateUser(migCtx2, "pw-admin@test.example.com", "PW Test Admin", nil)
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	if err := userService.UpdateUserRole(migCtx2, adminUser.ID, models.RoleAdmin); err != nil {
		t.Fatalf("promote admin: %v", err)
	}
	adminUser.Role = models.RoleAdmin

	seeder := templates.NewSeeder(templateRepo, adminUser.ID)
	if err := seeder.SeedDefaults(migCtx2); err != nil {
		t.Fatalf("seed templates: %v", err)
	}

	funcMap := buildFuncMap()

	dashboardService := events.NewDashboardService(eventRepo, inviteRepo, rsvpRepo)
	dashboardTemplates := mustParse(t, "dashboard.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "dashboard.html"),
	)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	dashboardHandler.SetTemplates(dashboardTemplates)

	adminSvc := admin.NewAdminService(userService, eventRepo, inviteRepo)
	adminTemplates := mustParse(t, "admin_dashboard.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_dashboard.html"),
	)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(adminSvc)
	adminDashboardHandler.SetTemplates(adminTemplates)

	settingsTemplates := mustParse(t, "admin_settings.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_settings.html"),
	)
	cfg := &config.Config{
		Server:   config.ServerConfig{BaseURL: "http://pw.test"},
		Database: config.DatabaseConfig{Type: "sqlite", Path: tmpFile.Name()},
		OIDC:     config.OIDCConfig{Enabled: false},
	}
	settingsHandler := handlers.NewSettingsHandler(cfg)
	settingsHandler.SetTemplates(settingsTemplates)

	metricsTemplates := mustParse(t, "admin_metrics.html", funcMap,
		filepath.Join(templateBase, "partials/base.html"),
		filepath.Join(templateBase, "partials/navigation.html"),
		filepath.Join(templateBase, "partials/page_header.html"),
		filepath.Join(templateBase, "admin_metrics.html"),
	)
	emailChecker := email.NewHealthChecker(emailQueueRepo, nil)
	metricsDataSource := handlers.NewMetricsDataSource(adminSvc, emailChecker, database)
	metricsHandler := handlers.NewMetricsHandler(metricsDataSource)
	metricsHandler.SetTemplates(metricsTemplates)

	metricsCollector := middleware.NewPrometheusMetrics()
	promMetricsHandler := middleware.MetricsHandler(metricsCollector)
	promMiddleware := middleware.PrometheusMetrics(metricsCollector)

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	staticFS := http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))

	router := handlers.NewRouter(&handlers.RouterHandlers{
		DashboardHandler:      dashboardHandler,
		AdminDashboardHandler: adminDashboardHandler,
		SettingsHandler:       settingsHandler,
		AdminMetricsHandler:   metricsHandler,
		AuthMiddleware:        middlewareAdapter,
		MetricsHandler:        promMetricsHandler,
		MetricsMiddleware:     promMiddleware,
		StaticFileServer:      staticFS,
	})

	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close() })

	return &PWServer{
		Server:   srv,
		AdminID:  adminUser.ID,
		Database: database,
	}
}

// NewBrowser launches a headless Chromium browser via Playwright.
// Sets up automatic cleanup.
func NewBrowser(t *testing.T) (pw.Browser, pw.BrowserContext) {
	t.Helper()

	pwInst, err := pw.Run()
	if err != nil {
		t.Fatalf("launch playwright: %v", err)
	}
	t.Cleanup(func() { pwInst.Stop() })

	browser, err := pwInst.Chromium.Launch(
		pw.BrowserTypeLaunchOptions{
			Args: []string{"--no-sandbox", "--disable-dev-shm-usage"},
		},
	)
	if err != nil {
		t.Fatalf("launch chromium: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	context, err := browser.NewContext()
	if err != nil {
		t.Fatalf("create context: %v", err)
	}
	t.Cleanup(func() { context.Close() })

	return browser, context
}

// AsAdminPage navigates to a URL as the admin user by injecting the
// X-Test-User-ID header on every request in this context.
func AsAdminPage(t *testing.T, ctx pw.BrowserContext, srv *PWServer, path string) pw.Page {
	t.Helper()

	if err := ctx.SetExtraHTTPHeaders(map[string]string{
		"X-Test-User-ID": srv.AdminIDStr(),
	}); err != nil {
		t.Fatalf("set extra headers on context: %v", err)
	}

	page, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}

	if _, err := page.Goto(srv.URL(path), pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
	}); err != nil {
		t.Fatalf("goto %s: %v", path, err)
	}

	return page
}

// AssertContainsText fails the test if the page's text content does not
// contain the given substring.
func AssertContainsText(t *testing.T, page pw.Page, want string) {
	t.Helper()
	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body text: %v", err)
	}
	if !strings.Contains(body, want) {
		t.Errorf("page body does not contain %q\nbody:\n%s", want, truncate(body, 500))
	}
}

// AssertNotContainsText fails the test if the page's text content contains
// the given substring (used for secret-leak checks).
func AssertNotContainsText(t *testing.T, page pw.Page, forbidden string) {
	t.Helper()
	body, err := page.Locator("body").TextContent()
	if err != nil {
		t.Fatalf("read body text: %v", err)
	}
	if strings.Contains(body, forbidden) {
		t.Errorf("page body contains forbidden text %q", forbidden)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "... (truncated)"
}

func mustParse(t *testing.T, name string, funcMap template.FuncMap, files ...string) *template.Template {
	t.Helper()
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return tmpl
}

func buildFuncMap() template.FuncMap {
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
		"until": func(n int) []int {
			r := make([]int, n)
			for i := range r {
				r[i] = i
			}
			return r
		},
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"dict": func(pairs ...interface{}) map[string]interface{} {
			m := make(map[string]interface{}, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				k, _ := pairs[i].(string)
				m[k] = pairs[i+1]
			}
			return m
		},
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"formatDateTime": func(t time.Time) string {
			return t.Format("January 2, 2006 at 3:04 PM")
		},
	}
}

var _ = fmt.Sprintf
