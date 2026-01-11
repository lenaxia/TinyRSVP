package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	"github.com/lenaxia/tinyrsvp/pkg/ics"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

func TestMain_RouterIntegration(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

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

	systemUser := &models.User{
		Email: "system@tinyrsvp.local",
		Name:  "System",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, systemUser); err != nil {
		t.Fatalf("Failed to create system user: %v", err)
	}

	seeder := templates.NewSeeder(templateRepo, systemUser.ID)
	if err := seeder.SeedDefaults(ctx); err != nil {
		t.Fatalf("Failed to seed templates: %v", err)
	}

	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, eventValidator, authChecker)

	questionValidator := events.NewQuestionValidator()
	questionService := events.NewQuestionService(eventRepo, questionRepo, questionValidator, authChecker)

	tokenGenerator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)
	individualInviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

	storageProvider, err := storage.NewProvider(&storage.Config{
		Type:     "local",
		BasePath: t.TempDir(),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Failed to create storage provider: %v", err)
	}

	imageService := assets.NewImageService(storageProvider)

	templateRenderer, err := email.NewTemplateRenderer(&email.TemplateConfig{
		TemplateDir:  "../../templates/email",
		CacheEnabled: false,
	})
	if err != nil {
		t.Fatalf("Failed to create template renderer: %v", err)
	}

	icsGenerator := ics.NewGenerator()
	emailService := email.NewConfirmationService(templateRenderer, emailQueueRepo, icsGenerator)

	rsvpService := rsvp.NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, emailService)

	fwdAuthCfg := &auth.ForwardAuthConfig{
		UserHeader:  "X-Forwarded-User",
		EmailHeader: "X-Forwarded-Email",
		TrustedIPs:  []string{"127.0.0.1"},
	}
	authenticator := auth.NewForwardAuthenticator(fwdAuthCfg, userService, sessionMgr)

	loginHandler := auth.NewLoginHandler(authenticator)
	callbackHandler := auth.NewCallbackHandler(authenticator, userService, sessionMgr)
	logoutHandler := auth.NewLogoutHandler(authenticator)

	healthHandler := handlers.NewHealthHandler("test-version")
	readinessHandler := handlers.NewReadinessHandler("test-version", database, migrator)

	eventHandlers := handlers.NewEventHandlers(eventService)
	questionHandlers := handlers.NewQuestionHandlers(questionService)
	templateHandlers := handlers.NewTemplateHandlers(templateService)
	imageHandlers := handlers.NewImageHandlers(imageService, eventService, authChecker)

	inviteHandlers := handlers.NewInviteHandlers(individualInviteService, "http://localhost:8080")
	importInviteHandlers := handlers.NewImportInviteHandlers(inviteService, eventRepo, "http://localhost:8080")
	manualInviteHandlers := handlers.NewManualInviteHandlers(inviteService, eventRepo, "http://localhost:8080")
	revokeInviteHandlers := handlers.NewRevokeInviteHandlers(inviteService, eventRepo)
	regenerateInviteHandlers := handlers.NewRegenerateInviteTokenHandlers(inviteService, eventRepo)
	listInviteHandlers := handlers.NewListInviteHandlers(inviteService, eventRepo)

	cleanupHandler := handlers.NewCleanupHandler(inviteService)

	rsvpHandler := handlers.NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	rsvpHandler.SetRSVPService(rsvpService)
	rsvpHandler.SetAnswerRepository(answerRepo)

	rsvpSummaryHandler := handlers.NewRSVPSummaryHandler(eventRepo, rsvpRepo, questionRepo, answerRepo)

	userHandler := handlers.NewUserHandler(userService, authChecker)
	assetHandler := handlers.NewAssetHandler(storageProvider)

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	staticFS := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static")))

	router := handlers.NewRouter(&handlers.RouterHandlers{
		LoginHandler:             loginHandler,
		CallbackHandler:          callbackHandler,
		LogoutHandler:            logoutHandler,
		HealthHandler:            healthHandler,
		ReadinessHandler:         readinessHandler,
		EventHandlers:            eventHandlers,
		QuestionHandlers:         questionHandlers,
		InviteHandlers:           inviteHandlers,
		ImportInviteHandlers:     importInviteHandlers,
		ManualInviteHandlers:     manualInviteHandlers,
		RevokeInviteHandlers:     revokeInviteHandlers,
		RegenerateInviteHandlers: regenerateInviteHandlers,
		ListInviteHandlers:       listInviteHandlers,
		ImageHandlers:            imageHandlers,
		RSVPHandler:              rsvpHandler,
		RSVPSummaryHandler:       rsvpSummaryHandler,
		UserHandler:              userHandler,
		TemplateHandlers:         templateHandlers,
		AssetHandler:             assetHandler,
		CleanupHandler:           cleanupHandler,
		AuthMiddleware:           middlewareAdapter,
		StaticFileServer:         staticFS,
	})

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
		description    string
	}{
		{
			name:           "health endpoint",
			method:         http.MethodGet,
			path:           "/health",
			wantStatusCode: http.StatusOK,
			description:    "Health check should return 200",
		},
		{
			name:           "readiness endpoint",
			method:         http.MethodGet,
			path:           "/ready",
			wantStatusCode: http.StatusOK,
			description:    "Readiness check should return 200",
		},
		{
			name:           "login endpoint",
			method:         http.MethodGet,
			path:           "/login",
			wantStatusCode: http.StatusInternalServerError,
			description:    "Login handler is called (fails in test due to missing OIDC config)",
		},
		{
			name:           "callback endpoint",
			method:         http.MethodGet,
			path:           "/auth/callback",
			wantStatusCode: http.StatusUnauthorized,
			description:    "Callback handler is called (fails in test due to untrusted IP)",
		},
		{
			name:           "logout endpoint",
			method:         http.MethodPost,
			path:           "/logout",
			wantStatusCode: http.StatusForbidden,
			description:    "Logout requires CSRF token (403 without token)",
		},
		{
			name:           "events list requires auth",
			method:         http.MethodGet,
			path:           "/api/events",
			wantStatusCode: http.StatusSeeOther,
			description:    "Events list redirects to login when unauthenticated",
		},
		{
			name:           "rsvp page public",
			method:         http.MethodGet,
			path:           "/rsvp/test-token-123",
			wantStatusCode: http.StatusNotFound,
			description:    "RSVP page should be publicly accessible (404 for invalid token is expected)",
		},
		{
			name:           "static files served",
			method:         http.MethodGet,
			path:           "/static/test.css",
			wantStatusCode: http.StatusNotFound,
			description:    "Static file route should be registered (404 for missing file is expected)",
		},
		{
			name:           "assets served",
			method:         http.MethodGet,
			path:           "/assets/test.jpg",
			wantStatusCode: http.StatusNotFound,
			description:    "Asset route should be registered (404 for missing asset is expected)",
		},
		{
			name:           "404 for unknown route",
			method:         http.MethodGet,
			path:           "/nonexistent",
			wantStatusCode: http.StatusNotFound,
			description:    "Unknown routes should return 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("%s: expected status %d, got %d. Body: %s",
					tt.description, tt.wantStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMain_RouterIntegration_AuthenticatedRoutes(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

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

	testUser := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	systemUser := &models.User{
		Email: "system@tinyrsvp.local",
		Name:  "System",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, systemUser); err != nil {
		t.Fatalf("Failed to create system user: %v", err)
	}

	seeder := templates.NewSeeder(templateRepo, systemUser.ID)
	if err := seeder.SeedDefaults(ctx); err != nil {
		t.Fatalf("Failed to seed templates: %v", err)
	}

	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	session, err := sessionMgr.CreateSession(ctx, testUser.ID, req)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, eventValidator, authChecker)

	questionValidator := events.NewQuestionValidator()
	questionService := events.NewQuestionService(eventRepo, questionRepo, questionValidator, authChecker)

	tokenGenerator := token.NewGenerator([]byte("test-secret-key-32-bytes-long!!"))
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)
	individualInviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

	storageProvider, err := storage.NewProvider(&storage.Config{
		Type:     "local",
		BasePath: t.TempDir(),
		BaseURL:  "http://localhost:8080",
	})
	if err != nil {
		t.Fatalf("Failed to create storage provider: %v", err)
	}

	imageService := assets.NewImageService(storageProvider)

	templateRenderer, err := email.NewTemplateRenderer(&email.TemplateConfig{
		TemplateDir:  "../../templates/email",
		CacheEnabled: false,
	})
	if err != nil {
		t.Fatalf("Failed to create template renderer: %v", err)
	}

	icsGenerator := ics.NewGenerator()
	emailService := email.NewConfirmationService(templateRenderer, emailQueueRepo, icsGenerator)

	rsvpService := rsvp.NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, emailService)

	fwdAuthCfg := &auth.ForwardAuthConfig{
		UserHeader:  "X-Forwarded-User",
		EmailHeader: "X-Forwarded-Email",
		TrustedIPs:  []string{"127.0.0.1"},
	}
	authenticator := auth.NewForwardAuthenticator(fwdAuthCfg, userService, sessionMgr)

	loginHandler := auth.NewLoginHandler(authenticator)
	callbackHandler := auth.NewCallbackHandler(authenticator, userService, sessionMgr)
	logoutHandler := auth.NewLogoutHandler(authenticator)

	healthHandler := handlers.NewHealthHandler("test-version")
	readinessHandler := handlers.NewReadinessHandler("test-version", database, migrator)

	eventHandlers := handlers.NewEventHandlers(eventService)
	questionHandlers := handlers.NewQuestionHandlers(questionService)
	templateHandlers := handlers.NewTemplateHandlers(templateService)
	imageHandlers := handlers.NewImageHandlers(imageService, eventService, authChecker)

	inviteHandlers := handlers.NewInviteHandlers(individualInviteService, "http://localhost:8080")
	importInviteHandlers := handlers.NewImportInviteHandlers(inviteService, eventRepo, "http://localhost:8080")
	manualInviteHandlers := handlers.NewManualInviteHandlers(inviteService, eventRepo, "http://localhost:8080")
	revokeInviteHandlers := handlers.NewRevokeInviteHandlers(inviteService, eventRepo)
	regenerateInviteHandlers := handlers.NewRegenerateInviteTokenHandlers(inviteService, eventRepo)
	listInviteHandlers := handlers.NewListInviteHandlers(inviteService, eventRepo)

	cleanupHandler := handlers.NewCleanupHandler(inviteService)

	rsvpHandler := handlers.NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	rsvpHandler.SetRSVPService(rsvpService)
	rsvpHandler.SetAnswerRepository(answerRepo)

	rsvpSummaryHandler := handlers.NewRSVPSummaryHandler(eventRepo, rsvpRepo, questionRepo, answerRepo)

	userHandler := handlers.NewUserHandler(userService, authChecker)
	assetHandler := handlers.NewAssetHandler(storageProvider)

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	staticFS := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static")))

	router := handlers.NewRouter(&handlers.RouterHandlers{
		LoginHandler:             loginHandler,
		CallbackHandler:          callbackHandler,
		LogoutHandler:            logoutHandler,
		HealthHandler:            healthHandler,
		ReadinessHandler:         readinessHandler,
		EventHandlers:            eventHandlers,
		QuestionHandlers:         questionHandlers,
		InviteHandlers:           inviteHandlers,
		ImportInviteHandlers:     importInviteHandlers,
		ManualInviteHandlers:     manualInviteHandlers,
		RevokeInviteHandlers:     revokeInviteHandlers,
		RegenerateInviteHandlers: regenerateInviteHandlers,
		ListInviteHandlers:       listInviteHandlers,
		ImageHandlers:            imageHandlers,
		RSVPHandler:              rsvpHandler,
		RSVPSummaryHandler:       rsvpSummaryHandler,
		UserHandler:              userHandler,
		TemplateHandlers:         templateHandlers,
		AssetHandler:             assetHandler,
		CleanupHandler:           cleanupHandler,
		AuthMiddleware:           middlewareAdapter,
		StaticFileServer:         staticFS,
	})

	tests := []struct {
		name           string
		method         string
		path           string
		addAuth        bool
		wantStatusCode int
		description    string
	}{
		{
			name:           "unauthenticated events list",
			method:         http.MethodGet,
			path:           "/api/events",
			addAuth:        false,
			wantStatusCode: http.StatusSeeOther,
			description:    "Unauthenticated request redirects to login - middleware is working",
		},
		{
			name:           "unauthenticated users list",
			method:         http.MethodGet,
			path:           "/api/users",
			addAuth:        false,
			wantStatusCode: http.StatusSeeOther,
			description:    "Unauthenticated request redirects to login - middleware is working",
		},
		{
			name:           "unauthenticated templates list",
			method:         http.MethodGet,
			path:           "/api/templates",
			addAuth:        false,
			wantStatusCode: http.StatusSeeOther,
			description:    "Unauthenticated request redirects to login - middleware is working",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.addAuth {
				cookie := &http.Cookie{
					Name:  "session",
					Value: session.ID,
				}
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("%s: expected status %d, got %d. Body: %s",
					tt.description, tt.wantStatusCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestMain_RouterIntegration_MiddlewareChain(t *testing.T) {
	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: ":memory:",
	})
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	eventRepo := repositories.NewEventRepository(database)

	testUser := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	sessionReq := httptest.NewRequest(http.MethodGet, "/", nil)
	session, err := sessionMgr.CreateSession(ctx, testUser.ID, sessionReq)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, eventValidator, authChecker)

	eventHandlers := handlers.NewEventHandlers(eventService)

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)
	middlewareAdapter := handlers.NewMiddlewareAdapter(requireAuth, requireAdmin)

	router := handlers.NewRouter(&handlers.RouterHandlers{
		EventHandlers:  eventHandlers,
		AuthMiddleware: middlewareAdapter,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected unauthenticated request to redirect to login, got status %d. Body: %s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if location != "/login?return=%2Fapi%2Fevents" {
		t.Errorf("Expected redirect to /login?return=%%2Fapi%%2Fevents, got %s", location)
	}

	_ = session
}
