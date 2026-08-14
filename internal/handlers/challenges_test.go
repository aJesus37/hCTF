package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ajesus37/hCTF/internal/database"
	"github.com/go-chi/chi/v5"
)

// withQuestionID injects the chi route context so handlers using
// chi.URLParam(r, "id") can resolve the question id in direct unit tests.
func withQuestionID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// withChallengeID injects the chi route context so handlers using
// chi.URLParam(r, "id") can resolve the challenge id in direct unit tests.
func withChallengeID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// withHintID injects the chi route context so handlers using
// chi.URLParam(r, "id") can resolve the hint id in direct unit tests.
func withHintID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestUpdateChallengePartial(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	dataset := "https://example.com/d.csv"
	hint := "schema hint"
	tags := "a,b"
	ch, err := db.CreateChallenge("Old", "desc", "web", "easy", &tags, true, true, &dataset, &hint, true, 1000, 100, 50, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/challenges/"+ch.ID, strings.NewReader(`{"name":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateChallenge(rr, withChallengeID(req, ch.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetChallengeByID(ch.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.Description != "desc" {
		t.Errorf("description: got %q", got.Description)
	}
	if !got.SQLEnabled || got.SQLDatasetURL == nil || got.SQLSchemaHint == nil {
		t.Error("sql fields were wiped")
	}
	if !got.DynamicScoring || got.InitialPoints != 1000 || got.MinimumPoints != 100 || got.DecayThreshold != 50 {
		t.Error("dynamic scoring fields were wiped")
	}
	if !got.Visible {
		t.Error("visible was wiped")
	}
	if got.Tags == nil || *got.Tags != tags {
		t.Error("tags were wiped")
	}
}

