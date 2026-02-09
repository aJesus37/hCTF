package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/hctf2/internal/auth"
	"github.com/yourusername/hctf2/internal/database"
)

type TeamHandler struct {
	db *database.DB
}

func NewTeamHandler(db *database.DB) *TeamHandler {
	return &TeamHandler{db: db}
}

type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTeam handles team creation
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user already in a team
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.TeamID != nil {
		http.Error(w, "You are already in a team. Leave your current team first.", http.StatusBadRequest)
		return
	}

	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Name == "" {
		http.Error(w, "Team name required", http.StatusBadRequest)
		return
	}

	// Create team
	team, err := h.db.CreateTeam(req.Name, req.Description, claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, "Team name already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create team", http.StatusInternalServerError)
		}
		return
	}

	// Add creator to team
	if err := h.db.JoinTeam(claims.UserID, team.ID); err != nil {
		http.Error(w, "Failed to join team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

// JoinTeam handles joining a team by ID
func (h *TeamHandler) JoinTeam(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	// Check if user already in a team
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.TeamID != nil {
		http.Error(w, "Already in a team", http.StatusBadRequest)
		return
	}

	// Check team exists
	_, err = h.db.GetTeamByID(teamID)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	// Join team
	if err := h.db.JoinTeam(claims.UserID, teamID); err != nil {
		http.Error(w, "Failed to join team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Joined team successfully"}`))
}

// LeaveTeam handles leaving current team
func (h *TeamHandler) LeaveTeam(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.TeamID == nil {
		http.Error(w, "Not in a team", http.StatusBadRequest)
		return
	}

	// Check if user is team owner
	team, err := h.db.GetTeamByID(*user.TeamID)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	if team.OwnerID == claims.UserID {
		http.Error(w, "Team owner cannot leave. Transfer ownership or disband team.", http.StatusForbidden)
		return
	}

	// Leave team
	if err := h.db.LeaveTeam(claims.UserID); err != nil {
		http.Error(w, "Failed to leave team", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Left team successfully"}`))
}

// ListTeams returns all teams
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.db.GetAllTeams()
	if err != nil {
		http.Error(w, "Failed to fetch teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// GetTeam returns team details with members
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	team, err := h.db.GetTeamByID(teamID)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	members, err := h.db.GetTeamMembers(teamID)
	if err != nil {
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"team":    team,
		"members": members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTeamScoreboard returns team rankings
func (h *TeamHandler) GetTeamScoreboard(w http.ResponseWriter, r *http.Request) {
	scoreboard, err := h.db.GetTeamScoreboard(50)
	if err != nil {
		http.Error(w, "Failed to fetch scoreboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scoreboard)
}
