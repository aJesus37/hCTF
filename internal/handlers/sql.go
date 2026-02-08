package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/hctf2/internal/database"
)

type SQLHandler struct {
	db *database.DB
}

func NewSQLHandler(db *database.DB) *SQLHandler {
	return &SQLHandler{db: db}
}

func (h *SQLHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.db.GetSQLSnapshot()
	if err != nil {
		http.Error(w, "Failed to generate snapshot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}
