package web

import (
	"strings"
	"testing"
	"time"
)

func getUserManagementData(users ...map[string]interface{}) map[string]interface{} {
	userList := []map[string]interface{}{}
	for _, u := range users {
		userList = append(userList, u)
	}
	return map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Users":     userList,
		"CSRFToken": "test-csrf-token",
	}
}

func TestUserManagement_RendersSuccessfully(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := getUserManagementData(map[string]interface{}{
		"ID":        int64(1),
		"Email":     "user1@example.com",
		"Name":      "User One",
		"Role":      "event_manager",
		"CreatedAt": time.Now(),
	})

	html := executeTemplate(t, tmpl, data)
	if html == "" {
		t.Error("Expected non-empty HTML output")
	}
}

func TestUserManagement_ContainsRequiredElements(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := getUserManagementData(map[string]interface{}{
		"ID":        int64(1),
		"Email":     "user1@example.com",
		"Name":      "User One",
		"Role":      "event_manager",
		"CreatedAt": time.Now(),
	})

	html := executeTemplate(t, tmpl, data)

	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>",
		"User Management",
		"<body>",
		"<nav",
		"<main",
		"<table",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Expected HTML to contain %q", elem)
		}
	}
}

func TestUserManagement_DisplaysUserList(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := getUserManagementData(
		map[string]interface{}{
			"ID":        int64(1),
			"Email":     "user1@example.com",
			"Name":      "User One",
			"Role":      "event_manager",
			"CreatedAt": time.Now(),
		},
		map[string]interface{}{
			"ID":        int64(2),
			"Email":     "user2@example.com",
			"Name":      "User Two",
			"Role":      "admin",
			"CreatedAt": time.Now(),
		},
	)

	html := executeTemplate(t, tmpl, data)

	if !strings.Contains(html, "user1@example.com") {
		t.Error("Expected HTML to contain user1 email")
	}
	if !strings.Contains(html, "User One") {
		t.Error("Expected HTML to contain user1 name")
	}
	if !strings.Contains(html, "user2@example.com") {
		t.Error("Expected HTML to contain user2 email")
	}
	if !strings.Contains(html, "User Two") {
		t.Error("Expected HTML to contain user2 name")
	}
}

func TestUserManagement_ContainsCSRFToken(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Users": []map[string]interface{}{
			{
				"ID":        int64(1),
				"Email":     "user1@example.com",
				"Name":      "User One",
				"Role":      "event_manager",
				"CreatedAt": time.Now(),
			},
		},
		"CSRFToken": "test-csrf-token-12345",
	}

	html := executeTemplate(t, tmpl, data)

	if !strings.Contains(html, "test-csrf-token-12345") {
		t.Error("Expected HTML to contain CSRF token")
	}
}

func TestUserManagement_HandlesEmptyUserList(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Users":     []map[string]interface{}{},
		"CSRFToken": "test-csrf-token",
	}

	html := executeTemplate(t, tmpl, data)
	if html == "" {
		t.Error("Expected non-empty HTML output")
	}
}

func TestUserManagement_ContainsActionButtons(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := getUserManagementData(map[string]interface{}{
		"ID":        int64(1),
		"Email":     "user1@example.com",
		"Name":      "User One",
		"Role":      "event_manager",
		"CreatedAt": time.Now(),
	})

	html := executeTemplate(t, tmpl, data)

	expectedActions := []string{
		"Edit",
		"Delete",
	}

	for _, action := range expectedActions {
		if !strings.Contains(html, action) {
			t.Errorf("Expected HTML to contain action %q", action)
		}
	}
}

func TestUserManagement_AccessibilityFeatures(t *testing.T) {
	tmpl, err := parseWithBase("user_management.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Users":     []map[string]interface{}{},
		"CSRFToken": "test-csrf-token",
	}

	html := executeTemplate(t, tmpl, data)

	accessibilityFeatures := []string{
		`role="navigation"`,
		`role="main"`,
		`aria-label`,
		`class="skip-link"`,
	}

	for _, feature := range accessibilityFeatures {
		if !strings.Contains(html, feature) {
			t.Errorf("Expected HTML to contain accessibility feature %q", feature)
		}
	}
}
