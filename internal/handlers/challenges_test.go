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
