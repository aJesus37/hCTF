package email

import (
	"strings"
	"testing"
)

func TestNewService_NoConfig(t *testing.T) {
	svc := NewService(Config{})
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.IsConfigured() {
		t.Error("service should not be configured without SMTP host")
	}
}

func TestNewService_WithConfig(t *testing.T) {
	svc := NewService(Config{
		Host:     "smtp.example.com",
		Port:     587,
		From:     "noreply@example.com",
		Username: "user",
		Password: "pass",
	})
	if !svc.IsConfigured() {
		t.Error("service should be configured with SMTP host")
	}
}

func TestBuildResetEmail(t *testing.T) {
	svc := NewService(Config{
		Host: "smtp.example.com",
		Port: 587,
		From: "noreply@example.com",
	})

	body := svc.buildResetEmail("user@example.com", "https://example.com/reset-password?token=abc123")
	if body == "" {
		t.Error("buildResetEmail returned empty string")
	}
	if !strings.Contains(body, "abc123") {
		t.Error("email body missing reset token URL")
	}
	if !strings.Contains(body, "user@example.com") {
		t.Error("email body missing recipient")
	}
}
