// Package integration provides end-to-end HTTP integration tests that verify
// the merged PRs (admin pages, metrics middleware, settings redaction) work
// correctly through the real router with real handlers against an in-process
// httptest.Server. Uses the X-Test-User-ID auth bypass via plain HTTP — no
// browser dependency.
package integration

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/admin"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

func setupVerificationDB(t *testing.T) (db.Database, repositories.UserRepository, repositories.EventRepository, repositories.InviteRepository, repositories.EmailQueueRepository, repositories.SessionRepository, int64) {
	t.Helper()
	database, err := db.NewDatabase(db.Config{
		Type:         "sqlite",
		Path:         ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
		MaxLifetime:  time.Hour,
	})
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("migrator: %v", err)
	}
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	emailQueueRepo := repositories.NewEmailQueueRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)

	adminUser := &models.User{
		Email: "admin@verify.test",
		Name:  "Verify Admin",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	return database, userRepo, eventRepo, inviteRepo, emailQueueRepo, sessionRepo, adminUser.ID
}

func mustParseTemplates(t *testing.T, name string, files ...string) *template.Template {
	t.Helper()
	funcMap := template.FuncMap{
		"dict": func(pairs ...interface{}) (map[string]interface{}, error) {
			if len(pairs)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			m := make(map[string]interface{}, len(pairs)/2)
			for i := 0; i < len(pairs); i += 2 {
				k, ok := pairs[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[k] = pairs[i+1]
			}
			return m, nil
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
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
	}
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return tmpl
}

func setupVerificationServer(t *testing.T) (*httptest.Server, int64, db.Database) {
	t.Helper()
	database, userRepo, eventRepo, inviteRepo, emailQueueRepo, sessionRepo, adminUserID := setupVerificationDB(t)

	userService := auth.NewUserService(userRepo)
	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	authChecker := auth.NewAuthorizationChecker()
	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	adminSvc := admin.NewAdminService(userService, eventRepo, inviteRepo)
	adminDashboardHandler := handlers.NewAdminDashboardHandler(adminSvc)
	adminDashboardHandler.SetTemplates(mustParseTemplates(t, "admin_dashboard.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
		"../../templates/web/partials/page_header.html",
		"../../templates/web/admin_dashboard.html",
	))

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "0.0.0.0", Port: 8080, BaseURL: "http://verify.test"},
		Database: config.DatabaseConfig{Type: "sqlite", Path: ":memory:", MaxOpenConns: 1, MaxIdleConns: 1},
		Email: config.EmailConfig{SMTPHost: "smtp.verify.test", SMTPPort: 587, SMTPUser: "u", SMTPPassword: "secret", FromEmail: "r@verify.test"},
		OIDC: config.OIDCConfig{Enabled: true, IssuerURL: "https://idp.verify.test", ClientID: "verify-client", ClientSecret: "oidc-secret-value"},
		Security: config.SecurityConfig{HMACSecretKey: "hmac-verify-key", SessionDuration: 7 * 24 * time.Hour},
		Token: config.TokenConfig{Secret: "token-secret-value", HashingEnabled: true},
	}
	settingsHandler := handlers.NewSettingsHandler(cfg)
	settingsHandler.SetTemplates(mustParseTemplates(t, "admin_settings.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
		"../../templates/web/partials/page_header.html",
		"../../templates/web/admin_settings.html",
	))

	emailChecker := email.NewHealthChecker(emailQueueRepo, nil)
	metricsDataSource := handlers.NewMetricsDataSource(adminSvc, emailChecker, database)
	metricsHandler := handlers.NewMetricsHandler(metricsDataSource)
	metricsHandler.SetTemplates(mustParseTemplates(t, "admin_metrics.html",
		"../../templates/web/partials/base.html",
		"../../templates/web/partials/navigation.html",
		"../../templates/web/partials/page_header.html",
		"../../templates/web/admin_metrics.html",
	))

	metricsCollector := middleware.NewPrometheusMetrics()
	promMetricsHandler := middleware.MetricsHandler(metricsCollector)
	promMiddleware := middleware.PrometheusMetrics(metricsCollector)

	router := handlers.NewRouter(&handlers.RouterHandlers{
		AdminDashboardHandler: adminDashboardHandler,
		SettingsHandler:       settingsHandler,
		AdminMetricsHandler:   metricsHandler,
		AuthMiddleware:        middlewareAdapter,
		MetricsHandler:        promMetricsHandler,
		MetricsMiddleware:     promMiddleware,
	})

	srv := httptest.NewServer(router)
	t.Cleanup(func() { srv.Close() })
	return srv, adminUserID, database
}

func adminReq(t *testing.T, method, url string, userID int64, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("X-Test-User-ID", strconv.FormatInt(userID, 10))
	return req
}

func TestVerification_AdminSettings_RendersWithoutError(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/settings", adminUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin/settings: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/admin/settings status = %d, want 200. Body: %s", resp.StatusCode, body)
	}

	bodyStr := string(body)

	for _, want := range []string{
		"System Settings",
		"smtp.verify.test",
		"verify-client",
		"••••••••",
	} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("/admin/settings body missing %q", want)
		}
	}

	for _, mustNotContain := range []string{
		"oidc-secret-value",
		"token-secret-value",
		"hmac-verify-key",
	} {
		if strings.Contains(bodyStr, mustNotContain) {
			t.Errorf("/admin settings leaked secret %q into HTML output", mustNotContain)
		}
	}
}

