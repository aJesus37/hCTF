package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/hctf2/internal/database"
)

type ScoreboardHandler struct {
	db *database.DB
}

func NewScoreboardHandler(db *database.DB) *ScoreboardHandler {
	return &ScoreboardHandler{db: db}
}

func (h *ScoreboardHandler) GetScoreboard(w http.ResponseWriter, r *http.Request) {
	entries, err := h.db.GetScoreboard(100) // Top 100
	if err != nil {
		http.Error(w, "Failed to fetch scoreboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
