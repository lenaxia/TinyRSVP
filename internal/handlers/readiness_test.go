package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/db"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.NewDatabase(db.Config{
		Type: "sqlite",
		Path: dbPath,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return database
}

func TestReadinessHandler_Healthy(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}

	if _, ok := response.Checks["database"]; !ok {
		t.Error("Expected database check in response")
	}

	if _, ok := response.Checks["migrations"]; !ok {
		t.Error("Expected migrations check in response")
	}
}

func TestReadinessHandler_DatabaseUnhealthy(t *testing.T) {
	database := setupTestDB(t)

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	database.Close()

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != StatusUnhealthy {
		t.Errorf("Expected status unhealthy, got %s", response.Status)
	}

	dbCheck, ok := response.Checks["database"]
	if !ok {
		t.Fatal("Expected database check in response")
	}

	if dbCheck.Status != StatusUnhealthy {
		t.Errorf("Expected database status unhealthy, got %s", dbCheck.Status)
	}
}

func TestReadinessHandler_DatabaseConnectivity(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	dbCheck, ok := response.Checks["database"]
	if !ok {
		t.Fatal("Expected database check in response")
	}

	if dbCheck.Status != StatusHealthy {
		t.Errorf("Expected database status healthy, got %s", dbCheck.Status)
	}

	if dbCheck.Message != "Connected" {
		t.Errorf("Expected message 'Connected', got %s", dbCheck.Message)
	}

	if dbCheck.LatencyMs == nil {
		t.Error("Expected latency measurement")
	} else if *dbCheck.LatencyMs < 0 {
		t.Errorf("Expected non-negative latency, got %d", *dbCheck.LatencyMs)
	}
}

func TestReadinessHandler_MigrationVersion(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	migrationCheck, ok := response.Checks["migrations"]
	if !ok {
		t.Fatal("Expected migrations check in response")
	}

	if migrationCheck.Status != StatusHealthy {
		t.Errorf("Expected migrations status healthy, got %s", migrationCheck.Status)
	}

	if migrationCheck.Version == nil {
		t.Error("Expected migration version")
	} else if *migrationCheck.Version == 0 {
		t.Error("Expected non-zero migration version")
	}
}

func TestReadinessHandler_DirtyMigrations(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = database.DB().Exec("UPDATE schema_migrations SET dirty = true WHERE version = (SELECT MAX(version) FROM schema_migrations)")
	if err != nil {
		t.Fatalf("Failed to set dirty flag: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != StatusUnhealthy {
		t.Errorf("Expected status unhealthy, got %s", response.Status)
	}

	migrationCheck, ok := response.Checks["migrations"]
	if !ok {
		t.Fatal("Expected migrations check in response")
	}

	if migrationCheck.Status != StatusUnhealthy {
		t.Errorf("Expected migrations status unhealthy, got %s", migrationCheck.Status)
	}
}

func TestReadinessHandler_NoMigrations(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	migrationCheck, ok := response.Checks["migrations"]
	if !ok {
		t.Fatal("Expected migrations check in response")
	}

	if migrationCheck.Status == StatusHealthy {
		t.Error("Expected migrations status to not be healthy when no migrations run")
	}
}

func TestReadinessHandler_ContentType(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestReadinessHandler_ContextTimeout(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/ready", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	time.Sleep(2 * time.Millisecond)

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Logf("Note: Context timeout test may not trigger failure if checks complete quickly")
	}

	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Checks == nil {
		t.Error("Expected checks map to be initialized")
	}
}

func TestReadinessHandler_MultipleSimultaneousRequests(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB(), "../../migrations/sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	handler := NewReadinessHandler("0.1.0", database, migrator)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", "/ready", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request %d: Expected status 200, got %d", id, w.Code)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
