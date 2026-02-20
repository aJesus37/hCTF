# New Features Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 5 features: UUIDv7 migration, OTel Prometheus+OTLP export, remove KNOWN_ISSUES.md, SMTP password reset, and Kubernetes-style healthcheck endpoints.

**Architecture:** Each feature is independent and can be implemented in any order. UUIDv7 changes the ID generation layer. OTel adds exporters to existing telemetry. SMTP adds an email service package. Healthcheck adds two simple endpoints.

**Tech Stack:** Go 1.24, google/uuid (UUIDv7), OTel Prometheus/OTLP exporters, net/smtp (stdlib), Chi router

---

### Task 1: Remove KNOWN_ISSUES.md

**Files:**
- Delete: `KNOWN_ISSUES.md`
- Modify: `README.md` (remove reference)

**Step 1: Remove the reference from README.md**

Find the line in `README.md` that links to KNOWN_ISSUES.md (around line 166 in the documentation table of contents) and remove it.

**Step 2: Delete KNOWN_ISSUES.md**

```bash
rm KNOWN_ISSUES.md
```

**Step 3: Commit**

```bash
git add KNOWN_ISSUES.md README.md
git commit -m "chore: remove obsolete KNOWN_ISSUES.md

The single documented issue (edit buttons on dynamic elements) was
fixed in commit 017ad17. File is no longer needed."
```

---

### Task 2: Kubernetes-style Healthcheck Endpoints

**Files:**
- Modify: `main.go` (add handlers and routes)
- Modify: `Dockerfile` (update HEALTHCHECK)
- Modify: `docker-compose.yml` (update healthcheck)

**Step 1: Write failing test for healthcheck endpoints**

Add to `handlers_test.go`:

