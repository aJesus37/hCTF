package email

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
}

// Service sends emails via SMTP
type Service struct {
	config Config
}

// NewService creates a new email service
func NewService(cfg Config) *Service {
	return &Service{config: cfg}
}

// IsConfigured returns true if SMTP settings are provided
func (s *Service) IsConfigured() bool {
	return s.config.Host != ""
}

// SendPasswordReset sends a password reset email or logs the link if SMTP is not configured
func (s *Service) SendPasswordReset(toEmail, resetURL string) error {
	if !s.IsConfigured() {
		log.Printf("[DEV] Password reset link for %s: %s", toEmail, resetURL)
		return nil
	}

	body := s.buildResetEmail(toEmail, resetURL)
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	return smtp.SendMail(addr, auth, s.config.From, []string{toEmail}, []byte(body))
}

func (s *Service) buildResetEmail(to, resetURL string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	b.WriteString("Subject: Password Reset - hCTF2\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; background: #1a1a2e; color: #e0e0e0; padding: 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: #16213e; border-radius: 8px; padding: 30px;">
    <h2 style="color: #a55eea; margin-top: 0;">Password Reset</h2>
    <p>You requested a password reset for your hCTF2 account (%s).</p>
    <p>Click the link below to reset your password. This link expires in 30 minutes.</p>
    <p style="text-align: center; margin: 25px 0;">
      <a href="%s" style="background: #a55eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;">Reset Password</a>
    </p>
    <p style="color: #888; font-size: 12px;">If you didn't request this, ignore this email.</p>
  </div>
</body>
</html>`, to, resetURL))
	return b.String()
}