func TestUpdateQuestionPartial(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	ch, err := db.CreateChallenge("C", "desc", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	dataset := "https://example.com/x.csv"
	q, err := db.CreateQuestion(ch.ID, "Old", "old desc", "flag{secret}", &dataset, true, 500, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	// PUT only {name} — everything else must be preserved.
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader(`{"name":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q want New", got.Name)
	}
	if got.Description != "old desc" {
		t.Errorf("description: got %q want 'old desc'", got.Description)
	}
	if got.Flag != "flag{secret}" {
		t.Errorf("flag: got %q want flag{secret}", got.Flag)
	}
	if got.Points != 500 {
		t.Errorf("points: got %d want 500", got.Points)
	}
	if !got.CaseSensitive {
		t.Error("case_sensitive: got false want true")
	}
	if got.FlagMask == nil || *got.FlagMask != dataset {
		t.Errorf("flag_mask: got %v want %q", got.FlagMask, dataset)
	}
}

func TestUpdateQuestionEmptyFlagKeepsExisting(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{secret}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader(`{"name":"Renamed","flag":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Flag != "flag{secret}" {
		t.Errorf("flag: got %q want flag{secret}", got.Flag)
	}
	if got.Name != "Renamed" {
		t.Errorf("name: got %q want Renamed", got.Name)
	}
}

func TestUpdateQuestionFlagRegeneratesMask(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	oldMask := "https://example.com/old.csv"
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{secret}", &oldMask, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	// New flag without a mask: flag_mask must be regenerated from the new flag,
	// not the old mask.
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader(`{"flag":"flag{new}"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Flag != "flag{new}" {
		t.Errorf("flag: got %q want flag{new}", got.Flag)
	}
	if got.FlagMask == nil || *got.FlagMask != "flag{***}" {
		t.Errorf("flag_mask: got %v want %q", got.FlagMask, "flag{***}")
	}
}

func TestUpdateQuestionFlagWithExplicitMask(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{secret}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	newMask := "https://example.com/new.csv"
	// New flag with an explicit mask: both the flag and the mask are stored.
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader(`{"flag":"flag{new}","flag_mask":"https://example.com/new.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Flag != "flag{new}" {
		t.Errorf("flag: got %q want flag{new}", got.Flag)
	}
	if got.FlagMask == nil || *got.FlagMask != newMask {
		t.Errorf("flag_mask: got %v want %q", got.FlagMask, newMask)
	}
}

func TestUpdateQuestionMaskOnly(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{secret}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	newMask := "https://example.com/mask.csv"
	// Mask-only update: the flag is preserved and the provided mask is stored.
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader(`{"flag_mask":"https://example.com/mask.csv"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Flag != "flag{secret}" {
		t.Errorf("flag: got %q want flag{secret}", got.Flag)
	}
	if got.FlagMask == nil || *got.FlagMask != newMask {
		t.Errorf("flag_mask: got %v want %q", got.FlagMask, newMask)
	}
}

func TestUpdateChallengeExplicitZerosApplied(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "desc", "web", "easy", nil, true, false, nil, nil, true, 1000, 100, 50, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	// Explicit false/zero values must be applied, not treated as "omitted".
	req := httptest.NewRequest("PUT", "/api/admin/challenges/"+ch.ID, strings.NewReader(`{"visible":false,"initial_points":0}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateChallenge(rr, withChallengeID(req, ch.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetChallengeByID(ch.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Visible {
		t.Error("visible: got true want false")
	}
	if got.InitialPoints != 0 {
		t.Errorf("initial_points: got %d want 0", got.InitialPoints)
	}
	if !got.DynamicScoring {
		t.Error("dynamic_scoring: got false want true (preserved)")
	}
	if got.MinimumPoints != 100 {
		t.Errorf("minimum_points: got %d want 100 (preserved)", got.MinimumPoints)
	}
	if got.Description != "desc" {
		t.Errorf("description: got %q want desc (preserved)", got.Description)
	}
}

func TestUpdateChallengeFormPreservesTags(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	tags := "a,b"
	ch, err := db.CreateChallenge("Old", "desc", "web", "easy", &tags, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/challenges/"+ch.ID, strings.NewReader("name=New&category=web&difficulty=easy&description=d"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateChallenge(rr, withChallengeID(req, ch.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetChallengeByID(ch.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.Tags == nil || *got.Tags != tags {
		t.Errorf("tags: got %v want %q", got.Tags, tags)
	}
}

func TestUpdateQuestionFormPreservesFileURL(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	fileURL := "https://example.com/x.csv"
	q, err := db.CreateQuestion(ch.ID, "Old", "", "flag{secret}", nil, false, 100, &fileURL)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/questions/"+q.ID, strings.NewReader("name=New&points=100"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.UpdateQuestion(rr, withQuestionID(req, q.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetQuestionByID(q.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if got.Name != "New" {
		t.Errorf("name: got %q want New", got.Name)
	}
	if got.FileURL == nil || *got.FileURL != fileURL {
		t.Errorf("file_url: got %v want %q", got.FileURL, fileURL)
	}
}

func TestUpdateChallengeNotFound(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/challenges/nonexistent", strings.NewReader(`{"name":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateChallenge(rr, withChallengeID(req, "nonexistent"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404, body: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateChallengeInvalidJSON(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "desc", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/challenges/"+ch.ID, strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateChallenge(rr, withChallengeID(req, ch.ID))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetChallengeByID(ch.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "C" || got.Description != "desc" {
		t.Errorf("challenge mutated on invalid JSON: got name %q desc %q", got.Name, got.Description)
	}
}

func TestUpdateHintPartial(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{x}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	hint, err := db.CreateHint(q.ID, "hint content", 20, 1)
	if err != nil {
		t.Fatalf("create hint: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	// PUT only {cost} — content and order must be preserved.
	req := httptest.NewRequest("PUT", "/api/admin/hints/"+hint.ID, strings.NewReader(`{"cost":50}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateHint(rr, withHintID(req, hint.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetHintByID(hint.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Cost != 50 {
		t.Errorf("cost: got %d want 50", got.Cost)
	}
	if got.Content != "hint content" {
		t.Errorf("content: got %q", got.Content)
	}
	if got.Order != 1 {
		t.Errorf("order: got %d want 1", got.Order)
	}
}

func TestUpdateHintNotFound(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/hints/nonexistent", strings.NewReader(`{"cost":50}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateHint(rr, withHintID(req, "nonexistent"))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404, body: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateHintExplicitZero(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{x}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	hint, err := db.CreateHint(q.ID, "hint content", 20, 2)
	if err != nil {
		t.Fatalf("create hint: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	// Explicit zero must be applied, not treated as "omitted".
	req := httptest.NewRequest("PUT", "/api/admin/hints/"+hint.ID, strings.NewReader(`{"cost":0}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateHint(rr, withHintID(req, hint.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetHintByID(hint.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Cost != 0 {
		t.Errorf("cost: got %d want 0", got.Cost)
	}
	if got.Content != "hint content" {
		t.Errorf("content: got %q want 'hint content' (preserved)", got.Content)
	}
	if got.Order != 2 {
		t.Errorf("order: got %d want 2 (preserved)", got.Order)
	}
}

func TestUpdateHintInvalidJSON(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	ch, err := db.CreateChallenge("C", "", "web", "easy", nil, true, false, nil, nil, false, 100, 0, 0, nil)
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	q, err := db.CreateQuestion(ch.ID, "Q", "", "flag{x}", nil, false, 100, nil)
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	hint, err := db.CreateHint(q.ID, "hint content", 20, 1)
	if err != nil {
		t.Fatalf("create hint: %v", err)
	}

	h := NewChallengeHandler(db, nil, nil, nil)
	req := httptest.NewRequest("PUT", "/api/admin/hints/"+hint.ID, strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.UpdateHint(rr, withHintID(req, hint.ID))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400, body: %s", rr.Code, rr.Body.String())
	}

	got, err := db.GetHintByID(hint.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Content != "hint content" || got.Cost != 20 || got.Order != 1 {
		t.Errorf("hint mutated on invalid JSON: got content %q cost %d order %d", got.Content, got.Cost, got.Order)
	}
}