func TestVerification_AdminMetrics_RendersWithoutError(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/metrics", adminUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin/metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/admin/metrics status = %d, want 200. Body: %s", resp.StatusCode, body)
	}

	bodyStr := string(body)

	for _, want := range []string{
		"System Metrics",
		"Business Metrics",
		"Database Connection Pool",
	} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("/admin/metrics body missing %q", want)
		}
	}
}

func TestVerification_AdminDashboard_HasLinks(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin", adminUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{
		`href="/admin/settings"`,
		`href="/admin/metrics"`,
	} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("/admin body missing link %q", want)
		}
	}
}

func TestVerification_PrometheusMiddleware_IncrementsCounters(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	for i := 0; i < 3; i++ {
		resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin", adminUserID, nil))
		if err != nil {
			t.Fatalf("GET /admin: %v", err)
		}
		resp.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "http_requests_total") {
		t.Fatal("/metrics does not contain http_requests_total")
	}

	for _, line := range strings.Split(bodyStr, "\n") {
		if strings.HasPrefix(line, "http_requests_total{") && !strings.Contains(line, "_total 0\n") {
			return
		}
	}

	t.Errorf("/metrics http_requests_total has no non-zero counts after 3 requests.\nRelevant lines:\n%s",
		func() string {
			var out []string
			for _, line := range strings.Split(bodyStr, "\n") {
				if strings.Contains(line, "http_requests_total") {
					out = append(out, line)
				}
			}
			return strings.Join(out, "\n")
		}())
}

func TestVerification_PrometheusMiddleware_DifferentPathsTracked(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	for i := 0; i < 3; i++ {
		resp1, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/settings", adminUserID, nil))
		if err != nil {
			t.Fatalf("GET settings: %v", err)
		}
		resp1.Body.Close()

		resp2, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/metrics", adminUserID, nil))
		if err != nil {
			t.Fatalf("GET metrics: %v", err)
		}
		resp2.Body.Close()
	}

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	hasSettings := false
	hasMetrics := false
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "http_requests_total{") {
			if strings.Contains(line, `path="/admin/settings"`) {
				parts := strings.Fields(strings.TrimSpace(line))
				if len(parts) > 0 && parts[len(parts)-1] != "0" {
					hasSettings = true
				}
			}
			if strings.Contains(line, `path="/admin/metrics"`) {
				parts := strings.Fields(strings.TrimSpace(line))
				if len(parts) > 0 && parts[len(parts)-1] != "0" {
					hasMetrics = true
				}
			}
		}
	}

	if !hasSettings {
		t.Error("/metrics did not track /admin/settings path with non-zero count after 3 requests")
	}
	if !hasMetrics {
		t.Error("/metrics did not track /admin/metrics path with non-zero count after 3 requests")
	}
}

func TestVerification_AdminSettings_NonAdminDenied(t *testing.T) {
	srv, adminUserID, database := setupVerificationServer(t)

	regularUser := &models.User{
		Email: "regular@verify.test",
		Name:  "Regular User",
		Role:  models.RoleEventManager,
	}
	_, err := database.Exec(context.Background(), `
		INSERT INTO users (email, name, role, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, regularUser.Email, regularUser.Name, regularUser.Role)
	if err != nil {
		t.Fatalf("create regular user: %v", err)
	}

	var regularUserID int64
	row := database.QueryRow(context.Background(), `SELECT id FROM users WHERE email = ?`, regularUser.Email)
	if err := row.Scan(&regularUserID); err != nil {
		t.Fatalf("get regular user ID: %v", err)
	}

	if regularUserID == adminUserID {
		t.Fatal("regular user ID should differ from admin")
	}

	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/settings", regularUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin/settings as non-admin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("non-admin /admin/settings: expected 403, got %d", resp.StatusCode)
	}
}

func TestVerification_AdminSettings_NonExistentEndpoint(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin/this-does-not-exist", adminUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin/this-does-not-exist: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		t.Error("non-existent admin endpoint returned 500 (should be 404 or similar, not a server error)")
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty body for non-existent endpoint (404 handler should render)")
	}
}

func TestVerification_PrometheusMetrics_EndpointScrapeable(t *testing.T) {
	srv, adminUserID, _ := setupVerificationServer(t)

	// Make one request so the counter vec has at least one observed label set.
	resp, err := http.DefaultClient.Do(adminReq(t, "GET", srv.URL+"/admin", adminUserID, nil))
	if err != nil {
		t.Fatalf("GET /admin: %v", err)
	}
	resp.Body.Close()

	// Now scrape /metrics — it should contain the metric.
	metricsResp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		t.Errorf("/metrics status = %d, want 200", metricsResp.StatusCode)
	}

	body, _ := io.ReadAll(metricsResp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "http_requests_total") {
		t.Errorf("/metrics does not contain http_requests_total after a request was made. Body (first 500 chars):\n%.500s", bodyStr)
	}
	if !strings.Contains(bodyStr, "http_request_duration_seconds") {
		t.Errorf("/metrics does not contain http_request_duration_seconds. Body (first 500 chars):\n%.500s", bodyStr)
	}
}
