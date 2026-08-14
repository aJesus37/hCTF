package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestQuestionUpdateMarshalOmitsNil(t *testing.T) {
	b, _ := json.Marshal(QuestionUpdate{Name: strPtr("New")})
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if len(m) != 1 || m["name"] != "New" {
		t.Fatalf("expected only name, got %v", m)
	}
	if _, ok := m["flag"]; ok {
		t.Fatal("flag should be omitted when nil")
	}
}

func TestChallengeUpdateMarshalOmitsNil(t *testing.T) {
	b, _ := json.Marshal(ChallengeUpdate{Points: intPtr(200)})
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if len(m) != 1 || m["initial_points"] != float64(200) {
		t.Fatalf("expected only initial_points, got %v", m)
	}
}

func TestUpdateCompetitionSendsOnlyProvidedFields(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "token")
	start := "2026-09-01T12:00"
	_, err := c.UpdateCompetition(7, CompetitionUpdate{Name: strPtr("New"), StartAt: &start})
	if err != nil {
		t.Fatalf("UpdateCompetition: %v", err)
	}
	if gotForm.Get("name") != "New" {
		t.Errorf("name: got %q", gotForm.Get("name"))
	}
	if gotForm.Get("start_at") != start {
		t.Errorf("start_at: got %q", gotForm.Get("start_at"))
	}
	// Absent fields must NOT be sent (so the server preserves them).
	for _, key := range []string{"description", "rules_html", "end_at", "registration_start", "registration_end"} {
		if _, ok := gotForm[key]; ok {
			t.Errorf("field %q should be omitted", key)
		}
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
