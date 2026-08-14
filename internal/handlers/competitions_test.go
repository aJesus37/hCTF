package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ajesus37/hCTF/internal/database"
	"github.com/ajesus37/hCTF/internal/models"
	"github.com/go-chi/chi/v5"
)

// withCompID injects the chi route context so handlers using
// chi.URLParam(r, "id") can resolve the competition id in direct unit tests.
func withCompID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestUpdateCompetitionPartial(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	comp, err := db.CreateCompetition("Old", "desc", "<p>rules</p>", &start, nil, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewCompetitionHandler(db)
	form := url.Values{"name": {"New"}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/"+strconv.FormatInt(comp.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, strconv.FormatInt(comp.ID, 10)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetCompetitionByID(comp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.RulesHTML != "<p>rules</p>" {
		t.Errorf("rules_html: got %q", got.RulesHTML)
	}
	if got.StartAt == nil || !got.StartAt.Equal(start) {
		t.Error("start_at was wiped")
	}
	if got.Description != "desc" {
		t.Errorf("description: got %q", got.Description)
	}

	var resp models.Competition
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "New" {
		t.Errorf("response name: got %q", resp.Name)
	}
	if resp.RulesHTML != "<p>rules</p>" {
		t.Errorf("response rules_html: got %q", resp.RulesHTML)
	}
	if resp.StartAt == nil || !resp.StartAt.Equal(start) {
		t.Error("response start_at was wiped")
	}
}

func TestUpdateCompetitionPreservesStatus(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	comp, err := db.CreateCompetition("Old", "desc", "<p>rules</p>", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if comp.Status != models.CompStatusDraft {
		t.Fatalf("expected default status %q, got %q", models.CompStatusDraft, comp.Status)
	}

	h := NewCompetitionHandler(db)
	form := url.Values{"name": {"New"}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/"+strconv.FormatInt(comp.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, strconv.FormatInt(comp.ID, 10)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetCompetitionByID(comp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q want New", got.Name)
	}
	if got.Status != models.CompStatusDraft {
		t.Errorf("status: got %q want %q (preserved when not submitted)", got.Status, models.CompStatusDraft)
	}
}

func TestUpdateCompetitionSetsTiming(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	comp, err := db.CreateCompetition("Old", "desc", "<p>rules</p>", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewCompetitionHandler(db)
	form := url.Values{"start_at": {"2026-09-01T12:00"}, "end_at": {"2026-09-02T12:00"}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/"+strconv.FormatInt(comp.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, strconv.FormatInt(comp.ID, 10)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetCompetitionByID(comp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	wantStart := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	if got.StartAt == nil || !got.StartAt.Equal(wantStart) {
		t.Errorf("start_at: got %v want %v", got.StartAt, wantStart)
	}
	if got.EndAt == nil || !got.EndAt.Equal(wantEnd) {
		t.Errorf("end_at: got %v want %v", got.EndAt, wantEnd)
	}
	if got.Name != "Old" {
		t.Errorf("name: got %q want Old (preserved)", got.Name)
	}
}

func TestUpdateCompetitionClearsTiming(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	comp, err := db.CreateCompetition("Old", "desc", "<p>rules</p>", &start, nil, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewCompetitionHandler(db)
	form := url.Values{"start_at": {""}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/"+strconv.FormatInt(comp.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, strconv.FormatInt(comp.ID, 10)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetCompetitionByID(comp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.StartAt != nil {
		t.Errorf("start_at: got %v want nil (cleared)", got.StartAt)
	}
	if got.Name != "Old" {
		t.Errorf("name: got %q want Old (preserved)", got.Name)
	}
}

func TestUpdateCompetitionNotFound(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	h := NewCompetitionHandler(db)
	form := url.Values{"name": {"New"}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/99999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, "99999"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404, body: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateCompetitionBadID(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	h := NewCompetitionHandler(db)
	form := url.Values{"name": {"New"}}
	req := httptest.NewRequest("PUT", "/api/admin/competitions/abc", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateCompetition(rr, withCompID(req, "abc"))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400, body: %s", rr.Code, rr.Body.String())
	}
}
