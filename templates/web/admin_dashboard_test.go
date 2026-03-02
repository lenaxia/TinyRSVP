package web

import (
	"strings"
	"testing"
)

func TestAdminDashboard_RendersSuccessfully(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   10,
			"TotalEvents":  5,
			"TotalInvites": 50,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if html == "" {
		t.Error("Expected non-empty HTML output")
	}
}

func TestAdminDashboard_ContainsRequiredElements(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   10,
			"TotalEvents":  5,
			"TotalInvites": 50,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	requiredElements := []string{
		"<!DOCTYPE html>",
		"<html",
		"<head>",
		"<title>",
		"Admin Dashboard",
		"<body>",
		"<nav",
		"<main",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Expected HTML to contain %q", elem)
		}
	}
}

func TestAdminDashboard_DisplaysStats(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   10,
			"TotalEvents":  5,
			"TotalInvites": 50,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "10") {
		t.Error("Expected HTML to contain total users count")
	}
	if !strings.Contains(html, "5") {
		t.Error("Expected HTML to contain total events count")
	}
	if !strings.Contains(html, "50") {
		t.Error("Expected HTML to contain total invites count")
	}
}

func TestAdminDashboard_ContainsNavigationLinks(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   10,
			"TotalEvents":  5,
			"TotalInvites": 50,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	expectedLinks := []string{
		"/admin/users",
		"/events",
	}

	for _, link := range expectedLinks {
		if !strings.Contains(html, link) {
			t.Errorf("Expected HTML to contain link %q", link)
		}
	}
}

func TestAdminDashboard_HandlesEmptyStats(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   0,
			"TotalEvents":  0,
			"TotalInvites": 0,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if html == "" {
		t.Error("Expected non-empty HTML output")
	}
}

func TestAdminDashboard_AccessibilityFeatures(t *testing.T) {
	tmpl, err := parseWithBase("admin_dashboard.html")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	data := map[string]interface{}{
		"User": map[string]interface{}{
			"Name":  "Admin User",
			"Email": "admin@example.com",
			"Role":  "admin",
		},
		"Stats": map[string]interface{}{
			"TotalUsers":   10,
			"TotalEvents":  5,
			"TotalInvites": 50,
		},
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

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