```go
func TestHealthEndpoints(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	server := newTestServer(db)
	router := newTestRouter(server)

	tests := []struct {
		name        string
		path        string
		statusCode  int
		contentMust []string
	}{
		{
			name:        "Liveness probe",
			path:        "/healthz",
			statusCode:  http.StatusOK,
			contentMust: []string{`"status":"ok"`},
		},
		{
			name:        "Readiness probe",
			path:        "/readyz",
			statusCode:  http.StatusOK,
			contentMust: []string{`"status":"ready"`, `"database":"ok"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("got status %d, want %d", w.Code, tt.statusCode)
			}

			body := w.Body.String()
			for _, must := range tt.contentMust {
				if !strings.Contains(body, must) {
					t.Errorf("response missing %q, got: %s", must, body)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
task test
```

Expected: FAIL - routes `/healthz` and `/readyz` not registered

**Step 3: Add healthcheck handlers to main.go**

Add these methods to the Server struct in `main.go`, after the existing handler methods:

```go
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","checks":{"database":"error: %s"}}`, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","checks":{"database":"ok"}}`))
}
```

**Step 4: Add DB Ping method**

Add to `internal/database/queries.go`:

```go
// Ping checks database connectivity
func (db *DB) Ping() error {
	return db.conn.Ping()
}
```

**Step 5: Register routes in main.go router setup**

Add before the public routes section (before `r.Get("/", ...)`):

```go
// Health check endpoints (no auth, no middleware)
r.Get("/healthz", s.handleHealthz)
r.Get("/readyz", s.handleReadyz)
```

**Step 6: Register routes in test router**

Add to `newTestRouter` in `handlers_test.go`:

```go
mux.HandleFunc("GET /healthz", s.handleHealthz)
mux.HandleFunc("GET /readyz", s.handleReadyz)
```

**Step 7: Run tests**

```bash
task test
```

Expected: PASS

**Step 8: Update Dockerfile healthcheck**

Change line 50-51 in `Dockerfile` from:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/
```
To:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/healthz
```

**Step 9: Update docker-compose.yml healthcheck**

Change the test command from:
```yaml
test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8090/"]
```
To:
```yaml
test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8090/healthz"]
```

**Step 10: Add swag annotations to handlers**

Add annotations above each handler function for OpenAPI docs.

**Step 11: Commit**

```bash
git add main.go internal/database/queries.go handlers_test.go Dockerfile docker-compose.yml
git commit -m "feat: add /healthz and /readyz health check endpoints

- /healthz: liveness probe (always 200 if process alive)
- /readyz: readiness probe (checks DB connectivity)
- Update Docker healthcheck to use /healthz
- Add DB Ping method for connectivity checks"
```

---

### Task 3: UUIDv7 Migration

**Files:**
- Create: `internal/database/id.go`
- Create: `internal/database/migrations/007_uuidv7.up.sql`
- Create: `internal/database/migrations/007_uuidv7.down.sql`
- Modify: `internal/database/queries.go` (all INSERT queries + imports)
- Modify: `go.mod` (promote google/uuid to direct)

**Step 1: Write test for ID generation**

Create `internal/database/id_test.go`:

```go
package database

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()

	// Must be valid UUID
	parsed, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("GenerateID() returned invalid UUID: %s, error: %v", id, err)
	}

	// Must be version 7
	if parsed.Version() != 7 {
		t.Errorf("GenerateID() version = %d, want 7", parsed.Version())
	}

	// Must be lowercase with hyphens
	if id != strings.ToLower(id) {
		t.Errorf("GenerateID() not lowercase: %s", id)
	}

	// Two IDs must be different
	id2 := GenerateID()
	if id == id2 {
		t.Errorf("GenerateID() returned duplicate: %s", id)
	}

	// Second ID should be >= first (time-ordered)
	if id2 < id {
		t.Errorf("GenerateID() not time-ordered: %s < %s", id2, id)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/database/ -run TestGenerateID -v
```

Expected: FAIL - GenerateID not defined

**Step 3: Create internal/database/id.go**

```go
package database

import "github.com/google/uuid"

// GenerateID returns a new UUIDv7 string.
// UUIDv7 is time-ordered, making it suitable for primary keys
// with better index locality than random UUIDs.
func GenerateID() string {
	return uuid.Must(uuid.NewV7()).String()
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/database/ -run TestGenerateID -v
```

Expected: PASS

**Step 5: Create migration 007_uuidv7.up.sql**

This migration removes the DEFAULT clause so new rows MUST provide an ID from Go code. Since this is a clean migration (no production data), we recreate all tables.

Create `internal/database/migrations/007_uuidv7.up.sql`:

```sql
-- UUIDv7 Migration: Remove hex(randomblob(16)) defaults
-- IDs are now generated in Go code using UUIDv7
-- This is a clean migration (no data preservation needed)

-- Disable foreign keys during migration
PRAGMA foreign_keys = OFF;

-- Users table
CREATE TABLE users_new (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    avatar_url TEXT,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    is_admin BOOLEAN DEFAULT 0,
    password_reset_token TEXT,
    password_reset_expires DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new SELECT id, email, password_hash, name, avatar_url, team_id, is_admin, password_reset_token, password_reset_expires, created_at, updated_at FROM users;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_reset_token ON users(password_reset_token);

-- Teams table
CREATE TABLE teams_new (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    owner_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    invite_id TEXT UNIQUE NOT NULL,
    invite_permission TEXT NOT NULL DEFAULT 'owner_only',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO teams_new SELECT id, name, description, owner_id, invite_id, invite_permission, created_at, updated_at FROM teams;
DROP TABLE teams;
ALTER TABLE teams_new RENAME TO teams;
CREATE INDEX idx_teams_invite_id ON teams(invite_id);

-- Challenges table
CREATE TABLE challenges_new (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    difficulty TEXT NOT NULL,
    tags JSON,
    visible BOOLEAN DEFAULT 1,
    sql_enabled BOOLEAN DEFAULT 0,
    sql_dataset_url TEXT,
    sql_schema_hint TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO challenges_new SELECT id, name, description, category, difficulty, tags, visible, sql_enabled, sql_dataset_url, sql_schema_hint, created_at, updated_at FROM challenges;
DROP TABLE challenges;
ALTER TABLE challenges_new RENAME TO challenges;

-- Questions table
CREATE TABLE questions_new (
    id TEXT PRIMARY KEY,
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    flag TEXT NOT NULL,
    flag_mask TEXT,
    case_sensitive BOOLEAN DEFAULT 0,
    points INTEGER DEFAULT 100,
    file_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO questions_new SELECT id, challenge_id, name, description, flag, flag_mask, case_sensitive, points, file_url, created_at, updated_at FROM questions;
DROP TABLE questions;
ALTER TABLE questions_new RENAME TO questions;

-- Hints table
CREATE TABLE hints_new (
    id TEXT PRIMARY KEY,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    cost INTEGER DEFAULT 0,
    "order" INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO hints_new SELECT id, question_id, content, cost, "order", created_at FROM hints;
DROP TABLE hints;
ALTER TABLE hints_new RENAME TO hints;

-- Submissions table
CREATE TABLE submissions_new (
    id TEXT PRIMARY KEY,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    submitted_flag TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, user_id)
);
INSERT INTO submissions_new SELECT id, question_id, user_id, team_id, submitted_flag, is_correct, created_at FROM submissions;
DROP TABLE submissions;
ALTER TABLE submissions_new RENAME TO submissions;

-- Hint unlocks table
CREATE TABLE hint_unlocks_new (
    id TEXT PRIMARY KEY,
    hint_id TEXT NOT NULL REFERENCES hints(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id TEXT REFERENCES teams(id) ON DELETE SET NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(hint_id, user_id)
);
INSERT INTO hint_unlocks_new SELECT id, hint_id, user_id, team_id, created_at FROM hint_unlocks;
DROP TABLE hint_unlocks;
ALTER TABLE hint_unlocks_new RENAME TO hint_unlocks;

-- Categories table
CREATE TABLE categories_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO categories_new SELECT id, name, sort_order, created_at FROM categories;
DROP TABLE categories;
ALTER TABLE categories_new RENAME TO categories;

-- Difficulties table
CREATE TABLE difficulties_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT 'bg-gray-600 text-gray-100',
    text_color TEXT NOT NULL DEFAULT 'text-gray-400',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO difficulties_new SELECT id, name, color, text_color, sort_order, created_at FROM difficulties;
DROP TABLE difficulties;
ALTER TABLE difficulties_new RENAME TO difficulties;

-- Re-enable foreign keys
PRAGMA foreign_keys = ON;
```

**Step 6: Create migration 007_uuidv7.down.sql**

Create `internal/database/migrations/007_uuidv7.down.sql`:

```sql
-- Rollback: This is a no-op since the data format is compatible
-- The only change was removing DEFAULT clauses, which SQLite handles fine
-- Old hex(randomblob(16)) IDs are still valid TEXT primary keys
SELECT 1;
```

**Step 7: Update all INSERT queries in queries.go**

Update each of the 9 CREATE/INSERT functions to generate and pass a UUIDv7 ID. The pattern for each is:

For `CreateUser` (around line 24):
```go
func (db *DB) CreateUser(email, passwordHash, name string, isAdmin bool) (*models.User, error) {
	id := GenerateID()
	query := `INSERT INTO users (id, email, password_hash, name, is_admin) VALUES (?, ?, ?, ?, ?) RETURNING id, email, name, is_admin, created_at, updated_at`
	// ... use id as first parameter
```

Apply same pattern to: `CreateChallenge`, `CreateQuestion`, `CreateSubmission`, `CreateHint`, `UnlockHint` (hint_unlocks), `CreateTeam`, `CreateCategory`, `CreateDifficulty`.

Also update `CreateUser` in `main.go` (the `createAdminUser` function) if it uses its own INSERT.

**Step 8: Promote google/uuid to direct dependency**

```bash
cd /home/jesus/Projects/hCTF2 && go get github.com/google/uuid@v1.6.0
```

This moves it from `// indirect` to direct in go.mod.

**Step 9: Run all tests**

```bash
task test
```

Expected: PASS (existing tests should still pass since UUIDv7 strings are valid TEXT primary keys)

**Step 10: Commit**

```bash
git add internal/database/id.go internal/database/id_test.go internal/database/migrations/007_uuidv7.up.sql internal/database/migrations/007_uuidv7.down.sql internal/database/queries.go go.mod go.sum main.go
git commit -m "feat: migrate ID generation from random hex to UUIDv7

- Add GenerateID() using google/uuid v7 (time-ordered)
- Update all 9 INSERT queries to pass Go-generated IDs
- Migration 007 removes SQLite DEFAULT hex(randomblob(16))
- Promote google/uuid to direct dependency"
```

---

### Task 4: OTel Prometheus + OTLP Export

**Files:**
- Modify: `internal/telemetry/telemetry.go` (add exporters, update Config)
- Modify: `main.go` (add CLI flags, register /metrics, pass config)
- Modify: `go.mod` (new OTel exporter dependencies)

**Step 1: Add new OTel dependencies**

```bash
cd /home/jesus/Projects/hCTF2
go get go.opentelemetry.io/otel/exporters/prometheus
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go get github.com/prometheus/client_golang/prometheus/promhttp
```

**Step 2: Update telemetry.Config struct**

In `internal/telemetry/telemetry.go`, update the Config struct:

```go
type Config struct {
	ServiceName          string
	ServiceVersion       string
	Environment          string
	EnableStdoutExporter bool
	EnablePrometheus     bool
	OTLPEndpoint         string // e.g. "localhost:4318"
}
```

**Step 3: Update Init function to support Prometheus**

Add Prometheus metric exporter to the Init function. The key change is creating a Prometheus exporter that registers with the default Prometheus registry, then creating an OTel MeterProvider that uses it:

```go
import (
	// existing imports...
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// In Init(), after creating the tracer provider, replace the meter setup:

// Create meter provider with appropriate exporters
var meterOpts []sdkmetric.Option
meterOpts = append(meterOpts, sdkmetric.WithResource(res))

if cfg.EnablePrometheus {
	promExp, err := promexporter.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}
	meterOpts = append(meterOpts, sdkmetric.WithReader(promExp))
}

mp := sdkmetric.NewMeterProvider(meterOpts...)
otel.SetMeterProvider(mp)
Meter = mp.Meter(cfg.ServiceName)
```

**Step 4: Add OTLP trace exporter support**

```go
import (
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otlpmetrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

// In Init(), when building trace exporters:
if cfg.OTLPEndpoint != "" {
	otlpExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp trace exporter: %w", err)
	}
	// Add to tracer provider options
	tpOpts = append(tpOpts, sdktrace.WithBatcher(otlpExp))

	// Also add OTLP metric exporter
	otlpMetricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp metric exporter: %w", err)
	}
	meterOpts = append(meterOpts, sdkmetric.WithReader(
		sdkmetric.NewPeriodicReader(otlpMetricExp),
	))
}
```

**Step 5: Update cleanup function to shutdown meter provider too**

```go
cleanup := func() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
	if err := mp.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down meter provider: %v", err)
	}
}
```

**Step 6: Add PrometheusHandler function**

Add to `internal/telemetry/telemetry.go`:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// PrometheusHandler returns the HTTP handler for /metrics endpoint.
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
```

**Step 7: Add CLI flags and /metrics route in main.go**

Add new flags in the `var` block:

```go
enablePrometheus = flag.Bool("metrics", false, "Enable Prometheus /metrics endpoint")
otlpEndpoint     = flag.String("otel-otlp-endpoint", "", "OTLP exporter endpoint (e.g. localhost:4318)")
```

Update telemetry.Init call:

```go
cleanupTelemetry, err := telemetry.Init(telemetry.Config{
	ServiceName:          "hctf2",
	ServiceVersion:       "0.5.0",
	Environment:          os.Getenv("ENVIRONMENT"),
	EnableStdoutExporter: os.Getenv("OTEL_EXPORTER_STDOUT") == "true",
	EnablePrometheus:     *enablePrometheus || os.Getenv("OTEL_METRICS_PROMETHEUS") == "true",
	OTLPEndpoint:         firstNonEmpty(*otlpEndpoint, os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
})
```

Add helper:

```go
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
```

Register /metrics route (after health routes, before public routes):

```go
// Prometheus metrics endpoint
if *enablePrometheus || os.Getenv("OTEL_METRICS_PROMETHEUS") == "true" {
	r.Handle("/metrics", telemetry.PrometheusHandler())
}
```

**Step 8: Run tests**

```bash
task test
```

Expected: PASS

**Step 9: Manual verification**

```bash
task rebuild && ./hctf2 --port 8090 --metrics
# In another terminal:
curl http://localhost:8090/metrics
```

Should return Prometheus-format metrics including `http_requests_total`, `http_request_duration_seconds`, etc.

**Step 10: Commit**

```bash
git add internal/telemetry/telemetry.go main.go go.mod go.sum
git commit -m "feat: add Prometheus /metrics and OTLP trace/metric exporters

- Prometheus: enable with --metrics flag or OTEL_METRICS_PROMETHEUS=true
- OTLP: configure with --otel-otlp-endpoint or OTEL_EXPORTER_OTLP_ENDPOINT
- Serves /metrics endpoint with standard Prometheus format
- Exports all 4 existing metrics + traces to OTel collectors"
```

---

### Task 5: SMTP Password Reset

**Files:**
- Create: `internal/email/email.go`
- Create: `internal/email/email_test.go`
- Modify: `internal/handlers/auth.go` (inject email service, update ForgotPassword)
- Modify: `main.go` (add SMTP flags, create email service, pass to handler)

**Step 1: Write test for email service**

Create `internal/email/email_test.go`:

```go
package email

import "testing"

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
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/email/ -v
```

Expected: FAIL - package doesn't exist

**Step 3: Create internal/email/email.go**

```go
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
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/email/ -v
```

Expected: PASS

**Step 5: Update AuthHandler to accept email service**

In `internal/handlers/auth.go`, update the struct and constructor:

```go
import "github.com/yourusername/hctf2/internal/email"

type AuthHandler struct {
	db       *database.DB
	emailSvc *email.Service
	baseURL  string
}

func NewAuthHandler(db *database.DB, emailSvc *email.Service, baseURL string) *AuthHandler {
	return &AuthHandler{db: db, emailSvc: emailSvc, baseURL: baseURL}
}
```

**Step 6: Update ForgotPassword handler**

Replace the TODO block (lines 258-260) with:

```go
// Send reset email (or log link in dev mode)
resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.baseURL, tokenStr)
if err := h.emailSvc.SendPasswordReset(email, resetURL); err != nil {
	log.Printf("Failed to send password reset email: %v", err)
}
```

Update the swag annotation to remove the "not yet implemented" note.

**Step 7: Add SMTP flags and wire up in main.go**

Add flags:

```go
smtpHost = flag.String("smtp-host", "", "SMTP server host")
smtpPort = flag.Int("smtp-port", 587, "SMTP server port")
smtpFrom = flag.String("smtp-from", "", "SMTP from address")
smtpUser = flag.String("smtp-user", "", "SMTP username")
smtpPass = flag.String("smtp-password", "", "SMTP password")
baseURL  = flag.String("base-url", "http://localhost:8090", "Base URL for links in emails")
```

Create email service before Server initialization:

```go
emailSvc := email.NewService(email.Config{
	Host:     firstNonEmpty(*smtpHost, os.Getenv("SMTP_HOST")),
	Port:     *smtpPort,
	From:     firstNonEmpty(*smtpFrom, os.Getenv("SMTP_FROM")),
	Username: firstNonEmpty(*smtpUser, os.Getenv("SMTP_USER")),
	Password: firstNonEmpty(*smtpPass, os.Getenv("SMTP_PASSWORD")),
})

if !emailSvc.IsConfigured() {
	log.Println("Warning: SMTP not configured. Password reset links will be logged to console.")
}
```

Update AuthHandler creation:

```go
authH: handlers.NewAuthHandler(db, emailSvc, *baseURL),
```

**Step 8: Update test helper**

In `handlers_test.go`, update `newTestServer` to pass email service:

```go
import "github.com/yourusername/hctf2/internal/email"

// In newTestServer:
authH: handlers.NewAuthHandler(db, email.NewService(email.Config{}), "http://localhost:8090"),
```

**Step 9: Run all tests**

```bash
task test
```

Expected: PASS

**Step 10: Commit**

```bash
git add internal/email/email.go internal/email/email_test.go internal/handlers/auth.go main.go handlers_test.go
git commit -m "feat: add SMTP email for password reset

- New internal/email package wraps net/smtp
- SMTP config via --smtp-host/--smtp-port/--smtp-from/--smtp-user/--smtp-password flags
- Also reads SMTP_HOST, SMTP_FROM, SMTP_USER, SMTP_PASSWORD env vars
- Dev mode: logs reset URL to console when SMTP not configured
- HTML email template matches app's dark theme"
```

---

### Task 6: Update OpenAPI spec and documentation

**Files:**
- Modify: `docs/openapi.yaml` (regenerated)
- Modify: `CLAUDE.md` (update version if needed)
- Modify: `CONFIGURATION.md` (add SMTP and OTel config docs)

**Step 1: Regenerate OpenAPI spec**

```bash
task generate-openapi
```

**Step 2: Update CONFIGURATION.md with new flags**

Add sections for:
- SMTP configuration (flags + env vars)
- OTel/Prometheus configuration (flags + env vars)
- Health check endpoints

**Step 3: Commit**

```bash
git add docs/openapi.yaml CONFIGURATION.md
git commit -m "docs: update OpenAPI spec and config docs for new features

- Add SMTP configuration section
- Add OTel/Prometheus configuration section
- Document /healthz and /readyz endpoints
- Regenerate OpenAPI spec with new endpoints"
```

---

### Task 7: Final integration test

**Step 1: Rebuild and start server**

```bash
task rebuild && ./hctf2 --port 8090 --metrics --admin-email admin@test.com --admin-password test123
```

**Step 2: Verify all new endpoints**

```bash
# Healthcheck
curl -s http://localhost:8090/healthz | jq .
curl -s http://localhost:8090/readyz | jq .

# Prometheus metrics
curl -s http://localhost:8090/metrics | head -20

# Password reset (should log URL to console)
curl -s -X POST -d "email=admin@test.com" http://localhost:8090/api/auth/forgot-password
```

**Step 3: Run full test suite**

```bash
task test
```

Expected: All tests PASS

**Step 4: Final commit (if any fixes needed)**

Only if integration testing reveals issues that need fixing.
